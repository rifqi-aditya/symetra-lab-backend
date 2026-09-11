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

// ComponentHandler menangani operasi CRUD untuk komponen tambahan (hardware non-3D print)
type ComponentHandler struct {
	db *gorm.DB
}

// NewComponentHandler membuat instance baru ComponentHandler
func NewComponentHandler(db *gorm.DB) *ComponentHandler {
	return &ComponentHandler{db: db}
}

// GetAllComponents mengambil daftar seluruh komponen milik user
// GET /api/v1/components
func (h *ComponentHandler) GetAllComponents(c *gin.Context) {
	userID := getUserID(c)

	var components []models.Component
	query := h.db.Where("user_id = ?", userID)

	// Filter pencarian berdasarkan nama atau deskripsi
	if search := strings.TrimSpace(c.Query("q")); search != "" {
		query = query.Where("LOWER(name) LIKE LOWER(?) OR LOWER(description) LIKE LOWER(?)", "%"+search+"%", "%"+search+"%")
	} else if name := strings.TrimSpace(c.Query("name")); name != "" {
		query = query.Where("LOWER(name) LIKE LOWER(?)", "%"+name+"%")
	}

	// Sorting
	sortBy := strings.ToLower(c.DefaultQuery("sort", "name"))
	order := strings.ToUpper(c.DefaultQuery("order", "ASC"))
	if order != "ASC" && order != "DESC" {
		order = "ASC"
	}

	switch sortBy {
	case "price":
		query = query.Order("price_per_unit " + order)
	case "created_at":
		query = query.Order("created_at " + order)
	default:
		query = query.Order("name " + order)
	}

	if err := query.Find(&components).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Gagal mengambil data komponen: " + err.Error()})
		return
	}

	responses := make([]models.ComponentResponse, len(components))
	for i, comp := range components {
		responses[i] = comp.ToResponse()
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   responses,
		"total":  len(responses),
	})
}

// GetComponentByID mengambil rincian 1 komponen berdasarkan ID
// GET /api/v1/components/:id
func (h *ComponentHandler) GetComponentByID(c *gin.Context) {
	userID := getUserID(c)
	id := c.Param("id")

	var comp models.Component
	if err := h.db.Where("id = ? AND user_id = ?", id, userID).First(&comp).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "Komponen tidak ditemukan"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Gagal mengambil data komponen: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   comp.ToResponse(),
	})
}

// CreateComponentRequest payload pembuatan komponen baru
type CreateComponentRequest struct {
	Name                 string   `json:"name" binding:"required"`
	PricePerUnit         float64  `json:"price_per_unit" binding:"gte=0"`
	DefaultMarkupPercent *float64 `json:"default_markup_percent"`
	Description          *string  `json:"description"`
}

// CreateComponent menambahkan komponen fisik baru ke inventori
// POST /api/v1/components
func (h *ComponentHandler) CreateComponent(c *gin.Context) {
	userID := getUserID(c)

	var req CreateComponentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Input tidak valid: " + err.Error()})
		return
	}

	markup := 0.0
	if req.DefaultMarkupPercent != nil {
		markup = *req.DefaultMarkupPercent
	}

	comp := models.Component{
		ID:                   uuid.New().String(),
		UserID:               userID,
		Name:                 strings.TrimSpace(req.Name),
		PricePerUnit:         req.PricePerUnit,
		DefaultMarkupPercent: markup,
		Description:          req.Description,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	if err := h.db.Create(&comp).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Gagal menyimpan data komponen: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"message": "Komponen berhasil ditambahkan",
		"data":    comp.ToResponse(),
	})
}

// UpdateComponentRequest payload pembaruan komponen
type UpdateComponentRequest struct {
	Name                 *string  `json:"name"`
	PricePerUnit         *float64 `json:"price_per_unit"`
	DefaultMarkupPercent *float64 `json:"default_markup_percent"`
	Description          *string  `json:"description"`
}

// UpdateComponent memperbarui data komponen
// PUT /api/v1/components/:id
func (h *ComponentHandler) UpdateComponent(c *gin.Context) {
	userID := getUserID(c)
	id := c.Param("id")

	var comp models.Component
	if err := h.db.Where("id = ? AND user_id = ?", id, userID).First(&comp).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "Komponen tidak ditemukan"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Gagal mengambil data komponen: " + err.Error()})
		return
	}

	var req UpdateComponentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Input tidak valid: " + err.Error()})
		return
	}

	if req.Name != nil && strings.TrimSpace(*req.Name) != "" {
		comp.Name = strings.TrimSpace(*req.Name)
	}
	if req.PricePerUnit != nil && *req.PricePerUnit >= 0 {
		comp.PricePerUnit = *req.PricePerUnit
	}
	if req.DefaultMarkupPercent != nil {
		comp.DefaultMarkupPercent = *req.DefaultMarkupPercent
	}
	if req.Description != nil {
		comp.Description = req.Description
	}
	comp.UpdatedAt = time.Now()

	if err := h.db.Save(&comp).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Gagal memperbarui data komponen: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Data komponen berhasil diperbarui",
		"data":    comp.ToResponse(),
	})
}

// DeleteComponent menghapus komponen dari inventori dengan proteksi penggunaan di produk
// DELETE /api/v1/components/:id
func (h *ComponentHandler) DeleteComponent(c *gin.Context) {
	userID := getUserID(c)
	id := c.Param("id")

	var comp models.Component
	if err := h.db.Where("id = ? AND user_id = ?", id, userID).First(&comp).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "Komponen tidak ditemukan"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Gagal mengambil data komponen: " + err.Error()})
		return
	}

	// Safety check: pastikan komponen tidak sedang aktif digunakan dalam resep produk
	force := c.Query("force") == "true"
	if !force && h.db.Migrator().HasTable("product_components") {
		var usageCount int64
		if err := h.db.Table("product_components").Where("component_id = ?", id).Count(&usageCount).Error; err == nil && usageCount > 0 {
			c.JSON(http.StatusConflict, gin.H{
				"status":      "error",
				"message":     fmt.Sprintf("Komponen tidak dapat dihapus karena sedang digunakan dalam %d resep produk. Gunakan ?force=true jika tetap ingin menghapus.", usageCount),
				"usage_count": usageCount,
			})
			return
		}
	}

	result := h.db.Delete(&comp)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Gagal menghapus komponen: " + result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Komponen berhasil dihapus",
	})
}
