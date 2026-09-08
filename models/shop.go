package models

import "time"

type Shop struct {
	ID                    uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	ShopID                uint64    `gorm:"uniqueIndex;notnull" json:"shop_id"`
	ShopName              string    `gorm:"size:255" json:"shop_name"`
	Region                string    `gorm:"size:10" json:"region"`
	AccessToken           string    `gorm:"type:text;not null" json:"-"`
	RefreshToken          string    `gorm:"type:text;not null" json:"-"`
	AccessTokenExpiresAt  time.Time `json:"access_token_expires_at"`
	RefreshTokenExpiresAt time.Time `json:"refresh_token_expires_at"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

func (s *Shop) IsTokenExpired() bool {
	return time.Now().Add(5 * time.Minute).After(s.AccessTokenExpiresAt)
}
