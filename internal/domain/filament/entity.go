package filament

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

// FilamentProfile represents technical printing profile presets for filaments.
type FilamentProfile struct {
	id                       uuid.UUID
	userID                   uuid.UUID
	brand                    string
	materialType             string
	diameterMM               float64
	emptySpoolWeightGrams    *float64
	spoolWeightGrams         float64
	spoolOuterDiameterMM     *float64
	spoolInnerHoleDiameterMM *float64
	createdAt                time.Time
	updatedAt                time.Time
}

func ReconstructProfile(
	id, userID uuid.UUID,
	brand, materialType string,
	diameterMM float64,
	emptySpoolWeightGrams *float64,
	spoolWeightGrams float64,
	spoolOuterDiameterMM, spoolInnerHoleDiameterMM *float64,
	createdAt, updatedAt time.Time,
) *FilamentProfile {
	return &FilamentProfile{
		id:                       id,
		userID:                   userID,
		brand:                    brand,
		materialType:             materialType,
		diameterMM:               diameterMM,
		emptySpoolWeightGrams:    emptySpoolWeightGrams,
		spoolWeightGrams:         spoolWeightGrams,
		spoolOuterDiameterMM:     spoolOuterDiameterMM,
		spoolInnerHoleDiameterMM: spoolInnerHoleDiameterMM,
		createdAt:                createdAt,
		updatedAt:                updatedAt,
	}
}

func (p *FilamentProfile) ID() uuid.UUID                      { return p.id }
func (p *FilamentProfile) UserID() uuid.UUID                  { return p.userID }
func (p *FilamentProfile) Brand() string                      { return p.brand }
func (p *FilamentProfile) MaterialType() string               { return p.materialType }
func (p *FilamentProfile) DiameterMM() float64                { return p.diameterMM }
func (p *FilamentProfile) EmptySpoolWeightGrams() *float64    { return p.emptySpoolWeightGrams }
func (p *FilamentProfile) SpoolWeightGrams() float64          { return p.spoolWeightGrams }
func (p *FilamentProfile) SpoolOuterDiameterMM() *float64     { return p.spoolOuterDiameterMM }
func (p *FilamentProfile) SpoolInnerHoleDiameterMM() *float64 { return p.spoolInnerHoleDiameterMM }
func (p *FilamentProfile) CreatedAt() time.Time               { return p.createdAt }
func (p *FilamentProfile) UpdatedAt() time.Time               { return p.updatedAt }

// Filament represents a physical spool of 3D printing filament.
type Filament struct {
	id                     uuid.UUID
	userID                 uuid.UUID
	profileID              *uuid.UUID
	colorName              string
	colorHex               string
	sku                    *string
	pricePerRoll           float64
	currentStockGrams      float64
	lowStockThresholdGrams float64
	lastWeighedGrams       *float64
	lastWeighedAt          *time.Time
	profile                *FilamentProfile
	createdAt              time.Time
	updatedAt              time.Time
}

func NewFilament(
	userID uuid.UUID,
	profileID *uuid.UUID,
	colorName, colorHex string,
	sku *string,
	pricePerRoll, currentStockGrams, lowStockThresholdGrams float64,
) (*Filament, error) {
	if userID == uuid.Nil {
		return nil, ErrUserIDRequired
	}
	cleanColor := strings.TrimSpace(colorName)
	if cleanColor == "" {
		return nil, ErrColorNameRequired
	}
	cleanHex := strings.TrimSpace(colorHex)
	if cleanHex == "" {
		cleanHex = "#FFFFFF"
	}
	now := time.Now()
	return &Filament{
		id:                     uuid.New(),
		userID:                 userID,
		profileID:              profileID,
		colorName:              cleanColor,
		colorHex:               cleanHex,
		sku:                    sku,
		pricePerRoll:           pricePerRoll,
		currentStockGrams:      currentStockGrams,
		lowStockThresholdGrams: lowStockThresholdGrams,
		createdAt:              now,
		updatedAt:              now,
	}, nil
}

func ReconstructFilament(
	id, userID uuid.UUID,
	profileID *uuid.UUID,
	colorName, colorHex string,
	sku *string,
	pricePerRoll, currentStockGrams, lowStockThresholdGrams float64,
	lastWeighedGrams *float64,
	lastWeighedAt *time.Time,
	profile *FilamentProfile,
	createdAt, updatedAt time.Time,
) *Filament {
	return &Filament{
		id:                     id,
		userID:                 userID,
		profileID:              profileID,
		colorName:              colorName,
		colorHex:               colorHex,
		sku:                    sku,
		pricePerRoll:           pricePerRoll,
		currentStockGrams:      currentStockGrams,
		lowStockThresholdGrams: lowStockThresholdGrams,
		lastWeighedGrams:       lastWeighedGrams,
		lastWeighedAt:          lastWeighedAt,
		profile:                profile,
		createdAt:              createdAt,
		updatedAt:              updatedAt,
	}
}

func (f *Filament) ID() uuid.UUID                  { return f.id }
func (f *Filament) UserID() uuid.UUID              { return f.userID }
func (f *Filament) ProfileID() *uuid.UUID          { return f.profileID }
func (f *Filament) ColorName() string              { return f.colorName }
func (f *Filament) ColorHex() string               { return f.colorHex }
func (f *Filament) SKU() *string                   { return f.sku }
func (f *Filament) PricePerRoll() float64          { return f.pricePerRoll }
func (f *Filament) CurrentStockGrams() float64      { return f.currentStockGrams }
func (f *Filament) LowStockThresholdGrams() float64 { return f.lowStockThresholdGrams }
func (f *Filament) LastWeighedGrams() *float64     { return f.lastWeighedGrams }
func (f *Filament) LastWeighedAt() *time.Time      { return f.lastWeighedAt }
func (f *Filament) Profile() *FilamentProfile      { return f.profile }
func (f *Filament) CreatedAt() time.Time           { return f.createdAt }
func (f *Filament) UpdatedAt() time.Time           { return f.updatedAt }

func (f *Filament) Update(
	profileID *uuid.UUID,
	colorName, colorHex string,
	sku *string,
	pricePerRoll, currentStock, threshold float64,
) {
	f.profileID = profileID
	if colorName != "" {
		f.colorName = colorName
	}
	if colorHex != "" {
		f.colorHex = colorHex
	}
	f.sku = sku
	f.pricePerRoll = pricePerRoll
	f.currentStockGrams = currentStock
	f.lowStockThresholdGrams = threshold
	f.updatedAt = time.Now()
}

func (f *Filament) SyncStock(grams float64, isWeighed bool) {
	f.currentStockGrams = grams
	now := time.Now()
	if isWeighed {
		f.lastWeighedGrams = &grams
		f.lastWeighedAt = &now
	}
	f.updatedAt = now
}

// FilamentMaterialRate represents global workshop pricing rates per gram of material.
type FilamentMaterialRate struct {
	id           uuid.UUID
	userID       uuid.UUID
	materialType string
	pricePerGram float64
	isDefault    bool
	description  *string
	createdAt    time.Time
	updatedAt    time.Time
}

func NewMaterialRate(
	userID uuid.UUID,
	materialType string,
	pricePerGram float64,
	isDefault bool,
	description *string,
) (*FilamentMaterialRate, error) {
	if userID == uuid.Nil {
		return nil, ErrUserIDRequired
	}
	now := time.Now()
	return &FilamentMaterialRate{
		id:           uuid.New(),
		userID:       userID,
		materialType: strings.ToUpper(strings.TrimSpace(materialType)),
		pricePerGram: pricePerGram,
		isDefault:    isDefault,
		description:  description,
		createdAt:    now,
		updatedAt:    now,
	}, nil
}

func ReconstructMaterialRate(
	id, userID uuid.UUID,
	materialType string,
	pricePerGram float64,
	isDefault bool,
	description *string,
	createdAt, updatedAt time.Time,
) *FilamentMaterialRate {
	return &FilamentMaterialRate{
		id:           id,
		userID:       userID,
		materialType: materialType,
		pricePerGram: pricePerGram,
		isDefault:    isDefault,
		description:  description,
		createdAt:    createdAt,
		updatedAt:    updatedAt,
	}
}

func (r *FilamentMaterialRate) ID() uuid.UUID          { return r.id }
func (r *FilamentMaterialRate) UserID() uuid.UUID      { return r.userID }
func (r *FilamentMaterialRate) MaterialType() string   { return r.materialType }
func (r *FilamentMaterialRate) PricePerGram() float64  { return r.pricePerGram }
func (r *FilamentMaterialRate) IsDefault() bool        { return r.isDefault }
func (r *FilamentMaterialRate) Description() *string   { return r.description }
func (r *FilamentMaterialRate) CreatedAt() time.Time   { return r.createdAt }
func (r *FilamentMaterialRate) UpdatedAt() time.Time   { return r.updatedAt }

func (r *FilamentMaterialRate) Update(pricePerGram float64, isDefault bool, description *string) {
	r.pricePerGram = pricePerGram
	r.isDefault = isDefault
	r.description = description
	r.updatedAt = time.Now()
}