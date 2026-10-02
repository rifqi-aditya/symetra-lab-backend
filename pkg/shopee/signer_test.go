package shopee

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func TestGeneratePublicSign(t *testing.T) {
	partnerID := int64(123456)
	path := "/api/v2/auth/token/get"
	timestamp := int64(1600000000)
	partnerKey := "secret-key"

	expectedBase := "123456/api/v2/auth/token/get1600000000"

	h := hmac.New(sha256.New, []byte(partnerKey))
	h.Write([]byte(expectedBase))
	expectedSign := hex.EncodeToString(h.Sum(nil))

	actualSign := GeneratePublicSign(partnerID, path, timestamp, partnerKey)

	if actualSign != expectedSign {
		t.Errorf("Expected %s, but got %s", expectedSign, actualSign)
	}
}

func TestGenerateShopSign(t *testing.T) {
	partnerID := int64(123456)
	path := "/api/v2/order/get"
	timestamp := int64(1600000000)
	accessToken := "access-token"
	shopID := uint64(789012)
	partnerKey := "secret-key"

	expectedBase := "123456/api/v2/order/get1600000000access-token789012"

	h := hmac.New(sha256.New, []byte(partnerKey))
	h.Write([]byte(expectedBase))
	expectedSign := hex.EncodeToString(h.Sum(nil))

	actualSign := GenerateShopSign(partnerID, path, timestamp, accessToken, shopID, partnerKey)

	if actualSign != expectedSign {
		t.Errorf("Expected %s, but got %s", expectedSign, actualSign)
	}
}
