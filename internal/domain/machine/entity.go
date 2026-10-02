package machine

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

// MachineMaintenancePart represents wear-and-tear printer spare parts.
type MachineMaintenancePart struct {
	id               uuid.UUID
	machineID        uuid.UUID
	partName         string
	costIDR          float64
	lifespanHours    float64
	stockQuantity    int
	hoursUsedCurrent float64
	lastReplacedAt   *time.Time
	createdAt        time.Time
	updatedAt        time.Time
}

func NewMaintenancePart(
	machineID uuid.UUID,
	partName string,
	costIDR, lifespanHours float64,
	stockQuantity int,
) (*MachineMaintenancePart, error) {
	cleanName := strings.TrimSpace(partName)
	if cleanName == "" {
		return nil, ErrPartNotFound
	}
	if lifespanHours <= 0 {
		lifespanHours = 500
	}
	now := time.Now()
	return &MachineMaintenancePart{
		id:               uuid.New(),
		machineID:        machineID,
		partName:         cleanName,
		costIDR:          costIDR,
		lifespanHours:    lifespanHours,
		stockQuantity:    stockQuantity,
		hoursUsedCurrent: 0,
		createdAt:        now,
		updatedAt:        now,
	}, nil
}

func ReconstructMaintenancePart(
	id, machineID uuid.UUID,
	partName string,
	costIDR, lifespanHours float64,
	stockQuantity int,
	hoursUsedCurrent float64,
	lastReplacedAt *time.Time,
	createdAt, updatedAt time.Time,
) *MachineMaintenancePart {
	return &MachineMaintenancePart{
		id:               id,
		machineID:        machineID,
		partName:         partName,
		costIDR:          costIDR,
		lifespanHours:    lifespanHours,
		stockQuantity:    stockQuantity,
		hoursUsedCurrent: hoursUsedCurrent,
		lastReplacedAt:   lastReplacedAt,
		createdAt:        createdAt,
		updatedAt:        updatedAt,
	}
}

func (p *MachineMaintenancePart) ID() uuid.UUID             { return p.id }
func (p *MachineMaintenancePart) MachineID() uuid.UUID      { return p.machineID }
func (p *MachineMaintenancePart) PartName() string          { return p.partName }
func (p *MachineMaintenancePart) CostIDR() float64          { return p.costIDR }
func (p *MachineMaintenancePart) LifespanHours() float64    { return p.lifespanHours }
func (p *MachineMaintenancePart) StockQuantity() int        { return p.stockQuantity }
func (p *MachineMaintenancePart) HoursUsedCurrent() float64 { return p.hoursUsedCurrent }
func (p *MachineMaintenancePart) LastReplacedAt() *time.Time { return p.lastReplacedAt }
func (p *MachineMaintenancePart) CreatedAt() time.Time      { return p.createdAt }
func (p *MachineMaintenancePart) UpdatedAt() time.Time      { return p.updatedAt }

func (p *MachineMaintenancePart) Update(name string, cost, lifespan float64, stock int) {
	if strings.TrimSpace(name) != "" {
		p.partName = strings.TrimSpace(name)
	}
	if cost >= 0 {
		p.costIDR = cost
	}
	if lifespan > 0 {
		p.lifespanHours = lifespan
	}
	if stock >= 0 {
		p.stockQuantity = stock
	}
	p.updatedAt = time.Now()
}

func (p *MachineMaintenancePart) Replace(cost float64) {
	p.hoursUsedCurrent = 0
	if p.stockQuantity > 0 {
		p.stockQuantity--
	}
	if cost > 0 {
		p.costIDR = cost
	}
	now := time.Now()
	p.lastReplacedAt = &now
	p.updatedAt = now
}

// Machine represents a 3D printer unit in the workshop.
type Machine struct {
	id                       uuid.UUID
	userID                   uuid.UUID
	name                     string
	brand                    *string
	totalPurchaseCost        float64
	salvageValue             *float64
	lifespanHours            int
	purchaseDate             *string
	totalHoursUsed           float64
	avgPowerWatts            int
	maintenanceBufferPerHour *float64
	failureRatePercent       *float64
	buildVolumeX             *float64
	buildVolumeY             *float64
	buildVolumeZ             *float64
	speedMultiplier          *float64
	currentState             string
	electricityCostPerHour   *float64
	maintenanceParts         []*MachineMaintenancePart
	createdAt                time.Time
	updatedAt                time.Time
}

func NewMachine(
	userID uuid.UUID,
	name string,
	brand *string,
	totalPurchaseCost float64,
	lifespanHours int,
	avgPowerWatts int,
) (*Machine, error) {
	if userID == uuid.Nil {
		return nil, ErrUserIDRequired
	}
	cleanName := strings.TrimSpace(name)
	if cleanName == "" {
		return nil, ErrMachineNameRequired
	}
	if lifespanHours <= 0 {
		lifespanHours = 5000
	}
	if avgPowerWatts <= 0 {
		avgPowerWatts = 200
	}
	now := time.Now()
	defaultState := "IDLE"
	defaultSpeed := 1.0
	return &Machine{
		id:                uuid.New(),
		userID:            userID,
		name:              cleanName,
		brand:             brand,
		totalPurchaseCost: totalPurchaseCost,
		lifespanHours:     lifespanHours,
		avgPowerWatts:     avgPowerWatts,
		speedMultiplier:   &defaultSpeed,
		currentState:      defaultState,
		createdAt:         now,
		updatedAt:         now,
	}, nil
}

func ReconstructMachine(
	id, userID uuid.UUID,
	name string,
	brand *string,
	totalPurchaseCost float64,
	salvageValue *float64,
	lifespanHours int,
	purchaseDate *string,
	totalHoursUsed float64,
	avgPowerWatts int,
	maintenanceBufferPerHour, failureRatePercent *float64,
	buildVolumeX, buildVolumeY, buildVolumeZ, speedMultiplier *float64,
	currentState string,
	electricityCostPerHour *float64,
	maintenanceParts []*MachineMaintenancePart,
	createdAt, updatedAt time.Time,
) *Machine {
	return &Machine{
		id:                       id,
		userID:                   userID,
		name:                     name,
		brand:                    brand,
		totalPurchaseCost:        totalPurchaseCost,
		salvageValue:             salvageValue,
		lifespanHours:            lifespanHours,
		purchaseDate:             purchaseDate,
		totalHoursUsed:           totalHoursUsed,
		avgPowerWatts:            avgPowerWatts,
		maintenanceBufferPerHour: maintenanceBufferPerHour,
		failureRatePercent:       failureRatePercent,
		buildVolumeX:             buildVolumeX,
		buildVolumeY:             buildVolumeY,
		buildVolumeZ:             buildVolumeZ,
		speedMultiplier:          speedMultiplier,
		currentState:             currentState,
		electricityCostPerHour:   electricityCostPerHour,
		maintenanceParts:         maintenanceParts,
		createdAt:                createdAt,
		updatedAt:                updatedAt,
	}
}

func (m *Machine) ID() uuid.UUID                                     { return m.id }
func (m *Machine) UserID() uuid.UUID                                 { return m.userID }
func (m *Machine) Name() string                                      { return m.name }
func (m *Machine) Brand() *string                                    { return m.brand }
func (m *Machine) TotalPurchaseCost() float64                        { return m.totalPurchaseCost }
func (m *Machine) SalvageValue() *float64                            { return m.salvageValue }
func (m *Machine) LifespanHours() int                                { return m.lifespanHours }
func (m *Machine) PurchaseDate() *string                             { return m.purchaseDate }
func (m *Machine) TotalHoursUsed() float64                           { return m.totalHoursUsed }
func (m *Machine) AvgPowerWatts() int                                { return m.avgPowerWatts }
func (m *Machine) MaintenanceBufferPerHour() *float64                { return m.maintenanceBufferPerHour }
func (m *Machine) FailureRatePercent() *float64                      { return m.failureRatePercent }
func (m *Machine) BuildVolumeX() *float64                            { return m.buildVolumeX }
func (m *Machine) BuildVolumeY() *float64                            { return m.buildVolumeY }
func (m *Machine) BuildVolumeZ() *float64                            { return m.buildVolumeZ }
func (m *Machine) SpeedMultiplier() *float64                         { return m.speedMultiplier }
func (m *Machine) CurrentState() string                              { return m.currentState }
func (m *Machine) ElectricityCostPerHour() *float64                  { return m.electricityCostPerHour }
func (m *Machine) MaintenanceParts() []*MachineMaintenancePart       { return m.maintenanceParts }
func (m *Machine) CreatedAt() time.Time                              { return m.createdAt }
func (m *Machine) UpdatedAt() time.Time                              { return m.updatedAt }

func (m *Machine) UpdateState(state string) {
	st := strings.ToUpper(strings.TrimSpace(state))
	if st == "IDLE" || st == "PRINTING" || st == "MAINTENANCE" || st == "OFFLINE" {
		m.currentState = st
		m.updatedAt = time.Now()
	}
}