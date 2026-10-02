package dto

import (
	"time"

	"github.com/google/uuid"

	"symetra-lab-backend-v2/internal/domain/shopconfig"
)

type UpdateShopConfigRequest struct {
	FilamentPricePerRoll    float64 `json:"filament_price_per_roll"`
	FilamentWeightGrams     float64 `json:"filament_weight_grams"`
	ElectricityTariffPerKwh float64 `json:"electricity_tariff_per_kwh"`
	PrinterPowerWatts       float64 `json:"printer_power_watts"`
	PrinterPrice            float64 `json:"printer_price"`
	PrinterLifespanHours    float64 `json:"printer_lifespan_hours"`
	FailureBufferPercent    float64 `json:"failure_buffer_percent"`
}

type ShopConfigResponse struct {
	ID                      uuid.UUID `json:"id"`
	UserID                  uuid.UUID `json:"user_id"`
	FilamentPricePerRoll    float64   `json:"filament_price_per_roll"`
	FilamentWeightGrams     float64   `json:"filament_weight_grams"`
	ElectricityTariffPerKwh float64   `json:"electricity_tariff_per_kwh"`
	PrinterPowerWatts       float64   `json:"printer_power_watts"`
	PrinterPrice            float64   `json:"printer_price"`
	PrinterLifespanHours    float64   `json:"printer_lifespan_hours"`
	FailureBufferPercent    float64   `json:"failure_buffer_percent"`
	UpdatedAt               time.Time `json:"updated_at"`
}

func ToShopConfigResponse(c *shopconfig.ShopConfig) ShopConfigResponse {
	return ShopConfigResponse{
		ID:                      c.ID(),
		UserID:                  c.UserID(),
		FilamentPricePerRoll:    c.FilamentPricePerRoll(),
		FilamentWeightGrams:     c.FilamentWeightGrams(),
		ElectricityTariffPerKwh: c.ElectricityTariffPerKwh(),
		PrinterPowerWatts:       c.PrinterPowerWatts(),
		PrinterPrice:            c.PrinterPrice(),
		PrinterLifespanHours:    c.PrinterLifespanHours(),
		FailureBufferPercent:    c.FailureBufferPercent(),
		UpdatedAt:               c.UpdatedAt(),
	}
}

type MarketplacePlatformRequest struct {
	Name                string  `json:"name"`
	CommissionPercent   float64 `json:"commission_percent"`
	PromoFeePercent     float64 `json:"promo_fee_percent"`
	FreeShippingPercent float64 `json:"free_shipping_percent"`
	OrderFeeIDR         float64 `json:"order_fee_idr"`
	IsActive            bool    `json:"is_active"`
}

type UpdateMarketplacePlatformRequest struct {
	Name                *string  `json:"name"`
	CommissionPercent   *float64 `json:"commission_percent"`
	PromoFeePercent     *float64 `json:"promo_fee_percent"`
	FreeShippingPercent *float64 `json:"free_shipping_percent"`
	OrderFeeIDR         *float64 `json:"order_fee_idr"`
	IsActive            *bool    `json:"is_active"`
}

type MarketplacePlatformResponse struct {
	ID                  uuid.UUID `json:"id"`
	UserID              uuid.UUID `json:"user_id"`
	Name                string    `json:"name"`
	CommissionPercent   float64   `json:"commission_percent"`
	PromoFeePercent     float64   `json:"promo_fee_percent"`
	FreeShippingPercent float64   `json:"free_shipping_percent"`
	OrderFeeIDR         float64   `json:"order_fee_idr"`
	IsActive            bool      `json:"is_active"`
	CreatedAt           time.Time `json:"created_at"`
}

func ToMarketplaceResponse(m *shopconfig.MarketplacePlatform) MarketplacePlatformResponse {
	return MarketplacePlatformResponse{
		ID:                  m.ID(),
		UserID:              m.UserID(),
		Name:                m.Name(),
		CommissionPercent:   m.CommissionPercent(),
		PromoFeePercent:     m.PromoFeePercent(),
		FreeShippingPercent: m.FreeShippingPercent(),
		OrderFeeIDR:         m.OrderFeeIDR(),
		IsActive:            m.IsActive(),
		CreatedAt:           m.CreatedAt(),
	}
}
