package order_test

import (
	"testing"

	"symetra-lab-backend-v2/internal/domain/order"
)

func TestNormalizeMarketplaceStatus(t *testing.T) {
	tests := []struct {
		channel  string
		raw      string
		expected string
	}{
		{order.ChannelShopee, "UNPAID", "PENDING"},
		{order.ChannelShopee, "READY_TO_SHIP", "PROCESSED"},
		{order.ChannelShopee, "PROCESSED", "PROCESSED"},
		{order.ChannelShopee, "SHIPPED", "SHIPPED"},
		{order.ChannelShopee, "COMPLETED", "COMPLETED"},
		{order.ChannelShopee, "TO_CONFIRM_RECEIVE", "COMPLETED"},
		{order.ChannelShopee, "CANCELLED", "CANCELLED"},
		{order.ChannelShopee, "IN_CANCEL", "CANCELLED"},
		{order.ChannelShopee, "UNKNOWN", "UNKNOWN"},
		{order.ChannelManual, "PROCESSED", "PROCESSED"},
	}

	for _, tt := range tests {
		got := order.NormalizeMarketplaceStatus(tt.channel, tt.raw)
		if got != tt.expected {
			t.Errorf("NormalizeMarketplaceStatus(%s, %s) = %s; want %s", tt.channel, tt.raw, got, tt.expected)
		}
	}
}
