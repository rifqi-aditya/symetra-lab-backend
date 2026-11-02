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

	// Candidate base string formats for HMAC verification
	candidates := []string{
		fullCallbackURL + bodyStr,          // 1. Format resmi Shopee v2: URL + body (tanpa pipe)
		fullCallbackURL + "|" + bodyStr,    // 2. Varian Shopee v1 / docs legacy: URL + "|" + body
		cleanURL + bodyStr,                 // 3. Normalisasi URL tanpa trailing slash + body
		cleanURL + "/" + bodyStr,           // 4. Normalisasi URL dengan trailing slash + body
		cleanURL + "|" + bodyStr,           // 5. Normalisasi URL + "|" + body
		cleanURL + "/|" + bodyStr,          // 6. Normalisasi URL dengan trailing slash + "|" + body
		bodyStr,                            // 7. Body saja
	}

	// Jaga-jaga jika URL terdaftar di console menggunakan http bukan https (atau sebaliknya)
	if strings.HasPrefix(fullCallbackURL, "https://") {
		httpURL := strings.Replace(fullCallbackURL, "https://", "http://", 1)
		cleanHTTP := strings.TrimRight(httpURL, "/")
		candidates = append(candidates,
			httpURL+bodyStr,
			httpURL+"|"+bodyStr,
			cleanHTTP+bodyStr,
			cleanHTTP+"/"+bodyStr,
			cleanHTTP+"|"+bodyStr,
		)
	} else if strings.HasPrefix(fullCallbackURL, "http://") {
		httpsURL := strings.Replace(fullCallbackURL, "http://", "https://", 1)
		cleanHTTPS := strings.TrimRight(httpsURL, "/")
		candidates = append(candidates,
			httpsURL+bodyStr,
			httpsURL+"|"+bodyStr,
			cleanHTTPS+bodyStr,
			cleanHTTPS+"/"+bodyStr,
			cleanHTTPS+"|"+bodyStr,
		)
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
