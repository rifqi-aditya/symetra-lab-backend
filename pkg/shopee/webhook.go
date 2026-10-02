package shopee

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
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
// Shopee menghitung HMAC-SHA256 dari string: fullCallbackURL + "|" + rawRequestBody dengan key: PartnerKey
func VerifyWebhookSignature(fullCallbackURL string, rawBody []byte, signatureHeader string, partnerKey string) bool {
	if signatureHeader == "" || partnerKey == "" {
		return false
	}

	baseString := fullCallbackURL + "|" + string(rawBody)
	h := hmac.New(sha256.New, []byte(partnerKey))
	h.Write([]byte(baseString))
	expectedSign := hex.EncodeToString(h.Sum(nil))

	return subtle.ConstantTimeCompare([]byte(signatureHeader), []byte(expectedSign)) == 1
}
