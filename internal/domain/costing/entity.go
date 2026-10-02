package costing

import "math"

// ProductCostBreakdown defines the exact HPP and pricing calculation breakdown.
type ProductCostBreakdown struct {
	FilamentCost           float64 `json:"filament_cost"`
	HardwareCost           float64 `json:"hardware_cost"`
	PackagingCost          float64 `json:"packaging_cost"`
	ElectricityCost        float64 `json:"electricity_cost"`
	MaintenanceCost        float64 `json:"maintenance_cost"`
	DepreciationCost       float64 `json:"depreciation_cost"`
	BaseHPP                float64 `json:"base_hpp"`
	TargetMarginPercent    int     `json:"target_margin_percent"`
	TargetProfitIDR        float64 `json:"target_profit_idr"`
	BaseSellingPrice       float64 `json:"base_selling_price"`
	ShopeeRecommendedPrice float64 `json:"shopee_recommended_price"`
}

type ShopConfigInput struct {
	FilamentPricePerRoll    float64
	FilamentWeightGrams     float64
	ElectricityTariffPerKwh float64
	PrinterPowerWatts       float64
	PrinterPrice            float64
	PrinterLifespanHours    float64
	FailureBufferPercent    float64
}

type MarketplacePlatformInput struct {
	CommissionPercent   float64
	PromoFeePercent     float64
	FreeShippingPercent float64
	OrderFeeIDR         float64
	IsActive            bool
}

type ComponentCostInput struct {
	PricePerUnit  float64
	Quantity      float64
	MarkupPercent float64
}

type PackagingCostInput struct {
	UnitCost     float64
	QuantityUsed float64
}

type MachineCostInput struct {
	TotalPurchaseCost float64
	LifespanHours     float64
	AvgPowerWatts     int
}

type CalculationInput struct {
	MaterialType           string
	DefaultWeightGrams     float64
	DefaultPrintTimeHours  float64
	BatchSize              int
	PackingFeeIDR          int
	TargetMarginPercent    int
	BaseHPP                float64
	BaseSellingPrice       float64
	CustomRatePerGram      float64
	Components             []ComponentCostInput
	PackagingItems         []PackagingCostInput
	HasPackagingPreset     bool
	PackagingPresetTotal   float64
	Machine                *MachineCostInput
	ShopConfig             *ShopConfigInput
	Marketplace            *MarketplacePlatformInput
}

func RoundTwo(val float64) float64 {
	return math.Round(val*100) / 100
}