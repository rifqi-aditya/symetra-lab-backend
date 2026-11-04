package finance

import "math"

// FundAllocationInput contains the revenue and cost elements needed to compute the 8-bucket cash allocation.
type FundAllocationInput struct {
	GrossAmount           float64
	ChannelFee            float64
	SellerVoucherDiscount float64
	NetAmount             float64 // Golden source: jika > 0, langsung gunakan nilai ini (misal: EscrowAmount riil dari Shopee)
	FilamentCost          float64 // Net base filament cost (without failure buffer)
	ComponentCost         float64
	PackagingCost         float64
	ElectricityCost       float64
	MaintenanceCost       float64
	DepreciationCost      float64
}

// FundAllocationResult holds the computed amounts for all 8 cash buckets and high-level order financials.
type FundAllocationResult struct {
	GrossAmount      float64
	ChannelFee       float64
	NetAmount        float64
	COGSAmount       float64
	NetProfit        float64 // Raw net profit = NetAmount - COGSAmount

	// The 8 Dedicated Cash Buckets:
	FundFilament     float64 // Pure filament allocation (without buffer)
	FundEmergency    float64 // Kas Cadangan Risiko = 10% filament buffer + 5% positive profit cut
	FundComponent    float64 // Hardware / accessories
	FundPackaging    float64 // Packaging items
	FundElectricity  float64 // Power & utility
	FundMaintenance  float64 // Machine wear-and-tear & maintenance parts
	FundDepreciation float64 // Machine replacement savings
	FundNetProfit    float64 // Final net profit retained after 5% emergency cut
}

// AllocateOrderFunds calculates the exact 8 cash fund allocations from order revenue and costs.
// It applies:
// 1. Filament failure buffer (10% of pure filament cost) -> allocated to Emergency Fund (Kas Cadangan Risiko)
// 2. Net profit emergency cut (5% of positive net profit) -> allocated to Emergency Fund
func AllocateOrderFunds(input FundAllocationInput) FundAllocationResult {
	var netAmount float64
	if input.NetAmount > 0 {
		netAmount = roundTwo(input.NetAmount)
	} else {
		netAmount = roundTwo(input.GrossAmount - input.ChannelFee - input.SellerVoucherDiscount)
	}

	// 1. Buffer kegagalan filamen: 10% dari filamen yang digunakan
	bufferEmergency := roundTwo(input.FilamentCost * FilamentFailureBufferPercent)
	netFilament := roundTwo(input.FilamentCost)

	// 2. Total COGS / HPP
	cogsAmount := roundTwo(netFilament + bufferEmergency + input.ComponentCost + input.PackagingCost +
		input.ElectricityCost + input.MaintenanceCost + input.DepreciationCost)

	// 3. Raw Net Profit
	rawProfit := roundTwo(netAmount - cogsAmount)

	// 4. Potongan 5% dari laba bersih jika untung
	profitCut := 0.0
	if rawProfit > 0 {
		profitCut = roundTwo(rawProfit * NetProfitEmergencyCutPercent)
	}

	// 5. Total Dana Darurat & Laba Bersih Akhir
	totalEmergency := roundTwo(bufferEmergency + profitCut)
	finalNetProfit := roundTwo(rawProfit - profitCut)

	return FundAllocationResult{
		GrossAmount:      roundTwo(input.GrossAmount),
		ChannelFee:       roundTwo(input.ChannelFee),
		NetAmount:        netAmount,
		COGSAmount:       cogsAmount,
		NetProfit:        rawProfit,
		FundFilament:     netFilament,
		FundEmergency:    totalEmergency,
		FundComponent:    roundTwo(input.ComponentCost),
		FundPackaging:    roundTwo(input.PackagingCost),
		FundElectricity:  roundTwo(input.ElectricityCost),
		FundMaintenance:  roundTwo(input.MaintenanceCost),
		FundDepreciation: roundTwo(input.DepreciationCost),
		FundNetProfit:    finalNetProfit,
	}
}

func roundTwo(val float64) float64 {
	return math.Round(val*100) / 100
}

