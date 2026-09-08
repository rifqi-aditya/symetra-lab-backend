package shopee

import (
	"net/http"
	"time"
)

const (
	ProductionBaseURL = "https://partner.shopeemobile.com"
	TestBaseURL       = "https://partner.test-stable.shopeemobile.com"
)

// Client menyimpan konfigurasi dan klien HTTP untuk berkomunikasi dengan Shopee Open Platform
type Client struct {
	PartnerID   int64
	PartnerKey  string
	BaseURL     string
	RedirectURL string
	HTTPClient  *http.Client
}

// NewClient membuat instance baru dari Client Shopee
func NewClient(partnerID int64, partnerKey string, isProduction bool, redirectURL string) *Client {
	baseURL := TestBaseURL
	if isProduction {
		baseURL = ProductionBaseURL
	}

	return &Client{
		PartnerID:   partnerID,
		PartnerKey:  partnerKey,
		BaseURL:     baseURL,
		RedirectURL: redirectURL,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}
