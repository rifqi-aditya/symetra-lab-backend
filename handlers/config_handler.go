package handlers

import (
	"net/http"
	"time"

	"symetra-lab-backend/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ConfigHandler struct {
	DB *gorm.DB
}

func NewConfigHandler(db *gorm.DB) *ConfigHandler {
	return &ConfigHandler{DB: db}
}

// GetShopConfig mengambil konfigurasi parameter bengkel aktif (tarif PLN, buffer gagal, dll)
// GET /api/v1/config/shop
func (h *ConfigHandler) GetShopConfig(c *gin.Context) {
	var cfg models.ShopConfig
	if err := h.DB.First(&cfg).Error; err != nil {
		// Jika belum ada, kembalikan default
		cfg = models.ShopConfig{
			ID:                      uuid.New().String(),
			UserID:                  DefaultAdminUserID,
			FilamentPricePerRoll:    200000,
			FilamentWeightGrams:     1000,
			ElectricityTariffPerKwh: 1700,
			PrinterPowerWatts:       200,
			PrinterPrice:            5000000,
			PrinterLifespanHours:    3000,
			FailureBufferPercent:    10,
			UpdatedAt:               time.Now(),
		}
		_ = h.DB.Create(&cfg)
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   cfg,
	})
}

// UpdateShopConfig memperbarui parameter operasional bengkel aktif
// PUT /api/v1/config/shop
func (h *ConfigHandler) UpdateShopConfig(c *gin.Context) {
	var cfg models.ShopConfig
	if err := h.DB.First(&cfg).Error; err != nil {
		cfg.ID = uuid.New().String()
		cfg.UserID = DefaultAdminUserID
	}

	var input models.ShopConfig
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format data tidak valid: " + err.Error(), "status": "error"})
		return
	}

	if input.FilamentPricePerRoll > 0 {
		cfg.FilamentPricePerRoll = input.FilamentPricePerRoll
	}
	if input.FilamentWeightGrams > 0 {
		cfg.FilamentWeightGrams = input.FilamentWeightGrams
	}
	if input.ElectricityTariffPerKwh > 0 {
		cfg.ElectricityTariffPerKwh = input.ElectricityTariffPerKwh
	}
	if input.PrinterPowerWatts > 0 {
		cfg.PrinterPowerWatts = input.PrinterPowerWatts
	}
	if input.PrinterPrice > 0 {
		cfg.PrinterPrice = input.PrinterPrice
	}
	if input.PrinterLifespanHours > 0 {
		cfg.PrinterLifespanHours = input.PrinterLifespanHours
	}
	if input.FailureBufferPercent >= 0 {
		cfg.FailureBufferPercent = input.FailureBufferPercent
	}
	cfg.UpdatedAt = time.Now()

	if err := h.DB.Save(&cfg).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan konfigurasi: " + err.Error(), "status": "error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Konfigurasi bengkel berhasil diperbarui",
		"status":  "success",
		"data":    cfg,
	})
}

// GetMarketplaces mengambil daftar platform marketplace dan skema fee potongannya
// GET /api/v1/config/marketplaces
func (h *ConfigHandler) GetMarketplaces(c *gin.Context) {
	var platforms []models.MarketplacePlatform
	if err := h.DB.Order("created_at ASC").Find(&platforms).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil platform marketplace", "status": "error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   platforms,
	})
}

// CreateMarketplace mendaftarkan skema potongan fee platform marketplace baru
// POST /api/v1/config/marketplaces
func (h *ConfigHandler) CreateMarketplace(c *gin.Context) {
	var platform models.MarketplacePlatform
	if err := c.ShouldBindJSON(&platform); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format data tidak valid: " + err.Error(), "status": "error"})
		return
	}

	platform.ID = uuid.New().String()
	platform.UserID = DefaultAdminUserID
	platform.CreatedAt = time.Now()

	if err := h.DB.Create(&platform).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat platform: " + err.Error(), "status": "error"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Platform marketplace berhasil ditambahkan",
		"status":  "success",
		"data":    platform,
	})
}

// UpdateMarketplace memperbarui skema potongan fee marketplace
// PUT /api/v1/config/marketplaces/:id
func (h *ConfigHandler) UpdateMarketplace(c *gin.Context) {
	id := c.Param("id")
	var platform models.MarketplacePlatform
	if err := h.DB.Where("id = ?", id).First(&platform).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Platform tidak ditemukan", "status": "error"})
		return
	}

	var input models.MarketplacePlatform
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format data tidak valid: " + err.Error(), "status": "error"})
		return
	}

	if input.Name != "" {
		platform.Name = input.Name
	}
	platform.CommissionPercent = input.CommissionPercent
	platform.PromoFeePercent = input.PromoFeePercent
	platform.FreeShippingPercent = input.FreeShippingPercent
	platform.OrderFeeIDR = input.OrderFeeIDR
	platform.IsActive = input.IsActive

	if err := h.DB.Save(&platform).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memperbarui platform", "status": "error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Platform marketplace berhasil diperbarui",
		"status":  "success",
		"data":    platform,
	})
}

// DeleteMarketplace menghapus skema platform marketplace
// DELETE /api/v1/config/marketplaces/:id
func (h *ConfigHandler) DeleteMarketplace(c *gin.Context) {
	id := c.Param("id")
	if err := h.DB.Where("id = ?", id).Delete(&models.MarketplacePlatform{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus platform", "status": "error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Platform berhasil dihapus",
		"status":  "success",
	})
}
