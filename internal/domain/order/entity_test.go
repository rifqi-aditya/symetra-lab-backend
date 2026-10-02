package order

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNewOrder_Success(t *testing.T) {
	userID := uuid.New()
	items := []OrderItem{
		ReconstructOrderItem(
			uuid.New(), uuid.Nil, nil, "Custom Phone Stand",
			2, 35000, 15000, 30, 1.0, nil, 0, 0, 0, 0, time.Now(),
		),
	}

	ord, err := NewOrder(
		userID,
		"Budi Santoso",
		"08123456789",
		"Harap packing bubble wrap tebal",
		"DIRECT_WHATSAPP",
		"UNPAID",
		items,
	)

	if err != nil {
		t.Fatalf("Expected nil err, got: %v", err)
	}
	if ord.TotalRevenue() != 70000 {
		t.Errorf("Expected total revenue 70000, got %f", ord.TotalRevenue())
	}
	if ord.TotalHPP() != 30000 {
		t.Errorf("Expected total HPP 30000, got %f", ord.TotalHPP())
	}
	if ord.TotalProfit() != 40000 {
		t.Errorf("Expected total profit 40000, got %f", ord.TotalProfit())
	}
	if ord.Status() != "PENDING" {
		t.Errorf("Expected status PENDING, got %s", ord.Status())
	}
}

func TestNewOrder_NoItemsError(t *testing.T) {
	userID := uuid.New()
	_, err := NewOrder(userID, "Budi", "08123", "", "MANUAL", "UNPAID", []OrderItem{})
	if err == nil {
		t.Error("Expected error when creating order with empty items, got nil")
	}
}
