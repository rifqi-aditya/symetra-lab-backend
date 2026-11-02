package shoptokopedia

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	domainOrder "symetra-lab-backend-v2/internal/domain/order"
)

type SearchOrdersRequest struct {
	PageSize  int    `json:"page_size"`
	PageToken string `json:"page_token,omitempty"`
}

type RawOrderSearchResponse struct {
	Code      int    `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
	Data      struct {
		NextPageToken string `json:"next_page_token"`
		Orders        []struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		} `json:"orders"`
		TotalCount int `json:"total_count"`
	} `json:"data"`
}

type RawOrderLineItem struct {
	ID                     string `json:"id"`
	ProductID              string `json:"product_id"`
	ProductName            string `json:"product_name"`
	SKUID                  string `json:"sku_id"`
	SellerSKU              string `json:"seller_sku"`
	Quantity               int    `json:"quantity"`
	OriginalPrice          string `json:"original_price"`
	SalePrice              string `json:"sale_price"`
	PlatformDiscount       string `json:"platform_discount"`
	SellerDiscount         string `json:"seller_discount"`
}

type RawOrderDetail struct {
	ID                   string             `json:"id"`
	Status               string             `json:"status"`
	BuyerEmail           string             `json:"buyer_email"`
	BuyerMessage         string             `json:"buyer_message"`
	RecipientAddress     struct {
		Name string `json:"name"`
	} `json:"recipient_address"`
	ShippingProvider     string             `json:"shipping_provider"`
	TrackingNumber       string             `json:"tracking_number"`
	CreateTime           int64              `json:"create_time"`
	UpdateTime           int64              `json:"update_time"`
	PaidTime             int64              `json:"paid_time"`
	LineItems            []RawOrderLineItem `json:"line_items"`
	Payment struct {
		Currency            string `json:"currency"`
		TotalAmount         string `json:"total_amount"`
		SubTotal            string `json:"sub_total"`
		ShippingFee         string `json:"shipping_fee"`
		PlatformDiscount    string `json:"platform_discount"`
		SellerDiscount      string `json:"seller_discount"`
	} `json:"payment"`
}

type RawOrderDetailResponse struct {
	Code      int              `json:"code"`
	Message   string           `json:"message"`
	RequestID string           `json:"request_id"`
	Data      struct {
		Orders []RawOrderDetail `json:"orders"`
	} `json:"data"`
}

// ConvertToStandardOrder transforms raw TikTok Shop / Tokopedia order into universal StandardMarketplaceOrder.
func ConvertToStandardOrder(raw RawOrderDetail, channel string, shopID uint64) domainOrder.StandardMarketplaceOrder {
	if channel == "" {
		channel = domainOrder.ChannelTokopedia
	}

	items := make([]domainOrder.StandardMarketplaceItem, len(raw.LineItems))
	for i, item := range raw.LineItems {
		origPrice, _ := strconv.ParseFloat(item.OriginalPrice, 64)
		salePrice, _ := strconv.ParseFloat(item.SalePrice, 64)
		itemID, _ := strconv.ParseUint(item.ProductID, 10, 64)
		modelID, _ := strconv.ParseUint(item.SKUID, 10, 64)

		items[i] = domainOrder.StandardMarketplaceItem{
			ChannelItemID:   itemID,
			ChannelModelID:  modelID,
			ItemName:        item.ProductName,
			ItemSKU:         item.SellerSKU,
			ModelSKU:        item.SellerSKU,
			Quantity:        item.Quantity,
			OriginalPrice:   origPrice,
			DiscountedPrice: salePrice,
			MappingStatus:   "UNMAPPED",
		}
	}

	subTotal, _ := strconv.ParseFloat(raw.Payment.SubTotal, 64)
	totalCustomer, _ := strconv.ParseFloat(raw.Payment.TotalAmount, 64)
	sellerDiscount, _ := strconv.ParseFloat(raw.Payment.SellerDiscount, 64)

	var paidAt *time.Time
	if raw.PaidTime > 0 {
		t := time.Unix(raw.PaidTime, 0)
		paidAt = &t
	}

	// Tokopedia / TikTok standard order representation
	return domainOrder.StandardMarketplaceOrder{
		Channel:           channel,
		ShopID:            shopID,
		OrderSN:           raw.ID,
		OrderStatus:       domainOrder.NormalizeMarketplaceStatus(channel, raw.Status),
		BuyerUsername:     raw.RecipientAddress.Name,
		MessageToSeller:   raw.BuyerMessage,
		ShippingCarrier:   raw.ShippingProvider,
		TrackingNumber:    raw.TrackingNumber,
		TotalCustomerPaid: totalCustomer,
		CreateTime:        raw.CreateTime,
		UpdateTime:        raw.UpdateTime,
		PayTime:           raw.PaidTime,
		PaidAt:            paidAt,
		Items:             items,
		Financials: domainOrder.StandardMarketplaceFinancials{
			GrossAmount:           subTotal,
			SellerVoucherDiscount: sellerDiscount,
			NetAmount:             subTotal - sellerDiscount, // preliminary before settlement sync
			FinancialStatus:       "PENDING",
		},
	}
}

// GetOrderDetails fetches full order details from Shop | Tokopedia Open API.
func (c *Client) GetOrderDetails(accessToken, shopCipher string, orderIDs []string) ([]RawOrderDetail, error) {
	if len(orderIDs) == 0 {
		return nil, nil
	}

	path := "/order/202309/orders"
	queryParams := map[string]string{
		"ids": fmt.Sprintf("%v", orderIDs),
	}

	resBytes, err := c.ExecuteSignedRequest("GET", path, accessToken, shopCipher, queryParams, nil)
	if err != nil {
		return nil, err
	}

	var resp RawOrderDetailResponse
	if err := json.Unmarshal(resBytes, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse order details: %w", err)
	}

	if resp.Code != 0 {
		return nil, fmt.Errorf("tiktok API error: %s (code: %d)", resp.Message, resp.Code)
	}

	return resp.Data.Orders, nil
}

