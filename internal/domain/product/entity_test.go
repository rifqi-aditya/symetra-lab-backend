package product

import (
	"testing"

	"github.com/google/uuid"
)

func TestNewParentProduct_Success(t *testing.T) {
	userID := uuid.New()
	parentSKU := "PAR-MOUSE-001"
	pp, err := NewParentProduct(
		userID,
		"Test Ergonomic Mouse Grip",
		parentSKU,
		nil,
		nil,
		nil,
		nil,
	)
	if err != nil {
		t.Fatalf("Expected nil err, got: %v", err)
	}
	if pp.ID() == uuid.Nil {
		t.Error("Expected valid parent product UUID")
	}
	if pp.Name() != "Test Ergonomic Mouse Grip" {
		t.Errorf("Expected name match, got %s", pp.Name())
	}
	if pp.ParentSKU() != "PAR-MOUSE-001" {
		t.Errorf("Expected SKU match, got %s", pp.ParentSKU())
	}
}

func TestNewProduct_Success(t *testing.T) {
	parentID := uuid.New()
	sku := "SKU-MOUSE-001"
	p, err := NewProduct(
		parentID,
		"Standard",
		&sku,
		"PLA",
		45.0,
		1.5,
	)

	if err != nil {
		t.Fatalf("Expected nil err, got: %v", err)
	}
	if p.ID() == uuid.Nil {
		t.Error("Expected valid product UUID")
	}
	if p.VariantName() != "Standard" {
		t.Errorf("Expected variant name match, got %s", p.VariantName())
	}
	if p.DefaultWeightGrams() != 45.0 {
		t.Errorf("Expected weight 45.0, got %f", p.DefaultWeightGrams())
	}
}

func TestNewProduct_ValidationError(t *testing.T) {
	_, err := NewProduct(uuid.Nil, "Standard", nil, "PLA", 45, 1)
	if err == nil {
		t.Error("Expected error for empty parent ID, got nil")
	}
}
