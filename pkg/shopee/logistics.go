package shopee

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// ShopeeShippingParameterResponse merepresentasikan respon dari /api/v2/logistics/get_shipping_parameter
type ShopeeShippingParameterResponse struct {
	Error     string `json:"error"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
	Response  struct {
		InfoNeeded struct {
			Dropoff []string `json:"dropoff"`
			Pickup  []string `json:"pickup"`
		} `json:"info_needed"`
		Dropoff struct {
			BranchList []struct {
				BranchID int64  `json:"branch_id"`
				Region   string `json:"region"`
				State    string `json:"state"`
				City     string `json:"city"`
				Address  string `json:"address"`
			} `json:"branch_list"`
		} `json:"dropoff"`
		Pickup struct {
			AddressList []struct {
				AddressID   int64    `json:"address_id"`
				Region      string   `json:"region"`
				State       string   `json:"state"`
				City        string   `json:"city"`
				Address     string   `json:"address"`
				TimeSlotList []struct {
					Date         int64  `json:"date"`
					PickupTimeID string `json:"pickup_time_id"`
				} `json:"time_slot_list"`
			} `json:"address_list"`
		} `json:"pickup"`
	} `json:"response"`
}

// ShopeeShipOrderRequest parameter body untuk /api/v2/logistics/ship_order
type ShopeeShipOrderRequest struct {
	OrderSN string                 `json:"order_sn"`
	Dropoff *ShopeeShipDropoff     `json:"dropoff,omitempty"`
	Pickup  *ShopeeShipPickup      `json:"pickup,omitempty"`
}

type ShopeeShipDropoff struct {
	BranchID   int64  `json:"branch_id,omitempty"`
	SenderRealName string `json:"sender_real_name,omitempty"`
}

type ShopeeShipPickup struct {
	AddressID    int64  `json:"address_id"`
	PickupTimeID string `json:"pickup_time_id,omitempty"`
}

// ShopeeShipOrderResponse respon dari /api/v2/logistics/ship_order
type ShopeeShipOrderResponse struct {
	Error     string `json:"error"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
}

// ShopeeTrackingNumberResponse respon dari /api/v2/logistics/get_tracking_number
type ShopeeTrackingNumberResponse struct {
	Error     string `json:"error"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
	Response  struct {
		TrackingNumber string `json:"tracking_number"`
	} `json:"response"`
}

// ShopeeCreateShippingDocumentResponse respon dari /api/v2/logistics/create_shipping_document
type ShopeeCreateShippingDocumentResponse struct {
	Error     string `json:"error"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
	Response  struct {
		ResultList []struct {
			OrderSN   string `json:"order_sn"`
			FailError string `json:"fail_error"`
			FailMsg   string `json:"fail_message"`
		} `json:"result_list"`
	} `json:"response"`
}

// GetShippingParameter memeriksa opsi logistik (dropoff vs pickup) untuk sebuah order
func (c *Client) GetShippingParameter(accessToken string, shopID uint64, orderSN string) (*ShopeeShippingParameterResponse, error) {
	path := "/api/v2/logistics/get_shipping_parameter"
	timestamp := time.Now().Unix()
	sign := GenerateShopSign(c.PartnerID, path, timestamp, accessToken, shopID, c.PartnerKey)

	params := url.Values{}
	params.Set("partner_id", fmt.Sprintf("%d", c.PartnerID))
	params.Set("timestamp", fmt.Sprintf("%d", timestamp))
	params.Set("access_token", accessToken)
	params.Set("shop_id", fmt.Sprintf("%d", shopID))
	params.Set("sign", sign)
	params.Set("order_sn", orderSN)

	fullURL := fmt.Sprintf("%s%s?%s", c.BaseURL, path, params.Encode())

	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat request shipping parameter: %w", err)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gagal menghubungi Shopee API: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("gagal membaca respons: %w", err)
	}

	var result ShopeeShippingParameterResponse
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return nil, fmt.Errorf("gagal decode JSON: %w", err)
	}

	if result.Error != "" {
		return nil, fmt.Errorf("shopee API error [%s]: %s", result.Error, result.Message)
	}

	return &result, nil
}

// ShipOrder mengonfirmasi kesiapan pengiriman paket ("Atur Pengiriman")
func (c *Client) ShipOrder(accessToken string, shopID uint64, shipReq ShopeeShipOrderRequest) (*ShopeeShipOrderResponse, error) {
	path := "/api/v2/logistics/ship_order"
	timestamp := time.Now().Unix()
	sign := GenerateShopSign(c.PartnerID, path, timestamp, accessToken, shopID, c.PartnerKey)

	params := url.Values{}
	params.Set("partner_id", fmt.Sprintf("%d", c.PartnerID))
	params.Set("timestamp", fmt.Sprintf("%d", timestamp))
	params.Set("access_token", accessToken)
	params.Set("shop_id", fmt.Sprintf("%d", shopID))
	params.Set("sign", sign)

	fullURL := fmt.Sprintf("%s%s?%s", c.BaseURL, path, params.Encode())

	reqBodyBytes, err := json.Marshal(shipReq)
	if err != nil {
		return nil, fmt.Errorf("gagal marshal request body: %w", err)
	}

	req, err := http.NewRequest("POST", fullURL, bytes.NewBuffer(reqBodyBytes))
	if err != nil {
		return nil, fmt.Errorf("gagal membuat request ship order: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gagal menghubungi Shopee API: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("gagal membaca respons: %w", err)
	}

	var result ShopeeShipOrderResponse
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return nil, fmt.Errorf("gagal decode JSON: %w", err)
	}

	if result.Error != "" {
		return nil, fmt.Errorf("shopee API error [%s]: %s", result.Error, result.Message)
	}

	return &result, nil
}

// GetTrackingNumber menarik nomor resi resmi pengiriman
func (c *Client) GetTrackingNumber(accessToken string, shopID uint64, orderSN string) (*ShopeeTrackingNumberResponse, error) {
	path := "/api/v2/logistics/get_tracking_number"
	timestamp := time.Now().Unix()
	sign := GenerateShopSign(c.PartnerID, path, timestamp, accessToken, shopID, c.PartnerKey)

	params := url.Values{}
	params.Set("partner_id", fmt.Sprintf("%d", c.PartnerID))
	params.Set("timestamp", fmt.Sprintf("%d", timestamp))
	params.Set("access_token", accessToken)
	params.Set("shop_id", fmt.Sprintf("%d", shopID))
	params.Set("sign", sign)
	params.Set("order_sn", orderSN)

	fullURL := fmt.Sprintf("%s%s?%s", c.BaseURL, path, params.Encode())

	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat request tracking number: %w", err)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gagal menghubungi Shopee API: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("gagal membaca respons: %w", err)
	}

	var result ShopeeTrackingNumberResponse
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return nil, fmt.Errorf("gagal decode JSON: %w", err)
	}

	if result.Error != "" {
		return nil, fmt.Errorf("shopee API error [%s]: %s", result.Error, result.Message)
	}

	return &result, nil
}

// CreateShippingDocument meminta pembuatan task dokumen label resi pengiriman
func (c *Client) CreateShippingDocument(accessToken string, shopID uint64, orderSN string) (*ShopeeCreateShippingDocumentResponse, error) {
	path := "/api/v2/logistics/create_shipping_document"
	timestamp := time.Now().Unix()
	sign := GenerateShopSign(c.PartnerID, path, timestamp, accessToken, shopID, c.PartnerKey)

	params := url.Values{}
	params.Set("partner_id", fmt.Sprintf("%d", c.PartnerID))
	params.Set("timestamp", fmt.Sprintf("%d", timestamp))
	params.Set("access_token", accessToken)
	params.Set("shop_id", fmt.Sprintf("%d", shopID))
	params.Set("sign", sign)

	fullURL := fmt.Sprintf("%s%s?%s", c.BaseURL, path, params.Encode())

	payload := map[string]interface{}{
		"order_list": []map[string]interface{}{
			{
				"order_sn":               orderSN,
				"shipping_document_type": "THERMAL_AIR_WAYBILL",
			},
		},
	}
	bodyBytes, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", fullURL, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("gagal membuat request create shipping document: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gagal menghubungi Shopee API: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("gagal membaca respons: %w", err)
	}

	var result ShopeeCreateShippingDocumentResponse
	if err := json.Unmarshal(respBytes, &result); err != nil {
		return nil, fmt.Errorf("gagal decode JSON: %w", err)
	}

	return &result, nil
}

// DownloadShippingDocument mendownload file biner PDF label pengiriman thermal 100x150 mm
func (c *Client) DownloadShippingDocument(accessToken string, shopID uint64, orderSN string) ([]byte, error) {
	path := "/api/v2/logistics/download_shipping_document"
	timestamp := time.Now().Unix()
	sign := GenerateShopSign(c.PartnerID, path, timestamp, accessToken, shopID, c.PartnerKey)

	params := url.Values{}
	params.Set("partner_id", fmt.Sprintf("%d", c.PartnerID))
	params.Set("timestamp", fmt.Sprintf("%d", timestamp))
	params.Set("access_token", accessToken)
	params.Set("shop_id", fmt.Sprintf("%d", shopID))
	params.Set("sign", sign)

	fullURL := fmt.Sprintf("%s%s?%s", c.BaseURL, path, params.Encode())

	payload := map[string]interface{}{
		"shipping_document_type": "THERMAL_AIR_WAYBILL",
		"order_list": []map[string]interface{}{
			{
				"order_sn": orderSN,
			},
		},
	}
	bodyBytes, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", fullURL, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("gagal membuat request download: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gagal menghubungi Shopee API: %w", err)
	}
	defer resp.Body.Close()

	pdfBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("gagal membaca file PDF: %w", err)
	}

	// Jika Shopee mengembalikan JSON error alih-alih file PDF
	if len(pdfBytes) > 0 && pdfBytes[0] == '{' {
		var errResp struct {
			Error   string `json:"error"`
			Message string `json:"message"`
		}
		if json.Unmarshal(pdfBytes, &errResp) == nil && errResp.Error != "" {
			return nil, fmt.Errorf("shopee API error [%s]: %s", errResp.Error, errResp.Message)
		}
	}

	return pdfBytes, nil
}
