package costing

import (
	"math"
	"regexp"
	"strings"

	"symetra-lab-backend/models"
)

var (
	nonAlphanumericRegex = regexp.MustCompile(`[^a-zA-Z0-9]+`)
)

// CalculateCostBreakdown menghitung rincian HPP dan rekomendasi harga jual secara akurat
func CalculateCostBreakdown(
	p *models.Product,
	cfg *models.ShopConfig,
	shopeePlatform *models.MarketplacePlatform,
) models.ProductCostBreakdown {
	// 1. Fallback Config jika cfg bernilai nil
	if cfg == nil {
		cfg = &models.ShopConfig{
			FilamentPricePerRoll:    200000,
			FilamentWeightGrams:     1000,
			ElectricityTariffPerKwh: 1700,
			PrinterPowerWatts:       200,
			PrinterPrice:            5000000,
			PrinterLifespanHours:    3000,
			FailureBufferPercent:    10,
		}
	}

	bufferMultiplier := 1.0 + (cfg.FailureBufferPercent / 100.0)

	// 2. Kalkulasi Biaya Filamen (Material)
	filamentCost := 0.0
	if len(p.Filaments) > 0 {
		for _, pf := range p.Filaments {
			pricePerRoll := cfg.FilamentPricePerRoll
			spoolWeight := cfg.FilamentWeightGrams

			if pf.Filament != nil && pf.Filament.PricePerRoll > 0 {
				pricePerRoll = pf.Filament.PricePerRoll
				if pf.Filament.Profile != nil && pf.Filament.Profile.SpoolWeightGrams > 0 {
					spoolWeight = pf.Filament.Profile.SpoolWeightGrams
				}
			}

			costPerGram := 0.0
			if spoolWeight > 0 {
				costPerGram = pricePerRoll / spoolWeight
			}
			weightWithBuffer := pf.WeightUsedGrams * bufferMultiplier
			filamentCost += weightWithBuffer * costPerGram
		}
	} else if p.DefaultWeightGrams > 0 {
		// Fallback ke default weight jika BOM filament belum diatur
		costPerGram := cfg.FilamentPricePerRoll / cfg.FilamentWeightGrams
		weightWithBuffer := p.DefaultWeightGrams * bufferMultiplier
		filamentCost = weightWithBuffer * costPerGram
	}

	// 3. Kalkulasi Biaya Komponen Hardware (Baut, switch, bearing)
	hardwareCost := 0.0
	for _, pc := range p.Components {
		compPrice := 0.0
		if pc.Component != nil {
			compPrice = pc.Component.PricePerUnit
		}
		// Tambahkan markup jika ada
		compMarkup := 1.0 + (pc.MarkupPercent / 100.0)
		hardwareCost += (compPrice * pc.Quantity) * compMarkup
	}

	// 4. Kalkulasi Biaya Kemasan (Packaging)
	packagingCost := 0.0
	if p.PackagingPreset != nil && len(p.PackagingPreset.Items) > 0 {
		for _, pi := range p.PackagingPreset.Items {
			unitCost := 0.0
			if pi.PackagingItem != nil {
				unitCost = pi.PackagingItem.UnitCost
				if unitCost <= 0 {
					unitCost = pi.PackagingItem.CalculateUnitCost()
				}
			}
			packagingCost += unitCost * pi.QuantityUsed
		}
	}
	for _, ppi := range p.PackagingItems {
		unitCost := 0.0
		if ppi.PackagingItem != nil {
			unitCost = ppi.PackagingItem.UnitCost
			if unitCost <= 0 {
				unitCost = ppi.PackagingItem.CalculateUnitCost()
			}
		}
		packagingCost += unitCost * ppi.QuantityUsed
	}
	packagingCost += float64(p.PackingFeeIDR)

	// 5. Kalkulasi Operasional Mesin & Utilitas
	// A. Listrik
	printerWatts := cfg.PrinterPowerWatts
	if p.DefaultMachine != nil && p.DefaultMachine.AvgPowerWatts > 0 {
		printerWatts = float64(p.DefaultMachine.AvgPowerWatts)
	}
	hours := p.DefaultPrintTimeHours
	kwhUsed := (printerWatts / 1000.0) * hours
	electricityCost := kwhUsed * cfg.ElectricityTariffPerKwh

	// B. Depresiasi / Kas Perawatan Mesin
	machinePrice := cfg.PrinterPrice
	lifespanHours := cfg.PrinterLifespanHours
	if p.DefaultMachine != nil {
		if p.DefaultMachine.TotalPurchaseCost > 0 {
			machinePrice = p.DefaultMachine.TotalPurchaseCost
		}
		if p.DefaultMachine.LifespanHours > 0 {
			lifespanHours = float64(p.DefaultMachine.LifespanHours)
		}
	}
	depreciationPerHour := 0.0
	if lifespanHours > 0 {
		depreciationPerHour = machinePrice / lifespanHours
	}
	depreciationCost := depreciationPerHour * hours

	// 6. Base HPP Total
	baseHPP := filamentCost + hardwareCost + packagingCost + electricityCost + depreciationCost

	// Jika baseHPP terhitung 0 namun di database sudah tersimpan base_hpp lama, gunakan fallback
	if baseHPP == 0 && p.BaseHPP > 0 {
		baseHPP = p.BaseHPP
	}

	// 7. Target Margin & Base Selling Price
	marginPercent := p.TargetMarginPercent
	if marginPercent <= 0 {
		marginPercent = 30
	}
	targetProfit := baseHPP * (float64(marginPercent) / 100.0)
	baseSellingPrice := baseHPP + targetProfit
	if p.BaseSellingPrice > 0 && p.BaseSellingPrice > baseHPP {
		baseSellingPrice = p.BaseSellingPrice
		targetProfit = baseSellingPrice - baseHPP
	}

	// 8. Rekomendasi Harga Shopee (Reverse Calculation)
	// H = (Net Desired + Flat Fee) / (1 - Total Percent Fees)
	totalFeePercent := 0.14 // default 14% (komisi + gratis ongkir xtra)
	flatFee := 1250.0       // default order processing fee shopee

	if shopeePlatform != nil && shopeePlatform.IsActive {
		pFees := (shopeePlatform.CommissionPercent + shopeePlatform.PromoFeePercent + shopeePlatform.FreeShippingPercent) / 100.0
		if pFees > 0 && pFees < 0.9 {
			totalFeePercent = pFees
		}
		if shopeePlatform.OrderFeeIDR > 0 {
			flatFee = shopeePlatform.OrderFeeIDR
		}
	}

	shopeeRecommended := (baseSellingPrice + flatFee) / (1.0 - totalFeePercent)
	// Bulatkan ke ratusan terdekat ke atas (misal 26.210 -> 26.500 atau 26.300)
	shopeeRecommended = math.Ceil(shopeeRecommended/100.0) * 100.0

	return models.ProductCostBreakdown{
		FilamentCost:        roundTwo(filamentCost),
		HardwareCost:        roundTwo(hardwareCost),
		PackagingCost:       roundTwo(packagingCost),
		ElectricityCost:     roundTwo(electricityCost),
		DepreciationCost:    roundTwo(depreciationCost),
		BaseHPP:             roundTwo(baseHPP),
		TargetMarginPercent: marginPercent,
		TargetProfitIDR:     roundTwo(targetProfit),
		BaseSellingPrice:    roundTwo(baseSellingPrice),
		ShopeeRecommended:   roundTwo(shopeeRecommended),
	}
}

// FormatScalableSKU menyusun SKU yang konsisten, human-readable, dan modular
// Format: [PARENT_SKU]-[VARIANT_SUFFIX]
func FormatScalableSKU(parentSKU string, variantSuffix string) (string, string) {
	cleanParent := CleanSKUPart(parentSKU)
	if cleanParent == "" {
		cleanParent = "PROD"
	}

	cleanSuffix := CleanSKUPart(variantSuffix)
	if cleanSuffix == "" {
		cleanSuffix = "STD"
	}

	fullSKU := cleanParent + "-" + cleanSuffix
	return cleanParent, fullSKU
}

// CleanSKUPart membersihkan string agar memenuhi standar kode SKU (Uppercase, alphanumeric & dashes)
func CleanSKUPart(input string) string {
	cleaned := strings.TrimSpace(input)
	cleaned = strings.ToUpper(cleaned)
	cleaned = nonAlphanumericRegex.ReplaceAllString(cleaned, "-")
	cleaned = strings.Trim(cleaned, "-")
	return cleaned
}

func roundTwo(val float64) float64 {
	return math.Round(val*100) / 100
}
