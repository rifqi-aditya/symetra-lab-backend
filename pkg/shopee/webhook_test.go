package shopee_test

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"symetra-lab-backend-v2/pkg/shopee"
)

func TestVerifyWebhookSignature(t *testing.T) {
	partnerKey := "my_partner_key_12345"
	url := "https://api.symetralab.com/api/v1/shopee/webhook"
	body := []byte(`{"code":3,"shop_id":711996297,"timestamp":1660123127,"data":{"ordersn":"220810QSK8S7BX","status":"READY_TO_SHIP"}}`)

	// Generate expected signature
	base := url + "|" + string(body)
	h := hmac.New(sha256.New, []byte(partnerKey))
	h.Write([]byte(base))
	validSign := hex.EncodeToString(h.Sum(nil))

	if !shopee.VerifyWebhookSignature(url, body, validSign, partnerKey) {
		t.Errorf("expected signature to be valid, but got false")
	}

	// Test invalid signature
	if shopee.VerifyWebhookSignature(url, body, "invalid_sign_hex", partnerKey) {
		t.Errorf("expected signature to fail with invalid signature header")
	}

	// Test invalid partner key
	if shopee.VerifyWebhookSignature(url, body, validSign, "wrong_partner_key") {
		t.Errorf("expected signature to fail with wrong partner key")
	}
}
