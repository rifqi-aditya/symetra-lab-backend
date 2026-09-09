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

// MachineHandler menangani operasi CRUD untuk 3D printer dan suku cadang perawatannya
type MachineHandler struct {
	db *gorm.DB
}

// NewMachineHandler membuat instance baru MachineHandler
func NewMachineHandler(db *gorm.DB) *MachineHandler {
	return &MachineHandler{db: db}
}

// GetAllMachines mengambil semua mesin 3D printer milik user beserta suku cadang perawatannya
// GET /api/v1/machines
func (h *MachineHandler) GetAllMachines(c *gin.Context) {
	userID := getUserID(c)

	var machines []models.Machine
	query := h.db.Where("user_id = ?", userID).Preload("MaintenanceParts").Order("created_at asc")

	if status := c.Query("state"); status != "" {
		query = query.Where("UPPER(current_state) = UPPER(?)", status)
	}

	if err := query.Find(&machines).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Gagal mengambil data mesin: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   machines,
		"total":  len(machines),
	})
}

// GetMachineByID mengambil detail 1 mesin printer
// GET /api/v1/machines/:id
func (h *MachineHandler) GetMachineByID(c *gin.Context) {
	userID := getUserID(c)
	id := c.Param("id")

	var machine models.Machine
	if err := h.db.Where("id = ? AND user_id = ?", id, userID).Preload("MaintenanceParts").First(&machine).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "Mesin tidak ditemukan"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Gagal mengambil data mesin: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   machine,
	})
}

// CreateMachineRequest payload pembuatan mesin baru
type CreateMachineRequest struct {
	Name                     string   `json:"name" binding:"required"`
	Brand                    *string  `json:"brand"`
	TotalPurchaseCost        float64  `json:"total_purchase_cost" binding:"required"`
	SalvageValue             *float64 `json:"salvage_value"`
	LifespanHours            *int     `json:"lifespan_hours"`
	PurchaseDate             *string  `json:"purchase_date"`
	TotalHoursUsed           *float64 `json:"total_hours_used"`
	AvgPowerWatts            int      `json:"avg_power_watts" binding:"required"`
	MaintenanceBufferPerHour *float64 `json:"maintenance_buffer_per_hour"`
	FailureRatePercent       *float64 `json:"failure_rate_percent"`
	BuildVolumeX             *float64 `json:"build_volume_x"`
	BuildVolumeY             *float64 `json:"build_volume_y"`
	BuildVolumeZ             *float64 `json:"build_volume_z"`
	SpeedMultiplier          *float64 `json:"speed_multiplier"`
	CurrentState             *string  `json:"current_state"`
	ElectricityCostPerHour   *float64 `json:"electricity_cost_per_hour"`
}

// CreateMachine menambahkan printer baru ke database
// POST /api/v1/machines
func (h *MachineHandler) CreateMachine(c *gin.Context) {
	userID := getUserID(c)

	var req CreateMachineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Input tidak valid: " + err.Error()})
		return
	}

	lifespan := 5000
	if req.LifespanHours != nil && *req.LifespanHours > 0 {
		lifespan = *req.LifespanHours
	}

	state := "IDLE"
	if req.CurrentState != nil && *req.CurrentState != "" {
		state = strings.ToUpper(*req.CurrentState)
	}

	speed := 1.0
	if req.SpeedMultiplier != nil && *req.SpeedMultiplier > 0 {
		speed = *req.SpeedMultiplier
	}

	totalHours := 0.0
	if req.TotalHoursUsed != nil {
		totalHours = *req.TotalHoursUsed
	}

	machine := models.Machine{
		ID:                       uuid.New().String(),
		UserID:                   userID,
		Name:                     req.Name,
		Brand:                    req.Brand,
		TotalPurchaseCost:        req.TotalPurchaseCost,
		SalvageValue:             req.SalvageValue,
		LifespanHours:            lifespan,
		PurchaseDate:             req.PurchaseDate,
		TotalHoursUsed:           totalHours,
		AvgPowerWatts:            req.AvgPowerWatts,
		MaintenanceBufferPerHour: req.MaintenanceBufferPerHour,
		FailureRatePercent:       req.FailureRatePercent,
		BuildVolumeX:             req.BuildVolumeX,
		BuildVolumeY:             req.BuildVolumeY,
		BuildVolumeZ:             req.BuildVolumeZ,
		SpeedMultiplier:          &speed,
		CurrentState:             state,
		ElectricityCostPerHour:   req.ElectricityCostPerHour,
		CreatedAt:                time.Now(),
		UpdatedAt:                time.Now(),
	}

	if err := h.db.Create(&machine).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Gagal menyimpan data mesin: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"message": "Mesin berhasil ditambahkan",
		"data":    machine,
	})
}

// UpdateMachineRequest payload pembaruan spesifikasi mesin
type UpdateMachineRequest struct {
	Name                     *string  `json:"name"`
	Brand                    *string  `json:"brand"`
	TotalPurchaseCost        *float64 `json:"total_purchase_cost"`
	SalvageValue             *float64 `json:"salvage_value"`
	LifespanHours            *int     `json:"lifespan_hours"`
	PurchaseDate             *string  `json:"purchase_date"`
	TotalHoursUsed           *float64 `json:"total_hours_used"`
	AvgPowerWatts            *int     `json:"avg_power_watts"`
	MaintenanceBufferPerHour *float64 `json:"maintenance_buffer_per_hour"`
	FailureRatePercent       *float64 `json:"failure_rate_percent"`
	BuildVolumeX             *float64 `json:"build_volume_x"`
	BuildVolumeY             *float64 `json:"build_volume_y"`
	BuildVolumeZ             *float64 `json:"build_volume_z"`
	SpeedMultiplier          *float64 `json:"speed_multiplier"`
	CurrentState             *string  `json:"current_state"`
	ElectricityCostPerHour   *float64 `json:"electricity_cost_per_hour"`
}

// UpdateMachine memperbarui spesifikasi mesin
// PUT /api/v1/machines/:id
func (h *MachineHandler) UpdateMachine(c *gin.Context) {
	userID := getUserID(c)
	id := c.Param("id")

	var machine models.Machine
	if err := h.db.Where("id = ? AND user_id = ?", id, userID).First(&machine).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "Mesin tidak ditemukan"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Gagal mengambil data mesin: " + err.Error()})
		return
	}

	var req UpdateMachineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Input tidak valid: " + err.Error()})
		return
	}

	if req.Name != nil {
		machine.Name = *req.Name
	}
	if req.Brand != nil {
		machine.Brand = req.Brand
	}
	if req.TotalPurchaseCost != nil {
		machine.TotalPurchaseCost = *req.TotalPurchaseCost
	}
	if req.SalvageValue != nil {
		machine.SalvageValue = req.SalvageValue
	}
	if req.LifespanHours != nil {
		machine.LifespanHours = *req.LifespanHours
	}
	if req.PurchaseDate != nil {
		machine.PurchaseDate = req.PurchaseDate
	}
	if req.TotalHoursUsed != nil {
		machine.TotalHoursUsed = *req.TotalHoursUsed
	}
	if req.AvgPowerWatts != nil {
		machine.AvgPowerWatts = *req.AvgPowerWatts
	}
	if req.MaintenanceBufferPerHour != nil {
		machine.MaintenanceBufferPerHour = req.MaintenanceBufferPerHour
	}
	if req.FailureRatePercent != nil {
		machine.FailureRatePercent = req.FailureRatePercent
	}
	if req.BuildVolumeX != nil {
		machine.BuildVolumeX = req.BuildVolumeX
	}
	if req.BuildVolumeY != nil {
		machine.BuildVolumeY = req.BuildVolumeY
	}
	if req.BuildVolumeZ != nil {
		machine.BuildVolumeZ = req.BuildVolumeZ
	}
	if req.SpeedMultiplier != nil {
		machine.SpeedMultiplier = req.SpeedMultiplier
	}
	if req.CurrentState != nil {
		machine.CurrentState = strings.ToUpper(*req.CurrentState)
	}
	if req.ElectricityCostPerHour != nil {
		machine.ElectricityCostPerHour = req.ElectricityCostPerHour
	}
	machine.UpdatedAt = time.Now()

	if err := h.db.Save(&machine).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Gagal memperbarui data mesin: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Data mesin berhasil diperbarui",
		"data":    machine,
	})
}

// UpdateStateRequest payload perubahan status mesin
type UpdateStateRequest struct {
	State string `json:"state" binding:"required"` // IDLE, PRINTING, MAINTENANCE, OFFLINE
}

// UpdateMachineState mengubah status mesin secara instan (misal saat mulai print)
// PATCH /api/v1/machines/:id/state
func (h *MachineHandler) UpdateMachineState(c *gin.Context) {
	userID := getUserID(c)
	id := c.Param("id")

	var req UpdateStateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Status tidak valid: " + err.Error()})
		return
	}

	targetState := strings.ToUpper(strings.TrimSpace(req.State))
	validStates := map[string]bool{
		"IDLE":        true,
		"PRINTING":    true,
		"MAINTENANCE": true,
		"OFFLINE":     true,
	}
	if !validStates[targetState] {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "State tidak valid. Pilihan: IDLE, PRINTING, MAINTENANCE, OFFLINE"})
		return
	}

	var machine models.Machine
	if err := h.db.Where("id = ? AND user_id = ?", id, userID).First(&machine).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "Mesin tidak ditemukan"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Gagal mengambil data mesin: " + err.Error()})
		return
	}

	machine.CurrentState = targetState
	machine.UpdatedAt = time.Now()

	if err := h.db.Save(&machine).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Gagal mengubah status mesin: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Status mesin berhasil diubah",
		"data": gin.H{
			"machine_id": machine.ID,
			"state":      machine.CurrentState,
			"updated_at": machine.UpdatedAt,
		},
	})
}

// DeleteMachine menghapus printer beserta relasi suku cadang
// DELETE /api/v1/machines/:id
func (h *MachineHandler) DeleteMachine(c *gin.Context) {
	userID := getUserID(c)
	id := c.Param("id")

	// Pastikan mesin milik user tersebut
	var machine models.Machine
	if err := h.db.Where("id = ? AND user_id = ?", id, userID).First(&machine).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "Mesin tidak ditemukan"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Gagal mengambil data mesin: " + err.Error()})
		return
	}

	// Hapus suku cadang terkait terlebih dahulu
	_ = h.db.Where("machine_id = ?", id).Delete(&models.MachineMaintenancePart{}).Error

	if err := h.db.Delete(&machine).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Gagal menghapus mesin: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Mesin dan suku cadang terkait berhasil dihapus",
	})
}

// --- SUB-RESOURCE: Suku Cadang Perawatan (Machine Maintenance Parts) ---

// GetPartsByMachine mengambil suku cadang untuk printer tertentu
// GET /api/v1/machines/:id/parts
func (h *MachineHandler) GetPartsByMachine(c *gin.Context) {
	userID := getUserID(c)
	machineID := c.Param("id")

	// Verifikasi kepemilikan mesin
	var machine models.Machine
	if err := h.db.Where("id = ? AND user_id = ?", machineID, userID).First(&machine).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "Mesin tidak ditemukan"})
		return
	}

	var parts []models.MachineMaintenancePart
	if err := h.db.Where("machine_id = ?", machineID).Order("part_name asc").Find(&parts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Gagal mengambil suku cadang: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   parts,
		"total":  len(parts),
	})
}

// CreatePartRequest payload pembuatan suku cadang
type CreatePartRequest struct {
	PartName         string   `json:"part_name" binding:"required"`
	CostIDR          float64  `json:"cost_idr"`
	LifespanHours    *float64 `json:"lifespan_hours"`
	StockQuantity    *int     `json:"stock_quantity"`
	HoursUsedCurrent *float64 `json:"hours_used_current"`
}

// CreatePart menambahkan suku cadang ke mesin
// POST /api/v1/machines/:id/parts
func (h *MachineHandler) CreatePart(c *gin.Context) {
	userID := getUserID(c)
	machineID := c.Param("id")

	var machine models.Machine
	if err := h.db.Where("id = ? AND user_id = ?", machineID, userID).First(&machine).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "Mesin tidak ditemukan"})
		return
	}

	var req CreatePartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Input suku cadang tidak valid: " + err.Error()})
		return
	}

	lifespan := 500.0
	if req.LifespanHours != nil && *req.LifespanHours > 0 {
		lifespan = *req.LifespanHours
	}

	stock := 0
	if req.StockQuantity != nil {
		stock = *req.StockQuantity
	}

	hoursUsed := 0.0
	if req.HoursUsedCurrent != nil {
		hoursUsed = *req.HoursUsedCurrent
	}

	part := models.MachineMaintenancePart{
		ID:               uuid.New().String(),
		MachineID:        machineID,
		PartName:         req.PartName,
		CostIDR:          req.CostIDR,
		LifespanHours:    lifespan,
		StockQuantity:    stock,
		HoursUsedCurrent: hoursUsed,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if err := h.db.Create(&part).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Gagal menyimpan suku cadang: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"message": "Suku cadang berhasil ditambahkan",
		"data":    part,
	})
}

// UpdatePartRequest payload pembaruan suku cadang
type UpdatePartRequest struct {
	PartName         *string  `json:"part_name"`
	CostIDR          *float64 `json:"cost_idr"`
	LifespanHours    *float64 `json:"lifespan_hours"`
	StockQuantity    *int     `json:"stock_quantity"`
	HoursUsedCurrent *float64 `json:"hours_used_current"`
}

// UpdatePart memperbarui data suku cadang
// PUT /api/v1/machines/parts/:part_id
func (h *MachineHandler) UpdatePart(c *gin.Context) {
	partID := c.Param("part_id")

	var part models.MachineMaintenancePart
	if err := h.db.Where("id = ?", partID).First(&part).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "Suku cadang tidak ditemukan"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Gagal mengambil data suku cadang: " + err.Error()})
		return
	}

	var req UpdatePartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Input tidak valid: " + err.Error()})
		return
	}

	if req.PartName != nil {
		part.PartName = *req.PartName
	}
	if req.CostIDR != nil {
		part.CostIDR = *req.CostIDR
	}
	if req.LifespanHours != nil {
		part.LifespanHours = *req.LifespanHours
	}
	if req.StockQuantity != nil {
		part.StockQuantity = *req.StockQuantity
	}
	if req.HoursUsedCurrent != nil {
		part.HoursUsedCurrent = *req.HoursUsedCurrent
	}
	part.UpdatedAt = time.Now()

	if err := h.db.Save(&part).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Gagal mengupdate suku cadang: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Suku cadang berhasil diperbarui",
		"data":    part,
	})
}

// ReplacePart mencatat penggantian suku cadang: mereset hours_used_current ke 0, kurangi stock_quantity, set last_replaced_at
// POST /api/v1/machines/parts/:part_id/replace
func (h *MachineHandler) ReplacePart(c *gin.Context) {
	partID := c.Param("part_id")

	var part models.MachineMaintenancePart
	if err := h.db.Where("id = ?", partID).First(&part).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "Suku cadang tidak ditemukan"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Gagal mengambil data suku cadang: " + err.Error()})
		return
	}

	now := time.Now()
	part.HoursUsedCurrent = 0
	if part.StockQuantity > 0 {
		part.StockQuantity--
	}
	part.LastReplacedAt = &now
	part.UpdatedAt = now

	if err := h.db.Save(&part).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Gagal mencatat penggantian suku cadang: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Penggantian suku cadang berhasil dicatat",
		"data": gin.H{
			"part_id":            part.ID,
			"part_name":          part.PartName,
			"hours_used_current": part.HoursUsedCurrent,
			"remaining_stock":    part.StockQuantity,
			"last_replaced_at":   part.LastReplacedAt,
		},
	})
}

// DeletePart menghapus suku cadang
// DELETE /api/v1/machines/parts/:part_id
func (h *MachineHandler) DeletePart(c *gin.Context) {
	partID := c.Param("part_id")

	result := h.db.Where("id = ?", partID).Delete(&models.MachineMaintenancePart{})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Gagal menghapus suku cadang: " + result.Error.Error()})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "Suku cadang tidak ditemukan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Suku cadang berhasil dihapus",
	})
}
