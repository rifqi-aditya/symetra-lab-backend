package filament

import (
	"testing"

	"github.com/google/uuid"
)

func TestNewFilament_Success(t *testing.T) {
	userID := uuid.New()
	profileID := uuid.New()
	sku := "FIL-RED-001"

	fil, err := NewFilament(
		userID,
		&profileID,
		"Signal Red",
		"#ff0000",
		&sku,
		165000,
		1000,
		150,
	)

	if err != nil {
		t.Fatalf("Expected nil err, got: %v", err)
	}
	if fil.ColorName() != "Signal Red" {
		t.Errorf("Expected color name 'Signal Red', got '%s'", fil.ColorName())
	}
	if fil.CurrentStockGrams() != 1000 {
		t.Errorf("Expected stock 1000, got %f", fil.CurrentStockGrams())
	}
}

func TestFilament_SyncStock(t *testing.T) {
	userID := uuid.New()
	profileID := uuid.New()

	fil, _ := NewFilament(
		userID,
		&profileID,
		"Blue",
		"#0000ff",
		nil,
		160000,
		800,
		200,
	)

	fil.SyncStock(450.0, true)
	if fil.CurrentStockGrams() != 450.0 {
		t.Errorf("Expected synced stock 450.0, got %f", fil.CurrentStockGrams())
	}
}
