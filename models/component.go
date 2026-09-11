package models

import (
	"math"
	"time"
)

// Component merepresentasikan komponen fisik non-3D print yang dipasang pada produk bengkel
// (contoh: baut, mur, magnet, gantungan kunci, bearing, lampu LED, modul USB, dll)
type Component struct {
	ID                   string    `gorm:"type:uuid;primaryKey" json:"id"`
	UserID               string    `gorm:"type:uuid;not null;index" json:"user_id"`
	Name                 string    `gorm:"size:255;not null" json:"name"`
	PricePerUnit         float64   `gorm:"type:numeric;default:0;not null" json:"price_per_unit"`
	DefaultMarkupPercent float64   `gorm:"type:numeric;default:0;not null" json:"default_markup_percent"`
	Description          *string   `gorm:"type:text" json:"description"`
	CreatedAt            time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt            time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (Component) TableName() string {
	return "components"
}

// CalculatedSellingPrice menghitung estimasi harga jual per unit setelah markup
func (c *Component) CalculatedSellingPrice() float64 {
	markup := c.DefaultMarkupPercent
	if markup < 0 {
		markup = 0
	}
	raw := c.PricePerUnit * (1.0 + (markup / 100.0))
	return math.Round(raw*100) / 100
}

// ComponentResponse format payload response JSON yang diperkaya dengan harga jual yang disarankan
type ComponentResponse struct {
	ID                     string    `json:"id"`
	UserID                 string    `json:"user_id"`
	Name                   string    `json:"name"`
	PricePerUnit           float64   `json:"price_per_unit"`
	DefaultMarkupPercent   float64   `json:"default_markup_percent"`
	CalculatedSellingPrice float64   `json:"calculated_selling_price"`
	Description            *string   `json:"description"`
	CreatedAt              time.Time `json:"created_at"`
	UpdatedAt              time.Time `json:"updated_at"`
}

// ToResponse mengubah model Component menjadi ComponentResponse
func (c *Component) ToResponse() ComponentResponse {
	return ComponentResponse{
		ID:                     c.ID,
		UserID:                 c.UserID,
		Name:                   c.Name,
		PricePerUnit:           c.PricePerUnit,
		DefaultMarkupPercent:   c.DefaultMarkupPercent,
		CalculatedSellingPrice: c.CalculatedSellingPrice(),
		Description:            c.Description,
		CreatedAt:              c.CreatedAt,
		UpdatedAt:              c.UpdatedAt,
	}
}
