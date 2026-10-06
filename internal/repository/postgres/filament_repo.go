package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"symetra-lab-backend-v2/internal/domain/filament"
)

type filamentProfileGORM struct {
	ID                       uuid.UUID `gorm:"column:id;primaryKey;type:uuid"`
	UserID                   uuid.UUID `gorm:"column:user_id;type:uuid"`
	Brand                    string    `gorm:"column:brand"`
	MaterialType             string    `gorm:"column:material_type"`
	DiameterMM               float64   `gorm:"column:diameter_mm"`
	EmptySpoolWeightGrams    *float64  `gorm:"column:empty_spool_weight_grams"`
	SpoolWeightGrams         float64   `gorm:"column:spool_weight_grams"`
	SpoolOuterDiameterMM     *float64  `gorm:"column:spool_outer_diameter_mm"`
	SpoolInnerHoleDiameterMM *float64  `gorm:"column:spool_inner_hole_diameter_mm"`
	CreatedAt                time.Time `gorm:"column:created_at"`
	UpdatedAt                time.Time `gorm:"column:updated_at"`
}

func (filamentProfileGORM) TableName() string {
	return "filament_profiles"
}

type filamentGORM struct {
	ID                     uuid.UUID            `gorm:"column:id;primaryKey;type:uuid"`
	UserID                 uuid.UUID            `gorm:"column:user_id;type:uuid"`
	ProfileID              *uuid.UUID           `gorm:"column:profile_id;type:uuid"`
	ColorName              string               `gorm:"column:color_name"`
	ColorHex               string               `gorm:"column:color_hex"`
	SKU                    *string              `gorm:"column:sku"`
	CurrentStockGrams      float64              `gorm:"column:current_stock_grams"`
	LowStockThresholdGrams float64              `gorm:"column:low_stock_threshold_grams"`
	PricePerRoll           float64              `gorm:"column:price_per_roll"`
	LastWeighedGrams       *float64             `gorm:"column:last_weighed_grams"`
	LastWeighedAt          *time.Time           `gorm:"column:last_weighed_at"`
	Profile                *filamentProfileGORM `gorm:"foreignKey:ProfileID"`
	CreatedAt              time.Time            `gorm:"column:created_at"`
	UpdatedAt              time.Time            `gorm:"column:updated_at"`
}

func (filamentGORM) TableName() string {
	return "filaments"
}

type filamentMaterialRateGORM struct {
	ID           uuid.UUID `gorm:"column:id;primaryKey;type:uuid"`
	UserID       uuid.UUID `gorm:"column:user_id;type:uuid"`
	MaterialType string    `gorm:"column:material_type"`
	PricePerGram float64   `gorm:"column:price_per_gram"`
	IsDefault    bool      `gorm:"column:is_default"`
	Description  *string   `gorm:"column:description"`
	CreatedAt    time.Time `gorm:"column:created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at"`
}

func (filamentMaterialRateGORM) TableName() string {
	return "filament_material_rates"
}

type FilamentRepository struct {
	db *gorm.DB
}

func NewFilamentRepository(db *gorm.DB) *FilamentRepository {
	return &FilamentRepository{db: db}
}

func mapProfileGORMToDomain(p *filamentProfileGORM) *filament.FilamentProfile {
	if p == nil {
		return nil
	}
	return filament.ReconstructProfile(
		p.ID,
		p.UserID,
		p.Brand,
		p.MaterialType,
		p.DiameterMM,
		p.EmptySpoolWeightGrams,
		p.SpoolWeightGrams,
		p.SpoolOuterDiameterMM,
		p.SpoolInnerHoleDiameterMM,
		p.CreatedAt,
		p.UpdatedAt,
	)
}

func mapFilamentGORMToDomain(f *filamentGORM) *filament.Filament {
	if f == nil {
		return nil
	}
	return filament.ReconstructFilament(
		f.ID,
		f.UserID,
		f.ProfileID,
		f.ColorName,
		f.ColorHex,
		f.SKU,
		f.PricePerRoll,
		f.CurrentStockGrams,
		f.LowStockThresholdGrams,
		f.LastWeighedGrams,
		f.LastWeighedAt,
		mapProfileGORMToDomain(f.Profile),
		f.CreatedAt,
		f.UpdatedAt,
	)
}

func (r *FilamentRepository) FindAll(ctx context.Context, userID uuid.UUID) ([]*filament.Filament, error) {
	var list []filamentGORM
	err := r.db.WithContext(ctx).Preload("Profile").Order("created_at DESC").Find(&list).Error
	if err != nil {
		return nil, err
	}

	result := make([]*filament.Filament, len(list))
	for i, f := range list {
		fCopy := f
		result[i] = mapFilamentGORMToDomain(&fCopy)
	}
	return result, nil
}

func (r *FilamentRepository) FindByID(ctx context.Context, userID, id uuid.UUID) (*filament.Filament, error) {
	var f filamentGORM
	err := r.db.WithContext(ctx).Preload("Profile").Where("id = ?", id).First(&f).Error
	if err != nil {
		return nil, err
	}
	return mapFilamentGORMToDomain(&f), nil
}

func (r *FilamentRepository) Create(ctx context.Context, f *filament.Filament) error {
	m := filamentGORM{
		ID:                     f.ID(),
		UserID:                 f.UserID(),
		ProfileID:              f.ProfileID(),
		ColorName:              f.ColorName(),
		ColorHex:               f.ColorHex(),
		SKU:                    f.SKU(),
		PricePerRoll:           f.PricePerRoll(),
		CurrentStockGrams:      f.CurrentStockGrams(),
		LowStockThresholdGrams: f.LowStockThresholdGrams(),
		LastWeighedGrams:       f.LastWeighedGrams(),
		LastWeighedAt:          f.LastWeighedAt(),
		CreatedAt:              f.CreatedAt(),
		UpdatedAt:              f.UpdatedAt(),
	}
	return r.db.WithContext(ctx).Create(&m).Error
}

func (r *FilamentRepository) Update(ctx context.Context, f *filament.Filament) error {
	return r.db.WithContext(ctx).Model(&filamentGORM{}).
		Where("id = ?", f.ID()).
		Updates(map[string]interface{}{
			"profile_id":                f.ProfileID(),
			"color_name":                f.ColorName(),
			"color_hex":                 f.ColorHex(),
			"sku":                       f.SKU(),
			"price_per_roll":            f.PricePerRoll(),
			"current_stock_grams":       f.CurrentStockGrams(),
			"low_stock_threshold_grams": f.LowStockThresholdGrams(),
			"last_weighed_grams":        f.LastWeighedGrams(),
			"last_weighed_at":           f.LastWeighedAt(),
			"updated_at":                f.UpdatedAt(),
		}).Error
}

func (r *FilamentRepository) Delete(ctx context.Context, userID, id uuid.UUID) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&filamentGORM{}).Error
}

func (r *FilamentRepository) SyncStock(ctx context.Context, userID, id uuid.UUID, stockGrams float64, isWeighed bool) (*filament.Filament, error) {
	f, err := r.FindByID(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	f.SyncStock(stockGrams, isWeighed)
	if err := r.Update(ctx, f); err != nil {
		return nil, err
	}
	return f, nil
}

func (r *FilamentRepository) FindAllProfiles(ctx context.Context, userID uuid.UUID) ([]*filament.FilamentProfile, error) {
	var list []filamentProfileGORM
	err := r.db.WithContext(ctx).Order("created_at DESC").Find(&list).Error
	if err != nil {
		return nil, err
	}
	result := make([]*filament.FilamentProfile, len(list))
	for i, p := range list {
		pCopy := p
		result[i] = mapProfileGORMToDomain(&pCopy)
	}
	return result, nil
}

func (r *FilamentRepository) FindProfileByID(ctx context.Context, userID, id uuid.UUID) (*filament.FilamentProfile, error) {
	var p filamentProfileGORM
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&p).Error
	if err != nil {
		return nil, err
	}
	return mapProfileGORMToDomain(&p), nil
}


func (r *FilamentRepository) FindOrCreateProfile(ctx context.Context, userID uuid.UUID, brand, materialType string, spoolWeight float64) (*filament.FilamentProfile, error) {
	if brand == "" {
		brand = "Generic"
	}
	if materialType == "" {
		materialType = "PLA"
	}
	if spoolWeight <= 0 {
		spoolWeight = 1000
	}

	var p filamentProfileGORM
	err := r.db.WithContext(ctx).
		Where("LOWER(brand) = LOWER(?) AND LOWER(material_type) = LOWER(?)", brand, materialType).
		First(&p).Error

	if err == nil {
		return mapProfileGORMToDomain(&p), nil
	}

	// Create new default profile
	newProfile := filamentProfileGORM{
		ID:               uuid.New(),
		UserID:           userID,
		Brand:            brand,
		MaterialType:     materialType,
		DiameterMM:       1.75,
		SpoolWeightGrams: spoolWeight,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if err := r.db.WithContext(ctx).Create(&newProfile).Error; err != nil {
		return nil, err
	}
	return mapProfileGORMToDomain(&newProfile), nil
}

func (r *FilamentRepository) FindAllMaterialRates(ctx context.Context, userID uuid.UUID) ([]*filament.FilamentMaterialRate, error) {
	var list []filamentMaterialRateGORM
	err := r.db.WithContext(ctx).Order("material_type ASC").Find(&list).Error
	if err != nil {
		return nil, err
	}
	result := make([]*filament.FilamentMaterialRate, len(list))
	for i, mr := range list {
		result[i] = filament.ReconstructMaterialRate(
			mr.ID,
			mr.UserID,
			mr.MaterialType,
			mr.PricePerGram,
			mr.IsDefault,
			mr.Description,
			mr.CreatedAt,
			mr.UpdatedAt,
		)
	}
	return result, nil
}

func (r *FilamentRepository) FindMaterialRateByID(ctx context.Context, userID, id uuid.UUID) (*filament.FilamentMaterialRate, error) {
	var mr filamentMaterialRateGORM
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&mr).Error
	if err != nil {
		return nil, err
	}
	return filament.ReconstructMaterialRate(
		mr.ID,
		mr.UserID,
		mr.MaterialType,
		mr.PricePerGram,
		mr.IsDefault,
		mr.Description,
		mr.CreatedAt,
		mr.UpdatedAt,
	), nil
}

func (r *FilamentRepository) CreateMaterialRate(ctx context.Context, rate *filament.FilamentMaterialRate) error {
	m := filamentMaterialRateGORM{
		ID:           rate.ID(),
		UserID:       rate.UserID(),
		MaterialType: rate.MaterialType(),
		PricePerGram: rate.PricePerGram(),
		IsDefault:    rate.IsDefault(),
		Description:  rate.Description(),
		CreatedAt:    rate.CreatedAt(),
		UpdatedAt:    rate.UpdatedAt(),
	}
	return r.db.WithContext(ctx).Create(&m).Error
}

func (r *FilamentRepository) UpdateMaterialRate(ctx context.Context, rate *filament.FilamentMaterialRate) error {
	return r.db.WithContext(ctx).Model(&filamentMaterialRateGORM{}).
		Where("id = ?", rate.ID()).
		Updates(map[string]interface{}{
			"price_per_gram": rate.PricePerGram(),
			"is_default":     rate.IsDefault(),
			"description":    rate.Description(),
			"updated_at":     rate.UpdatedAt(),
		}).Error
}

func (r *FilamentRepository) DeleteMaterialRate(ctx context.Context, userID, id uuid.UUID) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&filamentMaterialRateGORM{}).Error
}