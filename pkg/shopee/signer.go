package shopee

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

func GeneratePublicSign(partnerID int64, path string, timestamp int64, partnerKey string) string {
	baseString := fmt.Sprintf("%d%s%d", partnerID, path, timestamp)
	return calculateHmac(baseString, partnerKey)
}

func GenerateShopSign(partnerID int64, path string, timestamp int64, accessToken string, shopID uint64, partnerKey string) string {
	baseString := fmt.Sprintf("%d%s%d%s%d", partnerID, path, timestamp, accessToken, shopID)
	return calculateHmac(baseString, partnerKey)
}

func calculateHmac(message string, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(message))
	return hex.EncodeToString(h.Sum(nil))
}
