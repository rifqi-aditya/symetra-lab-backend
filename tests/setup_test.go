package tests

import (
	"testing"
	"time"

	"symetra-lab-backend/models"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

const DefaultAdminUserID = "4aecd5c5-c0a9-4f52-ba2f-b6f4272218b4"

// StringPtr helper untuk membuat pointer string
func StringPtr(s string) *string {
	return &s
}

// FloatPtr helper untuk membuat pointer float64
func FloatPtr(f float64) *float64 {
	return &f
}

// SetupTestDB membuat in-memory SQLite database terisolasi untuk unit testing seluruh modul
func SetupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	err = db.AutoMigrate(
		&models.Shop{},
		&models.ShopeeOrder{},
		&models.ShopeeOrderItem{},
		&models.ShopeeOrderEscrow{},
		&models.Order{},
		&models.OrderItem{},
		&models.OrderFilament{},
		&models.OrderComponent{},
		&models.ProductCategory{},
		&models.Product{},
		&models.ProductFilament{},
		&models.ProductComponent{},
		&models.ProductPackagingItem{},
		&models.FilamentProfile{},
		&models.Filament{},
		&models.Machine{},
		&models.MachineMaintenancePart{},
		&models.Component{},
		&models.PackagingItem{},
		&models.PackagingPreset{},
		&models.PackagingPresetItem{},
		&models.ShopConfig{},
		&models.MarketplacePlatform{},
	)
	assert.NoError(t, err)

	// Seed ShopConfig default
	db.Create(&models.ShopConfig{
		ID:                      "cfg-default-test",
		UserID:                  DefaultAdminUserID,
		FilamentPricePerRoll:    188000,
		FilamentWeightGrams:     1000,
		ElectricityTariffPerKwh: 1700,
		PrinterPowerWatts:       150,
		PrinterPrice:            7500000,
		PrinterLifespanHours:    10000,
		FailureBufferPercent:    10,
		UpdatedAt:               time.Now(),
	})

	// Seed MarketplacePlatform Shopee default
	db.Create(&models.MarketplacePlatform{
		ID:                  "mp-shopee-test",
		UserID:              DefaultAdminUserID,
		Name:                "Shopee",
		CommissionPercent:   9.5,
		PromoFeePercent:     4.5,
		FreeShippingPercent: 0,
		OrderFeeIDR:         1250,
		IsActive:            true,
		CreatedAt:           time.Now(),
	})

	return db
}
