package handler_test

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"

	"symetra-lab-backend-v2/config"
	"symetra-lab-backend-v2/internal/delivery/http/handler"
)

func TestShopeeHandler_HandlePushWebhook_SecuritySignature(t *testing.T) {
	partnerKey := "secret_partner_key_12345"
	cfg := &config.Config{
		ShopeePartnerKey: partnerKey,
	}

	h := handler.NewShopeeHandler(cfg, nil)

	t.Run("should reject when signature header is missing", func(t *testing.T) {
		e := echo.New()
		body := `{"code": 3, "shop_id": 12345, "data": {"ordersn": "ORDER123"}}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/shopee/webhook", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := h.HandlePushWebhook(c)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected status 401 Unauthorized, got %d (body: %s)", rec.Code, rec.Body.String())
		}
	})

	t.Run("should reject when signature is invalid / tampered", func(t *testing.T) {
		e := echo.New()
		body := `{"code": 3, "shop_id": 12345, "data": {"ordersn": "ORDER123"}}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/shopee/webhook", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "invalid_signature_hash")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := h.HandlePushWebhook(c)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected status 401 Unauthorized, got %d (body: %s)", rec.Code, rec.Body.String())
		}
	})

	t.Run("should accept and proceed when signature is valid", func(t *testing.T) {
		e := echo.New()
		body := `{"code": 999, "shop_id": 12345, "data": {}}`
		req := httptest.NewRequest(http.MethodPost, "http://example.com/api/v1/shopee/webhook", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		fullURL := "http://example.com/api/v1/shopee/webhook"
		baseString := fullURL + "|" + body
		mac := hmac.New(sha256.New, []byte(partnerKey))
		mac.Write([]byte(baseString))
		validSig := hex.EncodeToString(mac.Sum(nil))

		req.Header.Set("Authorization", validSig)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := h.HandlePushWebhook(c)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200 OK, got %d (body: %s)", rec.Code, rec.Body.String())
		}
	})
}

func TestShopeeHandler_HandleCallback_Redirect(t *testing.T) {
	cfg := &config.Config{
		FrontendURL: "https://symetralab.vercel.app",
	}
	h := handler.NewShopeeHandler(cfg, nil)

	t.Run("should redirect to frontend with error when Shopee returns error", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/shopee/callback?error=user_cancel&msg=User+denied+access", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := h.HandleCallback(c)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if rec.Code != http.StatusTemporaryRedirect {
			t.Fatalf("expected 307 redirect, got %d", rec.Code)
		}

		loc := rec.Header().Get("Location")
		if !strings.HasPrefix(loc, "https://symetralab.vercel.app/marketplace") {
			t.Fatalf("expected redirect to frontend /marketplace, got: %s", loc)
		}
		if !strings.Contains(loc, "status=error") || !strings.Contains(loc, "User+denied+access") {
			t.Fatalf("expected error query params in redirect location, got: %s", loc)
		}
	})

	t.Run("should redirect to frontend with error when state is invalid (CSRF check)", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/shopee/callback?code=abc&shop_id=123&state=fake_invalid_state", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := h.HandleCallback(c)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if rec.Code != http.StatusTemporaryRedirect {
			t.Fatalf("expected 307 redirect, got %d", rec.Code)
		}

		loc := rec.Header().Get("Location")
		if !strings.Contains(loc, "status=error") {
			t.Fatalf("expected CSRF error redirect, got: %s", loc)
		}
	})

	t.Run("should redirect to frontend with error when code or shop_id is missing", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/shopee/callback", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := h.HandleCallback(c)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if rec.Code != http.StatusTemporaryRedirect {
			t.Fatalf("expected 307 redirect, got %d", rec.Code)
		}

		loc := rec.Header().Get("Location")
		if !strings.Contains(loc, "status=error") {
			t.Fatalf("expected error redirect, got: %s", loc)
		}
	})
}
