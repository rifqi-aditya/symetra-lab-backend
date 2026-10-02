package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"symetra-lab-backend-v2/internal/domain/machine"
)

type machineMaintenancePartGORM struct {
	ID               uuid.UUID  `gorm:"column:id;primaryKey;type:uuid"`
	MachineID        uuid.UUID  `gorm:"column:machine_id;type:uuid"`
	PartName         string     `gorm:"column:part_name"`
	CostIDR          float64    `gorm:"column:cost_idr"`
	LifespanHours    float64    `gorm:"column:lifespan_hours"`
	StockQuantity    int        `gorm:"column:stock_quantity"`
	HoursUsedCurrent float64    `gorm:"column:hours_used_current"`
	LastReplacedAt   *time.Time `gorm:"column:last_replaced_at"`
	CreatedAt        time.Time  `gorm:"column:created_at"`
	UpdatedAt        time.Time  `gorm:"column:updated_at"`
}

func (machineMaintenancePartGORM) TableName() string {
	return "machine_maintenance_parts"
}

type machineGORM struct {
	ID                       uuid.UUID                    `gorm:"column:id;primaryKey;type:uuid"`
	UserID                   uuid.UUID                    `gorm:"column:user_id;type:uuid"`
	Name                     string                       `gorm:"column:name"`
	Brand                    *string                      `gorm:"column:brand"`
	TotalPurchaseCost        float64                      `gorm:"column:total_purchase_cost"`
	SalvageValue             *float64                     `gorm:"column:salvage_value"`
	LifespanHours            int                          `gorm:"column:lifespan_hours"`
	PurchaseDate             *string                      `gorm:"column:purchase_date"`
	TotalHoursUsed           float64                      `gorm:"column:total_hours_used"`
	AvgPowerWatts            int                          `gorm:"column:avg_power_watts"`
	MaintenanceBufferPerHour *float64                     `gorm:"column:maintenance_buffer_per_hour"`
	FailureRatePercent       *float64                     `gorm:"column:failure_rate_percent"`
	BuildVolumeX             *float64                     `gorm:"column:build_volume_x"`
	BuildVolumeY             *float64                     `gorm:"column:build_volume_y"`
	BuildVolumeZ             *float64                     `gorm:"column:build_volume_z"`
	SpeedMultiplier          *float64                     `gorm:"column:speed_multiplier"`
	CurrentState             string                       `gorm:"column:current_state"`
	ElectricityCostPerHour   *float64                     `gorm:"column:electricity_cost_per_hour"`
	MaintenanceParts         []machineMaintenancePartGORM `gorm:"foreignKey:MachineID"`
	CreatedAt                time.Time                    `gorm:"column:created_at"`
	UpdatedAt                time.Time                    `gorm:"column:updated_at"`
}

func (machineGORM) TableName() string {
	return "machines"
}

type MachineRepository struct {
	db *gorm.DB
}

func NewMachineRepository(db *gorm.DB) *MachineRepository {
	return &MachineRepository{db: db}
}

func mapPartGORMToDomain(p *machineMaintenancePartGORM) *machine.MachineMaintenancePart {
	if p == nil {
		return nil
	}
	return machine.ReconstructMaintenancePart(
		p.ID,
		p.MachineID,
		p.PartName,
		p.CostIDR,
		p.LifespanHours,
		p.StockQuantity,
		p.HoursUsedCurrent,
		p.LastReplacedAt,
		p.CreatedAt,
		p.UpdatedAt,
	)
}

func mapMachineGORMToDomain(m *machineGORM) *machine.Machine {
	if m == nil {
		return nil
	}
	parts := make([]*machine.MachineMaintenancePart, len(m.MaintenanceParts))
	for i, p := range m.MaintenanceParts {
		pCopy := p
		parts[i] = mapPartGORMToDomain(&pCopy)
	}

	return machine.ReconstructMachine(
		m.ID,
		m.UserID,
		m.Name,
		m.Brand,
		m.TotalPurchaseCost,
		m.SalvageValue,
		m.LifespanHours,
		m.PurchaseDate,
		m.TotalHoursUsed,
		m.AvgPowerWatts,
		m.MaintenanceBufferPerHour,
		m.FailureRatePercent,
		m.BuildVolumeX,
		m.BuildVolumeY,
		m.BuildVolumeZ,
		m.SpeedMultiplier,
		m.CurrentState,
		m.ElectricityCostPerHour,
		parts,
		m.CreatedAt,
		m.UpdatedAt,
	)
}

func (r *MachineRepository) FindAll(ctx context.Context, userID uuid.UUID) ([]*machine.Machine, error) {
	var list []machineGORM
	err := r.db.WithContext(ctx).Preload("MaintenanceParts").Order("created_at DESC").Find(&list).Error
	if err != nil {
		return nil, err
	}
	res := make([]*machine.Machine, len(list))
	for i, m := range list {
		mCopy := m
		res[i] = mapMachineGORMToDomain(&mCopy)
	}
	return res, nil
}

func (r *MachineRepository) FindByID(ctx context.Context, userID, id uuid.UUID) (*machine.Machine, error) {
	var m machineGORM
	err := r.db.WithContext(ctx).Preload("MaintenanceParts").Where("id = ?", id).First(&m).Error
	if err != nil {
		return nil, err
	}
	return mapMachineGORMToDomain(&m), nil
}

func (r *MachineRepository) Create(ctx context.Context, m *machine.Machine) error {
	g := machineGORM{
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
		CreatedAt:                m.CreatedAt(),
		UpdatedAt:                m.UpdatedAt(),
	}
	return r.db.WithContext(ctx).Create(&g).Error
}

func (r *MachineRepository) Update(ctx context.Context, m *machine.Machine) error {
	return r.db.WithContext(ctx).Model(&machineGORM{}).
		Where("id = ?", m.ID()).
		Updates(map[string]interface{}{
			"name":                        m.Name(),
			"brand":                       m.Brand(),
			"total_purchase_cost":         m.TotalPurchaseCost(),
			"salvage_value":               m.SalvageValue(),
			"lifespan_hours":              m.LifespanHours(),
			"purchase_date":               m.PurchaseDate(),
			"total_hours_used":            m.TotalHoursUsed(),
			"avg_power_watts":             m.AvgPowerWatts(),
			"maintenance_buffer_per_hour": m.MaintenanceBufferPerHour(),
			"failure_rate_percent":        m.FailureRatePercent(),
			"build_volume_x":              m.BuildVolumeX(),
			"build_volume_y":              m.BuildVolumeY(),
			"build_volume_z":              m.BuildVolumeZ(),
			"speed_multiplier":            m.SpeedMultiplier(),
			"current_state":               m.CurrentState(),
			"electricity_cost_per_hour":   m.ElectricityCostPerHour(),
			"updated_at":                  m.UpdatedAt(),
		}).Error
}

func (r *MachineRepository) UpdateState(ctx context.Context, userID, id uuid.UUID, state string) error {
	return r.db.WithContext(ctx).Model(&machineGORM{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"current_state": state,
			"updated_at":    time.Now(),
		}).Error
}

func (r *MachineRepository) Delete(ctx context.Context, userID, id uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("machine_id = ?", id).Delete(&machineMaintenancePartGORM{}).Error; err != nil {
			return err
		}
		return tx.Where("id = ?", id).Delete(&machineGORM{}).Error
	})
}

// Parts
func (r *MachineRepository) FindPartsByMachineID(ctx context.Context, machineID uuid.UUID) ([]*machine.MachineMaintenancePart, error) {
	var list []machineMaintenancePartGORM
	err := r.db.WithContext(ctx).Where("machine_id = ?", machineID).Order("created_at ASC").Find(&list).Error
	if err != nil {
		return nil, err
	}
	res := make([]*machine.MachineMaintenancePart, len(list))
	for i, p := range list {
		pCopy := p
		res[i] = mapPartGORMToDomain(&pCopy)
	}
	return res, nil
}

func (r *MachineRepository) FindPartByID(ctx context.Context, partID uuid.UUID) (*machine.MachineMaintenancePart, error) {
	var p machineMaintenancePartGORM
	err := r.db.WithContext(ctx).Where("id = ?", partID).First(&p).Error
	if err != nil {
		return nil, err
	}
	return mapPartGORMToDomain(&p), nil
}

func (r *MachineRepository) CreatePart(ctx context.Context, part *machine.MachineMaintenancePart) error {
	g := machineMaintenancePartGORM{
		ID:               part.ID(),
		MachineID:        part.MachineID(),
		PartName:         part.PartName(),
		CostIDR:          part.CostIDR(),
		LifespanHours:    part.LifespanHours(),
		StockQuantity:    part.StockQuantity(),
		HoursUsedCurrent: part.HoursUsedCurrent(),
		LastReplacedAt:   part.LastReplacedAt(),
		CreatedAt:        part.CreatedAt(),
		UpdatedAt:        part.UpdatedAt(),
	}
	return r.db.WithContext(ctx).Create(&g).Error
}

func (r *MachineRepository) UpdatePart(ctx context.Context, part *machine.MachineMaintenancePart) error {
	return r.db.WithContext(ctx).Model(&machineMaintenancePartGORM{}).
		Where("id = ?", part.ID()).
		Updates(map[string]interface{}{
			"part_name":      part.PartName(),
			"cost_idr":       part.CostIDR(),
			"lifespan_hours": part.LifespanHours(),
			"stock_quantity": part.StockQuantity(),
			"updated_at":     part.UpdatedAt(),
		}).Error
}

func (r *MachineRepository) ReplacePart(ctx context.Context, partID uuid.UUID, cost float64) (*machine.MachineMaintenancePart, error) {
	part, err := r.FindPartByID(ctx, partID)
	if err != nil {
		return nil, err
	}
	part.Replace(cost)
	now := time.Now()
	updates := map[string]interface{}{
		"hours_used_current": 0,
		"stock_quantity":    part.StockQuantity(),
		"last_replaced_at":  now,
		"updated_at":        now,
	}
	if cost > 0 {
		updates["cost_idr"] = cost
	}
	if err := r.db.WithContext(ctx).Model(&machineMaintenancePartGORM{}).Where("id = ?", partID).Updates(updates).Error; err != nil {
		return nil, err
	}
	return part, nil
}

func (r *MachineRepository) DeletePart(ctx context.Context, partID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("id = ?", partID).Delete(&machineMaintenancePartGORM{}).Error
}