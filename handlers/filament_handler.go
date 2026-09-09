package handlers

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"symetra-lab-backend/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	// DefaultAdminUserID adalah ID user default (admin@symetralab.com) di Supabase
	DefaultAdminUserID = "4aecd5c5-c0a9-4f52-ba2f-b6f4272218b4"
)

// FilamentHandler menangani operasi CRUD untuk filamen dan profil filamen
type FilamentHandler struct {
	db *gorm.DB
}

// NewFilamentHandler membuat instance baru FilamentHandler
func NewFilamentHandler(db *gorm.DB) *FilamentHandler {
	return &FilamentHandler{db: db}
}

// Helper untuk mendapatkan user_id aktif (dari header, query, atau default admin)
func getUserID(c *gin.Context) string {
	if uid := c.GetHeader("X-User-ID"); uid != "" {
		return strings.TrimSpace(uid)
	}
	if uid := c.Query("user_id"); uid != "" {
		return strings.TrimSpace(uid)
	}
	if val, exists := c.Get("user_id"); exists {
		if uidStr, ok := val.(string); ok && uidStr != "" {
			return uidStr
		}
	}
	return DefaultAdminUserID
}

// GetAllFilaments mengambil semua roll filamen milik user beserta preload relasi Profile
// GET /api/v1/filaments
func (h *FilamentHandler) GetAllFilaments(c *gin.Context) {
	userID := getUserID(c)

	var filaments []models.Filament
	query := h.db.Where("user_id = ?", userID).Preload("Profile").Order("created_at desc")

	// Filter opsional berdasarkan brand atau material
	if brand := c.Query("brand"); brand != "" {
		query = query.Joins("JOIN filament_profiles ON filament_profiles.id = filaments.profile_id").
			Where("LOWER(filament_profiles.brand) = LOWER(?)", brand)
	}
	if material := c.Query("material_type"); material != "" {
		query = query.Joins("JOIN filament_profiles ON filament_profiles.id = filaments.profile_id").
			Where("LOWER(filament_profiles.material_type) = LOWER(?)", material)
	}

	if err := query.Find(&filaments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Gagal mengambil data filamen: " + err.Error()})
		return
	}

	// Flattening ke bentuk FilamentResponse
	responses := make([]models.FilamentResponse, len(filaments))
	for i, f := range filaments {
		responses[i] = f.ToResponse()
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   responses,
		"total":  len(responses),
	})
}

// GetFilamentByID mengambil detail 1 roll filamen berdasarkan ID
// GET /api/v1/filaments/:id
func (h *FilamentHandler) GetFilamentByID(c *gin.Context) {
	userID := getUserID(c)
	id := c.Param("id")

	var filament models.Filament
	if err := h.db.Where("id = ? AND user_id = ?", id, userID).Preload("Profile").First(&filament).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "Filamen tidak ditemukan"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Gagal mengambil data filamen: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   filament.ToResponse(),
	})
}

// CreateFilamentRequest payload untuk membuat filamen baru
type CreateFilamentRequest struct {
	Brand                    string   `json:"brand" binding:"required"`
	MaterialType             string   `json:"material_type" binding:"required"`
	DiameterMM               *float64 `json:"diameter_mm"`
	NozzleTemp               *int     `json:"nozzle_temp"`
	BedTemp                  *int     `json:"bed_temp"`
	RetractionLength         *float64 `json:"retraction_length"`
	FlowRatio                *float64 `json:"flow_ratio"`
	PressureAdvance          *float64 `json:"pressure_advance"`
	CoolingFanPercent        *int     `json:"cooling_fan_percent"`
	MaxVolumetricSpeed       *float64 `json:"max_volumetric_speed"`
	EmptySpoolWeightGrams    *float64 `json:"empty_spool_weight_grams"`
	SpoolWeightGrams         *float64 `json:"spool_weight_grams"`
	SpoolOuterDiameterMM     *float64 `json:"spool_outer_diameter_mm"`
	SpoolInnerHoleDiameterMM *float64 `json:"spool_inner_hole_diameter_mm"`
	SpoolWidthMM             *float64 `json:"spool_width_mm"`

	// Roll / Varian Warna
	ProfileID              *string  `json:"profile_id"`
	ColorName              string   `json:"color_name" binding:"required"`
	ColorHex               string   `json:"color_hex"`
	SKU                    *string  `json:"sku"`
	PricePerRoll           float64  `json:"price_per_roll"`
	CurrentStockGrams      *float64 `json:"current_stock_grams"`
	LowStockThresholdGrams *float64 `json:"low_stock_threshold_grams"`
}

// CreateFilament menambahkan filamen baru (dan profil jika belum ada)
// POST /api/v1/filaments
func (h *FilamentHandler) CreateFilament(c *gin.Context) {
	userID := getUserID(c)

	var req CreateFilamentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Input tidak valid: " + err.Error()})
		return
	}

	var profile models.FilamentProfile

	// 1. Jika ProfileID dikirimkan, gunakan profil tersebut
	if req.ProfileID != nil && *req.ProfileID != "" {
		if err := h.db.Where("id = ? AND user_id = ?", *req.ProfileID, userID).First(&profile).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Profile ID tidak valid"})
			return
		}
	} else {
		// 2. Cari profil eksisting berdasarkan kombinasi (user_id, brand, material_type)
		err := h.db.Where("user_id = ? AND LOWER(brand) = LOWER(?) AND LOWER(material_type) = LOWER(?)",
			userID, req.Brand, req.MaterialType).First(&profile).Error

		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Buat profil baru
			profile = models.FilamentProfile{
				ID:                       uuid.New().String(),
				UserID:                   userID,
				Brand:                    req.Brand,
				MaterialType:             req.MaterialType,
				DiameterMM:               1.75,
				NozzleTemp:               req.NozzleTemp,
				BedTemp:                  req.BedTemp,
				RetractionLength:         req.RetractionLength,
				FlowRatio:                req.FlowRatio,
				PressureAdvance:          req.PressureAdvance,
				CoolingFanPercent:        req.CoolingFanPercent,
				MaxVolumetricSpeed:       req.MaxVolumetricSpeed,
				EmptySpoolWeightGrams:    req.EmptySpoolWeightGrams,
				SpoolOuterDiameterMM:     req.SpoolOuterDiameterMM,
				SpoolInnerHoleDiameterMM: req.SpoolInnerHoleDiameterMM,
				SpoolWidthMM:             req.SpoolWidthMM,
				CreatedAt:                time.Now(),
				UpdatedAt:                time.Now(),
			}
			if req.DiameterMM != nil && *req.DiameterMM > 0 {
				profile.DiameterMM = *req.DiameterMM
			}
			if req.SpoolWeightGrams != nil && *req.SpoolWeightGrams > 0 {
				profile.SpoolWeightGrams = *req.SpoolWeightGrams
			} else {
				profile.SpoolWeightGrams = 1000
			}

			if err := h.db.Create(&profile).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Gagal membuat profil filamen: " + err.Error()})
				return
			}
		} else if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Gagal memeriksa profil filamen: " + err.Error()})
			return
		}
	}

	// 3. Buat record roll filamen
	stockGrams := profile.SpoolWeightGrams
	if req.CurrentStockGrams != nil {
		stockGrams = *req.CurrentStockGrams
	}

	threshold := 200.0
	if req.LowStockThresholdGrams != nil {
		threshold = *req.LowStockThresholdGrams
	}

	filament := models.Filament{
		ID:                     uuid.New().String(),
		UserID:                 userID,
		ProfileID:              profile.ID,
		ColorName:              req.ColorName,
		ColorHex:               req.ColorHex,
		SKU:                    req.SKU,
		PricePerRoll:           req.PricePerRoll,
		CurrentStockGrams:      stockGrams,
		LowStockThresholdGrams: threshold,
		CreatedAt:              time.Now(),
		UpdatedAt:              time.Now(),
	}

	if err := h.db.Create(&filament).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Gagal menyimpan data filamen: " + err.Error()})
		return
	}

	filament.Profile = &profile

	c.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"message": "Filamen berhasil ditambahkan",
		"data":    filament.ToResponse(),
	})
}

// UpdateFilamentRequest payload untuk pembaruan filamen
type UpdateFilamentRequest struct {
	Brand                    *string  `json:"brand"`
	MaterialType             *string  `json:"material_type"`
	DiameterMM               *float64 `json:"diameter_mm"`
	NozzleTemp               *int     `json:"nozzle_temp"`
	BedTemp                  *int     `json:"bed_temp"`
	RetractionLength         *float64 `json:"retraction_length"`
	FlowRatio                *float64 `json:"flow_ratio"`
	PressureAdvance          *float64 `json:"pressure_advance"`
	CoolingFanPercent        *int     `json:"cooling_fan_percent"`
	MaxVolumetricSpeed       *float64 `json:"max_volumetric_speed"`
	EmptySpoolWeightGrams    *float64 `json:"empty_spool_weight_grams"`
	SpoolWeightGrams         *float64 `json:"spool_weight_grams"`
	SpoolOuterDiameterMM     *float64 `json:"spool_outer_diameter_mm"`
	SpoolInnerHoleDiameterMM *float64 `json:"spool_inner_hole_diameter_mm"`
	SpoolWidthMM             *float64 `json:"spool_width_mm"`

	ColorName              *string  `json:"color_name"`
	ColorHex               *string  `json:"color_hex"`
	SKU                    *string  `json:"sku"`
	PricePerRoll           *float64 `json:"price_per_roll"`
	CurrentStockGrams      *float64 `json:"current_stock_grams"`
	LowStockThresholdGrams *float64 `json:"low_stock_threshold_grams"`
}

// UpdateFilament memperbarui data roll filamen dan profilnya
// PUT /api/v1/filaments/:id
func (h *FilamentHandler) UpdateFilament(c *gin.Context) {
	userID := getUserID(c)
	id := c.Param("id")

	var filament models.Filament
	if err := h.db.Where("id = ? AND user_id = ?", id, userID).Preload("Profile").First(&filament).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "Filamen tidak ditemukan"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Gagal mengambil data filamen: " + err.Error()})
		return
	}

	var req UpdateFilamentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Input tidak valid: " + err.Error()})
		return
	}

	// Update field roll filamen
	if req.ColorName != nil {
		filament.ColorName = *req.ColorName
	}
	if req.ColorHex != nil {
		filament.ColorHex = *req.ColorHex
	}
	if req.SKU != nil {
		filament.SKU = req.SKU
	}
	if req.PricePerRoll != nil {
		filament.PricePerRoll = *req.PricePerRoll
	}
	if req.CurrentStockGrams != nil {
		filament.CurrentStockGrams = *req.CurrentStockGrams
	}
	if req.LowStockThresholdGrams != nil {
		filament.LowStockThresholdGrams = *req.LowStockThresholdGrams
	}
	filament.UpdatedAt = time.Now()

	if err := h.db.Save(&filament).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Gagal mengupdate roll filamen: " + err.Error()})
		return
	}

	// Update field profile jika ada perubahan spesifik
	if filament.Profile != nil {
		profileUpdated := false
		p := filament.Profile

		if req.Brand != nil {
			p.Brand = *req.Brand
			profileUpdated = true
		}
		if req.MaterialType != nil {
			p.MaterialType = *req.MaterialType
			profileUpdated = true
		}
		if req.DiameterMM != nil {
			p.DiameterMM = *req.DiameterMM
			profileUpdated = true
		}
		if req.NozzleTemp != nil {
			p.NozzleTemp = req.NozzleTemp
			profileUpdated = true
		}
		if req.BedTemp != nil {
			p.BedTemp = req.BedTemp
			profileUpdated = true
		}
		if req.RetractionLength != nil {
			p.RetractionLength = req.RetractionLength
			profileUpdated = true
		}
		if req.FlowRatio != nil {
			p.FlowRatio = req.FlowRatio
			profileUpdated = true
		}
		if req.PressureAdvance != nil {
			p.PressureAdvance = req.PressureAdvance
			profileUpdated = true
		}
		if req.CoolingFanPercent != nil {
			p.CoolingFanPercent = req.CoolingFanPercent
			profileUpdated = true
		}
		if req.MaxVolumetricSpeed != nil {
			p.MaxVolumetricSpeed = req.MaxVolumetricSpeed
			profileUpdated = true
		}
		if req.EmptySpoolWeightGrams != nil {
			p.EmptySpoolWeightGrams = req.EmptySpoolWeightGrams
			profileUpdated = true
		}
		if req.SpoolWeightGrams != nil {
			p.SpoolWeightGrams = *req.SpoolWeightGrams
			profileUpdated = true
		}
		if req.SpoolOuterDiameterMM != nil {
			p.SpoolOuterDiameterMM = req.SpoolOuterDiameterMM
			profileUpdated = true
		}
		if req.SpoolInnerHoleDiameterMM != nil {
			p.SpoolInnerHoleDiameterMM = req.SpoolInnerHoleDiameterMM
			profileUpdated = true
		}
		if req.SpoolWidthMM != nil {
			p.SpoolWidthMM = req.SpoolWidthMM
			profileUpdated = true
		}

		if profileUpdated {
			p.UpdatedAt = time.Now()
			_ = h.db.Save(p).Error
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Data filamen berhasil diperbarui",
		"data":    filament.ToResponse(),
	})
}

// DeleteFilament menghapus roll filamen
// DELETE /api/v1/filaments/:id
func (h *FilamentHandler) DeleteFilament(c *gin.Context) {
	userID := getUserID(c)
	id := c.Param("id")

	result := h.db.Where("id = ? AND user_id = ?", id, userID).Delete(&models.Filament{})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Gagal menghapus filamen: " + result.Error.Error()})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "Filamen tidak ditemukan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Filamen berhasil dihapus",
	})
}

// SyncStockRequest payload untuk mencatat penimbangan fisik spool
type SyncStockRequest struct {
	GrossWeightGrams float64 `json:"gross_weight_grams" binding:"required"` // Berat fisik di timbangan (spool + filamen)
}

// SyncStock mencatat berat timbangan fisik dan menghitung sisa filamen bersih
// POST /api/v1/filaments/:id/sync-stock
func (h *FilamentHandler) SyncStock(c *gin.Context) {
	userID := getUserID(c)
	id := c.Param("id")

	var filament models.Filament
	if err := h.db.Where("id = ? AND user_id = ?", id, userID).Preload("Profile").First(&filament).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "Filamen tidak ditemukan"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Gagal mengambil data filamen: " + err.Error()})
		return
	}

	var req SyncStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Input berat kotor tidak valid: " + err.Error()})
		return
	}

	// Hitung sisa filamen = gross_weight - empty_spool_weight
	emptyWeight := 0.0
	if filament.Profile != nil && filament.Profile.EmptySpoolWeightGrams != nil {
		emptyWeight = *filament.Profile.EmptySpoolWeightGrams
	}

	netWeight := req.GrossWeightGrams - emptyWeight
	if netWeight < 0 {
		netWeight = 0
	}

	now := time.Now()
	filament.CurrentStockGrams = netWeight
	filament.LastWeighedGrams = &req.GrossWeightGrams
	filament.LastWeighedAt = &now
	filament.UpdatedAt = now

	if err := h.db.Save(&filament).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Gagal menyimpan hasil timbangan: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Stok berhasil disinkronkan dari timbangan",
		"data": gin.H{
			"filament_id":              filament.ID,
			"gross_weight_grams":       req.GrossWeightGrams,
			"empty_spool_weight_grams": emptyWeight,
			"calculated_net_grams":     netWeight,
			"last_weighed_at":          filament.LastWeighedAt,
		},
	})
}

// GetProfiles mengambil master daftar profil teknis filamen
// GET /api/v1/filament-profiles
func (h *FilamentHandler) GetProfiles(c *gin.Context) {
	userID := getUserID(c)

	var profiles []models.FilamentProfile
	if err := h.db.Where("user_id = ?", userID).Order("brand asc, material_type asc").Find(&profiles).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Gagal mengambil profil filamen: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   profiles,
		"total":  len(profiles),
	})
}
