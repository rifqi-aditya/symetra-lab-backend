package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"symetra-lab-backend-v2/config"
	"symetra-lab-backend-v2/internal/delivery/http/dto"
	"symetra-lab-backend-v2/internal/delivery/http/handler"
	domainCred "symetra-lab-backend-v2/internal/domain/marketplacecred"
	credUC "symetra-lab-backend-v2/internal/usecase/marketplacecred"
)

type mockCredRepo struct {
	creds map[string]*domainCred.MarketplaceCredential
}

func newMockCredRepo() *mockCredRepo {
	return &mockCredRepo{creds: make(map[string]*domainCred.MarketplaceCredential)}
}

func (m *mockCredRepo) FindByPlatform(ctx context.Context, platform domainCred.Platform) (*domainCred.MarketplaceCredential, error) {
	c, ok := m.creds[string(platform)]
	if !ok {
		return nil, domainCred.ErrCredentialNotFound
	}
	return c, nil
}

func (m *mockCredRepo) FindAll(ctx context.Context) ([]*domainCred.MarketplaceCredential, error) {
	var list []*domainCred.MarketplaceCredential
	for _, c := range m.creds {
		list = append(list, c)
	}
	return list, nil
}

func (m *mockCredRepo) Save(ctx context.Context, cred *domainCred.MarketplaceCredential) error {
	m.creds[string(cred.Platform())] = cred
	return nil
}

func (m *mockCredRepo) Delete(ctx context.Context, id uuid.UUID) error {
	for k, v := range m.creds {
		if v.ID() == id {
			delete(m.creds, k)
			return nil
		}
	}
	return domainCred.ErrCredentialNotFound
}

func TestMarketplaceCredHandler(t *testing.T) {
	repo := newMockCredRepo()
	cfg := &config.Config{}
	uc := credUC.NewUseCase(repo, cfg)
	h := handler.NewMarketplaceCredHandler(uc)

	t.Run("should successfully save credential and return masked key", func(t *testing.T) {
		e := echo.New()
		payload := dto.SaveMarketplaceCredentialRequest{
			Platform:     "SHOPEE",
			PartnerID:    "2044497",
			PartnerKey:   "shpk676b516461596a455257534a557973446969466f44455268686f6d727756",
			RedirectURL:  "https://symetra.com/callback",
			IsProduction: true,
		}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/config/marketplace-credentials", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := h.SaveCredential(c)
		if err != nil {
			t.Fatalf("unexpected handler error: %v", err)
		}

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d (body: %s)", rec.Code, rec.Body.String())
		}

		var resp struct {
			Success bool                                `json:"success"`
			Data    dto.MarketplaceCredentialResponse  `json:"data"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if resp.Data.Platform != "SHOPEE" {
			t.Fatalf("expected platform SHOPEE, got %s", resp.Data.Platform)
		}
		if resp.Data.PartnerID != "2044497" {
			t.Fatalf("expected partner ID 2044497, got %s", resp.Data.PartnerID)
		}
		// Pastikan partner key asli TIDAK bocor di respons HTTP
		if resp.Data.MaskedPartnerKey == payload.PartnerKey {
			t.Fatalf("partner key must NOT be returned in plaintext")
		}
		if resp.Data.MaskedPartnerKey != "shpk...7756" {
			t.Fatalf("expected masked key shpk...7756, got: %s", resp.Data.MaskedPartnerKey)
		}
	})

	t.Run("should list credentials", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/config/marketplace-credentials", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := h.ListCredentials(c)
		if err != nil {
			t.Fatalf("unexpected handler error: %v", err)
		}

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}
	})
}

