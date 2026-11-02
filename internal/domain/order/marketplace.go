package order

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

// Supported marketplace channels
const (
	ChannelShopee    = "SHOPEE"
	ChannelTokopedia = "TOKOPEDIA"
	ChannelTikTok    = "TIKTOK"
	ChannelManual    = "MANUAL"
)

// StandardMarketplaceItem represents a line item from any marketplace platform.
type StandardMarketplaceItem struct {
	ChannelItemID   uint64
	ChannelModelID  uint64
	ItemName        string
	ItemSKU         string
	ModelName       string
	ModelSKU        string
	Quantity        int
	OriginalPrice   float64
	DiscountedPrice float64
	ProductID       *uuid.UUID
	MatchedSKU      string
	MappingStatus   string // MATCHED, UNMAPPED
}

// CandidateSKU returns the best candidate SKU for product matching (ModelSKU preferred over ItemSKU).
func (i *StandardMarketplaceItem) CandidateSKU() string {
	s := strings.TrimSpace(i.ModelSKU)
	if s != "" {
		return s
	}
	return strings.TrimSpace(i.ItemSKU)
}

// EffectivePrice returns the discounted unit price if positive, else the original unit price.
func (i *StandardMarketplaceItem) EffectivePrice() float64 {
	if i.DiscountedPrice > 0 {
		return i.DiscountedPrice
	}
	return i.OriginalPrice
}

// StandardMarketplaceFinancials represents escrow and fee breakdowns from any marketplace.
type StandardMarketplaceFinancials struct {
	GrossAmount              float64 // Total store merchandise price
	CommissionFee            float64
	ServiceFee               float64
	TransactionFee           float64
	OrderProcessingFee       float64
	SellerVoucherDiscount    float64
	TotalMarketplaceFee      float64 // Sum of all platform cuts
	NetAmount                float64 // Net settlement amount deposited to seller wallet
	FinancialStatus          string  // PENDING, RELEASED, REFUNDED
}

// StandardMarketplaceOrder is the universal marketplace order representation.
// All channel-specific adapters (Shopee, ShopTokopedia, etc.) normalize their payloads to this structure.
type StandardMarketplaceOrder struct {
	Channel           string // SHOPEE, TOKOPEDIA, TIKTOK
	ShopID            uint64
	OrderSN           string // Platform Order Serial / Order ID
	OrderStatus       string // UNPAID, READY_TO_SHIP, SHIPPED, COMPLETED, CANCELLED, REFUNDED
	BuyerUserID       uint64
	BuyerUsername     string
	MessageToSeller   string
	ShippingCarrier   string
	TrackingNumber    string
	ShipByDate        int64
	ShipByDateTime    *time.Time
	TotalCustomerPaid float64
	BuyerCancelReason string
	CreateTime        int64
	UpdateTime        int64
	PayTime           int64
	PaidAt            *time.Time
	Items             []StandardMarketplaceItem
	Financials        StandardMarketplaceFinancials
}

// NormalizeStatus translates channel-specific order status strings into a canonical internal status.
func NormalizeMarketplaceStatus(channel, rawStatus string) string {
	raw := strings.ToUpper(strings.TrimSpace(rawStatus))
	switch strings.ToUpper(channel) {
	case ChannelShopee:
		switch raw {
		case "UNPAID":
			return "UNPAID"
		case "READY_TO_SHIP", "PROCESSED":
			return "TO_SHIP"
		case "SHIPPED":
			return "SHIPPED"
		case "COMPLETED", "TO_CONFIRM_RECEIVE":
			return "COMPLETED"
		case "CANCELLED", "IN_CANCEL":
			return "CANCELLED"
		default:
			return raw
		}
	case ChannelTokopedia, ChannelTikTok:
		// TikTok Shop / Shop Tokopedia Partner API Statuses:
		// 100: UNPAID, 105: ON_HOLD, 111: AWAITING_SHIPMENT, 112: AWAITING_COLLECTION,
		// 114: PARTIALLY_SHIPPING, 121: IN_TRANSIT, 122: DELIVERED, 130: COMPLETED, 140: CANCELLED
		switch raw {
		case "100", "UNPAID":
			return "UNPAID"
		case "111", "112", "AWAITING_SHIPMENT", "AWAITING_COLLECTION":
			return "TO_SHIP"
		case "114", "121", "PARTIALLY_SHIPPING", "IN_TRANSIT":
			return "SHIPPED"
		case "122", "130", "DELIVERED", "COMPLETED":
			return "COMPLETED"
		case "140", "CANCELLED":
			return "CANCELLED"
		default:
			return raw
		}
	default:
		return raw
	}
}

// CalculateStoreSubtotal calculates the gross merchandise value of all items from the seller's perspective.
func (o *StandardMarketplaceOrder) CalculateStoreSubtotal() float64 {
	var total float64
	for _, it := range o.Items {
		total += it.EffectivePrice() * float64(it.Quantity)
	}
	return total
}

