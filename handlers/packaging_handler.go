package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"symetra-lab-backend/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PackagingHandler menangani operasi CRUD bahan kemasan dan preset kemasan
type PackagingHandler struct {
	db *gorm.DB
}

// NewPackagingHandler membuat instance baru PackagingHandler
func NewPackagingHandler(db *gorm.DB) *PackagingHandler {
	return &PackagingHandler{db: db}
}

// --- SUB-MODUL 1: BAHAN KEMASAN INDIVIDUAL (Packaging Items) ---

// GetAllPackagingItems mengambil seluruh bahan kemasan milik user
// GET /api/v1/packaging-items
func (h *PackagingHandler) GetAllPackagingItems(c *gin.Context) {
	userID := getUserID(c)

	var items []models.PackagingItem
	query := h.db.Where("user_id = ?", userID)

	// Filter berdasarkan kategori (BOX, PLASTIC, LABEL, OTHER)
	if category := strings.TrimSpace(c.Query("category")); category != "" {
		query = query.Where("UPPER(category) = UPPER(?)", category)
	}

	// Filter pencarian nama
	if search := strings.TrimSpace(c.Query("q")); search != "" {
		query = query.Where("LOWER(name) LIKE LOWER(?)", "%"+search+"%")
	}

	// Sorting
	sortBy := strings.ToLower(c.DefaultQuery("sort", "name"))
	order := strings.ToUpper(c.DefaultQuery("order", "ASC"))
	if order != "ASC" && order != "DESC" {
		order = "ASC"
	}

	switch sortBy {
	case "unit_cost":
		query = query.Order("unit_cost " + order)
	case "stock":
		query = query.Order("stock_quantity " + order)
	case "created_at":
		query = query.Order("created_at " + order)
	default:
		query = query.Order("name " + order)
	}

	if err := query.Find(&items).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Gagal mengambil data bahan kemasan: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   items,
		"total":  len(items),
	})
}

// GetPackagingItemByID mengambil rincian 1 bahan kemasan
// GET /api/v1/packaging-items/:id
func (h *PackagingHandler) GetPackagingItemByID(c *gin.Context) {
	userID := getUserID(c)
	id := c.Param("id")

	var item models.PackagingItem
	if err := h.db.Where("id = ? AND user_id = ?", id, userID).First(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "Bahan kemasan tidak ditemukan"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Gagal mengambil data bahan kemasan: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   item,
	})
}

// CreatePackagingItemRequest payload pembuatan bahan kemasan
type CreatePackagingItemRequest struct {
	Name             string   `json:"name" binding:"required"`
	Category         *string  `json:"category"`
	UnitType         *string  `json:"unit_type"`
	PurchasePrice    float64  `json:"purchase_price" binding:"gte=0"`
	PurchaseQuantity *float64 `json:"purchase_quantity"`
	StockQuantity    *float64 `json:"stock_quantity"`
}

// CreatePackagingItem menambahkan bahan kemasan baru
// POST /api/v1/packaging-items
func (h *PackagingHandler) CreatePackagingItem(c *gin.Context) {
	userID := getUserID(c)

	var req CreatePackagingItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Input tidak valid: " + err.Error()})
		return
	}

	cat := "OTHER"
	if req.Category != nil && strings.TrimSpace(*req.Category) != "" {
		cat = strings.ToUpper(strings.TrimSpace(*req.Category))
	}

	unitType := "PCS"
	if req.UnitType != nil && strings.TrimSpace(*req.UnitType) != "" {
		unitType = strings.ToUpper(strings.TrimSpace(*req.UnitType))
	}

	purchaseQty := 1.0
	if req.PurchaseQuantity != nil && *req.PurchaseQuantity > 0 {
		purchaseQty = *req.PurchaseQuantity
	}

	stockQty := 0.0
	if req.StockQuantity != nil && *req.StockQuantity >= 0 {
		stockQty = *req.StockQuantity
	}

	item := models.PackagingItem{
		ID:               uuid.New().String(),
		UserID:           userID,
		Name:             strings.TrimSpace(req.Name),
		Category:         cat,
		UnitType:         unitType,
		PurchasePrice:    req.PurchasePrice,
		PurchaseQuantity: purchaseQty,
		StockQuantity:    stockQty,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
	item.UnitCost = item.CalculateUnitCost()

	if err := h.db.Create(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Gagal menyimpan bahan kemasan: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"message": "Bahan kemasan berhasil ditambahkan",
		"data":    item,
	})
}

// UpdatePackagingItemRequest payload pembaruan bahan kemasan
type UpdatePackagingItemRequest struct {
	Name             *string  `json:"name"`
	Category         *string  `json:"category"`
	UnitType         *string  `json:"unit_type"`
	PurchasePrice    *float64 `json:"purchase_price"`
	PurchaseQuantity *float64 `json:"purchase_quantity"`
	StockQuantity    *float64 `json:"stock_quantity"`
}

// UpdatePackagingItem memperbarui data bahan kemasan
// PUT /api/v1/packaging-items/:id
func (h *PackagingHandler) UpdatePackagingItem(c *gin.Context) {
	userID := getUserID(c)
	id := c.Param("id")

	var item models.PackagingItem
	if err := h.db.Where("id = ? AND user_id = ?", id, userID).First(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "Bahan kemasan tidak ditemukan"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Gagal mengambil data bahan kemasan: " + err.Error()})
		return
	}

	var req UpdatePackagingItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Input tidak valid: " + err.Error()})
		return
	}

	recalc := false
	if req.Name != nil && strings.TrimSpace(*req.Name) != "" {
		item.Name = strings.TrimSpace(*req.Name)
	}
	if req.Category != nil {
		item.Category = strings.ToUpper(strings.TrimSpace(*req.Category))
	}
	if req.UnitType != nil {
		item.UnitType = strings.ToUpper(strings.TrimSpace(*req.UnitType))
	}
	if req.PurchasePrice != nil && *req.PurchasePrice >= 0 {
		item.PurchasePrice = *req.PurchasePrice
		recalc = true
	}
	if req.PurchaseQuantity != nil && *req.PurchaseQuantity > 0 {
		item.PurchaseQuantity = *req.PurchaseQuantity
		recalc = true
	}
	if req.StockQuantity != nil && *req.StockQuantity >= 0 {
		item.StockQuantity = *req.StockQuantity
	}

	if recalc {
		item.UnitCost = item.CalculateUnitCost()
	}
	item.UpdatedAt = time.Now()

	if err := h.db.Save(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Gagal memperbarui bahan kemasan: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Bahan kemasan berhasil diperbarui",
		"data":    item,
	})
}

// DeletePackagingItem menghapus bahan kemasan dengan safety check relasi
// DELETE /api/v1/packaging-items/:id
func (h *PackagingHandler) DeletePackagingItem(c *gin.Context) {
	userID := getUserID(c)
	id := c.Param("id")

	var item models.PackagingItem
	if err := h.db.Where("id = ? AND user_id = ?", id, userID).First(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "Bahan kemasan tidak ditemukan"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Gagal mengambil data bahan kemasan: " + err.Error()})
		return
	}

	force := c.Query("force") == "true"
	if !force {
		// 1. Cek apakah dipakai di preset
		var presetCount int64
		if h.db.Migrator().HasTable("packaging_preset_items") {
			_ = h.db.Table("packaging_preset_items").Where("packaging_item_id = ?", id).Count(&presetCount).Error
		}

		// 2. Cek apakah dipakai di resep produk
		var productCount int64
		if h.db.Migrator().HasTable("product_packaging_items") {
			_ = h.db.Table("product_packaging_items").Where("packaging_item_id = ?", id).Count(&productCount).Error
		}

		if presetCount > 0 || productCount > 0 {
			c.JSON(http.StatusConflict, gin.H{
				"status":        "error",
				"message":       fmt.Sprintf("Bahan kemasan sedang digunakan pada %d preset dan %d produk. Gunakan ?force=true jika tetap ingin menghapus.", presetCount, productCount),
				"preset_usage":  presetCount,
				"product_usage": productCount,
			})
			return
		}
	}

	if err := h.db.Delete(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Gagal menghapus bahan kemasan: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Bahan kemasan berhasil dihapus",
	})
}

// --- SUB-MODUL 2: PRESET KEMASAN / BUNDEL PACKING (Packaging Presets) ---

// GetAllPresets mengambil seluruh preset kemasan beserta komponen item dan total biayanya
// GET /api/v1/packaging-presets
func (h *PackagingHandler) GetAllPresets(c *gin.Context) {
	userID := getUserID(c)

	var presets []models.PackagingPreset
	if err := h.db.Where("user_id = ?", userID).
		Preload("Items.PackagingItem").
		Order("name asc").
		Find(&presets).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Gagal mengambil data preset: " + err.Error()})
		return
	}

	responses := make([]models.PackagingPresetResponse, len(presets))
	for i, p := range presets {
		responses[i] = p.ToResponse()
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   responses,
		"total":  len(responses),
	})
}

// GetPresetByID mengambil detail 1 preset kemasan
// GET /api/v1/packaging-presets/:id
func (h *PackagingHandler) GetPresetByID(c *gin.Context) {
	userID := getUserID(c)
	id := c.Param("id")

	var preset models.PackagingPreset
	if err := h.db.Where("id = ? AND user_id = ?", id, userID).
		Preload("Items.PackagingItem").
		First(&preset).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "Preset kemasan tidak ditemukan"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Gagal mengambil data preset: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   preset.ToResponse(),
	})
}

// CreatePresetItemInput item masukan saat membuat preset
type CreatePresetItemInput struct {
	PackagingItemID string  `json:"packaging_item_id" binding:"required"`
	QuantityUsed    float64 `json:"quantity_used" binding:"required,gt=0"`
}

// CreatePresetRequest payload pembuatan preset kemasan
type CreatePresetRequest struct {
	Name        string                  `json:"name" binding:"required"`
	Description *string                 `json:"description"`
	Items       []CreatePresetItemInput `json:"items"`
}

// CreatePreset membuat preset kemasan baru beserta rincian itemnya
// POST /api/v1/packaging-presets
func (h *PackagingHandler) CreatePreset(c *gin.Context) {
	userID := getUserID(c)

	var req CreatePresetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Input tidak valid: " + err.Error()})
		return
	}

	presetID := uuid.New().String()
	preset := models.PackagingPreset{
		ID:          presetID,
		UserID:      userID,
		Name:        strings.TrimSpace(req.Name),
		Description: req.Description,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Eksekusi dalam transaksi database
	err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&preset).Error; err != nil {
			return err
		}

		for _, itemInput := range req.Items {
			pItem := models.PackagingPresetItem{
				ID:              uuid.New().String(),
				PresetID:        presetID,
				PackagingItemID: itemInput.PackagingItemID,
				QuantityUsed:    itemInput.QuantityUsed,
				CreatedAt:       time.Now(),
			}
			if err := tx.Create(&pItem).Error; err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Gagal membuat preset kemasan: " + err.Error()})
		return
	}

	// Load kembali relasi lengkap untuk response
	_ = h.db.Where("id = ?", presetID).Preload("Items.PackagingItem").First(&preset).Error

	c.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"message": "Preset kemasan berhasil dibuat",
		"data":    preset.ToResponse(),
	})
}

// UpdatePresetRequest payload pembaruan preset kemasan
type UpdatePresetRequest struct {
	Name        *string                  `json:"name"`
	Description *string                  `json:"description"`
	Items       *[]CreatePresetItemInput `json:"items"`
}

// UpdatePreset memperbarui nama, deskripsi, atau susunan item preset
// PUT /api/v1/packaging-presets/:id
func (h *PackagingHandler) UpdatePreset(c *gin.Context) {
	userID := getUserID(c)
	id := c.Param("id")

	var preset models.PackagingPreset
	if err := h.db.Where("id = ? AND user_id = ?", id, userID).First(&preset).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "Preset kemasan tidak ditemukan"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Gagal mengambil data preset: " + err.Error()})
		return
	}

	var req UpdatePresetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Input tidak valid: " + err.Error()})
		return
	}

	err := h.db.Transaction(func(tx *gorm.DB) error {
		if req.Name != nil && strings.TrimSpace(*req.Name) != "" {
			preset.Name = strings.TrimSpace(*req.Name)
		}
		if req.Description != nil {
			preset.Description = req.Description
		}
		preset.UpdatedAt = time.Now()

		if err := tx.Save(&preset).Error; err != nil {
			return err
		}

		// Jika ada array items baru, ganti komposisi item preset
		if req.Items != nil {
			if err := tx.Where("preset_id = ?", id).Delete(&models.PackagingPresetItem{}).Error; err != nil {
				return err
			}

			for _, itemInput := range *req.Items {
				pItem := models.PackagingPresetItem{
					ID:              uuid.New().String(),
					PresetID:        id,
					PackagingItemID: itemInput.PackagingItemID,
					QuantityUsed:    itemInput.QuantityUsed,
					CreatedAt:       time.Now(),
				}
				if err := tx.Create(&pItem).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Gagal memperbarui preset: " + err.Error()})
		return
	}

	// Reload relasi lengkap
	_ = h.db.Where("id = ?", id).Preload("Items.PackagingItem").First(&preset).Error

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Preset kemasan berhasil diperbarui",
		"data":    preset.ToResponse(),
	})
}

// DeletePreset menghapus preset kemasan beserta relasi itemnya
// DELETE /api/v1/packaging-presets/:id
func (h *PackagingHandler) DeletePreset(c *gin.Context) {
	userID := getUserID(c)
	id := c.Param("id")

	var preset models.PackagingPreset
	if err := h.db.Where("id = ? AND user_id = ?", id, userID).First(&preset).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "Preset kemasan tidak ditemukan"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Gagal mengambil data preset: " + err.Error()})
		return
	}

	// Hapus item-item dalam preset terlebih dahulu
	_ = h.db.Where("preset_id = ?", id).Delete(&models.PackagingPresetItem{}).Error

	if err := h.db.Delete(&preset).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Gagal menghapus preset: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Preset kemasan berhasil dihapus",
	})
}
