package models

import (
	"time"
)

// Machine merepresentasikan unit 3D printer di bengkel Symetra Lab
type Machine struct {
	ID                       string                   `gorm:"type:uuid;primaryKey" json:"id"`
	UserID                   string                   `gorm:"type:uuid;not null;index" json:"user_id"`
	Name                     string                   `gorm:"size:255;not null" json:"name"`
	Brand                    *string                  `gorm:"size:255" json:"brand"`
	TotalPurchaseCost        float64                  `gorm:"type:numeric;not null" json:"total_purchase_cost"`
	SalvageValue             *float64                 `gorm:"type:numeric;default:0" json:"salvage_value"`
	LifespanHours            int                      `gorm:"default:5000;not null" json:"lifespan_hours"`
	PurchaseDate             *string                  `gorm:"type:date" json:"purchase_date"`
	TotalHoursUsed           float64                  `gorm:"type:numeric;default:0" json:"total_hours_used"`
	AvgPowerWatts            int                      `gorm:"not null" json:"avg_power_watts"`
	MaintenanceBufferPerHour *float64                 `gorm:"type:numeric;default:0" json:"maintenance_buffer_per_hour"`
	FailureRatePercent       *float64                 `gorm:"type:numeric;default:0" json:"failure_rate_percent"`
	BuildVolumeX             *float64                 `gorm:"type:numeric" json:"build_volume_x"`
	BuildVolumeY             *float64                 `gorm:"type:numeric" json:"build_volume_y"`
	BuildVolumeZ             *float64                 `gorm:"type:numeric" json:"build_volume_z"`
	SpeedMultiplier          *float64                 `gorm:"type:numeric;default:1.0" json:"speed_multiplier"`
	CurrentState             string                   `gorm:"size:32;default:'IDLE'" json:"current_state"` // IDLE, PRINTING, MAINTENANCE, OFFLINE
	ElectricityCostPerHour   *float64                 `gorm:"type:numeric;default:0" json:"electricity_cost_per_hour"`
	CreatedAt                time.Time                `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt                time.Time                `gorm:"autoUpdateTime" json:"updated_at"`

	// Relasi ke suku cadang perawatan
	MaintenanceParts         []MachineMaintenancePart `gorm:"foreignKey:MachineID;references:ID;constraint:OnDelete:CASCADE" json:"maintenance_parts,omitempty"`
}

func (Machine) TableName() string {
	return "machines"
}

// MachineMaintenancePart merepresentasikan komponen suku cadang / wear-and-tear printer (nozzle, PEI sheet, belt, dll)
type MachineMaintenancePart struct {
	ID               string     `gorm:"type:uuid;primaryKey" json:"id"`
	MachineID        string     `gorm:"type:uuid;not null;index" json:"machine_id"`
	PartName         string     `gorm:"type:text;not null" json:"part_name"`
	CostIDR          float64    `gorm:"type:numeric;default:0;not null" json:"cost_idr"`
	LifespanHours    float64    `gorm:"type:numeric;default:500;not null" json:"lifespan_hours"`
	StockQuantity    int        `gorm:"default:0;not null" json:"stock_quantity"`
	HoursUsedCurrent float64    `gorm:"type:numeric;default:0;not null" json:"hours_used_current"`
	LastReplacedAt   *time.Time `json:"last_replaced_at"`
	CreatedAt        time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

func (MachineMaintenancePart) TableName() string {
	return "machine_maintenance_parts"
}
