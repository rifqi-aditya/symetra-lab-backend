package models

import (
	"math"
	"time"
)

// PackagingItem merepresentasikan bahan kemasan individual untuk pesanan/produk
// (contoh: kardus diecut, plastik clip hologram, mailer putih, sticker label)
type PackagingItem struct {
	ID               string    `gorm:"type:uuid;primaryKey" json:"id"`
	UserID           string    `gorm:"type:uuid;not null;index" json:"user_id"`
	Name             string    `gorm:"size:255;not null" json:"name"`
	Category         string    `gorm:"size:64;default:'OTHER';not null" json:"category"` // BOX, PLASTIC, LABEL, OTHER
	UnitType         string    `gorm:"size:32;default:'PCS';not null" json:"unit_type"`  // PCS, ROLL, METER
	PurchasePrice    float64   `gorm:"type:numeric;default:0;not null" json:"purchase_price"`
	PurchaseQuantity float64   `gorm:"type:numeric;default:1;not null" json:"purchase_quantity"`
	UnitCost         float64   `gorm:"type:numeric;default:0;not null" json:"unit_cost"`
	StockQuantity    float64   `gorm:"type:numeric;default:0;not null" json:"stock_quantity"`
	CreatedAt        time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (PackagingItem) TableName() string {
	return "packaging_items"
}

// CalculateUnitCost menghitung biaya pokok satuan (purchase_price / purchase_quantity)
func (p *PackagingItem) CalculateUnitCost() float64 {
	if p.PurchaseQuantity <= 0 {
		return p.PurchasePrice
	}
	raw := p.PurchasePrice / p.PurchaseQuantity
	return math.Round(raw*100) / 100
}

// PackagingPreset merepresentasikan bundel kemasan standar (misal: "Keychain", "Mini Diorama Part")
type PackagingPreset struct {
	ID          string    `gorm:"type:uuid;primaryKey" json:"id"`
	UserID      string    `gorm:"type:uuid;not null;index" json:"user_id"`
	Name        string    `gorm:"size:255;not null" json:"name"`
	Description *string   `gorm:"type:text" json:"description"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	// Relasi ke item komponen kemasan
	Items []PackagingPresetItem `gorm:"foreignKey:PresetID;references:ID;constraint:OnDelete:CASCADE" json:"items,omitempty"`
}

func (PackagingPreset) TableName() string {
	return "packaging_presets"
}

// PackagingPresetItem merepresentasikan satu komponen bahan kemasan dalam preset
type PackagingPresetItem struct {
	ID              string    `gorm:"type:uuid;primaryKey" json:"id"`
	PresetID        string    `gorm:"type:uuid;not null;index" json:"preset_id"`
	PackagingItemID string    `gorm:"type:uuid;not null;index" json:"packaging_item_id"`
	QuantityUsed    float64   `gorm:"type:numeric;default:1;not null" json:"quantity_used"`
	CreatedAt       time.Time `gorm:"autoCreateTime" json:"created_at"`

	// Relasi ke master data bahan kemasan
	PackagingItem *PackagingItem `gorm:"foreignKey:PackagingItemID;references:ID" json:"packaging_item,omitempty"`
}

func (PackagingPresetItem) TableName() string {
	return "packaging_preset_items"
}

// PackagingPresetResponse format respons preset kemasan lengkap dengan kalkulasi total biaya
type PackagingPresetResponse struct {
	ID          string                    `json:"id"`
	UserID      string                    `json:"user_id"`
	Name        string                    `json:"name"`
	Description *string                   `json:"description"`
	TotalCost   float64                   `json:"total_cost"`
	Items       []PackagingPresetItemInfo `json:"items"`
	CreatedAt   time.Time                 `json:"created_at"`
	UpdatedAt   time.Time                 `json:"updated_at"`
}

// PackagingPresetItemInfo rincian per item dalam preset
type PackagingPresetItemInfo struct {
	ID              string  `json:"id"`
	PackagingItemID string  `json:"packaging_item_id"`
	ItemName        string  `json:"item_name"`
	Category        string  `json:"category"`
	UnitType        string  `json:"unit_type"`
	QuantityUsed    float64 `json:"quantity_used"`
	UnitCost        float64 `json:"unit_cost"`
	SubtotalCost    float64 `json:"subtotal_cost"`
}

// ToResponse mengubah PackagingPreset menjadi respons terstruktur dengan total biaya terhitung
func (p *PackagingPreset) ToResponse() PackagingPresetResponse {
	var totalCost float64
	items := make([]PackagingPresetItemInfo, 0, len(p.Items))

	for _, it := range p.Items {
		info := PackagingPresetItemInfo{
			ID:              it.ID,
			PackagingItemID: it.PackagingItemID,
			QuantityUsed:    it.QuantityUsed,
		}

		if it.PackagingItem != nil {
			info.ItemName = it.PackagingItem.Name
			info.Category = it.PackagingItem.Category
			info.UnitType = it.PackagingItem.UnitType
			info.UnitCost = it.PackagingItem.UnitCost
			info.SubtotalCost = math.Round(it.QuantityUsed*it.PackagingItem.UnitCost*100) / 100
			totalCost += info.SubtotalCost
		}

		items = append(items, info)
	}

	return PackagingPresetResponse{
		ID:          p.ID,
		UserID:      p.UserID,
		Name:        p.Name,
		Description: p.Description,
		TotalCost:   math.Round(totalCost*100) / 100,
		Items:       items,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}
