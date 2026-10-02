package costing

import (
	"testing"
)

func TestCalculate_DefaultValues(t *testing.T) {
	input := CalculationInput{
		MaterialType:          "PLA",
		DefaultWeightGrams:    50,
		DefaultPrintTimeHours: 2.0,
		TargetMarginPercent:   30,
	}

	result := Calculate(input)

	if result.FilamentCost <= 0 {
		t.Errorf("Expected filament cost > 0, got %f", result.FilamentCost)
	}
	if result.ElectricityCost <= 0 {
		t.Errorf("Expected electricity cost > 0, got %f", result.ElectricityCost)
	}
	if result.DepreciationCost <= 0 {
		t.Errorf("Expected depreciation cost > 0, got %f", result.DepreciationCost)
	}
	if result.BaseHPP <= 0 {
		t.Errorf("Expected base HPP > 0, got %f", result.BaseHPP)
	}
	if result.BaseSellingPrice <= result.BaseHPP {
		t.Errorf("Expected base selling price (%f) > base HPP (%f)", result.BaseSellingPrice, result.BaseHPP)
	}
	if result.ShopeeRecommendedPrice <= result.BaseSellingPrice {
		t.Errorf("Expected Shopee recommended price (%f) > base selling price (%f)", result.ShopeeRecommendedPrice, result.BaseSellingPrice)
	}
}

func TestCalculate_WithComponentsAndPackaging(t *testing.T) {
	input := CalculationInput{
		MaterialType:          "PETG",
		DefaultWeightGrams:    100,
		DefaultPrintTimeHours: 4.0,
		TargetMarginPercent:   40,
		Components: []ComponentCostInput{
			{PricePerUnit: 500, Quantity: 4, MarkupPercent: 20},
		},
		PackingFeeIDR: 2500,
	}

	result := Calculate(input)

	expectedHardware := (500.0 * 4.0) * 1.2 // 2400
	if result.HardwareCost != expectedHardware {
		t.Errorf("Expected hardware cost %f, got %f", expectedHardware, result.HardwareCost)
	}
	if result.PackagingCost != 2500 {
		t.Errorf("Expected packaging cost 2500, got %f", result.PackagingCost)
	}
}
