package packaging

import (
	"testing"

	"github.com/google/uuid"
)

func TestNewPackagingItem(t *testing.T) {
	userID := uuid.New()
	item, err := NewPackagingItem(
		userID,
		"Cardboard Box 20x20x10",
		"PCS",
		25.0,
		20,
	)

	if err != nil {
		t.Fatalf("Expected nil err, got: %v", err)
	}
	if item.Name() != "Cardboard Box 20x20x10" {
		t.Errorf("Expected name match, got %s", item.Name())
	}
	if item.UnitCost() != 25.0 {
		t.Errorf("Expected unit cost 25.0, got %f", item.UnitCost())
	}
}
