package order_test

import (
	"testing"

	"symetra-lab-backend-v2/internal/domain/order"
)

func TestStandardMarketplaceItem_CandidateSKU(t *testing.T) {
	item1 := order.StandardMarketplaceItem{
		ItemSKU:  "PARENT-SKU",
		ModelSKU: "CHILD-SKU-RED",
	}
	if item1.CandidateSKU() != "CHILD-SKU-RED" {
		t.Errorf("expected CHILD-SKU-RED, got %s", item1.CandidateSKU())
	}

	item2 := order.StandardMarketplaceItem{
		ItemSKU:  "PARENT-SKU",
		ModelSKU: "",
	}
	if item2.CandidateSKU() != "PARENT-SKU" {
		t.Errorf("expected PARENT-SKU, got %s", item2.CandidateSKU())
	}
}

func TestNormalizeMarketplaceStatus(t *testing.T) {
	tests := []struct {
		channel  string
		raw      string
		expected string
	}{
		{order.ChannelShopee, "READY_TO_SHIP", "TO_SHIP"},
		{order.ChannelShopee, "TO_CONFIRM_RECEIVE", "COMPLETED"},
		{order.ChannelShopee, "CANCELLED", "CANCELLED"},
		{order.ChannelTikTok, "100", "UNPAID"},
		{order.ChannelTikTok, "111", "TO_SHIP"},
		{order.ChannelTokopedia, "121", "SHIPPED"},
		{order.ChannelTokopedia, "130", "COMPLETED"},
		{order.ChannelTikTok, "140", "CANCELLED"},
	}

	for _, tt := range tests {
		got := order.NormalizeMarketplaceStatus(tt.channel, tt.raw)
		if got != tt.expected {
			t.Errorf("NormalizeMarketplaceStatus(%s, %s) = %s; want %s", tt.channel, tt.raw, got, tt.expected)
		}
	}
}

func TestStandardMarketplaceOrder_CalculateStoreSubtotal(t *testing.T) {
	ord := order.StandardMarketplaceOrder{
		Items: []order.StandardMarketplaceItem{
			{Quantity: 2, DiscountedPrice: 15000, OriginalPrice: 20000},
			{Quantity: 1, DiscountedPrice: 0, OriginalPrice: 50000},
		},
	}

	expected := (2 * 15000.0) + (1 * 50000.0)
	if ord.CalculateStoreSubtotal() != expected {
		t.Errorf("expected %f, got %f", expected, ord.CalculateStoreSubtotal())
	}
}

