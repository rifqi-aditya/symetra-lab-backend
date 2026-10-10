package dto

import (
	"symetra-lab-backend-v2/internal/domain/dashboard"
)

type DashboardOverviewResponse struct {
	TotalRevenue              float64                     `json:"total_revenue"`
	TotalProfit               float64                     `json:"total_profit"`
	AverageMargin             float64                     `json:"average_margin"`
	ActiveOrdersCount         int                         `json:"active_orders_count"`
	ActiveMachinesCount       int                         `json:"active_machines_count"`
	TotalMachinesCount        int                         `json:"total_machines_count"`
	ThisMonthRevenue          float64                     `json:"this_month_revenue"`
	LastMonthRevenue          float64                     `json:"last_month_revenue"`
	ThisMonthProfit           float64                     `json:"this_month_profit"`
	LastMonthProfit           float64                     `json:"last_month_profit"`
	TodayRevenue              float64                     `json:"today_revenue"`
	TodayProfit               float64                     `json:"today_profit"`
	TodayOrdersCount          int                         `json:"today_orders_count"`
	TodayMargin               float64                     `json:"today_margin"`
	QueuedOrdersCount         int                         `json:"queued_orders_count"`
	PrintingOrdersCount       int                         `json:"printing_orders_count"`
	ReadyToShipOrdersCount    int                         `json:"ready_to_ship_orders_count"`
	ShippedOrdersCount        int                         `json:"shipped_orders_count"`
	UnmappedOrdersCount       int                         `json:"unmapped_orders_count"`
	UrgentShippingOrdersCount int                         `json:"urgent_shipping_orders_count"`
	PendingEscrowAmount       float64                     `json:"pending_escrow_amount"`
	PendingEscrowOrdersCount  int                         `json:"pending_escrow_orders_count"`
	ChartData7D               []dashboard.DailySalesData  `json:"chart_data_7d"`
	ChartData30D              []dashboard.DailySalesData  `json:"chart_data_30d"`
	TopProducts               []dashboard.TopProductStat  `json:"top_products"`
	OrdersBySource            []dashboard.OrderSourceStat `json:"orders_by_source"`
	RecentOrders              []dashboard.RecentOrderStat `json:"recent_orders"`
	Machines                  []dashboard.MachineStat     `json:"machines"`
}

func ToDashboardOverviewResponse(d *dashboard.DashboardOverview) DashboardOverviewResponse {
	return DashboardOverviewResponse{
		TotalRevenue:              d.TotalRevenue,
		TotalProfit:               d.TotalProfit,
		AverageMargin:             d.AverageMargin,
		ActiveOrdersCount:         d.ActiveOrdersCount,
		ActiveMachinesCount:       d.ActiveMachinesCount,
		TotalMachinesCount:        d.TotalMachinesCount,
		ThisMonthRevenue:          d.ThisMonthRevenue,
		LastMonthRevenue:          d.LastMonthRevenue,
		ThisMonthProfit:           d.ThisMonthProfit,
		LastMonthProfit:           d.LastMonthProfit,
		TodayRevenue:              d.TodayRevenue,
		TodayProfit:               d.TodayProfit,
		TodayOrdersCount:          d.TodayOrdersCount,
		TodayMargin:               d.TodayMargin,
		QueuedOrdersCount:         d.QueuedOrdersCount,
		PrintingOrdersCount:       d.PrintingOrdersCount,
		ReadyToShipOrdersCount:    d.ReadyToShipOrdersCount,
		ShippedOrdersCount:        d.ShippedOrdersCount,
		UnmappedOrdersCount:       d.UnmappedOrdersCount,
		UrgentShippingOrdersCount: d.UrgentShippingOrdersCount,
		PendingEscrowAmount:       d.PendingEscrowAmount,
		PendingEscrowOrdersCount:  d.PendingEscrowOrdersCount,
		ChartData7D:               d.ChartData7D,
		ChartData30D:              d.ChartData30D,
		TopProducts:               d.TopProducts,
		OrdersBySource:            d.OrdersBySource,
		RecentOrders:              d.RecentOrders,
		Machines:                  d.Machines,
	}
}
