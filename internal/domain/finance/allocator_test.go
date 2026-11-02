package finance

import (
	"math"
	"testing"
)

func TestAllocateOrderFunds_Profitable(t *testing.T) {
	input := FundAllocationInput{
		GrossAmount:      100000,
		ChannelFee:       5000,
		FilamentCost:     20000, // 10% buffer = 2000
		ComponentCost:    10000,
		PackagingCost:    5000,
		ElectricityCost:  3000,
		MaintenanceCost:  2000,
		DepreciationCost: 4000,
	}

	res := AllocateOrderFunds(input)

	// NetAmount = 100000 - 5000 = 95000
	if res.NetAmount != 95000 {
		t.Errorf("Expected NetAmount 95000, got %f", res.NetAmount)
	}

	// COGS = 20000 (filament) + 2000 (10% buffer) + 10000 + 5000 + 3000 + 2000 + 4000 = 46000
	if res.COGSAmount != 46000 {
		t.Errorf("Expected COGS 46000, got %f", res.COGSAmount)
	}

	// RawProfit = 95000 - 46000 = 49000
	if res.NetProfit != 49000 {
		t.Errorf("Expected NetProfit 49000, got %f", res.NetProfit)
	}

	// 5% cut from profit = 49000 * 0.05 = 2450
	// TotalEmergency = 2000 (buffer) + 2450 (profit cut) = 4450
	if res.FundEmergency != 4450 {
		t.Errorf("Expected FundEmergency 4450, got %f", res.FundEmergency)
	}

	// Final net profit = 49000 - 2450 = 46550
	if res.FundNetProfit != 46550 {
		t.Errorf("Expected FundNetProfit 46550, got %f", res.FundNetProfit)
	}

	// Verify FundFilament is pure filament (20000)
	if res.FundFilament != 20000 {
		t.Errorf("Expected FundFilament 20000, got %f", res.FundFilament)
	}

	// Crucial invariant: Sum of all 8 buckets MUST equal NetAmount (95000)
	sumBuckets := res.FundFilament + res.FundEmergency + res.FundComponent +
		res.FundPackaging + res.FundElectricity + res.FundMaintenance +
		res.FundDepreciation + res.FundNetProfit

	if math.Abs(sumBuckets-res.NetAmount) > 0.01 {
		t.Errorf("Invariance failed: sum of 8 buckets (%f) != NetAmount (%f)", sumBuckets, res.NetAmount)
	}
}

func TestAllocateOrderFunds_BreakEven(t *testing.T) {
	input := FundAllocationInput{
		GrossAmount:      50000,
		ChannelFee:       0,
		FilamentCost:     20000, // 10% buffer = 2000
		ComponentCost:    10000,
		PackagingCost:    5000,
		ElectricityCost:  5000,
		MaintenanceCost:  3000,
		DepreciationCost: 5000,
	}
	// Total COGS = 20000 + 2000 + 10000 + 5000 + 5000 + 3000 + 5000 = 50000
	// NetAmount = 50000
	// RawProfit = 0
	res := AllocateOrderFunds(input)

	if res.NetProfit != 0 {
		t.Errorf("Expected NetProfit 0, got %f", res.NetProfit)
	}
	// No profit cut, emergency is just the 10% buffer = 2000
	if res.FundEmergency != 2000 {
		t.Errorf("Expected FundEmergency 2000, got %f", res.FundEmergency)
	}
	if res.FundNetProfit != 0 {
		t.Errorf("Expected FundNetProfit 0, got %f", res.FundNetProfit)
	}

	sumBuckets := res.FundFilament + res.FundEmergency + res.FundComponent +
		res.FundPackaging + res.FundElectricity + res.FundMaintenance +
		res.FundDepreciation + res.FundNetProfit

	if math.Abs(sumBuckets-res.NetAmount) > 0.01 {
		t.Errorf("Invariance failed: sum of 8 buckets (%f) != NetAmount (%f)", sumBuckets, res.NetAmount)
	}
}

func TestAllocateOrderFunds_Loss(t *testing.T) {
	input := FundAllocationInput{
		GrossAmount:      30000,
		ChannelFee:       0,
		FilamentCost:     20000, // 10% buffer = 2000
		ComponentCost:    10000,
		PackagingCost:    5000,
		ElectricityCost:  3000,
		MaintenanceCost:  2000,
		DepreciationCost: 4000,
	}
	// Total COGS = 46000
	// NetAmount = 30000
	// RawProfit = -16000
	res := AllocateOrderFunds(input)

	if res.NetProfit != -16000 {
		t.Errorf("Expected NetProfit -16000, got %f", res.NetProfit)
	}
	// When at a loss, no profit cut, emergency remains buffer = 2000
	if res.FundEmergency != 2000 {
		t.Errorf("Expected FundEmergency 2000, got %f", res.FundEmergency)
	}
	if res.FundNetProfit != -16000 {
		t.Errorf("Expected FundNetProfit -16000, got %f", res.FundNetProfit)
	}

	sumBuckets := res.FundFilament + res.FundEmergency + res.FundComponent +
		res.FundPackaging + res.FundElectricity + res.FundMaintenance +
		res.FundDepreciation + res.FundNetProfit

	if math.Abs(sumBuckets-res.NetAmount) > 0.01 {
		t.Errorf("Invariance failed: sum of 8 buckets (%f) != NetAmount (%f)", sumBuckets, res.NetAmount)
	}
}

