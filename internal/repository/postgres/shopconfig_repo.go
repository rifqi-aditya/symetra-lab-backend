package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"symetra-lab-backend-v2/internal/domain/shopconfig"
)

type shopConfigGORM struct {
	ID                      uuid.UUID `gorm:"column:id;primaryKey;type:uuid"`
	UserID                  uuid.UUID `gorm:"column:user_id;type:uuid"`
	FilamentPricePerRoll    float64   `gorm:"column:filament_price_per_roll"`
	FilamentWeightGrams     float64   `gorm:"column:filament_weight_grams"`
	ElectricityTariffPerKwh float64   `gorm:"column:electricity_tariff_per_kwh"`
	PrinterPowerWatts       float64   `gorm:"column:printer_power_watts"`
	PrinterPrice            float64   `gorm:"column:printer_price"`
	PrinterLifespanHours    float64   `gorm:"column:printer_lifespan_hours"`
	FailureBufferPercent    float64   `gorm:"column:failure_buffer_percent"`
	UpdatedAt               time.Time `gorm:"column:updated_at"`
}

func (shopConfigGORM) TableName() string {
	return "shop_configs"
}

type marketplacePlatformGORM struct {
	ID                  uuid.UUID `gorm:"column:id;primaryKey;type:uuid"`
	UserID              uuid.UUID `gorm:"column:user_id;type:uuid"`
	Name                string    `gorm:"column:name"`
	CommissionPercent   float64   `gorm:"column:commission_percent"`
	PromoFeePercent     float64   `gorm:"column:promo_fee_percent"`
	FreeShippingPercent float64   `gorm:"column:free_shipping_percent"`
	OrderFeeIDR         float64   `gorm:"column:order_fee_idr"`
	IsActive            bool      `gorm:"column:is_active"`
	CreatedAt           time.Time `gorm:"column:created_at"`
}

func (marketplacePlatformGORM) TableName() string {
	return "marketplace_platforms"
}

type ShopConfigRepository struct {
	db *gorm.DB
}

func NewShopConfigRepository(db *gorm.DB) *ShopConfigRepository {
	return &ShopConfigRepository{db: db}
}

func (r *ShopConfigRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*shopconfig.ShopConfig, error) {
	var g shopConfigGORM
	err := r.db.WithContext(ctx).First(&g).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Initialize default workshop configuration (matching v1 behavior)
			g = shopConfigGORM{
				ID:                      uuid.New(),
				UserID:                  userID,
				FilamentPricePerRoll:    200000,
				FilamentWeightGrams:     1000,
				ElectricityTariffPerKwh: 1700,
				PrinterPowerWatts:       200,
				PrinterPrice:            5000000,
				PrinterLifespanHours:    3000,
				FailureBufferPercent:    10,
				UpdatedAt:               time.Now(),
			}
			_ = r.db.WithContext(ctx).Create(&g)
		} else {
			return nil, err
		}
	}
	return shopconfig.ReconstructShopConfig(
		g.ID,
		g.UserID,
		g.FilamentPricePerRoll,
		g.FilamentWeightGrams,
		g.ElectricityTariffPerKwh,
		g.PrinterPowerWatts,
		g.PrinterPrice,
		g.PrinterLifespanHours,
		g.FailureBufferPercent,
		g.UpdatedAt,
	), nil
}

func (r *ShopConfigRepository) Save(ctx context.Context, cfg *shopconfig.ShopConfig) error {
	g := shopConfigGORM{
		ID:                      cfg.ID(),
		UserID:                  cfg.UserID(),
		FilamentPricePerRoll:    cfg.FilamentPricePerRoll(),
		FilamentWeightGrams:     cfg.FilamentWeightGrams(),
		ElectricityTariffPerKwh: cfg.ElectricityTariffPerKwh(),
		PrinterPowerWatts:       cfg.PrinterPowerWatts(),
		PrinterPrice:            cfg.PrinterPrice(),
		PrinterLifespanHours:    cfg.PrinterLifespanHours(),
		FailureBufferPercent:    cfg.FailureBufferPercent(),
		UpdatedAt:               cfg.UpdatedAt(),
	}
	return r.db.WithContext(ctx).Save(&g).Error
}

func (r *ShopConfigRepository) GetActiveMarketplace(ctx context.Context, userID uuid.UUID, platformName string) (*shopconfig.MarketplacePlatform, error) {
	var g marketplacePlatformGORM
	err := r.db.WithContext(ctx).Where("LOWER(name) = LOWER(?) AND is_active = ?", platformName, true).First(&g).Error
	if err != nil {
		return nil, err
	}
	return shopconfig.ReconstructMarketplacePlatform(
		g.ID,
		g.UserID,
		g.Name,
		g.CommissionPercent,
		g.PromoFeePercent,
		g.FreeShippingPercent,
		g.OrderFeeIDR,
		g.IsActive,
		g.CreatedAt,
	), nil
}

func (r *ShopConfigRepository) ListMarketplaces(ctx context.Context, userID uuid.UUID) ([]*shopconfig.MarketplacePlatform, error) {
	var list []marketplacePlatformGORM
	if err := r.db.WithContext(ctx).Order("created_at ASC").Find(&list).Error; err != nil {
		return nil, err
	}
	res := make([]*shopconfig.MarketplacePlatform, len(list))
	for i, g := range list {
		res[i] = shopconfig.ReconstructMarketplacePlatform(
			g.ID,
			g.UserID,
			g.Name,
			g.CommissionPercent,
			g.PromoFeePercent,
			g.FreeShippingPercent,
			g.OrderFeeIDR,
			g.IsActive,
			g.CreatedAt,
		)
	}
	return res, nil
}

func (r *ShopConfigRepository) GetMarketplaceByID(ctx context.Context, id uuid.UUID) (*shopconfig.MarketplacePlatform, error) {
	var g marketplacePlatformGORM
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&g).Error; err != nil {
		return nil, err
	}
	return shopconfig.ReconstructMarketplacePlatform(
		g.ID,
		g.UserID,
		g.Name,
		g.CommissionPercent,
		g.PromoFeePercent,
		g.FreeShippingPercent,
		g.OrderFeeIDR,
		g.IsActive,
		g.CreatedAt,
	), nil
}

func (r *ShopConfigRepository) CreateMarketplace(ctx context.Context, p *shopconfig.MarketplacePlatform) error {
	g := marketplacePlatformGORM{
		ID:                  p.ID(),
		UserID:              p.UserID(),
		Name:                p.Name(),
		CommissionPercent:   p.CommissionPercent(),
		PromoFeePercent:     p.PromoFeePercent(),
		FreeShippingPercent: p.FreeShippingPercent(),
		OrderFeeIDR:         p.OrderFeeIDR(),
		IsActive:            p.IsActive(),
		CreatedAt:           p.CreatedAt(),
	}
	return r.db.WithContext(ctx).Create(&g).Error
}

func (r *ShopConfigRepository) UpdateMarketplace(ctx context.Context, p *shopconfig.MarketplacePlatform) error {
	g := marketplacePlatformGORM{
		ID:                  p.ID(),
		UserID:              p.UserID(),
		Name:                p.Name(),
		CommissionPercent:   p.CommissionPercent(),
		PromoFeePercent:     p.PromoFeePercent(),
		FreeShippingPercent: p.FreeShippingPercent(),
		OrderFeeIDR:         p.OrderFeeIDR(),
		IsActive:            p.IsActive(),
		CreatedAt:           p.CreatedAt(),
	}
	return r.db.WithContext(ctx).Save(&g).Error
}

func (r *ShopConfigRepository) DeleteMarketplace(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&marketplacePlatformGORM{}).Error
}
