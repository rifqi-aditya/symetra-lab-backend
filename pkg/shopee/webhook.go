package shopee

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"strings"
)

// WebhookPushPayload represents the webhook payload pushed by Shopee Open Platform.
type WebhookPushPayload struct {
	ShopID    int64             `json:"shop_id"`
	Code      int               `json:"code"`
	Timestamp int64             `json:"timestamp"`
	Data      WebhookPushDetail `json:"data"`
}

// WebhookPushDetail holds event-specific fields in the push payload.
type WebhookPushDetail struct {
	OrderSN    string `json:"ordersn"`
	OrderSNAlt string `json:"order_sn"`
	Status     string `json:"status"`
	UpdateTime int64  `json:"update_time"`
}

func (d *WebhookPushDetail) GetOrderSN() string {
	if d.OrderSN != "" {
		return d.OrderSN
	}
	return d.OrderSNAlt
}

// VerifyWebhookSignature memverifikasi keaslian webhook Shopee menggunakan HMAC-SHA256
// Mendukung format Shopee v2 resmi (URL + body) maupun varian v1/legacy (URL + "|" + body)
func VerifyWebhookSignature(fullCallbackURL string, rawBody []byte, signatureHeader string, partnerKey string) bool {
	if signatureHeader == "" || partnerKey == "" {
		return false
	}

	sig := strings.TrimSpace(signatureHeader)
	sig = strings.TrimPrefix(sig, "SHA256 ")
	sig = strings.TrimPrefix(sig, "sha256 ")

	bodyStr := string(rawBody)
	cleanURL := strings.TrimRight(fullCallbackURL, "/")

	candidates := []string{
		fullCallbackURL + bodyStr,       // 1. Format resmi Shopee v2: URL + body (tanpa pipe)
		fullCallbackURL + "|" + bodyStr, // 2. Varian Shopee v1 / docs legacy: URL + "|" + body
		cleanURL + bodyStr,              // 3. Normalisasi URL tanpa trailing slash + body
		cleanURL + "/" + bodyStr,        // 4. Normalisasi URL dengan trailing slash + body
		cleanURL + "|" + bodyStr,        // 5. Normalisasi URL + "|" + body
		bodyStr,                         // 6. Body saja
	}

	keyBytes := []byte(partnerKey)
	for _, base := range candidates {
		h := hmac.New(sha256.New, keyBytes)
		h.Write([]byte(base))
		expectedSign := hex.EncodeToString(h.Sum(nil))

		if subtle.ConstantTimeCompare([]byte(strings.ToLower(sig)), []byte(strings.ToLower(expectedSign))) == 1 {
			return true
		}
	}

	return false
}
