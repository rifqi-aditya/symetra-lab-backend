package shoptokopedia

import (
	"time"
)

// ShopTokopediaStore represents an authorized TikTok Shop / Shop | Tokopedia merchant account.
type ShopTokopediaStore struct {
	id                    uint64
	openID                string
	sellerName            string
	region                string // "ID"
	accessToken           string
	refreshToken          string
	accessTokenExpiresAt  time.Time
	refreshTokenExpiresAt time.Time
	createdAt             time.Time
	updatedAt             time.Time
}

func ReconstructStore(
	id uint64,
	openID, sellerName, region, accessToken, refreshToken string,
	accessExpires, refreshExpires, createdAt, updatedAt time.Time,
) *ShopTokopediaStore {
	return &ShopTokopediaStore{
		id:                    id,
		openID:                openID,
		sellerName:            sellerName,
		region:                region,
		accessToken:           accessToken,
		refreshToken:          refreshToken,
		accessTokenExpiresAt:  accessExpires,
		refreshTokenExpiresAt: refreshExpires,
		createdAt:             createdAt,
		updatedAt:             updatedAt,
	}
}

func (s *ShopTokopediaStore) ID() uint64                    { return s.id }
func (s *ShopTokopediaStore) OpenID() string                { return s.openID }
func (s *ShopTokopediaStore) SellerName() string            { return s.sellerName }
func (s *ShopTokopediaStore) Region() string                { return s.region }
func (s *ShopTokopediaStore) AccessToken() string           { return s.accessToken }
func (s *ShopTokopediaStore) RefreshToken() string          { return s.refreshToken }
func (s *ShopTokopediaStore) AccessTokenExpiresAt() time.Time  { return s.accessTokenExpiresAt }
func (s *ShopTokopediaStore) RefreshTokenExpiresAt() time.Time { return s.refreshTokenExpiresAt }
func (s *ShopTokopediaStore) CreatedAt() time.Time          { return s.createdAt }
func (s *ShopTokopediaStore) UpdatedAt() time.Time          { return s.updatedAt }

func (s *ShopTokopediaStore) IsTokenExpired() bool {
	return time.Now().Add(5 * time.Minute).After(s.accessTokenExpiresAt)
}

func (s *ShopTokopediaStore) UpdateTokens(access, refresh string, accessExpires, refreshExpires time.Time) {
	s.accessToken = access
	s.refreshToken = refresh
	s.accessTokenExpiresAt = accessExpires
	s.refreshTokenExpiresAt = refreshExpires
	s.updatedAt = time.Now()
}

