package costing

import (
	"math"
	"strings"
)

// Calculate computes the full HPP, margin, and marketplace recommended pricing.
func Calculate(input CalculationInput) ProductCostBreakdown {
	cfg := input.ShopConfig
	if cfg == nil {
		cfg = &ShopConfigInput{
			FilamentPricePerRoll:    175000,
			FilamentWeightGrams:     1000,
			ElectricityTariffPerKwh: 1700,
			PrinterPowerWatts:       200,
			PrinterPrice:            5000000,
			PrinterLifespanHours:    3000,
			FailureBufferPercent:    10,
		}
	}

	bufferMultiplier := 1.0 + (cfg.FailureBufferPercent / 100.0)

	// 1. Filament Cost
	filamentCost := 0.0
	if input.DefaultWeightGrams > 0 {
		costPerGram := input.CustomRatePerGram
		if costPerGram <= 0 {
			switch strings.ToUpper(strings.TrimSpace(input.MaterialType)) {
			case "PLA SILK":
				costPerGram = 210
			case "PETG":
				costPerGram = 190
			case "TPU":
				costPerGram = 250
			case "ABS":
				costPerGram = 180
			case "PLA":
				costPerGram = 175
			default:
				if cfg.FilamentWeightGrams > 0 && cfg.FilamentPricePerRoll > 0 {
					costPerGram = cfg.FilamentPricePerRoll / cfg.FilamentWeightGrams
				} else {
					costPerGram = 175
				}
			}
		}
		filamentCost = input.DefaultWeightGrams * bufferMultiplier * costPerGram
	}

	// 2. Hardware / Components Cost
	hardwareCost := 0.0
	for _, c := range input.Components {
		markup := 1.0 + (c.MarkupPercent / 100.0)
		hardwareCost += (c.PricePerUnit * c.Quantity) * markup
	}

	// 3. Packaging Cost
	packagingCost := 0.0
	if input.HasPackagingPreset && input.PackagingPresetTotal > 0 {
		packagingCost = input.PackagingPresetTotal
	} else if len(input.PackagingItems) > 0 {
		for _, pi := range input.PackagingItems {
			packagingCost += pi.UnitCost * pi.QuantityUsed
		}
	} else if input.PackingFeeIDR > 0 {
		packagingCost = float64(input.PackingFeeIDR)
	}

	// 4. Electricity & Machine Depreciation
	printerWatts := cfg.PrinterPowerWatts
	if input.Machine != nil && input.Machine.AvgPowerWatts > 0 {
		printerWatts = float64(input.Machine.AvgPowerWatts)
	}
	hours := input.DefaultPrintTimeHours
	kwhUsed := (printerWatts / 1000.0) * hours
	electricityCost := kwhUsed * cfg.ElectricityTariffPerKwh

	machinePrice := cfg.PrinterPrice
	lifespanHours := cfg.PrinterLifespanHours
	if input.Machine != nil {
		if input.Machine.TotalPurchaseCost > 0 {
			machinePrice = input.Machine.TotalPurchaseCost
		}
		if input.Machine.LifespanHours > 0 {
			lifespanHours = input.Machine.LifespanHours
		}
	}
	depreciationPerHour := 0.0
	if lifespanHours > 0 {
		depreciationPerHour = machinePrice / lifespanHours
	}
	depreciationCost := depreciationPerHour * hours

	// 5. Base HPP
	baseHPP := filamentCost + hardwareCost + packagingCost + electricityCost + depreciationCost
	if baseHPP == 0 && input.BaseHPP > 0 {
		baseHPP = input.BaseHPP
	}

	// 6. Target Profit & Selling Price
	marginPercent := input.TargetMarginPercent
	if marginPercent <= 0 {
		marginPercent = 30
	}
	targetProfit := baseHPP * (float64(marginPercent) / 100.0)
	baseSellingPrice := baseHPP + targetProfit
	if input.BaseSellingPrice > 0 && input.BaseSellingPrice > baseHPP {
		baseSellingPrice = input.BaseSellingPrice
		targetProfit = baseSellingPrice - baseHPP
	}

	// 7. Shopee Recommended Price
	totalFeePercent := 0.14
	flatFee := 1250.0
	mp := input.Marketplace
	if mp != nil && mp.IsActive {
		pFees := (mp.CommissionPercent + mp.PromoFeePercent + mp.FreeShippingPercent) / 100.0
		if pFees > 0 && pFees < 0.9 {
			totalFeePercent = pFees
		}
		if mp.OrderFeeIDR > 0 {
			flatFee = mp.OrderFeeIDR
		}
	}

	shopeeRecommended := (baseSellingPrice + flatFee) / (1.0 - totalFeePercent)
	shopeeRecommended = math.Ceil(shopeeRecommended/100.0) * 100.0

	return ProductCostBreakdown{
		FilamentCost:           RoundTwo(filamentCost),
		HardwareCost:           RoundTwo(hardwareCost),
		PackagingCost:          RoundTwo(packagingCost),
		ElectricityCost:        RoundTwo(electricityCost),
		MaintenanceCost:        RoundTwo(depreciationCost * 0.5),
		DepreciationCost:       RoundTwo(depreciationCost),
		BaseHPP:                RoundTwo(baseHPP),
		TargetMarginPercent:    marginPercent,
		TargetProfitIDR:        RoundTwo(targetProfit),
		BaseSellingPrice:       RoundTwo(baseSellingPrice),
		ShopeeRecommendedPrice: RoundTwo(shopeeRecommended),
	}
}