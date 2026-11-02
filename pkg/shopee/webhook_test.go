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

	// 1. Test Format v1 dengan pipe: url + "|" + body
	baseWithPipe := url + "|" + string(body)
	h1 := hmac.New(sha256.New, []byte(partnerKey))
	h1.Write([]byte(baseWithPipe))
	signWithPipe := hex.EncodeToString(h1.Sum(nil))

	if !shopee.VerifyWebhookSignature(url, body, signWithPipe, partnerKey) {
		t.Errorf("expected signature with pipe to be valid, but got false")
	}

	// 2. Test Format v2 resmi tanpa pipe: url + body
	baseDirect := url + string(body)
	h2 := hmac.New(sha256.New, []byte(partnerKey))
	h2.Write([]byte(baseDirect))
	signDirect := hex.EncodeToString(h2.Sum(nil))

	if !shopee.VerifyWebhookSignature(url, body, signDirect, partnerKey) {
		t.Errorf("expected signature direct (no pipe) to be valid, but got false")
	}

	// 3. Test invalid signature
	if shopee.VerifyWebhookSignature(url, body, "invalid_sign_hex", partnerKey) {
		t.Errorf("expected signature to fail with invalid signature header")
	}

	// 4. Test invalid partner key
	if shopee.VerifyWebhookSignature(url, body, signDirect, "wrong_partner_key") {
		t.Errorf("expected signature to fail with wrong partner key")
	}
}
