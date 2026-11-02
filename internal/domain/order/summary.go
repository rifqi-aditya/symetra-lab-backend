package order

// OrderAllocationSummary represents the aggregated financial and 8-bucket cash breakdown
// for orders across channels and date ranges.
type OrderAllocationSummary struct {
	TotalOrders      int64   `json:"total_orders"`
	TotalGrossSales  float64 `json:"total_gross_sales"`
	TotalChannelFees float64 `json:"total_channel_fees"`
	TotalNetRevenue  float64 `json:"total_net_revenue"`
	TotalCOGS        float64 `json:"total_cogs"`
	TotalNetProfit   float64 `json:"total_net_profit"`

	// Net Available Fund Balances (Saldo Tersedia = Allocated - Spent)
	FundFilament     float64 `json:"fund_filament"`
	FundEmergency    float64 `json:"fund_emergency"` // Kas Cadangan Risiko
	FundComponent    float64 `json:"fund_component"`
	FundPackaging    float64 `json:"fund_packaging"`
	FundElectricity  float64 `json:"fund_electricity"`
	FundMaintenance  float64 `json:"fund_maintenance"`
	FundDepreciation float64 `json:"fund_depreciation"`
	FundNetProfit    float64 `json:"fund_net_profit"`

	// Allocated Inflow from Orders
	AllocatedFilament     float64 `json:"allocated_filament"`
	AllocatedEmergency    float64 `json:"allocated_emergency"`
	AllocatedComponent    float64 `json:"allocated_component"`
	AllocatedPackaging    float64 `json:"allocated_packaging"`
	AllocatedElectricity  float64 `json:"allocated_electricity"`
	AllocatedMaintenance  float64 `json:"allocated_maintenance"`
	AllocatedDepreciation float64 `json:"allocated_depreciation"`
	AllocatedNetProfit    float64 `json:"allocated_net_profit"`

	// Actual Spent Outflow (PO & Expenses)
	SpentFilament     float64 `json:"spent_filament"`
	SpentEmergency    float64 `json:"spent_emergency"`
	SpentComponent    float64 `json:"spent_component"`
	SpentPackaging    float64 `json:"spent_packaging"`
	SpentElectricity  float64 `json:"spent_electricity"`
	SpentMaintenance  float64 `json:"spent_maintenance"`
	SpentDepreciation float64 `json:"spent_depreciation"`
	SpentNetProfit    float64 `json:"spent_net_profit"`

	AverageProfitMargin float64 `json:"average_profit_margin"`
	UnmappedItemsCount  int64   `json:"unmapped_items_count"`

	// Pending Marketplace Escrow (Dana Mengendap / Belum Cair)
	PendingOrdersCount  int64   `json:"pending_orders_count"`
	PendingGrossSales   float64 `json:"pending_gross_sales"`
	PendingEscrowAmount float64 `json:"pending_escrow_amount"`
	PendingNetProfit    float64 `json:"pending_net_profit"`
}

