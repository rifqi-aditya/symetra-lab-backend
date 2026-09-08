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

// TokenResponseBody memetakan data respon sukses dari Shopee
type TokenResponseBody struct {
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
	ExpireIn     int      `json:"expire_in"` // Durasi dalam detik (biasanya 14400 / 4 jam)
	ShopIDList   []uint64 `json:"shop_id_list"`
}

// ShopeeResponse membungkus struktur JSON standar dari Shopee Open Platform
type ShopeeResponse struct {
	Error     string             `json:"error"`
	Message   string             `json:"message"`
	Response  *TokenResponseBody `json:"response,omitempty"`
	RequestID string             `json:"request_id"`
}

// BuildAuthURL membuat link otorisasi Shopee untuk diklik oleh pemilik toko
func (c *Client) BuildAuthURL() (string, error) {
	path := "/api/v2/shop/auth_partner"
	timestamp := time.Now().Unix()
	sign := GeneratePublicSign(c.PartnerID, path, timestamp, c.PartnerKey)

	u, err := url.Parse(c.BaseURL + path)
	if err != nil {
		return "", err
	}

	// Masukkan parameter query yang diwajibkan Shopee
	q := u.Query()
	q.Set("partner_id", fmt.Sprintf("%d", c.PartnerID))
	q.Set("timestamp", fmt.Sprintf("%d", timestamp))
	q.Set("sign", sign)
	q.Set("redirect", c.RedirectURL)
	u.RawQuery = q.Encode()

	return u.String(), nil
}

// GetAccessToken menukar authorization code menjadi access token & refresh token
func (c *Client) GetAccessToken(code string, shopID uint64) (*TokenResponseBody, error) {
	path := "/api/v2/auth/token/get"
	timestamp := time.Now().Unix()
	sign := GeneratePublicSign(c.PartnerID, path, timestamp, c.PartnerKey)

	endpoint := fmt.Sprintf("%s%s?partner_id=%d&timestamp=%d&sign=%s", c.BaseURL, path, c.PartnerID, timestamp, sign)

	payload := map[string]interface{}{
		"code":       code,
		"shop_id":    shopID,
		"partner_id": c.PartnerID,
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("gagal marshal request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("gagal membuat request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gagal request ke Shopee: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("gagal membaca response body: %w", err)
	}

	var shopeeResp ShopeeResponse
	if err := json.Unmarshal(respBytes, &shopeeResp); err != nil {
		return nil, fmt.Errorf("gagal parse JSON Shopee: %w (raw: %s)", err, string(respBytes))
	}

	if shopeeResp.Error != "" {
		return nil, fmt.Errorf("error dari Shopee [%s]: %s (request_id: %s)", shopeeResp.Error, shopeeResp.Message, shopeeResp.RequestID)
	}

	if shopeeResp.Response == nil {
		return nil, fmt.Errorf("data response Shopee kosong")
	}

	return shopeeResp.Response, nil
}

// RefreshAccessToken memperbarui access token menggunakan refresh token
func (c *Client) RefreshAccessToken(refreshToken string, shopID uint64) (*TokenResponseBody, error) {
	path := "/api/v2/auth/access_token/get"
	timestamp := time.Now().Unix()
	sign := GeneratePublicSign(c.PartnerID, path, timestamp, c.PartnerKey)

	endpoint := fmt.Sprintf("%s%s?partner_id=%d&timestamp=%d&sign=%s", c.BaseURL, path, c.PartnerID, timestamp, sign)

	payload := map[string]interface{}{
		"refresh_token": refreshToken,
		"shop_id":       shopID,
		"partner_id":    c.PartnerID,
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("gagal marshal request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("gagal membuat request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gagal request ke Shopee: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("gagal membaca response body: %w", err)
	}

	var shopeeResp ShopeeResponse
	if err := json.Unmarshal(respBytes, &shopeeResp); err != nil {
		return nil, fmt.Errorf("gagal parse JSON Shopee: %w", err)
	}

	if shopeeResp.Error != "" {
		return nil, fmt.Errorf("error dari Shopee [%s]: %s (request_id: %s)", shopeeResp.Error, shopeeResp.Message, shopeeResp.RequestID)
	}

	if shopeeResp.Response == nil {
		return nil, fmt.Errorf("data response Shopee kosong")
	}

	return shopeeResp.Response, nil
}
