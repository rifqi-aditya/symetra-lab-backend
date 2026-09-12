package models

import (
	"time"
)

// ProductCategory merepresentasikan kategori produk (contoh: Functional Item, Keychain, Flexi)
type ProductCategory struct {
	ID        string    `gorm:"type:uuid;primaryKey" json:"id"`
	UserID    string    `gorm:"type:uuid;not null;index" json:"user_id"`
	Name      string    `gorm:"type:text;not null" json:"name"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (ProductCategory) TableName() string {
	return "product_categories"
}

// Product merepresentasikan master data produk dan cetak biru fisik bengkel (3D print)
type Product struct {
	ID                     string    `gorm:"type:uuid;primaryKey" json:"id"`
	UserID                 string    `gorm:"type:uuid;not null;index" json:"user_id"`
	Name                   string    `gorm:"size:255;not null" json:"name"`
	ParentSKU              *string   `gorm:"size:50;index" json:"parent_sku"`
	SKU                    *string   `gorm:"size:50;uniqueIndex" json:"sku"`
	Description            *string   `gorm:"type:text" json:"description"`
	Category               string    `gorm:"size:128" json:"category"`
	ThumbnailURL           *string   `gorm:"type:text" json:"thumbnail_url"`
	DesignLink             *string   `gorm:"type:text" json:"design_link"`
	DefaultWeightGrams     float64   `gorm:"type:numeric;default:0" json:"default_weight_grams"`
	DefaultPrintTimeHours  float64   `gorm:"type:numeric;default:0" json:"default_print_time_hours"`
	DefaultMachineID       *string   `gorm:"type:uuid" json:"default_machine_id"`
	PackagingPresetID      *string   `gorm:"type:uuid" json:"packaging_preset_id"`
	BatchSize              int       `gorm:"default:1" json:"batch_size"`
	PackingFeeIDR          int       `gorm:"default:0" json:"packing_fee_idr"`
	BaseHPP                float64   `gorm:"type:numeric;default:0" json:"base_hpp"`
	BaseSellingPrice       float64   `gorm:"type:numeric;default:0" json:"base_selling_price"`
	TargetMarginPercent    int       `gorm:"default:30" json:"target_margin_percent"`
	TimesOrdered           int       `gorm:"default:0" json:"times_ordered"`
	CreatedAt              time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt              time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	// Relasi BOM (Bill of Materials)
	Filaments       []ProductFilament      `gorm:"foreignKey:ProductID;references:ID" json:"filaments,omitempty"`
	Components      []ProductComponent     `gorm:"foreignKey:ProductID;references:ID" json:"components,omitempty"`
	PackagingItems  []ProductPackagingItem `gorm:"foreignKey:ProductID;references:ID" json:"packaging_items,omitempty"`
	DefaultMachine  *Machine               `gorm:"foreignKey:DefaultMachineID;references:ID" json:"default_machine,omitempty"`
	PackagingPreset *PackagingPreset       `gorm:"foreignKey:PackagingPresetID;references:ID" json:"packaging_preset,omitempty"`
}

func (Product) TableName() string {
	return "products"
}

// ProductFilament mencatat filamen apa saja dan berapa gram yang dipakai untuk 1 unit produk
type ProductFilament struct {
	ID               string    `gorm:"type:uuid;primaryKey" json:"id"`
	ProductID        string    `gorm:"type:uuid;not null;index" json:"product_id"`
	FilamentID       *string   `gorm:"type:uuid;index" json:"filament_id"`
	WeightUsedGrams  float64   `gorm:"type:numeric;default:0;not null" json:"weight_used_grams"`
	CreatedAt        time.Time `gorm:"autoCreateTime" json:"created_at"`

	Filament *Filament `gorm:"foreignKey:FilamentID;references:ID" json:"filament,omitempty"`
}

func (ProductFilament) TableName() string {
	return "product_filaments"
}

// ProductComponent mencatat komponen tambahan (baut, bearing, magnet, switch) untuk 1 unit produk
type ProductComponent struct {
	ID            string    `gorm:"type:uuid;primaryKey" json:"id"`
	ProductID     string    `gorm:"type:uuid;not null;index" json:"product_id"`
	ComponentID   *string   `gorm:"type:uuid;index" json:"component_id"`
	Quantity      float64   `gorm:"type:numeric;default:1;not null" json:"quantity"`
	MarkupPercent float64   `gorm:"type:numeric;default:0;not null" json:"markup_percent"`
	CreatedAt     time.Time `gorm:"autoCreateTime" json:"created_at"`

	Component *Component `gorm:"foreignKey:ComponentID;references:ID" json:"component,omitempty"`
}

func (ProductComponent) TableName() string {
	return "product_components"
}

// ProductPackagingItem mencatat kemasan khusus per produk jika tidak menggunakan preset umum
type ProductPackagingItem struct {
	ID              string    `gorm:"type:uuid;primaryKey" json:"id"`
	ProductID       string    `gorm:"type:uuid;not null;index" json:"product_id"`
	PackagingItemID string    `gorm:"type:uuid;not null;index" json:"packaging_item_id"`
	QuantityUsed    float64   `gorm:"type:numeric;default:1;not null" json:"quantity_used"`
	CreatedAt       time.Time `gorm:"autoCreateTime" json:"created_at"`

	PackagingItem *PackagingItem `gorm:"foreignKey:PackagingItemID;references:ID" json:"packaging_item,omitempty"`
}

func (ProductPackagingItem) TableName() string {
	return "product_packaging_items"
}

// ShopConfig menyimpan konfigurasi default operasional bengkel (listrik, depresiasi, buffer kegagalan)
type ShopConfig struct {
	ID                      string    `gorm:"type:uuid;primaryKey" json:"id"`
	UserID                  string    `gorm:"type:uuid;not null;index" json:"user_id"`
	FilamentPricePerRoll    float64   `gorm:"type:numeric;default:200000" json:"filament_price_per_roll"`
	FilamentWeightGrams     float64   `gorm:"type:numeric;default:1000" json:"filament_weight_grams"`
	ElectricityTariffPerKwh float64   `gorm:"type:numeric;default:1700" json:"electricity_tariff_per_kwh"`
	PrinterPowerWatts       float64   `gorm:"type:numeric;default:200" json:"printer_power_watts"`
	PrinterPrice            float64   `gorm:"type:numeric;default:5000000" json:"printer_price"`
	PrinterLifespanHours    float64   `gorm:"type:numeric;default:3000" json:"printer_lifespan_hours"`
	FailureBufferPercent    float64   `gorm:"type:numeric;default:10" json:"failure_buffer_percent"`
	UpdatedAt               time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (ShopConfig) TableName() string {
	return "shop_configs"
}

// MarketplacePlatform menyimpan aturan persentase potongan fee platform e-commerce (Shopee, Tokopedia, dll.)
type MarketplacePlatform struct {
	ID                  string    `gorm:"type:uuid;primaryKey" json:"id"`
	UserID              string    `gorm:"type:uuid;not null;index" json:"user_id"`
	Name                string    `gorm:"type:text;not null" json:"name"`
	CommissionPercent   float64   `gorm:"type:numeric;default:0" json:"commission_percent"`
	PromoFeePercent     float64   `gorm:"type:numeric;default:0" json:"promo_fee_percent"`
	FreeShippingPercent float64   `gorm:"type:numeric;default:0" json:"free_shipping_percent"`
	OrderFeeIDR         float64   `gorm:"type:numeric;default:0" json:"order_fee_idr"`
	IsActive            bool      `gorm:"default:true" json:"is_active"`
	CreatedAt           time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (MarketplacePlatform) TableName() string {
	return "marketplace_platforms"
}

// ProductCostBreakdown merangkum seluruh kalkulasi HPP dan rekomendasi harga pasar secara presisi
type ProductCostBreakdown struct {
	FilamentCost         float64 `json:"filament_cost"`          // Biaya filamen riil + buffer
	HardwareCost         float64 `json:"hardware_cost"`          // Biaya baut, bearing, switch
	PackagingCost        float64 `json:"packaging_cost"`         // Biaya preset / kemasan
	ElectricityCost      float64 `json:"electricity_cost"`       // Biaya listrik cetak
	DepreciationCost     float64 `json:"depreciation_cost"`      // Tabungan mesin / depresiasi
	BaseHPP              float64 `json:"base_hpp"`               // Total modal pokok produksi
	TargetMarginPercent  int     `json:"target_margin_percent"`  // Target untung bersih
	TargetProfitIDR      float64 `json:"target_profit_idr"`      // Nominal untung bersih
	BaseSellingPrice     float64 `json:"base_selling_price"`     // Rekomendasi harga jual non-marketplace (offline)
	ShopeeRecommended    float64 `json:"shopee_recommended_price"` // Rekomendasi pasang di Shopee (termasuk admin + flat fee)
}

// ProductResponse merepresentasikan DTO lengkap produk untuk frontend dan API
type ProductResponse struct {
	Product
	CostBreakdown ProductCostBreakdown `json:"cost_breakdown"`
}
