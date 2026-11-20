package order

import (
	"strings"
)

// Supported marketplace channels
const (
	ChannelShopee = "SHOPEE"
	ChannelManual = "MANUAL"
)

// NormalizeMarketplaceStatus translates channel-specific order status strings into the 4 canonical internal statuses:
// PENDING, PROCESSED, SHIPPED, COMPLETED (or CANCELLED).
func NormalizeMarketplaceStatus(channel, rawStatus string) string {
	raw := strings.ToUpper(strings.TrimSpace(rawStatus))
	switch strings.ToUpper(channel) {
	case ChannelShopee:
		switch raw {
		case "UNPAID":
			return "PENDING"
		case "READY_TO_SHIP", "PROCESSED":
			return "PROCESSED"
		case "SHIPPED":
			return "SHIPPED"
		case "COMPLETED", "TO_CONFIRM_RECEIVE":
			return "COMPLETED"
		case "CANCELLED", "IN_CANCEL":
			return "CANCELLED"
		default:
			return raw
		}
	default:
		return raw
	}
}
