package shopee

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// GetOrderList mengambil daftar order_sn dari Shopee Open Platform (/api/v2/order/get_order_list)
func (c *Client) GetOrderList(accessToken string, shopID uint64, timeFrom, timeTo int64, orderStatus string, cursor string) (*ShopeeOrderListResponse, error) {
	path := "/api/v2/order/get_order_list"
	timestamp := time.Now().Unix()
	sign := GenerateShopSign(c.PartnerID, path, timestamp, accessToken, shopID, c.PartnerKey)

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
