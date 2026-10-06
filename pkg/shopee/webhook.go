package shopee

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"strings"
)

// WebhookPushPayload merepresentasikan root payload push notification dari Shopee Open Platform
type WebhookPushPayload struct {
	Code      int                `json:"code"`
	ShopID    uint64             `json:"shop_id"`
	Timestamp int64              `json:"timestamp"`
	Data      OrderStatusPushData `json:"data"`
}

// OrderStatusPushData merepresentasikan detail saat Code == 3 (order_status_push)
type OrderStatusPushData struct {
	OrderSN           string        `json:"ordersn"`
	Status            string        `json:"status"`
	CompletedScenario string        `json:"completed_scenario,omitempty"`
	UpdateTime        int64         `json:"update_time"`
	Items             []interface{} `json:"items,omitempty"`
}

// VerifyWebhookSignature memverifikasi keaslian webhook Shopee menggunakan HMAC-SHA256
func VerifyWebhookSignature(fullCallbackURL string, rawBody []byte, signatureHeader string, partnerKey string) bool {
	if signatureHeader == "" || partnerKey == "" {
		return false
	}

	sig := strings.TrimSpace(signatureHeader)
	sig = strings.TrimPrefix(sig, "SHA256 ")
	sig = strings.TrimPrefix(sig, "sha256 ")

	// 1. Format standar Shopee: fullCallbackURL + "|" + rawBody
	baseString := fullCallbackURL + "|" + string(rawBody)
	h := hmac.New(sha256.New, []byte(partnerKey))
	h.Write([]byte(baseString))
	expectedSign := hex.EncodeToString(h.Sum(nil))

	if subtle.ConstantTimeCompare([]byte(strings.ToLower(sig)), []byte(strings.ToLower(expectedSign))) == 1 {
		return true
	}

	// 2. Format alternatif: rawBody saja
	h2 := hmac.New(sha256.New, []byte(partnerKey))
	h2.Write(rawBody)
	expectedSign2 := hex.EncodeToString(h2.Sum(nil))

	return subtle.ConstantTimeCompare([]byte(strings.ToLower(sig)), []byte(strings.ToLower(expectedSign2))) == 1
}
