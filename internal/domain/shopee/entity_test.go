package shopee

import (
	"testing"
	"time"
)

func TestShopeeOrder_Reconstruct(t *testing.T) {
	now := time.Now()
	order := ReconstructOrder(
		"240929ABCDEF12",
		1234567,
		"PROCESSED",
		8888,
		"buyer_symetra",
		"Please pack safely",
		now.Unix(),
		&now,
		"J&T Express",
		"JT123456789",
		150000,
		"",
		now.Unix(),
		now.Unix(),
		nil,
		nil,
		now,
		now,
	)

	if order.OrderSN() != "240929ABCDEF12" {
		t.Errorf("Expected OrderSN match, got %s", order.OrderSN())
	}
	if order.OrderStatus() != "PROCESSED" {
		t.Errorf("Expected OrderStatus PROCESSED, got %s", order.OrderStatus())
	}
	if order.TotalAmount() != 150000 {
		t.Errorf("Expected TotalAmount 150000, got %f", order.TotalAmount())
	}
}
