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

// GetOrderDetail mengambil detail lengkap hingga 50 order_sn sekaligus dari Shopee (/api/v2/order/get_order_detail)
func (c *Client) GetOrderDetail(accessToken string, shopID uint64, orderSNList []string) (*ShopeeOrderDetailResponse, error) {
	if len(orderSNList) == 0 {
		return &ShopeeOrderDetailResponse{}, nil
	}

	path := "/api/v2/order/get_order_detail"
	timestamp := time.Now().Unix()
	sign := GenerateShopSign(c.PartnerID, path, timestamp, accessToken, shopID, c.PartnerKey)

	optionalFields := "buyer_user_id,buyer_username,item_list,message_to_seller,ship_by_date,shipping_carrier,checkout_shipping_carrier,package_list,actual_shipping_carrier,total_amount,buyer_cancel_reason"

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

// CallShopeeAPIRaw memanggil endpoint Shopee apapun dengan method GET dan mengembalikan raw body bytes
func (c *Client) CallShopeeAPIRaw(path string, accessToken string, shopID uint64, extraParams map[string]string) ([]byte, error) {
	timestamp := time.Now().Unix()
	sign := GenerateShopSign(c.PartnerID, path, timestamp, accessToken, shopID, c.PartnerKey)

	params := url.Values{}
	params.Set("partner_id", fmt.Sprintf("%d", c.PartnerID))
	params.Set("timestamp", fmt.Sprintf("%d", timestamp))
	params.Set("access_token", accessToken)
	params.Set("shop_id", fmt.Sprintf("%d", shopID))
	params.Set("sign", sign)

	for k, v := range extraParams {
		params.Set(k, v)
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

	return io.ReadAll(resp.Body)
}
