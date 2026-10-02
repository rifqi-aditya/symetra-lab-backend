package shopconfig

import (
	"testing"

	"github.com/google/uuid"
)

func TestShopConfig_Update(t *testing.T) {
	cfg := NewShopConfig(
		uuid.New(),
		200000, 1000, 1700, 200, 5000000, 3000, 10,
	)

	cfg.Update(210000, 1000, 1800, 250, 5500000, 3500, 12)

	if cfg.ElectricityTariffPerKwh() != 1800 {
		t.Errorf("Expected tariff 1800, got %f", cfg.ElectricityTariffPerKwh())
	}
	if cfg.PrinterPowerWatts() != 250 {
		t.Errorf("Expected power 250, got %f", cfg.PrinterPowerWatts())
	}
}

func TestMarketplacePlatform_Update(t *testing.T) {
	mp := NewMarketplacePlatform(
		uuid.New(),
		"Shopee Star Seller",
		6.5, 3.5, 4.0, 1000, true,
	)

	newName := "Shopee Mall"
	newComm := 8.5
	mp.Update(&newName, &newComm, nil, nil, nil, nil)

	if mp.Name() != "Shopee Mall" {
		t.Errorf("Expected name 'Shopee Mall', got '%s'", mp.Name())
	}
	if mp.CommissionPercent() != 8.5 {
		t.Errorf("Expected commission 8.5, got %f", mp.CommissionPercent())
	}
}
