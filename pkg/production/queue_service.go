package production

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"time"

	"symetra-lab-backend/models"

	"gorm.io/gorm"
)

// GetUnifiedProductionQueue mengambil seluruh antrean cetak gabungan dari Shopee dan Manual Orders
func GetUnifiedProductionQueue(db *gorm.DB) ([]models.ProductionQueueItem, error) {
	var queue []models.ProductionQueueItem
	now := time.Now()

	// 1. Ambil pesanan Shopee yang butuh dicetak (READY_TO_SHIP atau PROCESSED)
	var shopeeOrders []models.ShopeeOrder
	if err := db.Preload("Items.Product.DefaultMachine").
		Where("order_status IN ('READY_TO_SHIP', 'PROCESSED')").
		Find(&shopeeOrders).Error; err == nil {

		for _, so := range shopeeOrders {
			for _, it := range so.Items {
				var weight float64 = 20.0
				var printHours float64 = 1.0
				var machineID *string
				var machineName string

				if it.Product != nil {
					if it.Product.DefaultWeightGrams > 0 {
						weight = it.Product.DefaultWeightGrams
					}
					if it.Product.DefaultPrintTimeHours > 0 {
						printHours = it.Product.DefaultPrintTimeHours
					}
					if it.Product.DefaultMachineID != nil {
						machineID = it.Product.DefaultMachineID
						if it.Product.DefaultMachine != nil {
							machineName = it.Product.DefaultMachine.Name
						}
					}
				}

				totalHours := math.Round(printHours*float64(it.Quantity)*100) / 100

				var deadline *time.Time = so.ShipByDateTime
				isUrgent := false
				if deadline != nil {
					// Urgent jika sisa waktu pengiriman kurang dari 24 jam
					if deadline.Sub(now) < 24*time.Hour {
						isUrgent = true
					}
				}

				variant := it.ModelName
				if it.MatchedSKU != "" {
					variant = it.MatchedSKU
				} else if it.ModelSKU != "" {
					variant = it.ModelSKU
				}

				queue = append(queue, models.ProductionQueueItem{
					JobID:             fmt.Sprintf("%d", it.ID),
					Source:            "SHOPEE",
					OrderIdentifier:   so.OrderSN,
					CustomerName:      so.BuyerUsername,
					ItemName:          it.ItemName,
					VariantSKU:        variant,
					Quantity:          it.Quantity,
					WeightGrams:       weight,
					PrintTimeHours:    printHours,
					TotalPrintHours:   totalHours,
					AssignedMachineID: machineID,
					AssignedMachine:   machineName,
					Status:            so.OrderStatus,
					Deadline:          deadline,
					IsUrgent:          isUrgent,
					CreatedAt:         so.CreatedAt,
				})
			}
		}
	}

	// 2. Ambil pesanan Manual / Offline (status PENDING atau IN_PRODUCTION)
	var manualOrders []models.Order
	if err := db.Preload("Items.Product.DefaultMachine").
		Preload("Items.Machine").
		Where("status IN ('PENDING', 'IN_PRODUCTION')").
		Find(&manualOrders).Error; err == nil {

		for _, mo := range manualOrders {
			for _, it := range mo.Items {
				var machineID *string = it.MachineID
				var machineName string
				if it.Machine != nil {
					machineName = it.Machine.Name
				} else if it.Product != nil && it.Product.DefaultMachine != nil {
					machineID = it.Product.DefaultMachineID
					machineName = it.Product.DefaultMachine.Name
				}

				weight := it.WeightGrams
				if weight <= 0 && it.Product != nil {
					weight = it.Product.DefaultWeightGrams
				}
				printHours := it.PrintTimeHours
				if printHours <= 0 && it.Product != nil {
					printHours = it.Product.DefaultPrintTimeHours
				}
				totalHours := math.Round(printHours*float64(it.Quantity)*100) / 100

				// Default SLA manual order 48 jam dari tanggal dibuat
				orderDeadline := mo.CreatedAt.Add(48 * time.Hour)
				isUrgent := false
				if orderDeadline.Sub(now) < 24*time.Hour {
					isUrgent = true
				}

				variant := ""
				if it.Product != nil && it.Product.SKU != nil {
					variant = *it.Product.SKU
				}

				queue = append(queue, models.ProductionQueueItem{
					JobID:             it.ID,
					Source:            "MANUAL",
					OrderIdentifier:   mo.OrderNumber,
					CustomerName:      mo.CustomerName,
					ItemName:          it.ProductName,
					VariantSKU:        variant,
					Quantity:          it.Quantity,
					WeightGrams:       weight,
					PrintTimeHours:    printHours,
					TotalPrintHours:   totalHours,
					AssignedMachineID: machineID,
					AssignedMachine:   machineName,
					Status:            mo.Status,
					Deadline:          &orderDeadline,
					IsUrgent:          isUrgent,
					CreatedAt:         mo.CreatedAt,
				})
			}
		}
	}

	// 3. Urutkan antrean: Urgent paling depan, lalu deadline terdekat
	sort.Slice(queue, func(i, j int) bool {
		if queue[i].IsUrgent != queue[j].IsUrgent {
			return queue[i].IsUrgent // Yang urgent didahulukan
		}
		if queue[i].Deadline != nil && queue[j].Deadline != nil {
			return queue[i].Deadline.Before(*queue[j].Deadline)
		}
		return queue[i].CreatedAt.Before(queue[j].CreatedAt)
	})

	return queue, nil
}

// CompletePrintJob memproses selesai cetak: potong stok filamen dan tambah jam pakai mesin
func CompletePrintJob(db *gorm.DB, req models.CompletePrintJobRequest) (*models.CompletePrintJobResponse, error) {
	resp := &models.CompletePrintJobResponse{
		Status:            "success",
		DeductedFilaments: []string{},
	}

	var printHours float64
	var quantity int = 1
	var targetMachineID string = req.MachineID
	var filamentDeductions []struct {
		FilamentID string
		Grams      float64
	}

	if req.Source == "MANUAL" {
		var item models.OrderItem
		if err := db.Preload("Product.Filaments").
			Preload("Filaments").
			Where("id = ?", req.JobID).
			First(&item).Error; err != nil {
			return nil, fmt.Errorf("item pesanan manual tidak ditemukan: %w", err)
		}

		quantity = item.Quantity
		printHours = item.PrintTimeHours

		if targetMachineID == "" && item.MachineID != nil {
			targetMachineID = *item.MachineID
		}

		// Kumpulkan filamen yang akan dipotong
		if len(item.Filaments) > 0 {
			for _, f := range item.Filaments {
				if f.FilamentID != nil && *f.FilamentID != "" {
					filamentDeductions = append(filamentDeductions, struct {
						FilamentID string
						Grams      float64
					}{
						FilamentID: *f.FilamentID,
						Grams:      f.WeightUsedGrams * float64(quantity),
					})
				}
			}
		} else if item.Product != nil && len(item.Product.Filaments) > 0 {
			for _, pf := range item.Product.Filaments {
				if pf.FilamentID != nil && *pf.FilamentID != "" {
					filamentDeductions = append(filamentDeductions, struct {
						FilamentID string
						Grams      float64
					}{
						FilamentID: *pf.FilamentID,
						Grams:      pf.WeightUsedGrams * float64(quantity),
					})
				}
			}
		}

		// Update status order jika belum COMPLETED
		var order models.Order
		if err := db.Where("id = ?", item.OrderID).First(&order).Error; err == nil {
			now := time.Now()
			order.Status = "COMPLETED"
			order.CompletedAt = &now
			_ = db.Save(&order)
		}

	} else if req.Source == "SHOPEE" {
		itemID, err := strconv.ParseUint(req.JobID, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("format job_id Shopee tidak valid: %w", err)
		}

		var item models.ShopeeOrderItem
		if err := db.Preload("Product.Filaments").
			Where("id = ?", itemID).
			First(&item).Error; err != nil {
			return nil, fmt.Errorf("item pesanan Shopee tidak ditemukan: %w", err)
		}

		quantity = item.Quantity
		if item.Product != nil {
			printHours = item.Product.DefaultPrintTimeHours
			if targetMachineID == "" && item.Product.DefaultMachineID != nil {
				targetMachineID = *item.Product.DefaultMachineID
			}

			for _, pf := range item.Product.Filaments {
				if pf.FilamentID != nil && *pf.FilamentID != "" {
					filamentDeductions = append(filamentDeductions, struct {
						FilamentID string
						Grams      float64
					}{
						FilamentID: *pf.FilamentID,
						Grams:      pf.WeightUsedGrams * float64(quantity),
					})
				}
			}
		}
	} else {
		return nil, fmt.Errorf("source harus SHOPEE atau MANUAL")
	}

	// 1. Eksekusi Pemotongan Stok Filamen
	for _, fd := range filamentDeductions {
		var fil models.Filament
		if err := db.Where("id = ?", fd.FilamentID).First(&fil).Error; err == nil {
			fil.CurrentStockGrams -= fd.Grams
			if fil.CurrentStockGrams < 0 {
				fil.CurrentStockGrams = 0
			}
			_ = db.Save(&fil)

			resp.DeductedFilaments = append(resp.DeductedFilaments,
				fmt.Sprintf("%s: -%.2fg (Sisa: %.2fg)", fil.ColorName, fd.Grams, fil.CurrentStockGrams))

			threshold := fil.LowStockThresholdGrams
			if threshold <= 0 {
				threshold = 200
			}
			if fil.CurrentStockGrams <= threshold {
				resp.LowStockWarning = true
			}
		}
	}

	// 2. Eksekusi Penambahan Jam Operasional Mesin
	totalHoursToAdd := math.Round(printHours*float64(quantity)*100) / 100
	if targetMachineID != "" && totalHoursToAdd > 0 {
		var machine models.Machine
		if err := db.Preload("MaintenanceParts").Where("id = ?", targetMachineID).First(&machine).Error; err == nil {
			machine.TotalHoursUsed += totalHoursToAdd
			machine.CurrentState = "IDLE"
			_ = db.Save(&machine)
			resp.AddedMachineHours = totalHoursToAdd

			// Cek suku cadang perawatan
			for _, part := range machine.MaintenanceParts {
				if part.LifespanHours > 0 && part.HoursUsedCurrent >= part.LifespanHours {
					resp.MaintenanceWarning = true
				}
			}
		}
	}

	resp.Message = fmt.Sprintf("Pekerjaan cetak selesai. Filamen terpotong untuk %d pcs dan mesin beroperasi +%.2f jam.", quantity, totalHoursToAdd)
	return resp, nil
}
