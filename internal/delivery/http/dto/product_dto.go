package dto

import (
	"math"
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
	ParentID              *string                 `json:"parent_id,omitempty"`
	VariantName           string                  `json:"variant_name,omitempty"`
	Name                  string                  `json:"name"`
	ParentSKU             *string                 `json:"parent_sku"`
	SKU                   *string                 `json:"sku"`
	Description           *string                 `json:"description"`
	Category              string                  `json:"category"`
	CategoryID            *string                 `json:"category_id,omitempty"`
	ThumbnailURL          *string                 `json:"thumbnail_url"`
	DesignLink            *string                 `json:"design_link"`

	// =========================================================================
	// Data Bobot & Waktu Cetak (Unit Level vs Slicer Build Plate Level)
	// =========================================================================
	// Database selalu menyimpan data dalam satuan per-unit (1 pcs):
	// - DefaultWeightGrams: Berat per-unit (1 pcs).
	// - DefaultPrintTimeHours: Waktu cetak per-unit (1 pcs).
	//
	// Opsi Virtual Field (Slicer Raw Input):
	// - PlateWeightGrams: Total berat 1 plate slicer (misal 100 gram).
	// - PlatePrintTimeHours: Total durasi 1 plate slicer (misal 5 jam).
	// Jika Plate* dikirim, Backend otomatis membaginya dengan batch_size sehingga
	// Frontend tidak perlu repot melakukan pembagian manual di UI.
	DefaultWeightGrams    float64                 `json:"default_weight_grams"`
	PlateWeightGrams      *float64                `json:"plate_weight_grams,omitempty"`
	MaterialType          string                  `json:"material_type"`
	DefaultPrintTimeHours float64                 `json:"default_print_time_hours"`
	PlatePrintTimeHours   *float64                `json:"plate_print_time_hours,omitempty"`

	PackagingPresetID     *string                 `json:"packaging_preset_id,omitempty"`
	BatchSize             int                     `json:"batch_size"`
	PackingFeeIDR         int                     `json:"packing_fee_idr,omitempty"`
	TargetMarginPercent   int                     `json:"target_margin_percent"`
	Components            []ProductComponentInput `json:"components"`
	PackagingItems        []ProductPackagingInput `json:"packaging_items"`
}

type UpdateProductRequest struct {
	VariantName           *string                  `json:"variant_name,omitempty"`
	Name                  *string                  `json:"name"`
	ParentSKU             *string                  `json:"parent_sku"`
	SKU                   *string                  `json:"sku"`
	Description           *string                  `json:"description"`
	Category              *string                  `json:"category"`
	CategoryID            *string                  `json:"category_id,omitempty"`
	ThumbnailURL          *string                  `json:"thumbnail_url"`
	DesignLink            *string                  `json:"design_link"`

	// Data Bobot & Waktu Cetak (Unit Level vs Slicer Build Plate Level)
	DefaultWeightGrams    *float64                 `json:"default_weight_grams"`
	PlateWeightGrams      *float64                 `json:"plate_weight_grams,omitempty"`
	MaterialType          *string                  `json:"material_type"`
	DefaultPrintTimeHours *float64                 `json:"default_print_time_hours"`
	PlatePrintTimeHours   *float64                 `json:"plate_print_time_hours,omitempty"`

	PackagingPresetID     *string                  `json:"packaging_preset_id,omitempty"`
	BatchSize             *int                     `json:"batch_size"`
	PackingFeeIDR         *int                     `json:"packing_fee_idr,omitempty"`
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
	ParentID              uuid.UUID                      `json:"parent_id"`
	VariantName           string                         `json:"variant_name"`
	UserID                uuid.UUID                      `json:"user_id"`
	Name                  string                         `json:"name"`
	ParentSKU             *string                        `json:"parent_sku"`
	SKU                   *string                        `json:"sku"`
	Description           *string                        `json:"description"`
	Category              string                         `json:"category"`
	CategoryID            *uuid.UUID                     `json:"category_id"`
	ThumbnailURL          *string                        `json:"thumbnail_url"`
	DesignLink            *string                        `json:"design_link"`

	// =========================================================================
	// Weight & Print Time: Unit vs Slicer Plate
	// =========================================================================
	// DefaultWeightGrams & DefaultPrintTimeHours adalah nilai murni per-unit (1 pcs) yang tersimpan di DB.
	// PlateWeightGrams & PlatePrintTimeHours adalah virtual field (Unit * BatchSize)
	// untuk memudahkan Frontend menampilkan form slicer 1 plate tanpa perlu rumus manual.
	DefaultWeightGrams    float64                        `json:"default_weight_grams"`
	PlateWeightGrams      float64                        `json:"plate_weight_grams"`
	MaterialType          string                         `json:"material_type"`
	DefaultPrintTimeHours float64                        `json:"default_print_time_hours"`
	PlatePrintTimeHours   float64                        `json:"plate_print_time_hours"`

	PackagingPresetID     *uuid.UUID                     `json:"packaging_preset_id,omitempty"`
	BatchSize             int                            `json:"batch_size"`
	PackingFeeIDR         int                            `json:"packing_fee_idr,omitempty"`
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

	// =========================================================================
	// Virtual Field Calculation: Slicer Build Plate
	// =========================================================================
	// Menghitung kembali angka 1 plate slicer utuh (Unit * BatchSize).
	// Menggunakan pembulatan 2 desimal untuk berat dan 3 desimal untuk jam
	// agar terhindar dari floating point precision error di sisi client.
	bSize := p.BatchSize()
	if bSize <= 0 {
		bSize = 1
	}
	plateWeight := math.Round(p.DefaultWeightGrams()*float64(bSize)*100) / 100
	plateHours := math.Round(p.DefaultPrintTimeHours()*float64(bSize)*1000) / 1000

	return ProductResponse{
		ID:                    p.ID(),
		ParentID:              p.ParentID(),
		VariantName:           p.VariantName(),
		UserID:                p.UserID(),
		Name:                  p.Name(),
		ParentSKU:             p.ParentSKU(),
		SKU:                   p.SKU(),
		Description:           p.Description(),
		Category:              p.Category(),
		CategoryID:            p.CategoryID(),
		ThumbnailURL:          p.ThumbnailURL(),
		DesignLink:            p.DesignLink(),
		DefaultWeightGrams:    p.DefaultWeightGrams(),
		PlateWeightGrams:      plateWeight,
		MaterialType:          p.MaterialType(),
		DefaultPrintTimeHours: p.DefaultPrintTimeHours(),
		PlatePrintTimeHours:   plateHours,
		PackagingPresetID:     nil,
		BatchSize:             p.BatchSize(),
		PackingFeeIDR:         0,
		BaseHPP:               p.BaseHPP(),
		BaseSellingPrice:      p.BaseSellingPrice(),
		TargetMarginPercent:   p.TargetMarginPercent(),
		TimesOrdered:          totalSold,
		CreatedAt:             p.CreatedAt(),
		UpdatedAt:             p.UpdatedAt(),
		Components:            comps,
		PackagingItems:        packs,
		CostBreakdown:         cb,
		TotalSold:             totalSold,
	}
}

type ProductListItemResponse struct {
	ID                    uuid.UUID  `json:"id"`
	ParentID              uuid.UUID  `json:"parent_id"`
	VariantName           string     `json:"variant_name"`
	Name                  string     `json:"name"`
	ParentSKU             *string    `json:"parent_sku"`
	SKU                   *string    `json:"sku"`
	Category              string     `json:"category"`
	CategoryID            *uuid.UUID `json:"category_id"`
	ThumbnailURL          *string    `json:"thumbnail_url"`
	MaterialType          string     `json:"material_type"`
	DefaultWeightGrams    float64    `json:"default_weight_grams"`
	PlateWeightGrams      float64    `json:"plate_weight_grams"`
	DefaultPrintTimeHours float64    `json:"default_print_time_hours"`
	PlatePrintTimeHours   float64    `json:"plate_print_time_hours"`
	PackagingPresetID     *uuid.UUID `json:"packaging_preset_id,omitempty"`
	BatchSize             int        `json:"batch_size"`
	PackingFeeIDR         int        `json:"packing_fee_idr,omitempty"`
	BaseHPP               float64    `json:"base_hpp"`
	BaseSellingPrice      float64    `json:"base_selling_price"`
	TargetMarginPercent   int        `json:"target_margin_percent"`
	TotalSold             int        `json:"total_sold"`
	CreatedAt             time.Time  `json:"created_at"`
}

func ToProductListItemResponse(p *product.Product, totalSold int) ProductListItemResponse {
	bSize := p.BatchSize()
	if bSize <= 0 {
		bSize = 1
	}
	plateWeight := math.Round(p.DefaultWeightGrams()*float64(bSize)*100) / 100
	plateHours := math.Round(p.DefaultPrintTimeHours()*float64(bSize)*1000) / 1000

	return ProductListItemResponse{
		ID:                    p.ID(),
		ParentID:              p.ParentID(),
		VariantName:           p.VariantName(),
		Name:                  p.Name(),
		ParentSKU:             p.ParentSKU(),
		SKU:                   p.SKU(),
		Category:              p.Category(),
		CategoryID:            p.CategoryID(),
		ThumbnailURL:          p.ThumbnailURL(),
		MaterialType:          p.MaterialType(),
		DefaultWeightGrams:    p.DefaultWeightGrams(),
		PlateWeightGrams:      plateWeight,
		DefaultPrintTimeHours: p.DefaultPrintTimeHours(),
		PlatePrintTimeHours:   plateHours,
		PackagingPresetID:     nil,
		BatchSize:             p.BatchSize(),
		PackingFeeIDR:         0,
		BaseHPP:               p.BaseHPP(),
		BaseSellingPrice:      p.BaseSellingPrice(),
		TargetMarginPercent:   p.TargetMarginPercent(),
		TotalSold:             totalSold,
		CreatedAt:             p.CreatedAt(),
	}
}
