package dto

import (
	"time"

	"github.com/google/uuid"

	"symetra-lab-backend-v2/internal/domain/costing"
	"symetra-lab-backend-v2/internal/domain/product"
)

type ProductComponentInput struct {
	ComponentID   *string `json:"component_id"`
	Quantity      float64 `json:"quantity"`
	MarkupPercent float64 `json:"markup_percent"`
}

type ProductPackagingInput struct {
	PackagingItemID string  `json:"packaging_item_id"`
	QuantityUsed    float64 `json:"quantity_used"`
}

type CreateProductRequest struct {
	Name                  string                  `json:"name"`
	ParentSKU             *string                 `json:"parent_sku"`
	SKU                   *string                 `json:"sku"`
	Description           *string                 `json:"description"`
	Category              string                  `json:"category"`
	ThumbnailURL          *string                 `json:"thumbnail_url"`
	DesignLink            *string                 `json:"design_link"`
	DefaultWeightGrams    float64                 `json:"default_weight_grams"`
	MaterialType          string                  `json:"material_type"`
	DefaultPrintTimeHours float64                 `json:"default_print_time_hours"`
	DefaultMachineID      *string                 `json:"default_machine_id"`
	PackagingPresetID     *string                 `json:"packaging_preset_id"`
	BatchSize             int                     `json:"batch_size"`
	PackingFeeIDR         int                     `json:"packing_fee_idr"`
	TargetMarginPercent   int                     `json:"target_margin_percent"`
	Components            []ProductComponentInput `json:"components"`
	PackagingItems        []ProductPackagingInput `json:"packaging_items"`
}

type UpdateProductRequest struct {
	Name                  *string                  `json:"name"`
	ParentSKU             *string                  `json:"parent_sku"`
	SKU                   *string                  `json:"sku"`
	Description           *string                  `json:"description"`
	Category              *string                  `json:"category"`
	ThumbnailURL          *string                  `json:"thumbnail_url"`
	DesignLink            *string                  `json:"design_link"`
	DefaultWeightGrams    *float64                 `json:"default_weight_grams"`
	MaterialType          *string                  `json:"material_type"`
	DefaultPrintTimeHours *float64                 `json:"default_print_time_hours"`
	DefaultMachineID      *string                  `json:"default_machine_id"`
	PackagingPresetID     *string                  `json:"packaging_preset_id"`
	BatchSize             *int                     `json:"batch_size"`
	PackingFeeIDR         *int                     `json:"packing_fee_idr"`
	TargetMarginPercent   *int                     `json:"target_margin_percent"`
	Components            *[]ProductComponentInput `json:"components"`
	PackagingItems        *[]ProductPackagingInput `json:"packaging_items"`
}

type ProductComponentResponse struct {
	ID            uuid.UUID  `json:"id"`
	ProductID     uuid.UUID  `json:"product_id"`
	ComponentID   *uuid.UUID `json:"component_id"`
	Quantity      float64    `json:"quantity"`
	MarkupPercent float64    `json:"markup_percent"`
	CreatedAt     time.Time  `json:"created_at"`
}

type ProductPackagingItemResponse struct {
	ID              uuid.UUID `json:"id"`
	ProductID       uuid.UUID `json:"product_id"`
	PackagingItemID uuid.UUID `json:"packaging_item_id"`
	QuantityUsed    float64   `json:"quantity_used"`
	CreatedAt       time.Time `json:"created_at"`
}

type ProductResponse struct {
	ID                    uuid.UUID                      `json:"id"`
	UserID                uuid.UUID                      `json:"user_id"`
	Name                  string                         `json:"name"`
	ParentSKU             *string                        `json:"parent_sku"`
	SKU                   *string                        `json:"sku"`
	Description           *string                        `json:"description"`
	Category              string                         `json:"category"`
	ThumbnailURL          *string                        `json:"thumbnail_url"`
	DesignLink            *string                        `json:"design_link"`
	DefaultWeightGrams    float64                        `json:"default_weight_grams"`
	MaterialType          string                         `json:"material_type"`
	DefaultPrintTimeHours float64                        `json:"default_print_time_hours"`
	DefaultMachineID      *uuid.UUID                     `json:"default_machine_id"`
	PackagingPresetID     *uuid.UUID                     `json:"packaging_preset_id"`
	BatchSize             int                            `json:"batch_size"`
	PackingFeeIDR         int                            `json:"packing_fee_idr"`
	BaseHPP               float64                        `json:"base_hpp"`
	BaseSellingPrice      float64                        `json:"base_selling_price"`
	TargetMarginPercent   int                            `json:"target_margin_percent"`
	TimesOrdered          int                            `json:"times_ordered"`
	CreatedAt             time.Time                      `json:"created_at"`
	UpdatedAt             time.Time                      `json:"updated_at"`
	Components            []ProductComponentResponse     `json:"components,omitempty"`
	PackagingItems        []ProductPackagingItemResponse `json:"packaging_items,omitempty"`
	CostBreakdown         costing.ProductCostBreakdown   `json:"cost_breakdown"`
	TotalSold             int                            `json:"total_sold"`
}

func ToProductResponse(p *product.Product, totalSold int, cb costing.ProductCostBreakdown) ProductResponse {
	comps := make([]ProductComponentResponse, len(p.Components()))
	for i, c := range p.Components() {
		comps[i] = ProductComponentResponse{
			ID:            c.ID(),
			ProductID:     c.ProductID(),
			ComponentID:   c.ComponentID(),
			Quantity:      c.Quantity(),
			MarkupPercent: c.MarkupPercent(),
			CreatedAt:     c.CreatedAt(),
		}
	}

	packs := make([]ProductPackagingItemResponse, len(p.PackagingItems()))
	for i, it := range p.PackagingItems() {
		packs[i] = ProductPackagingItemResponse{
			ID:              it.ID(),
			ProductID:       it.ProductID(),
			PackagingItemID: it.PackagingItemID(),
			QuantityUsed:    it.QuantityUsed(),
			CreatedAt:       it.CreatedAt(),
		}
	}

	return ProductResponse{
		ID:                    p.ID(),
		UserID:                p.UserID(),
		Name:                  p.Name(),
		ParentSKU:             p.ParentSKU(),
		SKU:                   p.SKU(),
		Description:           p.Description(),
		Category:              p.Category(),
		ThumbnailURL:          p.ThumbnailURL(),
		DesignLink:            p.DesignLink(),
		DefaultWeightGrams:    p.DefaultWeightGrams(),
		MaterialType:          p.MaterialType(),
		DefaultPrintTimeHours: p.DefaultPrintTimeHours(),
		DefaultMachineID:      p.DefaultMachineID(),
		PackagingPresetID:     p.PackagingPresetID(),
		BatchSize:             p.BatchSize(),
		PackingFeeIDR:         p.PackingFeeIDR(),
		BaseHPP:               p.BaseHPP(),
		BaseSellingPrice:      p.BaseSellingPrice(),
		TargetMarginPercent:   p.TargetMarginPercent(),
		TimesOrdered:          p.TimesOrdered(),
		CreatedAt:             p.CreatedAt(),
		UpdatedAt:             p.UpdatedAt(),
		Components:            comps,
		PackagingItems:        packs,
		CostBreakdown:         cb,
		TotalSold:             totalSold,
	}
}

type ProductListItemResponse struct {
	ID                    uuid.UUID `json:"id"`
	Name                  string    `json:"name"`
	ParentSKU             *string   `json:"parent_sku"`
	SKU                   *string   `json:"sku"`
	Category              string    `json:"category"`
	ThumbnailURL          *string   `json:"thumbnail_url"`
	MaterialType          string    `json:"material_type"`
	DefaultWeightGrams    float64   `json:"default_weight_grams"`
	DefaultPrintTimeHours float64   `json:"default_print_time_hours"`
	BatchSize             int       `json:"batch_size"`
	BaseHPP               float64   `json:"base_hpp"`
	BaseSellingPrice      float64   `json:"base_selling_price"`
	TargetMarginPercent   int       `json:"target_margin_percent"`
	TotalSold             int       `json:"total_sold"`
	CreatedAt             time.Time `json:"created_at"`
}

func ToProductListItemResponse(p *product.Product, totalSold int) ProductListItemResponse {
	return ProductListItemResponse{
		ID:                    p.ID(),
		Name:                  p.Name(),
		ParentSKU:             p.ParentSKU(),
		SKU:                   p.SKU(),
		Category:              p.Category(),
		ThumbnailURL:          p.ThumbnailURL(),
		MaterialType:          p.MaterialType(),
		DefaultWeightGrams:    p.DefaultWeightGrams(),
		DefaultPrintTimeHours: p.DefaultPrintTimeHours(),
		BatchSize:             p.BatchSize(),
		BaseHPP:               p.BaseHPP(),
		BaseSellingPrice:      p.BaseSellingPrice(),
		TargetMarginPercent:   p.TargetMarginPercent(),
		TotalSold:             totalSold,
		CreatedAt:             p.CreatedAt(),
	}
}
