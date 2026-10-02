package product

import (
	"testing"

	"github.com/google/uuid"
)

func TestNewProduct_Success(t *testing.T) {
	userID := uuid.New()
	sku := "SKU-MOUSE-001"
	p, err := NewProduct(
		userID,
		"Test Ergonomic Mouse Grip",
		&sku,
		"Accessories",
		45.0,
		1.5,
		"PLA",
	)

	if err != nil {
		t.Fatalf("Expected nil err, got: %v", err)
	}
	if p.ID() == uuid.Nil {
		t.Error("Expected valid product UUID")
	}
	if p.Name() != "Test Ergonomic Mouse Grip" {
		t.Errorf("Expected name match, got %s", p.Name())
	}
	if p.DefaultWeightGrams() != 45.0 {
		t.Errorf("Expected weight 45.0, got %f", p.DefaultWeightGrams())
	}
}

func TestNewProduct_ValidationError(t *testing.T) {
	userID := uuid.New()
	_, err := NewProduct(userID, "", nil, "Accessories", 45, 1, "PLA")
	if err == nil {
		t.Error("Expected error for empty name, got nil")
	}
}
