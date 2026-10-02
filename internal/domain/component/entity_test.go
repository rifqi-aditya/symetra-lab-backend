package component

import (
	"testing"

	"github.com/google/uuid"
)

func TestNewComponent(t *testing.T) {
	userID := uuid.New()
	desc := "High speed skate bearing"
	comp, err := NewComponent(
		userID,
		"608ZZ Ball Bearing",
		3500,
		25,
		&desc,
	)

	if err != nil {
		t.Fatalf("Expected nil err, got: %v", err)
	}
	if comp.Name() != "608ZZ Ball Bearing" {
		t.Errorf("Expected name match, got %s", comp.Name())
	}
	if comp.PricePerUnit() != 3500 {
		t.Errorf("Expected price 3500, got %f", comp.PricePerUnit())
	}
}
