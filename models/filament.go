package models

import (
	"time"
)

// FilamentProfile merepresentasikan master profil teknis filamen (suhu, ekstrusi, berat spool)
type FilamentProfile struct {
	ID                       string    `gorm:"type:uuid;primaryKey" json:"id"`
	UserID                   string    `gorm:"type:uuid;not null;index" json:"user_id"`
	Brand                    string    `gorm:"type:text;not null" json:"brand"`
	MaterialType             string    `gorm:"type:text;not null" json:"material_type"`
	DiameterMM               float64   `gorm:"type:numeric;default:1.75" json:"diameter_mm"`
	NozzleTemp               *int      `json:"nozzle_temp"`
	BedTemp                  *int      `json:"bed_temp"`
	RetractionLength         *float64  `gorm:"type:numeric" json:"retraction_length"`
	FlowRatio                *float64  `gorm:"type:numeric" json:"flow_ratio"`
	PressureAdvance          *float64  `gorm:"type:numeric" json:"pressure_advance"`
	CoolingFanPercent        *int      `json:"cooling_fan_percent"`
	MaxVolumetricSpeed       *float64  `gorm:"type:numeric" json:"max_volumetric_speed"`
	EmptySpoolWeightGrams    *float64  `gorm:"type:numeric" json:"empty_spool_weight_grams"`
	SpoolWeightGrams         float64   `gorm:"type:numeric;default:1000" json:"spool_weight_grams"`
	SpoolOuterDiameterMM     *float64  `gorm:"type:numeric" json:"spool_outer_diameter_mm"`
	SpoolInnerHoleDiameterMM *float64  `gorm:"type:numeric" json:"spool_inner_hole_diameter_mm"`
	SpoolWidthMM             *float64  `gorm:"type:numeric" json:"spool_width_mm"`
	CreatedAt                time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt                time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (FilamentProfile) TableName() string {
	return "filament_profiles"
}

// Filament merepresentasikan satu unit spool / varian warna fisik di bengkel
type Filament struct {
	ID                     string           `gorm:"type:uuid;primaryKey" json:"id"`
	UserID                 string           `gorm:"type:uuid;not null;index" json:"user_id"`
	ProfileID              string           `gorm:"type:uuid;not null;index" json:"profile_id"`
	ColorName              string           `gorm:"size:255" json:"color_name"`
	ColorHex               string           `gorm:"size:255" json:"color_hex"`
	SKU                    *string          `gorm:"size:255" json:"sku"`
	PricePerRoll           float64          `gorm:"type:numeric;default:0;not null" json:"price_per_roll"`
	CurrentStockGrams      float64          `gorm:"type:numeric" json:"current_stock_grams"`
	LowStockThresholdGrams float64          `gorm:"type:numeric;default:200" json:"low_stock_threshold_grams"`
	LastWeighedGrams       *float64         `gorm:"type:numeric" json:"last_weighed_grams"`
	LastWeighedAt          *time.Time       `json:"last_weighed_at"`
	CreatedAt              time.Time        `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt              time.Time        `gorm:"autoUpdateTime" json:"updated_at"`

	// Relasi ke FilamentProfile
	Profile *FilamentProfile `gorm:"foreignKey:ProfileID;references:ID" json:"profile,omitempty"`
}

func (Filament) TableName() string {
	return "filaments"
}

// FilamentResponse merepresentasikan format JSON gabungan (flattened) yang siap dikonsumsi langsung oleh React frontend
type FilamentResponse struct {
	ID                       string     `json:"id"`
	UserID                   string     `json:"user_id"`
	ProfileID                string     `json:"profile_id"`
	Brand                    string     `json:"brand"`
	MaterialType             string     `json:"material_type"`
	DiameterMM               float64    `json:"diameter_mm"`
	NozzleTemp               *int       `json:"nozzle_temp"`
	BedTemp                  *int       `json:"bed_temp"`
	RetractionLength         *float64   `json:"retraction_length"`
	FlowRatio                *float64   `json:"flow_ratio"`
	PressureAdvance          *float64   `json:"pressure_advance"`
	CoolingFanPercent        *int       `json:"cooling_fan_percent"`
	MaxVolumetricSpeed       *float64   `json:"max_volumetric_speed"`
	EmptySpoolWeightGrams    *float64   `json:"empty_spool_weight_grams"`
	SpoolWeightGrams         float64    `json:"spool_weight_grams"`
	SpoolOuterDiameterMM     *float64   `json:"spool_outer_diameter_mm"`
	SpoolInnerHoleDiameterMM *float64   `json:"spool_inner_hole_diameter_mm"`
	SpoolWidthMM             *float64   `json:"spool_width_mm"`
	ColorName                string     `json:"color_name"`
	ColorHex                 string     `json:"color_hex"`
	SKU                      *string    `json:"sku"`
	PricePerRoll             float64    `json:"price_per_roll"`
	CurrentStockGrams        float64    `json:"current_stock_grams"`
	LowStockThresholdGrams   float64    `json:"low_stock_threshold_grams"`
	LastWeighedGrams         *float64   `json:"last_weighed_grams"`
	LastWeighedAt            *time.Time `json:"last_weighed_at"`
	CreatedAt                time.Time  `json:"created_at"`
	UpdatedAt                time.Time  `json:"updated_at"`
}

// ToResponse mengubah struct Filament & relasi Profile-nya menjadi format flattened
func (f *Filament) ToResponse() FilamentResponse {
	resp := FilamentResponse{
		ID:                     f.ID,
		UserID:                 f.UserID,
		ProfileID:              f.ProfileID,
		ColorName:              f.ColorName,
		ColorHex:               f.ColorHex,
		SKU:                    f.SKU,
		PricePerRoll:           f.PricePerRoll,
		CurrentStockGrams:      f.CurrentStockGrams,
		LowStockThresholdGrams: f.LowStockThresholdGrams,
		LastWeighedGrams:       f.LastWeighedGrams,
		LastWeighedAt:          f.LastWeighedAt,
		CreatedAt:              f.CreatedAt,
		UpdatedAt:              f.UpdatedAt,
	}

	if f.Profile != nil {
		resp.Brand = f.Profile.Brand
		resp.MaterialType = f.Profile.MaterialType
		resp.DiameterMM = f.Profile.DiameterMM
		resp.NozzleTemp = f.Profile.NozzleTemp
		resp.BedTemp = f.Profile.BedTemp
		resp.RetractionLength = f.Profile.RetractionLength
		resp.FlowRatio = f.Profile.FlowRatio
		resp.PressureAdvance = f.Profile.PressureAdvance
		resp.CoolingFanPercent = f.Profile.CoolingFanPercent
		resp.MaxVolumetricSpeed = f.Profile.MaxVolumetricSpeed
		resp.EmptySpoolWeightGrams = f.Profile.EmptySpoolWeightGrams
		resp.SpoolWeightGrams = f.Profile.SpoolWeightGrams
		resp.SpoolOuterDiameterMM = f.Profile.SpoolOuterDiameterMM
		resp.SpoolInnerHoleDiameterMM = f.Profile.SpoolInnerHoleDiameterMM
		resp.SpoolWidthMM = f.Profile.SpoolWidthMM
	}

	return resp
}
