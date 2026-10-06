package dto

import (
	"time"

	"github.com/google/uuid"

	"symetra-lab-backend-v2/internal/domain/filament"
)

type CreateFilamentRequest struct {
	Brand                  *string  `json:"brand"`
	MaterialType           *string  `json:"material_type"`
	DiameterMM             *float64 `json:"diameter_mm"`
	SpoolWeightGrams       *float64 `json:"spool_weight_grams"`
	ProfileID              *string  `json:"profile_id"`
	ColorName              string   `json:"color_name"`
	ColorHex               string   `json:"color_hex"`
	SKU                    *string  `json:"sku"`
	PricePerRoll           float64  `json:"price_per_roll"`
	CurrentStockGrams      *float64 `json:"current_stock_grams"`
	RemainingWeightGrams   *float64 `json:"remaining_weight_grams"`
	LowStockThresholdGrams *float64 `json:"low_stock_threshold_grams"`
}

type UpdateFilamentRequest struct {
	ProfileID              *string  `json:"profile_id"`
	ColorName              *string  `json:"color_name"`
	ColorHex               *string  `json:"color_hex"`
	SKU                    *string  `json:"sku"`
	PricePerRoll           *float64 `json:"price_per_roll"`
	CurrentStockGrams      *float64 `json:"current_stock_grams"`
	LowStockThresholdGrams *float64 `json:"low_stock_threshold_grams"`
}

type SyncStockRequest struct {
	GrossWeightGrams     *float64 `json:"gross_weight_grams"`
	RemainingWeightGrams *float64 `json:"remaining_weight_grams"`
	CurrentStockGrams    *float64 `json:"current_stock_grams"`
	IsWeighed            bool     `json:"is_weighed"`
}

type CreateMaterialRateRequest struct {
	MaterialType string  `json:"material_type"`
	PricePerGram float64 `json:"price_per_gram"`
	IsDefault    bool    `json:"is_default"`
	Description  *string `json:"description"`
}

type UpdateMaterialRateRequest struct {
	PricePerGram float64 `json:"price_per_gram"`
	IsDefault    bool    `json:"is_default"`
	Description  *string `json:"description"`
}

type FilamentMaterialRateResponse struct {
	ID           uuid.UUID `json:"id"`
	UserID       uuid.UUID `json:"user_id"`
	MaterialType string    `json:"material_type"`
	PricePerGram float64   `json:"price_per_gram"`
	IsDefault    bool      `json:"is_default"`
	Description  *string   `json:"description"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func ToMaterialRateResponse(r *filament.FilamentMaterialRate) FilamentMaterialRateResponse {
	return FilamentMaterialRateResponse{
		ID:           r.ID(),
		UserID:       r.UserID(),
		MaterialType: r.MaterialType(),
		PricePerGram: r.PricePerGram(),
		IsDefault:    r.IsDefault(),
		Description:  r.Description(),
		CreatedAt:    r.CreatedAt(),
		UpdatedAt:    r.UpdatedAt(),
	}
}

type FilamentProfileResponse struct {
	ID                       uuid.UUID `json:"id"`
	UserID                   uuid.UUID `json:"user_id"`
	Brand                    string    `json:"brand"`
	MaterialType             string    `json:"material_type"`
	DiameterMM               float64   `json:"diameter_mm"`
	EmptySpoolWeightGrams    *float64  `json:"empty_spool_weight_grams"`
	SpoolWeightGrams         float64   `json:"spool_weight_grams"`
	SpoolOuterDiameterMM     *float64  `json:"spool_outer_diameter_mm"`
	SpoolInnerHoleDiameterMM *float64  `json:"spool_inner_hole_diameter_mm"`
	CreatedAt                time.Time `json:"created_at"`
	UpdatedAt                time.Time `json:"updated_at"`
}

func ToFilamentProfileResponse(p *filament.FilamentProfile) FilamentProfileResponse {
	return FilamentProfileResponse{
		ID:                       p.ID(),
		UserID:                   p.UserID(),
		Brand:                    p.Brand(),
		MaterialType:             p.MaterialType(),
		DiameterMM:               p.DiameterMM(),
		EmptySpoolWeightGrams:    p.EmptySpoolWeightGrams(),
		SpoolWeightGrams:         p.SpoolWeightGrams(),
		SpoolOuterDiameterMM:     p.SpoolOuterDiameterMM(),
		SpoolInnerHoleDiameterMM: p.SpoolInnerHoleDiameterMM(),
		CreatedAt:                p.CreatedAt(),
		UpdatedAt:                p.UpdatedAt(),
	}
}

type FilamentResponse struct {
	ID                       uuid.UUID  `json:"id"`
	UserID                   uuid.UUID  `json:"user_id"`
	ProfileID                *uuid.UUID `json:"profile_id"`
	Brand                    string     `json:"brand"`
	MaterialType             string     `json:"material_type"`
	DiameterMM               float64    `json:"diameter_mm"`
	EmptySpoolWeightGrams    *float64   `json:"empty_spool_weight_grams"`
	SpoolWeightGrams         float64    `json:"spool_weight_grams"`
	SpoolOuterDiameterMM     *float64   `json:"spool_outer_diameter_mm"`
	SpoolInnerHoleDiameterMM *float64   `json:"spool_inner_hole_diameter_mm"`
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

func ToFilamentResponse(f *filament.Filament) FilamentResponse {
	resp := FilamentResponse{
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

	if f.Profile() != nil {
		prof := f.Profile()
		resp.Brand = prof.Brand()
		resp.MaterialType = prof.MaterialType()
		resp.DiameterMM = prof.DiameterMM()
		resp.EmptySpoolWeightGrams = prof.EmptySpoolWeightGrams()
		resp.SpoolWeightGrams = prof.SpoolWeightGrams()
		resp.SpoolOuterDiameterMM = prof.SpoolOuterDiameterMM()
		resp.SpoolInnerHoleDiameterMM = prof.SpoolInnerHoleDiameterMM()
	}

	return resp
}