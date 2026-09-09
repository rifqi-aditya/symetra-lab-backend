package shopee

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// ShopeeOrderListResponse merepresentasikan respon dari /api/v2/order/get_order_list
type ShopeeOrderListResponse struct {
	Error     string `json:"error"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
	Response  struct {
		More       bool   `json:"more"`
		NextCursor string `json:"next_cursor"`
		OrderList  []struct {
			OrderSN     string `json:"order_sn"`
			OrderStatus string `json:"order_status"`
		} `json:"order_list"`
	} `json:"response"`
}

// ShopeeOrderDetailResponse merepresentasikan respon dari /api/v2/order/get_order_detail
type ShopeeOrderDetailResponse struct {
	Error     string `json:"error"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
	Response  struct {
		OrderList []ShopeeOrderDetailItem `json:"order_list"`
	} `json:"response"`
}

type ShopeeOrderDetailItem struct {
	OrderSN           string           `json:"order_sn"`
	OrderStatus       string           `json:"order_status"`
	BuyerUserID       uint64           `json:"buyer_user_id"`
	BuyerUsername     string           `json:"buyer_username"`
	MessageToSeller   string           `json:"message_to_seller"`
	ShipByDate        int64            `json:"ship_by_date"`
	ShippingCarrier   string           `json:"shipping_carrier"`
	TotalAmount       float64          `json:"total_amount"`
	BuyerCancelReason string           `json:"buyer_cancel_reason"`
	CreateTime        int64            `json:"create_time"`
	UpdateTime        int64            `json:"update_time"`
	ItemList          []ShopeeItemInfo `json:"item_list"`
}

type ShopeeItemInfo struct {
	ItemID                 uint64  `json:"item_id"`
	ItemName               string  `json:"item_name"`
	ItemSKU                string  `json:"item_sku"`
	ModelID                uint64  `json:"model_id"`
	ModelName              string  `json:"model_name"`
	ModelSKU               string  `json:"model_sku"`
	ModelQuantityPurchased int     `json:"model_quantity_purchased"`
	ModelOriginalPrice     float64 `json:"model_original_price"`
	ModelDiscountedPrice   float64 `json:"model_discounted_price"`
}

// GetOrderList mengambil daftar order_sn dari Shopee
func (c *Client) GetOrderList(accessToken string, shopID uint64, timeFrom, timeTo int64, orderStatus string, cursor string) (*ShopeeOrderListResponse, error) {
	path := "/api/v2/order/get_order_list"
	timestamp := time.Now().Unix()
	sign := GenerateShopSign(c.PartnerID, path, timestamp, accessToken, shopID, c.PartnerKey)

	// Parameter query URL
	params := url.Values{}
	params.Set("partner_id", fmt.Sprintf("%d", c.PartnerID))
	params.Set("timestamp", fmt.Sprintf("%d", timestamp))
	params.Set("access_token", accessToken)
	params.Set("shop_id", fmt.Sprintf("%d", shopID))
	params.Set("sign", sign)
	params.Set("time_range_field", "create_time")
	params.Set("time_from", fmt.Sprintf("%d", timeFrom))
	params.Set("time_to", fmt.Sprintf("%d", timeTo))
	params.Set("page_size", "50")
	params.Set("response_optional_fields", "order_status")

	if orderStatus != "" {
		params.Set("order_status", orderStatus)
	}
	if cursor != "" {
		params.Set("cursor", cursor)
	}

	fullURL := fmt.Sprintf("%s%s?%s", c.BaseURL, path, params.Encode())

	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat HTTP request: %w", err)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gagal menghubungi Shopee API: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("gagal membaca respons Shopee: %w", err)
	}

	var result ShopeeOrderListResponse
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return nil, fmt.Errorf("gagal decode JSON order list: %w, raw: %s", err, string(bodyBytes))
	}

	if result.Error != "" {
		return nil, fmt.Errorf("shopee API error [%s]: %s (req_id: %s)", result.Error, result.Message, result.RequestID)
	}

	return &result, nil
}

// GetOrderDetail mengambil detail lengkap hingga 50 order_sn sekaligus
func (c *Client) GetOrderDetail(accessToken string, shopID uint64, orderSNList []string) (*ShopeeOrderDetailResponse, error) {
	if len(orderSNList) == 0 {
		return &ShopeeOrderDetailResponse{}, nil
	}

	path := "/api/v2/order/get_order_detail"
	timestamp := time.Now().Unix()
	sign := GenerateShopSign(c.PartnerID, path, timestamp, accessToken, shopID, c.PartnerKey)

	optionalFields := "buyer_user_id,buyer_username,item_list,message_to_seller,ship_by_date,shipping_carrier,total_amount,buyer_cancel_reason"

	params := url.Values{}
	params.Set("partner_id", fmt.Sprintf("%d", c.PartnerID))
	params.Set("timestamp", fmt.Sprintf("%d", timestamp))
	params.Set("access_token", accessToken)
	params.Set("shop_id", fmt.Sprintf("%d", shopID))
	params.Set("sign", sign)
	params.Set("order_sn_list", strings.Join(orderSNList, ","))
	params.Set("response_optional_fields", optionalFields)

	fullURL := fmt.Sprintf("%s%s?%s", c.BaseURL, path, params.Encode())

	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat HTTP request: %w", err)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gagal menghubungi Shopee API: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("gagal membaca respons Shopee: %w", err)
	}

	var result ShopeeOrderDetailResponse
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return nil, fmt.Errorf("gagal decode JSON order detail: %w, raw: %s", err, string(bodyBytes))
	}

	if result.Error != "" {
		return nil, fmt.Errorf("shopee API error [%s]: %s (req_id: %s)", result.Error, result.Message, result.RequestID)
	}

	return &result, nil
}
