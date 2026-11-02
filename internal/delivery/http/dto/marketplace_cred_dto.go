package dto

import (
	"time"

	"github.com/google/uuid"

	domainCred "symetra-lab-backend-v2/internal/domain/marketplacecred"
)

type SaveMarketplaceCredentialRequest struct {
	Platform     string `json:"platform"`
	PartnerID    string `json:"partner_id"`
	PartnerKey   string `json:"partner_key"`
	RedirectURL  string `json:"redirect_url"`
	IsProduction bool   `json:"is_production"`
}

type MarketplaceCredentialResponse struct {
	ID               uuid.UUID `json:"id"`
	Platform         string    `json:"platform"`
	PartnerID        string    `json:"partner_id"`
	MaskedPartnerKey string    `json:"masked_partner_key"` // PartnerKey asli tidak pernah diekspos ke client
	RedirectURL      string    `json:"redirect_url"`
	IsProduction     bool      `json:"is_production"`
	IsActive         bool      `json:"is_active"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

func ToMarketplaceCredentialResponse(m *domainCred.MarketplaceCredential) MarketplaceCredentialResponse {
	return MarketplaceCredentialResponse{
		ID:               m.ID(),
		Platform:         string(m.Platform()),
		PartnerID:        m.PartnerID(),
		MaskedPartnerKey: m.MaskedPartnerKey(),
		RedirectURL:      m.RedirectURL(),
		IsProduction:     m.IsProduction(),
		IsActive:         m.IsActive(),
		CreatedAt:        m.CreatedAt(),
		UpdatedAt:        m.UpdatedAt(),
	}
}

