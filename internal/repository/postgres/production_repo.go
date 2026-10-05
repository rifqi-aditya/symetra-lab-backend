package postgres

import (
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"symetra-lab-backend-v2/internal/domain/production"
)

type prodProductGORM struct {
	ID                    uuid.UUID        `gorm:"column:id;primaryKey;type:uuid"`
	SKU                   *string          `gorm:"column:sku"`
	DefaultWeightGrams    float64          `gorm:"column:default_weight_grams"`
	DefaultPrintTimeHours float64          `gorm:"column:default_print_time_hours"`
	DefaultMachineID      *uuid.UUID       `gorm:"column:default_machine_id;type:uuid"`
	DefaultMachine        *prodMachineGORM `gorm:"foreignKey:DefaultMachineID;references:ID"`
}

func (prodProductGORM) TableName() string {
	return "products"
}

type prodMachinePartGORM struct {
	ID               uuid.UUID `gorm:"column:id;primaryKey;type:uuid"`
	PartName         string    `gorm:"column:part_name"`
	MachineID        uuid.UUID `gorm:"column:machine_id;type:uuid"`
	LifespanHours    float64   `gorm:"column:lifespan_hours"`
	HoursUsedCurrent float64   `gorm:"column:hours_used_current"`
}

func (prodMachinePartGORM) TableName() string {
	return "machine_maintenance_parts"
}

type prodMachineGORM struct {
	ID             uuid.UUID             `gorm:"column:id;primaryKey;type:uuid"`
	Name           string                `gorm:"column:name"`
	TotalHoursUsed float64               `gorm:"column:total_hours_used"`
	CurrentState   string                `gorm:"column:current_state"`
	Parts          []prodMachinePartGORM `gorm:"foreignKey:MachineID;references:ID"`
}

func (prodMachineGORM) TableName() string {
	return "machines"
}

type prodFilamentGORM struct {
	ID                     uuid.UUID `gorm:"column:id;primaryKey;type:uuid"`
	ColorName              string    `gorm:"column:color_name"`
	CurrentStockGrams      float64   `gorm:"column:current_stock_grams"`
	LowStockThresholdGrams float64   `gorm:"column:low_stock_threshold_grams"`
}

func (prodFilamentGORM) TableName() string {
	return "filaments"
}

type prodOrderFilamentGORM struct {
	ID              uuid.UUID  `gorm:"column:id;primaryKey;type:uuid"`
	OrderItemID     uuid.UUID  `gorm:"column:order_item_id;type:uuid"`
	FilamentID      *uuid.UUID `gorm:"column:filament_id;type:uuid"`
	WeightUsedGrams float64    `gorm:"column:weight_used_grams"`
}

func (prodOrderFilamentGORM) TableName() string {
	return "order_filaments"
}

type prodOrderItemGORM struct {
	ID             uuid.UUID               `gorm:"column:id;primaryKey;type:uuid"`
	OrderID        uuid.UUID               `gorm:"column:order_id;type:uuid"`
	ProductID      *uuid.UUID              `gorm:"column:product_id;type:uuid"`
	ProductName    string                  `gorm:"column:product_name"`
	Quantity       int                     `gorm:"column:quantity"`
	WeightGrams    float64                 `gorm:"column:weight_grams"`
	PrintTimeHours float64                 `gorm:"column:print_time_hours"`
	MachineID      *uuid.UUID              `gorm:"column:machine_id;type:uuid"`
	MatchedSKU     string                  `gorm:"column:matched_sku"`
	ItemSKU        string                  `gorm:"column:item_sku"`
	Product        *prodProductGORM        `gorm:"foreignKey:ProductID;references:ID"`
	Machine        *prodMachineGORM        `gorm:"foreignKey:MachineID;references:ID"`
	Filaments      []prodOrderFilamentGORM `gorm:"foreignKey:OrderItemID;references:ID"`
}

func (prodOrderItemGORM) TableName() string {
	return "order_items"
}

type prodMarketplaceGORM struct {
	OrderID        uuid.UUID  `gorm:"column:order_id;primaryKey;type:uuid"`
	OrderSN        string     `gorm:"column:order_sn"`
	BuyerUsername  string     `gorm:"column:buyer_username"`
	ShipByDateTime *time.Time `gorm:"column:ship_by_date_time"`
}

func (prodMarketplaceGORM) TableName() string {
	return "order_marketplace_details"
}

type prodOrderGORM struct {
	ID           uuid.UUID            `gorm:"column:id;primaryKey;type:uuid"`
	OrderNumber  string               `gorm:"column:order_number"`
	CustomerName string               `gorm:"column:customer_name"`
	Status       string               `gorm:"column:status"`
	Channel      string               `gorm:"column:channel"`
	CreatedAt    time.Time            `gorm:"column:created_at"`
	CompletedAt  *time.Time           `gorm:"column:completed_at"`
	Items        []prodOrderItemGORM  `gorm:"foreignKey:OrderID;references:ID"`
	Marketplace  *prodMarketplaceGORM `gorm:"foreignKey:OrderID;references:ID"`
}

func (prodOrderGORM) TableName() string {
	return "orders"
}

type ProductionRepository struct {
	db *gorm.DB
}

func NewProductionRepository(db *gorm.DB) *ProductionRepository {
	return &ProductionRepository{db: db}
}

func (r *ProductionRepository) GetUnifiedQueue() ([]*production.ProductionQueueItem, error) {
	var queue []*production.ProductionQueueItem
	now := time.Now()

	var activeOrders []prodOrderGORM
	if err := r.db.Preload("Items.Product.DefaultMachine").
		Preload("Items.Machine").
		Preload("Marketplace").
		Where("status IN ('PENDING', 'IN_PRODUCTION', 'READY_TO_SHIP', 'PROCESSED')").
		Find(&activeOrders).Error; err != nil {
		return nil, err
	}

	for _, o := range activeOrders {
		channel := o.Channel
		if channel == "" {
			channel = "MANUAL"
		}

		for _, it := range o.Items {
			var machineID *string
			var machineName string
			if it.MachineID != nil {
				mid := it.MachineID.String()
				machineID = &mid
			}
			if it.Machine != nil {
				machineName = it.Machine.Name
			} else if it.Product != nil && it.Product.DefaultMachine != nil {
				mid := it.Product.DefaultMachineID.String()
				machineID = &mid
				machineName = it.Product.DefaultMachine.Name
			}

			weight := it.WeightGrams
			if weight <= 0 && it.Product != nil {
				weight = it.Product.DefaultWeightGrams
			}
			if weight <= 0 {
				weight = 20.0
			}

			printHours := it.PrintTimeHours
			if printHours <= 0 && it.Product != nil {
				printHours = it.Product.DefaultPrintTimeHours
			}
			if printHours <= 0 {
				printHours = 1.0
			}

			totalHours := math.Round(printHours*float64(it.Quantity)*100) / 100

			var deadline *time.Time
			isUrgent := false

			if channel == "SHOPEE" && o.Marketplace != nil && o.Marketplace.ShipByDateTime != nil {
				deadline = o.Marketplace.ShipByDateTime
			} else {
				orderDeadline := o.CreatedAt.Add(48 * time.Hour)
				deadline = &orderDeadline
			}

			if deadline != nil && deadline.Sub(now) < 24*time.Hour {
				isUrgent = true
			}

			variant := it.ItemSKU
			if it.MatchedSKU != "" {
				variant = it.MatchedSKU
			} else if it.Product != nil && it.Product.SKU != nil {
				variant = *it.Product.SKU
			}

			orderIdentifier := o.OrderNumber
			customerName := o.CustomerName
			if o.Marketplace != nil {
				if o.Marketplace.OrderSN != "" {
					orderIdentifier = o.Marketplace.OrderSN
				}
				if o.Marketplace.BuyerUsername != "" {
					customerName = o.Marketplace.BuyerUsername
				}
			}

			queue = append(queue, &production.ProductionQueueItem{
				JobID:             it.ID.String(),
				Source:            channel,
				OrderIdentifier:   orderIdentifier,
				CustomerName:      customerName,
				ItemName:          it.ProductName,
				VariantSKU:        variant,
				Quantity:          it.Quantity,
				WeightGrams:       weight,
				PrintTimeHours:    printHours,
				TotalPrintHours:   totalHours,
				AssignedMachineID: machineID,
				AssignedMachine:   machineName,
				Status:            o.Status,
				Deadline:          deadline,
				IsUrgent:          isUrgent,
				CreatedAt:         o.CreatedAt,
			})
		}
	}

	// Sort Queue
	sort.Slice(queue, func(i, j int) bool {
		if queue[i].IsUrgent != queue[j].IsUrgent {
			return queue[i].IsUrgent
		}
		if queue[i].Deadline != nil && queue[j].Deadline != nil {
			return queue[i].Deadline.Before(*queue[j].Deadline)
		}
		return queue[i].CreatedAt.Before(queue[j].CreatedAt)
	})

	return queue, nil
}

func (r *ProductionRepository) CompleteJob(source, jobID, machineID, filamentID string) (*production.CompleteJobResult, error) {
	resp := &production.CompleteJobResult{
		Status:            "success",
		DeductedFilaments: []string{},
	}

	itemUUID, err := uuid.Parse(jobID)
	if err != nil {
		return nil, fmt.Errorf("invalid job_id format: %w", err)
	}

	var item prodOrderItemGORM
	if err := r.db.Preload("Filaments").Preload("Product").
		Where("id = ?", itemUUID).
		First(&item).Error; err != nil {
		return nil, fmt.Errorf("order item not found: %w", err)
	}

	quantity := item.Quantity
	printHours := item.PrintTimeHours
	itemWeight := item.WeightGrams
	targetMachineID := machineID

	if item.Product != nil {
		if printHours <= 0 {
			printHours = item.Product.DefaultPrintTimeHours
		}
		if itemWeight <= 0 {
			itemWeight = item.Product.DefaultWeightGrams
		}
		if targetMachineID == "" && item.Product.DefaultMachineID != nil {
			targetMachineID = item.Product.DefaultMachineID.String()
		}
	}
	if printHours <= 0 {
		printHours = 1.0
	}
	if itemWeight <= 0 {
		itemWeight = 20.0
	}

	if targetMachineID == "" && item.MachineID != nil {
		targetMachineID = item.MachineID.String()
	}

	var filamentDeductions []struct {
		FilamentID string
		Grams      float64
	}

	if len(item.Filaments) > 0 {
		for _, f := range item.Filaments {
			if f.FilamentID != nil {
				filamentDeductions = append(filamentDeductions, struct {
					FilamentID string
					Grams      float64
				}{
					FilamentID: f.FilamentID.String(),
					Grams:      f.WeightUsedGrams * float64(quantity),
				})
			}
		}
	}

	// Update order status if not COMPLETED
	var order prodOrderGORM
	if err := r.db.Where("id = ?", item.OrderID).First(&order).Error; err == nil {
		now := time.Now()
		order.Status = "COMPLETED"
		order.CompletedAt = &now
		_ = r.db.Save(&order)
	}

	// Fallback filament deduction if caller supplied filamentID or if item has weight
	if len(filamentDeductions) == 0 && filamentID != "" {
		deductWeight := itemWeight
		if deductWeight <= 0 {
			deductWeight = 20.0
		}
		filamentDeductions = append(filamentDeductions, struct {
			FilamentID string
			Grams      float64
		}{
			FilamentID: filamentID,
			Grams:      deductWeight * float64(quantity),
		})
	}

	// 1. Deduct Filament Stock
	for _, fd := range filamentDeductions {
		filUUID, err := uuid.Parse(fd.FilamentID)
		if err != nil {
			continue
		}
		var fil prodFilamentGORM
		if err := r.db.Where("id = ?", filUUID).First(&fil).Error; err == nil {
			fil.CurrentStockGrams -= fd.Grams
			if fil.CurrentStockGrams < 0 {
				fil.CurrentStockGrams = 0
			}
			_ = r.db.Save(&fil)

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

	// 2. Increment Machine Hours
	totalHoursToAdd := math.Round(printHours*float64(quantity)*100) / 100
	if targetMachineID != "" && totalHoursToAdd > 0 {
		mUUID, err := uuid.Parse(targetMachineID)
		if err == nil {
			var machine prodMachineGORM
			if err := r.db.Preload("Parts").Where("id = ?", mUUID).First(&machine).Error; err == nil {
				newTotalHours := machine.TotalHoursUsed + totalHoursToAdd
				_ = r.db.Model(&prodMachineGORM{}).Where("id = ?", mUUID).Updates(map[string]interface{}{
					"total_hours_used": newTotalHours,
					"current_state":    "IDLE",
				})
				resp.AddedMachineHours = totalHoursToAdd

				for _, part := range machine.Parts {
					if part.LifespanHours > 0 && part.HoursUsedCurrent >= part.LifespanHours {
						resp.MaintenanceWarning = true
					}
				}
			}
		}
	}

	resp.Message = fmt.Sprintf("Pekerjaan cetak selesai. Filamen terpotong untuk %d pcs dan mesin beroperasi +%.2f jam.", quantity, totalHoursToAdd)
	return resp, nil
}
