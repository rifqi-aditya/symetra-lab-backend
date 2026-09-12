package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"symetra-lab-backend/models"
	"symetra-lab-backend/pkg/costing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ProductHandler menangani request HTTP untuk master produk, BOM, dan kalkulasi HPP
type ProductHandler struct {
	db *gorm.DB
}

// NewProductHandler membuat instance baru dari ProductHandler
func NewProductHandler(db *gorm.DB) *ProductHandler {
	return &ProductHandler{db: db}
}

// getActiveConfigAndShopee mengambil konfigurasi bengkel dan parameter fee Shopee
func (h *ProductHandler) getActiveConfigAndShopee(userID string) (*models.ShopConfig, *models.MarketplacePlatform) {
	if h.db == nil {
		return nil, nil
	}

	var cfg models.ShopConfig
	if err := h.db.Where("user_id = ?", userID).First(&cfg).Error; err != nil {
		// Fallback ke first row apapun jika ada
		h.db.First(&cfg)
	}

	var shopee models.MarketplacePlatform
	if err := h.db.Where("LOWER(name) = ? AND is_active = ?", "shopee", true).First(&shopee).Error; err != nil {
		h.db.Where("is_active = ?", true).First(&shopee)
	}

	return &cfg, &shopee
}

// ListProducts mengambil daftar produk dengan opsi pencarian dan filter
func (h *ProductHandler) ListProducts(c *gin.Context) {
	if h.db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Koneksi database tidak tersedia"})
		return
	}

	userID := getUserID(c)

	category := c.Query("category")
	search := c.Query("search")
	parentSKU := c.Query("parent_sku")

	query := h.db.Model(&models.Product{})
	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}

	if category != "" {
		query = query.Where("category = ?", category)
	}

	if parentSKU != "" {
		query = query.Where("parent_sku = ?", parentSKU)
	}

	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("name ILIKE ? OR sku ILIKE ? OR parent_sku ILIKE ?", searchPattern, searchPattern, searchPattern)
	}

	var products []models.Product
	err := query.
		Preload("Filaments.Filament.Profile").
		Preload("Components.Component").
		Preload("PackagingItems.PackagingItem").
		Preload("PackagingPreset.Items.PackagingItem").
		Preload("DefaultMachine").
		Order("created_at DESC").
		Find(&products).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data produk: " + err.Error()})
		return
	}

	cfg, shopee := h.getActiveConfigAndShopee(userID)

	responses := make([]models.ProductResponse, len(products))
	for i := range products {
		breakdown := costing.CalculateCostBreakdown(&products[i], cfg, shopee)
		responses[i] = models.ProductResponse{
			Product:       products[i],
			CostBreakdown: breakdown,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   responses,
		"total":  len(responses),
	})
}

// GetProduct mengambil detail 1 produk lengkap dengan resep BOM dan rincian HPP
func (h *ProductHandler) GetProduct(c *gin.Context) {
	id := c.Param("id")
	userID := getUserID(c)

	var product models.Product
	err := h.db.
		Preload("Filaments.Filament.Profile").
		Preload("Components.Component").
		Preload("PackagingItems.PackagingItem").
		Preload("PackagingPreset.Items.PackagingItem").
		Preload("DefaultMachine").
		Where("id = ?", id).
		First(&product).Error

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Produk tidak ditemukan"})
		return
	}

	cfg, shopee := h.getActiveConfigAndShopee(userID)
	breakdown := costing.CalculateCostBreakdown(&product, cfg, shopee)

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": models.ProductResponse{
			Product:       product,
			CostBreakdown: breakdown,
		},
	})
}

// GetProductBySKU mencari produk instan berdasarkan SKU atau Parent SKU (vital untuk integrasi pesanan Shopee)
func (h *ProductHandler) GetProductBySKU(c *gin.Context) {
	skuQuery := strings.TrimSpace(c.Param("sku"))
	if skuQuery == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Parameter SKU tidak boleh kosong"})
		return
	}

	userID := getUserID(c)

	var product models.Product
	err := h.db.
		Preload("Filaments.Filament.Profile").
		Preload("Components.Component").
		Preload("PackagingItems.PackagingItem").
		Preload("PackagingPreset.Items.PackagingItem").
		Preload("DefaultMachine").
		Where("sku = ? OR parent_sku = ?", skuQuery, skuQuery).
		First(&product).Error

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": fmt.Sprintf("Produk dengan SKU '%s' tidak ditemukan di database bengkel", skuQuery),
		})
		return
	}

	cfg, shopee := h.getActiveConfigAndShopee(userID)
	breakdown := costing.CalculateCostBreakdown(&product, cfg, shopee)

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": models.ProductResponse{
			Product:       product,
			CostBreakdown: breakdown,
		},
	})
}

type CreateProductFilamentItem struct {
	FilamentID      *string `json:"filament_id"`
	WeightUsedGrams float64 `json:"weight_used_grams"`
}

type CreateProductComponentItem struct {
	ComponentID   string  `json:"component_id"`
	Quantity      float64 `json:"quantity"`
	MarkupPercent float64 `json:"markup_percent"`
}

type CreateProductPackagingItem struct {
	PackagingItemID string  `json:"packaging_item_id"`
	QuantityUsed    float64 `json:"quantity_used"`
}

// CreateProductPayload merepresentasikan data input pembuatan produk
type CreateProductPayload struct {
	Name                  string                       `json:"name" binding:"required"`
	ParentSKU             *string                      `json:"parent_sku"`
	SKU                   *string                      `json:"sku"`
	Description           *string                      `json:"description"`
	Category              string                       `json:"category"`
	ThumbnailURL          *string                      `json:"thumbnail_url"`
	DesignLink            *string                      `json:"design_link"`
	DefaultWeightGrams    float64                      `json:"default_weight_grams"`
	DefaultPrintTimeHours float64                      `json:"default_print_time_hours"`
	DefaultMachineID      *string                      `json:"default_machine_id"`
	PackagingPresetID     *string                      `json:"packaging_preset_id"`
	BatchSize             int                          `json:"batch_size"`
	PackingFeeIDR         int                          `json:"packing_fee_idr"`
	TargetMarginPercent   int                          `json:"target_margin_percent"`
	Filaments             []CreateProductFilamentItem  `json:"filaments"`
	Components            []CreateProductComponentItem `json:"components"`
	PackagingItems        []CreateProductPackagingItem `json:"packaging_items"`
}

// CreateProduct membuat produk baru dengan validasi SKU dan kalkulasi otomatis
func (h *ProductHandler) CreateProduct(c *gin.Context) {
	userID := getUserID(c)

	var payload CreateProductPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Payload tidak valid: " + err.Error()})
		return
	}

	// Tangani SKU & Parent SKU yang scalable
	var parentSKU string
	var finalSKU string

	if payload.ParentSKU != nil && strings.TrimSpace(*payload.ParentSKU) != "" {
		parentSKU = costing.CleanSKUPart(*payload.ParentSKU)
	}

	if payload.SKU != nil && strings.TrimSpace(*payload.SKU) != "" {
		finalSKU = costing.CleanSKUPart(*payload.SKU)
		if parentSKU == "" {
			parts := strings.Split(finalSKU, "-")
			if len(parts) >= 2 {
				parentSKU = strings.Join(parts[:len(parts)-1], "-")
			} else {
				parentSKU = finalSKU
			}
		}
	} else if parentSKU != "" {
		_, full := costing.FormatScalableSKU(parentSKU, "STD")
		finalSKU = full
	}

	// Cek keunikan SKU jika terisi
	if finalSKU != "" {
		var existing models.Product
		if err := h.db.Where("sku = ?", finalSKU).First(&existing).Error; err == nil {
			c.JSON(http.StatusConflict, gin.H{"error": fmt.Sprintf("SKU '%s' sudah digunakan oleh produk lain", finalSKU)})
			return
		}
	}

	batchSize := payload.BatchSize
	if batchSize <= 0 {
		batchSize = 1
	}

	margin := payload.TargetMarginPercent
	if margin <= 0 {
		margin = 30
	}

	product := models.Product{
		ID:                    uuid.New().String(),
		UserID:                userID,
		Name:                  strings.TrimSpace(payload.Name),
		ParentSKU:             nilIfEmpty(parentSKU),
		SKU:                   nilIfEmpty(finalSKU),
		Description:           payload.Description,
		Category:              strings.TrimSpace(payload.Category),
		ThumbnailURL:          payload.ThumbnailURL,
		DesignLink:            payload.DesignLink,
		DefaultWeightGrams:    payload.DefaultWeightGrams,
		DefaultPrintTimeHours: payload.DefaultPrintTimeHours,
		DefaultMachineID:      payload.DefaultMachineID,
		PackagingPresetID:     payload.PackagingPresetID,
		BatchSize:             batchSize,
		PackingFeeIDR:         payload.PackingFeeIDR,
		TargetMarginPercent:   margin,
	}

	err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&product).Error; err != nil {
			return err
		}
		for _, f := range payload.Filaments {
			pf := models.ProductFilament{
				ID:              uuid.New().String(),
				ProductID:       product.ID,
				FilamentID:      f.FilamentID,
				WeightUsedGrams: f.WeightUsedGrams,
			}
			if err := tx.Create(&pf).Error; err != nil {
				return err
			}
		}
		for _, comp := range payload.Components {
			pc := models.ProductComponent{
				ID:            uuid.New().String(),
				ProductID:     product.ID,
				ComponentID:   nilIfEmpty(comp.ComponentID),
				Quantity:      comp.Quantity,
				MarkupPercent: comp.MarkupPercent,
			}
			if err := tx.Create(&pc).Error; err != nil {
				return err
			}
		}
		for _, pack := range payload.PackagingItems {
			pi := models.ProductPackagingItem{
				ID:              uuid.New().String(),
				ProductID:       product.ID,
				PackagingItemID: pack.PackagingItemID,
				QuantityUsed:    pack.QuantityUsed,
			}
			if err := tx.Create(&pi).Error; err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan produk: " + err.Error()})
		return
	}

	cfg, shopee := h.getActiveConfigAndShopee(userID)
	breakdown := costing.CalculateCostBreakdown(&product, cfg, shopee)
	product.BaseHPP = breakdown.BaseHPP
	product.BaseSellingPrice = breakdown.BaseSellingPrice
	_ = h.db.Model(&product).Updates(map[string]interface{}{
		"base_hpp":           breakdown.BaseHPP,
		"base_selling_price": breakdown.BaseSellingPrice,
	})

	c.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"message": "Produk berhasil dibuat",
		"data": models.ProductResponse{
			Product:       product,
			CostBreakdown: breakdown,
		},
	})
}

// UpdateProduct memperbarui data produk, SKU, atau konfigurasi print
func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	id := c.Param("id")
	userID := getUserID(c)

	var product models.Product
	if err := h.db.Where("id = ?", id).First(&product).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Produk tidak ditemukan"})
		return
	}

	var payload CreateProductPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Payload tidak valid: " + err.Error()})
		return
	}

	// Update fields
	product.Name = strings.TrimSpace(payload.Name)
	if payload.Category != "" {
		product.Category = strings.TrimSpace(payload.Category)
	}
	product.Description = payload.Description
	product.ThumbnailURL = payload.ThumbnailURL
	product.DesignLink = payload.DesignLink
	product.DefaultWeightGrams = payload.DefaultWeightGrams
	product.DefaultPrintTimeHours = payload.DefaultPrintTimeHours
	product.DefaultMachineID = payload.DefaultMachineID
	product.PackagingPresetID = payload.PackagingPresetID
	if payload.BatchSize > 0 {
		product.BatchSize = payload.BatchSize
	}
	product.PackingFeeIDR = payload.PackingFeeIDR
	if payload.TargetMarginPercent > 0 {
		product.TargetMarginPercent = payload.TargetMarginPercent
	}

	// SKU updates
	if payload.SKU != nil && strings.TrimSpace(*payload.SKU) != "" {
		cleanSKU := costing.CleanSKUPart(*payload.SKU)
		var existing models.Product
		if err := h.db.Where("sku = ? AND id != ?", cleanSKU, id).First(&existing).Error; err == nil {
			c.JSON(http.StatusConflict, gin.H{"error": fmt.Sprintf("SKU '%s' sudah digunakan oleh produk lain", cleanSKU)})
			return
		}
		product.SKU = &cleanSKU
	}

	if payload.ParentSKU != nil && strings.TrimSpace(*payload.ParentSKU) != "" {
		cleanParent := costing.CleanSKUPart(*payload.ParentSKU)
		product.ParentSKU = &cleanParent
	}

	err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&product).Error; err != nil {
			return err
		}
		if payload.Filaments != nil {
			tx.Where("product_id = ?", id).Delete(&models.ProductFilament{})
			for _, f := range payload.Filaments {
				pf := models.ProductFilament{
					ID:              uuid.New().String(),
					ProductID:       product.ID,
					FilamentID:      f.FilamentID,
					WeightUsedGrams: f.WeightUsedGrams,
				}
				if err := tx.Create(&pf).Error; err != nil {
					return err
				}
			}
		}
		if payload.Components != nil {
			tx.Where("product_id = ?", id).Delete(&models.ProductComponent{})
			for _, comp := range payload.Components {
				pc := models.ProductComponent{
					ID:            uuid.New().String(),
					ProductID:     product.ID,
					ComponentID:   nilIfEmpty(comp.ComponentID),
					Quantity:      comp.Quantity,
					MarkupPercent: comp.MarkupPercent,
				}
				if err := tx.Create(&pc).Error; err != nil {
					return err
				}
			}
		}
		if payload.PackagingItems != nil {
			tx.Where("product_id = ?", id).Delete(&models.ProductPackagingItem{})
			for _, pack := range payload.PackagingItems {
				pi := models.ProductPackagingItem{
					ID:              uuid.New().String(),
					ProductID:       product.ID,
					PackagingItemID: pack.PackagingItemID,
					QuantityUsed:    pack.QuantityUsed,
				}
				if err := tx.Create(&pi).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memperbarui produk: " + err.Error()})
		return
	}

	// Reload BOM relationships untuk re-kalkulasi
	h.db.
		Preload("Filaments.Filament.Profile").
		Preload("Components.Component").
		Preload("PackagingItems.PackagingItem").
		Preload("PackagingPreset.Items.PackagingItem").
		Preload("DefaultMachine").
		Where("id = ?", id).
		First(&product)

	cfg, shopee := h.getActiveConfigAndShopee(userID)
	breakdown := costing.CalculateCostBreakdown(&product, cfg, shopee)
	product.BaseHPP = breakdown.BaseHPP
	product.BaseSellingPrice = breakdown.BaseSellingPrice
	_ = h.db.Model(&product).Updates(map[string]interface{}{
		"base_hpp":           breakdown.BaseHPP,
		"base_selling_price": breakdown.BaseSellingPrice,
	})

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Produk berhasil diperbarui",
		"data": models.ProductResponse{
			Product:       product,
			CostBreakdown: breakdown,
		},
	})
}

// DeleteProduct menghapus produk secara aman (safe delete) jika belum pernah ada di transaksi order
func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	id := c.Param("id")

	var product models.Product
	if err := h.db.Where("id = ?", id).First(&product).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Produk tidak ditemukan"})
		return
	}

	// Cek referensi di tabel order_items jika tabel ada
	if h.db.Migrator().HasTable("order_items") {
		var orderItemCount int64
		h.db.Table("order_items").Where("product_id = ?", id).Count(&orderItemCount)
		if orderItemCount > 0 {
			c.JSON(http.StatusConflict, gin.H{
				"error": fmt.Sprintf("Produk tidak dapat dihapus karena tercatat pada %d riwayat pesanan (order_items)", orderItemCount),
			})
			return
		}
	}

	// Hapus relasi BOM produk terlebih dahulu
	h.db.Where("product_id = ?", id).Delete(&models.ProductFilament{})
	h.db.Where("product_id = ?", id).Delete(&models.ProductComponent{})
	h.db.Where("product_id = ?", id).Delete(&models.ProductPackagingItem{})

	// Hapus produk
	if err := h.db.Delete(&product).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus produk: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Produk beserta resep BOM berhasil dihapus",
	})
}

// AutoGenerateDraftSKUs memberikan draft SKU otomatis yang scalable untuk produk yang belum memiliki SKU
func (h *ProductHandler) AutoGenerateDraftSKUs(c *gin.Context) {
	var products []models.Product
	err := h.db.Where("sku IS NULL OR sku = ''").Find(&products).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil produk: " + err.Error()})
		return
	}

	updatedCount := 0
	for _, p := range products {
		catCode := getCategoryCode(p.Category)
		nameCode := getProductNameCode(p.Name)
		parentSKU := fmt.Sprintf("%s-%s", catCode, nameCode)
		fullSKU := fmt.Sprintf("%s-STD", parentSKU)

		// Cek apakah fullSKU sudah dipakai produk lain
		var count int64
		h.db.Model(&models.Product{}).Where("sku = ?", fullSKU).Count(&count)
		if count > 0 {
			fullSKU = fmt.Sprintf("%s-%d", fullSKU, p.TimesOrdered+1)
		}

		p.ParentSKU = &parentSKU
		p.SKU = &fullSKU
		if err := h.db.Model(&models.Product{}).Where("id = ?", p.ID).Updates(map[string]interface{}{
			"parent_sku": parentSKU,
			"sku":        fullSKU,
		}).Error; err == nil {
			updatedCount++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       fmt.Sprintf("Berhasil membuat draft SKU untuk %d produk", updatedCount),
		"total_updated": updatedCount,
	})
}

// Helper untuk menyingkat nama kategori menjadi 3 huruf kapital standar
func getCategoryCode(cat string) string {
	c := strings.ToUpper(strings.TrimSpace(cat))
	switch {
	case strings.Contains(c, "BOX") || strings.Contains(c, "FUNCTIONAL"):
		return "BOX"
	case strings.Contains(c, "CLICKER") || strings.Contains(c, "FIDGET"):
		return "CLK"
	case strings.Contains(c, "KEYCHAIN"):
		return "KEY"
	case strings.Contains(c, "FLEXI") || strings.Contains(c, "ARTICULAT"):
		return "TOY"
	case strings.Contains(c, "DIORAMA") || strings.Contains(c, "MINIATURE"):
		return "DEC"
	case strings.Contains(c, "ORGANIZER") || strings.Contains(c, "SKADIS"):
		return "ORG"
	default:
		clean := costing.CleanSKUPart(c)
		if len(clean) >= 3 {
			return clean[:3]
		}
		return "PRD"
	}
}

// Helper untuk menyingkat nama produk menjadi 3-6 huruf kode model
func getProductNameCode(name string) string {
	cleaned := costing.CleanSKUPart(name)
	words := strings.Split(cleaned, "-")
	var codeParts []string
	for _, w := range words {
		// Abaikan kata umum non-identitas
		if w == "DAN" || w == "DENGAN" || w == "FOR" || w == "THE" || w == "ITEM" {
			continue
		}
		if len(w) > 4 {
			codeParts = append(codeParts, w[:4])
		} else {
			codeParts = append(codeParts, w)
		}
		if len(strings.Join(codeParts, "")) >= 6 {
			break
		}
	}
	res := strings.Join(codeParts, "")
	if len(res) > 8 {
		return res[:8]
	}
	if res == "" {
		return "ITEM"
	}
	return res
}

func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
