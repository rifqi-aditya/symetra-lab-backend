package dto

import (
	"time"

	"github.com/google/uuid"

	"symetra-lab-backend-v2/internal/domain/machine"
)

type CreateMachineRequest struct {
	Name              string  `json:"name"`
	Brand             *string `json:"brand"`
	TotalPurchaseCost float64 `json:"total_purchase_cost"`
	LifespanHours     int     `json:"lifespan_hours"`
	AvgPowerWatts     int     `json:"avg_power_watts"`
}

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

type UpdateMachineStateRequest struct {
	State string `json:"state"`
}

type CreatePartRequest struct {
	PartName      string  `json:"part_name"`
	CostIDR       float64 `json:"cost_idr"`
	LifespanHours float64 `json:"lifespan_hours"`
	StockQuantity int     `json:"stock_quantity"`
}

type UpdatePartRequest struct {
	PartName      *string  `json:"part_name"`
	CostIDR       *float64 `json:"cost_idr"`
	LifespanHours *float64 `json:"lifespan_hours"`
	StockQuantity *int     `json:"stock_quantity"`
}

type ReplacePartRequest struct {
	CostIDR *float64 `json:"cost_idr"`
}

type MachineMaintenancePartResponse struct {
	ID               uuid.UUID  `json:"id"`
	MachineID        uuid.UUID  `json:"machine_id"`
	PartName         string     `json:"part_name"`
	CostIDR          float64    `json:"cost_idr"`
	LifespanHours    float64    `json:"lifespan_hours"`
	StockQuantity    int        `json:"stock_quantity"`
	HoursUsedCurrent float64    `json:"hours_used_current"`
	LastReplacedAt   *time.Time `json:"last_replaced_at"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

func ToPartResponse(p *machine.MachineMaintenancePart) MachineMaintenancePartResponse {
	return MachineMaintenancePartResponse{
		ID:               p.ID(),
		MachineID:        p.MachineID(),
		PartName:         p.PartName(),
		CostIDR:          p.CostIDR(),
		LifespanHours:    p.LifespanHours(),
		StockQuantity:    p.StockQuantity(),
		HoursUsedCurrent: p.HoursUsedCurrent(),
		LastReplacedAt:   p.LastReplacedAt(),
		CreatedAt:        p.CreatedAt(),
		UpdatedAt:        p.UpdatedAt(),
	}
}

type MachineResponse struct {
	ID                       uuid.UUID                        `json:"id"`
	UserID                   uuid.UUID                        `json:"user_id"`
	Name                     string                           `json:"name"`
	Brand                    *string                          `json:"brand"`
	TotalPurchaseCost        float64                          `json:"total_purchase_cost"`
	SalvageValue             *float64                         `json:"salvage_value"`
	LifespanHours            int                              `json:"lifespan_hours"`
	PurchaseDate             *string                          `json:"purchase_date"`
	TotalHoursUsed           float64                          `json:"total_hours_used"`
	AvgPowerWatts            int                              `json:"avg_power_watts"`
	MaintenanceBufferPerHour *float64                         `json:"maintenance_buffer_per_hour"`
	FailureRatePercent       *float64                         `json:"failure_rate_percent"`
	BuildVolumeX             *float64                         `json:"build_volume_x"`
	BuildVolumeY             *float64                         `json:"build_volume_y"`
	BuildVolumeZ             *float64                         `json:"build_volume_z"`
	SpeedMultiplier          *float64                         `json:"speed_multiplier"`
	CurrentState             string                           `json:"current_state"`
	ElectricityCostPerHour   *float64                         `json:"electricity_cost_per_hour"`
	MaintenanceParts         []MachineMaintenancePartResponse `json:"maintenance_parts,omitempty"`
	CreatedAt                time.Time                        `json:"created_at"`
	UpdatedAt                time.Time                        `json:"updated_at"`
}

func ToMachineResponse(m *machine.Machine) MachineResponse {
	parts := make([]MachineMaintenancePartResponse, len(m.MaintenanceParts()))
	for i, p := range m.MaintenanceParts() {
		parts[i] = ToPartResponse(p)
	}

	return MachineResponse{
		ID:                       m.ID(),
		UserID:                   m.UserID(),
		Name:                     m.Name(),
		Brand:                    m.Brand(),
		TotalPurchaseCost:        m.TotalPurchaseCost(),
		SalvageValue:             m.SalvageValue(),
		LifespanHours:            m.LifespanHours(),
		PurchaseDate:             m.PurchaseDate(),
		TotalHoursUsed:           m.TotalHoursUsed(),
		AvgPowerWatts:            m.AvgPowerWatts(),
		MaintenanceBufferPerHour: m.MaintenanceBufferPerHour(),
		FailureRatePercent:       m.FailureRatePercent(),
		BuildVolumeX:             m.BuildVolumeX(),
		BuildVolumeY:             m.BuildVolumeY(),
		BuildVolumeZ:             m.BuildVolumeZ(),
		SpeedMultiplier:          m.SpeedMultiplier(),
		CurrentState:             m.CurrentState(),
		ElectricityCostPerHour:   m.ElectricityCostPerHour(),
		MaintenanceParts:         parts,
		CreatedAt:                m.CreatedAt(),
		UpdatedAt:                m.UpdatedAt(),
	}
}