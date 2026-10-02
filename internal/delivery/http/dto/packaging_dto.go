package dto

import (
	"math"
	"time"

	"github.com/google/uuid"

	"symetra-lab-backend-v2/internal/domain/packaging"
)

type CreatePackagingItemRequest struct {
	Name             string  `json:"name"`
	Category         string  `json:"category"`
	UnitType         string  `json:"unit_type"`
	PurchasePrice    float64 `json:"purchase_price"`
	PurchaseQuantity float64 `json:"purchase_quantity"`
	StockQuantity    float64 `json:"stock_quantity"`
}

type UpdatePackagingItemRequest struct {
	Name             *string  `json:"name"`
	Category         *string  `json:"category"`
	UnitType         *string  `json:"unit_type"`
	PurchasePrice    *float64 `json:"purchase_price"`
	PurchaseQuantity *float64 `json:"purchase_quantity"`
	StockQuantity    *float64 `json:"stock_quantity"`
}

type PackagingItemResponse struct {
	ID               uuid.UUID `json:"id"`
	UserID           uuid.UUID `json:"user_id"`
	Name             string    `json:"name"`
	Category         string    `json:"category"`
	UnitType         string    `json:"unit_type"`
	PurchasePrice    float64   `json:"purchase_price"`
	PurchaseQuantity float64   `json:"purchase_quantity"`
	UnitCost         float64   `json:"unit_cost"`
	StockQuantity    float64   `json:"stock_quantity"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

func ToPackagingItemResponse(p *packaging.PackagingItem) PackagingItemResponse {
	return PackagingItemResponse{
		ID:               p.ID(),
		UserID:           p.UserID(),
		Name:             p.Name(),
		Category:         p.Category(),
		UnitType:         p.UnitType(),
		PurchasePrice:    p.PurchasePrice(),
		PurchaseQuantity: p.PurchaseQuantity(),
		UnitCost:         p.UnitCost(),
		StockQuantity:    p.StockQuantity(),
		CreatedAt:        p.CreatedAt(),
		UpdatedAt:        p.UpdatedAt(),
	}
}

type PresetItemRequest struct {
	PackagingItemID string  `json:"packaging_item_id"`
	QuantityUsed    float64 `json:"quantity_used"`
}

type CreatePresetRequest struct {
	Name        string              `json:"name"`
	Description *string             `json:"description"`
	Items       []PresetItemRequest `json:"items"`
}

type UpdatePresetRequest struct {
	Name        string              `json:"name"`
	Description *string             `json:"description"`
	Items       []PresetItemRequest `json:"items"`
}

type PackagingPresetItemInfo struct {
	ID              uuid.UUID `json:"id"`
	PackagingItemID uuid.UUID `json:"packaging_item_id"`
	ItemName        string    `json:"item_name"`
	Category        string    `json:"category"`
	UnitType        string    `json:"unit_type"`
	QuantityUsed    float64   `json:"quantity_used"`
	UnitCost        float64   `json:"unit_cost"`
	SubtotalCost    float64   `json:"subtotal_cost"`
}

type PackagingPresetResponse struct {
	ID          uuid.UUID                 `json:"id"`
	UserID      uuid.UUID                 `json:"user_id"`
	Name        string                    `json:"name"`
	Description *string                   `json:"description"`
	TotalCost   float64                   `json:"total_cost"`
	Items       []PackagingPresetItemInfo `json:"items"`
	CreatedAt   time.Time                 `json:"created_at"`
	UpdatedAt   time.Time                 `json:"updated_at"`
}

func ToPackagingPresetResponse(p *packaging.PackagingPreset) PackagingPresetResponse {
	var totalCost float64
	items := make([]PackagingPresetItemInfo, 0, len(p.Items()))

	for _, it := range p.Items() {
		info := PackagingPresetItemInfo{
			ID:              it.ID(),
			PackagingItemID: it.PackagingItemID(),
			QuantityUsed:    it.QuantityUsed(),
		}

		if it.Item() != nil {
			info.ItemName = it.Item().Name()
			info.Category = it.Item().Category()
			info.UnitType = it.Item().UnitType()
			info.UnitCost = it.Item().UnitCost()
			info.SubtotalCost = math.Round(it.QuantityUsed()*it.Item().UnitCost()*100) / 100
			totalCost += info.SubtotalCost
		}

		items = append(items, info)
	}

	return PackagingPresetResponse{
		ID:          p.ID(),
		UserID:      p.UserID(),
		Name:        p.Name(),
		Description: p.Description(),
		TotalCost:   math.Round(totalCost*100) / 100,
		Items:       items,
		CreatedAt:   p.CreatedAt(),
		UpdatedAt:   p.UpdatedAt(),
	}
}