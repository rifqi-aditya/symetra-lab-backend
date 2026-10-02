package shopconfig

import (
	"time"

	"github.com/google/uuid"
)

type ShopConfig struct {
	id                      uuid.UUID
	userID                  uuid.UUID
	filamentPricePerRoll    float64
	filamentWeightGrams     float64
	electricityTariffPerKwh float64
	printerPowerWatts       float64
	printerPrice            float64
	printerLifespanHours    float64
	failureBufferPercent    float64
	updatedAt               time.Time
}

func NewShopConfig(
	userID uuid.UUID,
	filamentPricePerRoll, filamentWeightGrams, elecTariff, powerWatts, printerPrice, lifespanHours, failureBuffer float64,
) *ShopConfig {
	return &ShopConfig{
		id:                      uuid.New(),
		userID:                  userID,
		filamentPricePerRoll:    filamentPricePerRoll,
		filamentWeightGrams:     filamentWeightGrams,
		electricityTariffPerKwh: elecTariff,
		printerPowerWatts:       powerWatts,
		printerPrice:            printerPrice,
		printerLifespanHours:    lifespanHours,
		failureBufferPercent:    failureBuffer,
		updatedAt:               time.Now(),
	}
}

func ReconstructShopConfig(
	id, userID uuid.UUID,
	filamentPricePerRoll, filamentWeightGrams, elecTariff, powerWatts, printerPrice, lifespanHours, failureBuffer float64,
	updatedAt time.Time,
) *ShopConfig {
	return &ShopConfig{
		id:                      id,
		userID:                  userID,
		filamentPricePerRoll:    filamentPricePerRoll,
		filamentWeightGrams:     filamentWeightGrams,
		electricityTariffPerKwh: elecTariff,
		printerPowerWatts:       powerWatts,
		printerPrice:            printerPrice,
		printerLifespanHours:    lifespanHours,
		failureBufferPercent:    failureBuffer,
		updatedAt:               updatedAt,
	}
}

func (s *ShopConfig) ID() uuid.UUID                     { return s.id }
func (s *ShopConfig) UserID() uuid.UUID                 { return s.userID }
func (s *ShopConfig) FilamentPricePerRoll() float64     { return s.filamentPricePerRoll }
func (s *ShopConfig) FilamentWeightGrams() float64     { return s.filamentWeightGrams }
func (s *ShopConfig) ElectricityTariffPerKwh() float64 { return s.electricityTariffPerKwh }
func (s *ShopConfig) PrinterPowerWatts() float64       { return s.printerPowerWatts }
func (s *ShopConfig) PrinterPrice() float64            { return s.printerPrice }
func (s *ShopConfig) PrinterLifespanHours() float64    { return s.printerLifespanHours }
func (s *ShopConfig) FailureBufferPercent() float64    { return s.failureBufferPercent }
func (s *ShopConfig) UpdatedAt() time.Time             { return s.updatedAt }

func (s *ShopConfig) Update(
	filamentPricePerRoll, filamentWeightGrams, elecTariff, powerWatts, printerPrice, lifespanHours, failureBuffer float64,
) {
	if filamentPricePerRoll > 0 {
		s.filamentPricePerRoll = filamentPricePerRoll
	}
	if filamentWeightGrams > 0 {
		s.filamentWeightGrams = filamentWeightGrams
	}
	if elecTariff > 0 {
		s.electricityTariffPerKwh = elecTariff
	}
	if powerWatts > 0 {
		s.printerPowerWatts = powerWatts
	}
	if printerPrice > 0 {
		s.printerPrice = printerPrice
	}
	if lifespanHours > 0 {
		s.printerLifespanHours = lifespanHours
	}
	if failureBuffer >= 0 {
		s.failureBufferPercent = failureBuffer
	}
	s.updatedAt = time.Now()
}

type MarketplacePlatform struct {
	id                  uuid.UUID
	userID              uuid.UUID
	name                string
	commissionPercent   float64
	promoFeePercent     float64
	freeShippingPercent float64
	orderFeeIDR         float64
	isActive            bool
	createdAt           time.Time
}

func NewMarketplacePlatform(
	userID uuid.UUID,
	name string,
	commission, promo, freeShipping, orderFee float64,
	active bool,
) *MarketplacePlatform {
	return &MarketplacePlatform{
		id:                  uuid.New(),
		userID:              userID,
		name:                name,
		commissionPercent:   commission,
		promoFeePercent:     promo,
		freeShippingPercent: freeShipping,
		orderFeeIDR:         orderFee,
		isActive:            active,
		createdAt:           time.Now(),
	}
}

func ReconstructMarketplacePlatform(
	id, userID uuid.UUID,
	name string,
	commission, promo, freeShipping, orderFee float64,
	active bool,
	createdAt time.Time,
) *MarketplacePlatform {
	return &MarketplacePlatform{
		id:                  id,
		userID:              userID,
		name:                name,
		commissionPercent:   commission,
		promoFeePercent:     promo,
		freeShippingPercent: freeShipping,
		orderFeeIDR:         orderFee,
		isActive:            active,
		createdAt:           createdAt,
	}
}

func (m *MarketplacePlatform) ID() uuid.UUID                { return m.id }
func (m *MarketplacePlatform) UserID() uuid.UUID            { return m.userID }
func (m *MarketplacePlatform) Name() string                 { return m.name }
func (m *MarketplacePlatform) CommissionPercent() float64   { return m.commissionPercent }
func (m *MarketplacePlatform) PromoFeePercent() float64     { return m.promoFeePercent }
func (m *MarketplacePlatform) FreeShippingPercent() float64 { return m.freeShippingPercent }
func (m *MarketplacePlatform) OrderFeeIDR() float64         { return m.orderFeeIDR }
func (m *MarketplacePlatform) IsActive() bool               { return m.isActive }
func (m *MarketplacePlatform) CreatedAt() time.Time         { return m.createdAt }

func (m *MarketplacePlatform) Update(
	name *string,
	commission, promo, freeShipping, orderFee *float64,
	active *bool,
) {
	if name != nil {
		m.name = *name
	}
	if commission != nil {
		m.commissionPercent = *commission
	}
	if promo != nil {
		m.promoFeePercent = *promo
	}
	if freeShipping != nil {
		m.freeShippingPercent = *freeShipping
	}
	if orderFee != nil {
		m.orderFeeIDR = *orderFee
	}
	if active != nil {
		m.isActive = *active
	}
}
