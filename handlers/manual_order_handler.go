package handlers

import (
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"symetra-lab-backend/models"
	"symetra-lab-backend/pkg/costing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ManualOrderHandler struct {
	DB *gorm.DB
}

func NewManualOrderHandler(db *gorm.DB) *ManualOrderHandler {
	return &ManualOrderHandler{DB: db}
}

// generateOrderNumber membuat nomor pesanan unik misal ORD-20260912-7819
func generateOrderNumber() string {
	dateStr := time.Now().Format("20060102")
	randomNum := rand.Intn(9000) + 1000
	return fmt.Sprintf("ORD-%s-%d", dateStr, randomNum)
}

// GetOrders mengambil daftar pesanan manual / offline dari database
// GET /api/v1/orders
func (h *ManualOrderHandler) GetOrders(c *gin.Context) {
	userID := getUserID(c)
	query := h.DB.Model(&models.Order{}).Where("user_id = ?", userID)

	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	if paymentStatus := c.Query("payment_status"); paymentStatus != "" {
		query = query.Where("payment_status = ?", paymentStatus)
	}
	if search := c.Query("search"); search != "" {
		pat := "%" + search + "%"
		query = query.Where("LOWER(order_number) LIKE LOWER(?) OR LOWER(customer_name) LIKE LOWER(?) OR LOWER(notes) LIKE LOWER(?)", pat, pat, pat)
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	var total int64
	query.Count(&total)

	var orders []models.Order
	if err := query.Preload("Items.Product").
		Preload("Items.Machine").
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&orders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil daftar pesanan", "status": "error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"orders":      orders,
			"total":       total,
			"page":        page,
			"page_size":   pageSize,
			"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

// GetOrderByID mengambil detail lengkap 1 pesanan manual
// GET /api/v1/orders/:id
func (h *ManualOrderHandler) GetOrderByID(c *gin.Context) {
	id := c.Param("id")
	var order models.Order
	if err := h.DB.Preload("Items.Product").
		Preload("Items.Machine").
		Preload("Items.Filaments.Filament").
		Preload("Items.Components.Component").
		Where("id = ?", id).
		First(&order).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Pesanan tidak ditemukan", "status": "error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   order,
	})
}

// CreateOrder membuat pesanan manual baru dan otomatis mengkalkulasi HPP & modal fisik
// POST /api/v1/orders
func (h *ManualOrderHandler) CreateOrder(c *gin.Context) {
	var req models.CreateManualOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validasi form gagal: " + err.Error(), "status": "error"})
		return
	}

	var cfg models.ShopConfig
	_ = h.DB.First(&cfg).Error

	orderID := uuid.New().String()
	orderNum := generateOrderNumber()

	status := "PENDING"
	paymentStatus := "UNPAID"
	if req.PaymentStatus != "" {
		paymentStatus = req.PaymentStatus
	}
	source := "MANUAL"
	if req.Source != "" {
		source = req.Source
	}

	var totalRevenue float64
	var totalHPP float64
	var orderItems []models.OrderItem

	for _, itemInput := range req.Items {
		itemUUID := uuid.New().String()
		qty := itemInput.Quantity
		if qty < 1 {
			qty = 1
		}

		unitPrice := itemInput.SellingPrice
		itemRevenue := unitPrice * float64(qty)
		totalRevenue += itemRevenue

		orderItem := models.OrderItem{
			ID:           itemUUID,
			OrderID:      orderID,
			ProductID:    itemInput.ProductID,
			ProductName:  itemInput.ProductName,
			Quantity:     qty,
			SellingPrice: unitPrice,
			CreatedAt:    time.Now(),
		}

		// Jika item berasal dari Master Product katalog, hitung HPP dari BOM
		if itemInput.ProductID != nil && *itemInput.ProductID != "" {
			var prod models.Product
			err := h.DB.Preload("Filaments.Filament.Profile").
				Preload("Components.Component").
				Preload("PackagingPreset.Items.PackagingItem").
				Preload("PackagingItems.PackagingItem").
				Preload("DefaultMachine").
				Where("id = ?", *itemInput.ProductID).
				First(&prod).Error

			if err == nil {
				breakdown := costing.CalculateCostBreakdown(&prod, &cfg, nil)

				orderItem.HPP = math.Round(breakdown.BaseHPP*100) / 100
				orderItem.WeightGrams = prod.DefaultWeightGrams
				orderItem.PrintTimeHours = prod.DefaultPrintTimeHours
				orderItem.MachineID = prod.DefaultMachineID
				orderItem.EnergyCost = breakdown.ElectricityCost
				orderItem.DepreciationCost = breakdown.DepreciationCost

				// Snapshot relasi pemakaian filamen
				for _, pf := range prod.Filaments {
					filSnapshot := "Filamen"
					if pf.Filament != nil {
						filSnapshot = pf.Filament.ColorName
					}
					costGram := (cfg.FilamentPricePerRoll / 1000.0)
					if pf.Filament != nil && pf.Filament.PricePerRoll > 0 {
						costGram = pf.Filament.PricePerRoll / 1000.0
					}
					orderItem.Filaments = append(orderItem.Filaments, models.OrderFilament{
						ID:                   uuid.New().String(),
						OrderItemID:          itemUUID,
						FilamentID:           pf.FilamentID,
						FilamentSnapshotName: filSnapshot,
						WeightUsedGrams:      pf.WeightUsedGrams,
						CostContribution:     math.Round(pf.WeightUsedGrams*costGram*100) / 100,
						CreatedAt:            time.Now(),
					})
				}

				// Snapshot relasi pemakaian komponen
				for _, pc := range prod.Components {
					compName := "Komponen"
					costUnit := 0.0
					if pc.Component != nil {
						compName = pc.Component.Name
						costUnit = pc.Component.PricePerUnit
					}
					orderItem.Components = append(orderItem.Components, models.OrderComponent{
						ID:                    uuid.New().String(),
						OrderItemID:           itemUUID,
						ComponentID:           pc.ComponentID,
						ComponentSnapshotName: compName,
						Quantity:              pc.Quantity,
						CostContribution:      math.Round(costUnit*pc.Quantity*100) / 100,
						CreatedAt:             time.Now(),
					})
				}
			}
		} else {
			// Custom order tanpa katalog: gunakan estimasi jam & berat yang diisi
			weight := 20.0
			if itemInput.CustomWeight != nil && *itemInput.CustomWeight > 0 {
				weight = *itemInput.CustomWeight
			}
			printHours := 1.0
			if itemInput.CustomPrintTime != nil && *itemInput.CustomPrintTime > 0 {
				printHours = *itemInput.CustomPrintTime
			}

			costPerGram := cfg.FilamentPricePerRoll / 1000.0
			matCost := weight * costPerGram
			energyCost := (float64(cfg.PrinterPowerWatts) / 1000.0) * printHours * cfg.ElectricityTariffPerKwh
			deprecCost := (cfg.PrinterPrice / float64(cfg.PrinterLifespanHours)) * printHours

			unitHPP := matCost + energyCost + deprecCost
			orderItem.HPP = math.Round(unitHPP*100) / 100
			orderItem.WeightGrams = weight
			orderItem.PrintTimeHours = printHours
			orderItem.MachineID = itemInput.MachineID
			orderItem.EnergyCost = energyCost
			orderItem.DepreciationCost = deprecCost
		}

		totalHPP += orderItem.HPP * float64(qty)
		orderItems = append(orderItems, orderItem)
	}

	totalProfit := totalRevenue - totalHPP

	userID := getUserID(c)

	order := models.Order{
		ID:              orderID,
		UserID:          userID,
		OrderNumber:     orderNum,
		CustomerName:    req.CustomerName,
		CustomerContact: req.CustomerContact,
		TotalRevenue:    math.Round(totalRevenue*100) / 100,
		TotalHPP:        math.Round(totalHPP*100) / 100,
		TotalProfit:     math.Round(totalProfit*100) / 100,
		Status:          status,
		Notes:           req.Notes,
		Source:          source,
		PaymentStatus:   paymentStatus,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		Items:           orderItems,
	}

	if err := h.DB.Create(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan pesanan: " + err.Error(), "status": "error"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": fmt.Sprintf("Pesanan manual %s berhasil dibuat", orderNum),
		"status":  "success",
		"data":    order,
	})
}

// UpdateOrderStatus memperbarui status pengerjaan pesanan (PENDING, IN_PRODUCTION, COMPLETED, dll)
// PATCH /api/v1/orders/:id/status
func (h *ManualOrderHandler) UpdateOrderStatus(c *gin.Context) {
	id := c.Param("id")
	var req models.UpdateOrderStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Status wajib diisi", "status": "error"})
		return
	}

	var order models.Order
	if err := h.DB.Where("id = ?", id).First(&order).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Pesanan tidak ditemukan", "status": "error"})
		return
	}

	order.Status = req.Status
	now := time.Now()
	if req.Status == "IN_PRODUCTION" && order.StartedAt == nil {
		order.StartedAt = &now
	}
	if req.Status == "COMPLETED" && order.CompletedAt == nil {
		order.CompletedAt = &now
	}
	order.UpdatedAt = now

	if err := h.DB.Save(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memperbarui status: " + err.Error(), "status": "error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Status pesanan diperbarui menjadi %s", req.Status),
		"status":  "success",
		"data":    order,
	})
}

// UpdatePaymentStatus memperbarui status lunas pesanan (PAID, UNPAID)
// PATCH /api/v1/orders/:id/payment
func (h *ManualOrderHandler) UpdatePaymentStatus(c *gin.Context) {
	id := c.Param("id")
	var req models.UpdatePaymentStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Payment status wajib diisi", "status": "error"})
		return
	}

	var order models.Order
	if err := h.DB.Where("id = ?", id).First(&order).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Pesanan tidak ditemukan", "status": "error"})
		return
	}

	order.PaymentStatus = req.PaymentStatus
	order.UpdatedAt = time.Now()

	if err := h.DB.Save(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memperbarui pembayaran", "status": "error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Status pembayaran diperbarui menjadi %s", req.PaymentStatus),
		"status":  "success",
		"data":    order,
	})
}

// DeleteOrder menghapus pesanan manual
// DELETE /api/v1/orders/:id
func (h *ManualOrderHandler) DeleteOrder(c *gin.Context) {
	id := c.Param("id")
	if err := h.DB.Where("id = ?", id).Delete(&models.Order{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus pesanan", "status": "error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Pesanan berhasil dihapus",
		"status":  "success",
	})
}
