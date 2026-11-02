package shoptokopedia

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

const (
	ProductionBaseURL = "https://open-api.tiktokglobalshop.com"
	SandboxBaseURL    = "https://open-api.sandbox.tiktokglobalshop.com"
	AuthURL           = "https://services.tiktokshop.com/open/authorize"
)

// Client handles communication with the unified TikTok Shop / Shop | Tokopedia Open API.
type Client struct {
	AppKey      string
	AppSecret   string
	BaseURL     string
	AuthBaseURL string
	RedirectURL string
	HTTPClient  *http.Client
}

// NewClient creates a new Shop | Tokopedia API client.
func NewClient(appKey, appSecret string, isProduction bool, redirectURL string) *Client {
	baseURL := SandboxBaseURL
	if isProduction {
		baseURL = ProductionBaseURL
	}

	return &Client{
		AppKey:      appKey,
		AppSecret:   appSecret,
		BaseURL:     baseURL,
		AuthBaseURL: AuthURL,
		RedirectURL: redirectURL,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// BuildAuthURL generates the seller authorization URL for Shop | Tokopedia.
func (c *Client) BuildAuthURL(serviceID, state string) (string, error) {
	if c.AppKey == "" {
		return "", fmt.Errorf("app_key is required")
	}
	u, err := url.Parse(c.AuthBaseURL)
	if err != nil {
		return "", err
	}

	q := u.Query()
	q.Set("service_id", serviceID)
	q.Set("state", state)
	u.RawQuery = q.Encode()

	return u.String(), nil
}

// GenerateSignature calculates the HMAC-SHA256 signature for TikTok Shop Partner API.
// Rule:
// 1. Extract all query parameters except 'sign' and 'access_token'.
// 2. Sort parameter keys alphabetically.
// 3. Concatenate: path + key1 + val1 + key2 + val2 + ... + body.
// 4. Wrap with HMAC-SHA256 using app_secret as key and hex-encode.
func (c *Client) GenerateSignature(path string, queryParams map[string]string, body []byte) string {
	keys := make([]string, 0, len(queryParams))
	for k := range queryParams {
		if k == "sign" || k == "access_token" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var sb strings.Builder
	sb.WriteString(path)
	for _, k := range keys {
		sb.WriteString(k)
		sb.WriteString(queryParams[k])
	}
	if len(body) > 0 {
		sb.Write(body)
	}

	h := hmac.New(sha256.New, []byte(c.AppSecret))
	h.Write([]byte(sb.String()))
	return hex.EncodeToString(h.Sum(nil))
}

type TokenResponse struct {
	Code      int    `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
	Data      struct {
		AccessToken          string `json:"access_token"`
		AccessTokenExpireIn  int64  `json:"access_token_expire_in"`
		RefreshToken         string `json:"refresh_token"`
		RefreshTokenExpireIn int64  `json:"refresh_token_expire_in"`
		OpenID               string `json:"open_id"`
		SellerName           string `json:"seller_name"`
		SellerBaseRegion     string `json:"seller_base_region"`
		UserType             int    `json:"user_type"`
	} `json:"data"`
}

// GetAccessToken exchanges authorization code for access and refresh tokens.
func (c *Client) GetAccessToken(authCode string) (*TokenResponse, error) {
	endpoint := "/api/v2/token/get"
	u, _ := url.Parse(c.BaseURL + endpoint)
	q := u.Query()
	q.Set("app_key", c.AppKey)
	q.Set("app_secret", c.AppSecret)
	q.Set("auth_code", authCode)
	q.Set("grant_type", "authorized_code")
	u.RawQuery = q.Encode()

	resp, err := c.HTTPClient.Get(u.String())
	if err != nil {
		return nil, fmt.Errorf("failed to exchange token: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(bodyBytes, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	if tokenResp.Code != 0 {
		return nil, fmt.Errorf("tiktok API error: %s (code: %d)", tokenResp.Message, tokenResp.Code)
	}

	return &tokenResp, nil
}

// RefreshAccessToken refreshes an expired access token using refresh_token.
func (c *Client) RefreshAccessToken(refreshToken string) (*TokenResponse, error) {
	endpoint := "/api/v2/token/refresh"
	u, _ := url.Parse(c.BaseURL + endpoint)
	q := u.Query()
	q.Set("app_key", c.AppKey)
	q.Set("app_secret", c.AppSecret)
	q.Set("refresh_token", refreshToken)
	q.Set("grant_type", "refresh_token")
	u.RawQuery = q.Encode()

	resp, err := c.HTTPClient.Get(u.String())
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(bodyBytes, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse refresh response: %w", err)
	}

	if tokenResp.Code != 0 {
		return nil, fmt.Errorf("tiktok API error: %s (code: %d)", tokenResp.Message, tokenResp.Code)
	}

	return &tokenResp, nil
}

// ExecuteSignedRequest sends a signed request to TikTok Shop Open API.
func (c *Client) ExecuteSignedRequest(
	method, path string,
	accessToken string,
	shopCipher string,
	queryParams map[string]string,
	bodyPayload interface{},
) ([]byte, error) {
	if queryParams == nil {
		queryParams = make(map[string]string)
	}
	queryParams["app_key"] = c.AppKey
	queryParams["timestamp"] = fmt.Sprintf("%d", time.Now().Unix())
	if shopCipher != "" {
		queryParams["shop_cipher"] = shopCipher
	}

	var bodyBytes []byte
	if bodyPayload != nil {
		var err error
		bodyBytes, err = json.Marshal(bodyPayload)
		if err != nil {
			return nil, err
		}
	}

	sign := c.GenerateSignature(path, queryParams, bodyBytes)
	queryParams["sign"] = sign

	u, _ := url.Parse(c.BaseURL + path)
	q := u.Query()
	for k, v := range queryParams {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()

	var reqBody io.Reader
	if len(bodyBytes) > 0 {
		reqBody = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequest(method, u.String(), reqBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if accessToken != "" {
		req.Header.Set("x-tts-access-token", accessToken)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

