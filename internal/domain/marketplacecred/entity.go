package marketplacecred

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type Platform string

const (
	PlatformShopee    Platform = "SHOPEE"
	PlatformTokopedia Platform = "TOKOPEDIA"
	PlatformTikTok    Platform = "TIKTOK"
)

// MarketplaceCredential menyimpan data otorisasi partner level untuk masing-masing marketplace
type MarketplaceCredential struct {
	id           uuid.UUID
	platform     Platform
	partnerID    string
	partnerKey   string // Dalam domain entity ini berisi plaintext; di layer repository dienkripsi AES-256
	pushKey      string // Khusus Shopee Push Mechanism (dienkripsi AES-256 di layer repo)
	redirectURL  string
	isProduction bool
	isActive     bool
	createdAt    time.Time
	updatedAt    time.Time
}

func NewMarketplaceCredential(
	platform Platform,
	partnerID, partnerKey, redirectURL string,
	isProduction bool,
) (*MarketplaceCredential, error) {
	platformNorm := Platform(strings.ToUpper(strings.TrimSpace(string(platform))))
	if platformNorm == "" {
		return nil, ErrPlatformRequired
	}
	partnerIDTrim := strings.TrimSpace(partnerID)
	if partnerIDTrim == "" {
		return nil, ErrPartnerIDRequired
	}
	partnerKeyTrim := strings.TrimSpace(partnerKey)
	if partnerKeyTrim == "" {
		return nil, ErrPartnerKeyRequired
	}

	now := time.Now()
	return &MarketplaceCredential{
		id:           uuid.New(),
		platform:     platformNorm,
		partnerID:    partnerIDTrim,
		partnerKey:   partnerKeyTrim,
		redirectURL:  strings.TrimSpace(redirectURL),
		isProduction: isProduction,
		isActive:     true,
		createdAt:    now,
		updatedAt:    now,
	}, nil
}

func ReconstructMarketplaceCredential(
	id uuid.UUID,
	platform Platform,
	partnerID, partnerKey, pushKey, redirectURL string,
	isProduction, isActive bool,
	createdAt, updatedAt time.Time,
) *MarketplaceCredential {
	return &MarketplaceCredential{
		id:           id,
		platform:     platform,
		partnerID:    partnerID,
		partnerKey:   partnerKey,
		pushKey:      pushKey,
		redirectURL:  redirectURL,
		isProduction: isProduction,
		isActive:     isActive,
		createdAt:    createdAt,
		updatedAt:    updatedAt,
	}
}

func (m *MarketplaceCredential) ID() uuid.UUID       { return m.id }
func (m *MarketplaceCredential) Platform() Platform  { return m.platform }
func (m *MarketplaceCredential) PartnerID() string   { return m.partnerID }
func (m *MarketplaceCredential) PartnerKey() string  { return m.partnerKey }
func (m *MarketplaceCredential) PushKey() string     { return m.pushKey }
func (m *MarketplaceCredential) RedirectURL() string { return m.redirectURL }
func (m *MarketplaceCredential) IsProduction() bool  { return m.isProduction }
func (m *MarketplaceCredential) IsActive() bool      { return m.isActive }
func (m *MarketplaceCredential) CreatedAt() time.Time{ return m.createdAt }
func (m *MarketplaceCredential) UpdatedAt() time.Time{ return m.updatedAt }

func (m *MarketplaceCredential) SetPushKey(key string) {
	m.pushKey = strings.TrimSpace(key)
	m.updatedAt = time.Now()
}

// MaskedPartnerKey mengembalikan key yang disamarkan untuk tampilan aman di UI / response DTO
func (m *MarketplaceCredential) MaskedPartnerKey() string {
	if len(m.partnerKey) <= 8 {
		return "********"
	}
	prefix := m.partnerKey[:4]
	suffix := m.partnerKey[len(m.partnerKey)-4:]
	return prefix + "..." + suffix
}

// MaskedPushKey mengembalikan PushKey yang disamarkan untuk tampilan aman di UI / response DTO
func (m *MarketplaceCredential) MaskedPushKey() string {
	if m.pushKey == "" {
		return ""
	}
	if len(m.pushKey) <= 8 {
		return "********"
	}
	prefix := m.pushKey[:4]
	suffix := m.pushKey[len(m.pushKey)-4:]
	return prefix + "..." + suffix
}

func (m *MarketplaceCredential) Update(
	partnerID, partnerKey, redirectURL *string,
	isProduction, isActive *bool,
) {
	if partnerID != nil && strings.TrimSpace(*partnerID) != "" {
		m.partnerID = strings.TrimSpace(*partnerID)
	}
	if partnerKey != nil && strings.TrimSpace(*partnerKey) != "" {
		m.partnerKey = strings.TrimSpace(*partnerKey)
	}
	if redirectURL != nil {
		m.redirectURL = strings.TrimSpace(*redirectURL)
	}
	if isProduction != nil {
		m.isProduction = *isProduction
	}
	if isActive != nil {
		m.isActive = *isActive
	}
	m.updatedAt = time.Now()
}

