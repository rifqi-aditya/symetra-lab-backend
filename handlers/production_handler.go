package handlers

import (
	"math"
	"net/http"

	"symetra-lab-backend/models"
	"symetra-lab-backend/pkg/production"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ProductionHandler struct {
	DB *gorm.DB
}

func NewProductionHandler(db *gorm.DB) *ProductionHandler {
	return &ProductionHandler{DB: db}
}

// GetQueue mengambil papan antrean cetak terpadu bengkel (Shopee SLA + Manual Order)
// GET /api/v1/production/queue
func (h *ProductionHandler) GetQueue(c *gin.Context) {
	queue, err := production.GetUnifiedProductionQueue(h.DB)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil antrean produksi: " + err.Error(), "status": "error"})
		return
	}

	var totalPrintHours float64
	urgentCount := 0
	for _, q := range queue {
		totalPrintHours += q.TotalPrintHours
		if q.IsUrgent {
			urgentCount++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"queue":               queue,
			"total_jobs":          len(queue),
			"total_hours_waiting": math.Round(totalPrintHours*100) / 100,
			"urgent_jobs_count":   urgentCount,
		},
	})
}

// CompleteJob menandai 1 pekerjaan cetak selesai, otomatis memotong stok filamen & menambah jam pakai mesin
// POST /api/v1/production/complete-job
func (h *ProductionHandler) CompleteJob(c *gin.Context) {
	var req models.CompletePrintJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validasi input gagal: " + err.Error(), "status": "error"})
		return
	}

	resp, err := production.CompletePrintJob(h.DB, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "status": "error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   resp,
	})
}
