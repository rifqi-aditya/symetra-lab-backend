package dashboard

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type DailySalesData struct {
	Date   string  `json:"date"`
	Omzet  float64 `json:"omzet"`
	Profit float64 `json:"profit"`
}

type TopProductStat struct {
	ProductName  string  `json:"product_name"`
	TotalQty     int     `json:"total_qty"`
	TotalRevenue float64 `json:"total_revenue"`
}

type OrderSourceStat struct {
	Source  string  `json:"source"`
	Count   int     `json:"count"`
	Revenue float64 `json:"revenue"`
}

type MachineStat struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	CurrentState string    `json:"current_state"`
	Model        string    `json:"model"`
}

type RecentOrderStat struct {
	ID           uuid.UUID `json:"id"`
	OrderNumber  string    `json:"order_number"`
	CustomerName string    `json:"customer_name"`
	Source       string    `json:"source"`
	Status       string    `json:"status"`
	GrossAmount  float64   `json:"gross_amount"`
	NetProfit    float64   `json:"net_profit"`
	CreatedAt    time.Time `json:"created_at"`
}

type DashboardOverview struct {
	TotalRevenue              float64           `json:"total_revenue"`
	TotalProfit               float64           `json:"total_profit"`
	AverageMargin             float64           `json:"average_margin"`
	ActiveOrdersCount         int               `json:"active_orders_count"`
	ActiveMachinesCount       int               `json:"active_machines_count"`
	TotalMachinesCount        int               `json:"total_machines_count"`
	ThisMonthRevenue          float64           `json:"this_month_revenue"`
	LastMonthRevenue          float64           `json:"last_month_revenue"`
	ThisMonthProfit           float64           `json:"this_month_profit"`
	LastMonthProfit           float64           `json:"last_month_profit"`
	TodayRevenue              float64           `json:"today_revenue"`
	TodayProfit               float64           `json:"today_profit"`
	TodayOrdersCount          int               `json:"today_orders_count"`
	TodayMargin               float64           `json:"today_margin"`
	QueuedOrdersCount         int               `json:"queued_orders_count"`
	PrintingOrdersCount       int               `json:"printing_orders_count"`
	ReadyToShipOrdersCount    int               `json:"ready_to_ship_orders_count"`
	ShippedOrdersCount        int               `json:"shipped_orders_count"`
	UnmappedOrdersCount       int               `json:"unmapped_orders_count"`
	UrgentShippingOrdersCount int               `json:"urgent_shipping_orders_count"`
	PendingEscrowAmount       float64           `json:"pending_escrow_amount"`
	PendingEscrowOrdersCount  int               `json:"pending_escrow_orders_count"`
	ChartData7D               []DailySalesData  `json:"chart_data_7d"`
	ChartData30D              []DailySalesData  `json:"chart_data_30d"`
	TopProducts               []TopProductStat  `json:"top_products"`
	OrdersBySource            []OrderSourceStat `json:"orders_by_source"`
	RecentOrders              []RecentOrderStat `json:"recent_orders"`
	Machines                  []MachineStat     `json:"machines"`
}

type Repository interface {
	GetOverview(ctx context.Context) (*DashboardOverview, error)
}
