package handlers

import (
	"net/http"
	"strings"
	"time"

	"symetra-lab-backend/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CategoryHandler menangani operasi CRUD untuk kategori produk
type CategoryHandler struct {
	db *gorm.DB
}

func NewCategoryHandler(db *gorm.DB) *CategoryHandler {
	return &CategoryHandler{db: db}
}

type CategoryRequest struct {
	Name string `json:"name" binding:"required"`
}

// GetAll mengambil semua kategori produk milik user
// GET /api/v1/categories
func (h *CategoryHandler) GetAll(c *gin.Context) {
	userID := getUserID(c)

	var categories []models.ProductCategory
	if err := h.db.Where("user_id = ?", userID).Order("name ASC").Find(&categories).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Gagal mengambil kategori: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   categories,
		"total":  len(categories),
	})
}

// Create membuat kategori produk baru
// POST /api/v1/categories
func (h *CategoryHandler) Create(c *gin.Context) {
	var req CategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Nama kategori wajib diisi"})
		return
	}

	userID := getUserID(c)
	cat := models.ProductCategory{
		ID:        uuid.New().String(),
		UserID:    userID,
		Name:      strings.TrimSpace(req.Name),
		CreatedAt: time.Now(),
	}

	if err := h.db.Create(&cat).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Gagal menyimpan kategori: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status": "success",
		"data":   cat,
	})
}

// Update mengubah nama kategori produk
// PUT /api/v1/categories/:id
func (h *CategoryHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req CategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Nama kategori wajib diisi"})
		return
	}

	userID := getUserID(c)
	var cat models.ProductCategory
	if err := h.db.Where("id = ? AND user_id = ?", id, userID).First(&cat).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "Kategori tidak ditemukan"})
		return
	}

	cat.Name = strings.TrimSpace(req.Name)
	if err := h.db.Save(&cat).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Gagal memperbarui kategori"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   cat,
	})
}

// Delete menghapus kategori produk
// DELETE /api/v1/categories/:id
func (h *CategoryHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	userID := getUserID(c)

	if err := h.db.Where("id = ? AND user_id = ?", id, userID).Delete(&models.ProductCategory{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Gagal menghapus kategori: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Kategori berhasil dihapus",
	})
}
