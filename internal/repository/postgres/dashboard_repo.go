package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"symetra-lab-backend-v2/internal/domain/dashboard"
)

type DashboardRepository struct {
	db *gorm.DB
}

func NewDashboardRepository(db *gorm.DB) *DashboardRepository {
	return &DashboardRepository{db: db}
}

var indoMonths = map[time.Month]string{
	time.January:   "Jan",
	time.February:  "Feb",
	time.March:     "Mar",
	time.April:     "Apr",
	time.May:       "Mei",
	time.June:      "Jun",
	time.July:      "Jul",
	time.August:     "Agu",
	time.September: "Sep",
	time.October:   "Okt",
	time.November:  "Nov",
	time.December:  "Des",
}

func formatIndoDate(t time.Time) string {
	return fmt.Sprintf("%02d %s", t.Day(), indoMonths[t.Month()])
}

func (r *DashboardRepository) GetOverview(ctx context.Context) (*dashboard.DashboardOverview, error) {
	locWIB := time.FixedZone("WIB", 7*3600)
	nowWIB := time.Now().In(locWIB)

	startOfTodayWIB := time.Date(nowWIB.Year(), nowWIB.Month(), nowWIB.Day(), 0, 0, 0, 0, locWIB)
	endOfTodayWIB := time.Date(nowWIB.Year(), nowWIB.Month(), nowWIB.Day(), 23, 59, 59, 999999999, locWIB)

	startOfThisMonthWIB := time.Date(nowWIB.Year(), nowWIB.Month(), 1, 0, 0, 0, 0, locWIB)
	lastMonthTime := startOfThisMonthWIB.AddDate(0, -1, 0)
	startOfLastMonthWIB := time.Date(lastMonthTime.Year(), lastMonthTime.Month(), 1, 0, 0, 0, 0, locWIB)
	endOfLastMonthWIB := startOfThisMonthWIB.Add(-time.Nanosecond)

	overview := &dashboard.DashboardOverview{}

	// 1. KPI Utama, Denyut Hari Ini, Bulan Ini, dan Pipeline Antrean
	type kpiRow struct {
		TotalRevenue           float64 `gorm:"column:total_revenue"`
		TotalProfit            float64 `gorm:"column:total_profit"`
		ActiveOrdersCount      int     `gorm:"column:active_orders_count"`
		QueuedOrdersCount      int     `gorm:"column:queued_orders_count"`
		PrintingOrdersCount    int     `gorm:"column:printing_orders_count"`
		ReadyToShipOrdersCount int     `gorm:"column:ready_to_ship_orders_count"`
		ShippedOrdersCount     int     `gorm:"column:shipped_orders_count"`

		TodayRevenue     float64 `gorm:"column:today_revenue"`
		TodayProfit      float64 `gorm:"column:today_profit"`
		TodayOrdersCount int     `gorm:"column:today_orders_count"`

		ThisMonthRevenue float64 `gorm:"column:this_month_revenue"`
		ThisMonthProfit  float64 `gorm:"column:this_month_profit"`

		LastMonthRevenue float64 `gorm:"column:last_month_revenue"`
		LastMonthProfit  float64 `gorm:"column:last_month_profit"`

		PendingEscrowAmount      float64 `gorm:"column:pending_escrow_amount"`
		PendingEscrowOrdersCount int     `gorm:"column:pending_escrow_orders_count"`
	}

	var kpi kpiRow
	kpiQuery := `
		SELECT
			COALESCE(SUM(CASE WHEN status != 'CANCELLED' THEN gross_amount ELSE 0 END), 0) AS total_revenue,
			COALESCE(SUM(CASE WHEN status != 'CANCELLED' THEN net_profit ELSE 0 END), 0) AS total_profit,
			COUNT(CASE WHEN status NOT IN ('COMPLETED', 'CANCELLED', 'DELIVERED') THEN 1 END) AS active_orders_count,

			COUNT(CASE WHEN status IN ('QUEUED', 'PROCESSED', 'PENDING') THEN 1 END) AS queued_orders_count,
			COUNT(CASE WHEN status IN ('PRINTING', 'IN_PRODUCTION') THEN 1 END) AS printing_orders_count,
			COUNT(CASE WHEN status IN ('READY_TO_SHIP', 'DONE') THEN 1 END) AS ready_to_ship_orders_count,
			COUNT(CASE WHEN status IN ('SHIPPED', 'TO_CONFIRM_RECEIVE') THEN 1 END) AS shipped_orders_count,

			COALESCE(SUM(CASE WHEN status != 'CANCELLED' AND created_at >= ? AND created_at <= ? THEN gross_amount ELSE 0 END), 0) AS today_revenue,
			COALESCE(SUM(CASE WHEN status != 'CANCELLED' AND created_at >= ? AND created_at <= ? THEN net_profit ELSE 0 END), 0) AS today_profit,
			COUNT(CASE WHEN status != 'CANCELLED' AND created_at >= ? AND created_at <= ? THEN 1 END) AS today_orders_count,

			COALESCE(SUM(CASE WHEN status != 'CANCELLED' AND created_at >= ? AND created_at <= ? THEN gross_amount ELSE 0 END), 0) AS this_month_revenue,
			COALESCE(SUM(CASE WHEN status != 'CANCELLED' AND created_at >= ? AND created_at <= ? THEN net_profit ELSE 0 END), 0) AS this_month_profit,

			COALESCE(SUM(CASE WHEN status != 'CANCELLED' AND created_at >= ? AND created_at <= ? THEN gross_amount ELSE 0 END), 0) AS last_month_revenue,
			COALESCE(SUM(CASE WHEN status != 'CANCELLED' AND created_at >= ? AND created_at <= ? THEN net_profit ELSE 0 END), 0) AS last_month_profit,

			COALESCE(SUM(CASE WHEN channel = 'SHOPEE' AND payment_status != 'PAID' AND status NOT IN ('CANCELLED', 'COMPLETED') THEN net_amount ELSE 0 END), 0) AS pending_escrow_amount,
			COUNT(CASE WHEN channel = 'SHOPEE' AND payment_status != 'PAID' AND status NOT IN ('CANCELLED', 'COMPLETED') THEN 1 END) AS pending_escrow_orders_count
		FROM orders;
	`
	if err := r.db.WithContext(ctx).Raw(kpiQuery,
		startOfTodayWIB, endOfTodayWIB,
		startOfTodayWIB, endOfTodayWIB,
		startOfTodayWIB, endOfTodayWIB,
		startOfThisMonthWIB, nowWIB,
		startOfThisMonthWIB, nowWIB,
		startOfLastMonthWIB, endOfLastMonthWIB,
		startOfLastMonthWIB, endOfLastMonthWIB,
	).Scan(&kpi).Error; err != nil {
		return nil, fmt.Errorf("failed to query kpi overview: %w", err)
	}

	overview.TotalRevenue = kpi.TotalRevenue
	overview.TotalProfit = kpi.TotalProfit
	if overview.TotalRevenue > 0 {
		overview.AverageMargin = (overview.TotalProfit / overview.TotalRevenue) * 100
	}
	overview.ActiveOrdersCount = kpi.ActiveOrdersCount
	overview.QueuedOrdersCount = kpi.QueuedOrdersCount
	overview.PrintingOrdersCount = kpi.PrintingOrdersCount
	overview.ReadyToShipOrdersCount = kpi.ReadyToShipOrdersCount
	overview.ShippedOrdersCount = kpi.ShippedOrdersCount

	overview.TodayRevenue = kpi.TodayRevenue
	overview.TodayProfit = kpi.TodayProfit
	overview.TodayOrdersCount = kpi.TodayOrdersCount
	if overview.TodayRevenue > 0 {
		overview.TodayMargin = (overview.TodayProfit / overview.TodayRevenue) * 100
	}

	overview.ThisMonthRevenue = kpi.ThisMonthRevenue
	overview.ThisMonthProfit = kpi.ThisMonthProfit
	overview.LastMonthRevenue = kpi.LastMonthRevenue
	overview.LastMonthProfit = kpi.LastMonthProfit

	overview.PendingEscrowAmount = kpi.PendingEscrowAmount
	overview.PendingEscrowOrdersCount = kpi.PendingEscrowOrdersCount

	// 2. Urgent Shipping Count (Hanya pesanan yang benar-benar belum diserahkan ke kurir)
	var urgentShippingCount int64
	urgentShippingQuery := `
		SELECT COUNT(*)
		FROM orders o
		LEFT JOIN order_marketplace_details m ON m.order_id = o.id
		WHERE o.status NOT IN ('SHIPPED', 'TO_CONFIRM_RECEIVE', 'DELIVERED', 'COMPLETED', 'CANCELLED')
		  AND (
		    (m.ship_by_date_time IS NOT NULL AND m.ship_by_date_time <= ?)
		    OR (m.ship_by_date IS NOT NULL AND to_timestamp(m.ship_by_date) <= ?)
		  );
	`
	if err := r.db.WithContext(ctx).Raw(urgentShippingQuery, endOfTodayWIB, endOfTodayWIB).Scan(&urgentShippingCount).Error; err != nil {
		return nil, fmt.Errorf("failed to query urgent shipping count: %w", err)
	}
	overview.UrgentShippingOrdersCount = int(urgentShippingCount)

	// 3. Unmapped Orders Count
	var unmappedCount int64
	unmappedQuery := `
		SELECT COUNT(DISTINCT o.id)
		FROM orders o
		JOIN order_items it ON it.order_id = o.id
		WHERE o.status != 'CANCELLED' AND it.mapping_status = 'UNMAPPED';
	`
	if err := r.db.WithContext(ctx).Raw(unmappedQuery).Scan(&unmappedCount).Error; err != nil {
		return nil, fmt.Errorf("failed to query unmapped count: %w", err)
	}
	overview.UnmappedOrdersCount = int(unmappedCount)

	// 4. Armada 3D Printer (Machines)
	type machineRow struct {
		ID           uuid.UUID `gorm:"column:id"`
		Name         string    `gorm:"column:name"`
		CurrentState string    `gorm:"column:current_state"`
		Brand        *string   `gorm:"column:brand"`
	}
	var machines []machineRow
	if err := r.db.WithContext(ctx).Raw(`SELECT id, name, COALESCE(current_state, 'IDLE') as current_state, brand FROM machines ORDER BY name ASC;`).Scan(&machines).Error; err != nil {
		return nil, fmt.Errorf("failed to query machines: %w", err)
	}

	overview.TotalMachinesCount = len(machines)
	overview.Machines = make([]dashboard.MachineStat, len(machines))
	for i, m := range machines {
		model := ""
		if m.Brand != nil {
			model = *m.Brand
		}
		if m.CurrentState == "PRINTING" {
			overview.ActiveMachinesCount++
		}
		overview.Machines[i] = dashboard.MachineStat{
			ID:           m.ID,
			Name:         m.Name,
			CurrentState: m.CurrentState,
			Model:        model,
		}
	}

	// 5. Orders by Source / Channel
	var ordersBySource []dashboard.OrderSourceStat
	if err := r.db.WithContext(ctx).Raw(`
		SELECT 
			COALESCE(channel, 'MANUAL') AS source,
			COUNT(*) AS count,
			COALESCE(SUM(gross_amount), 0) AS revenue
		FROM orders
		WHERE status != 'CANCELLED'
		GROUP BY channel
		ORDER BY revenue DESC;
	`).Scan(&ordersBySource).Error; err != nil {
		return nil, fmt.Errorf("failed to query orders by source: %w", err)
	}
	overview.OrdersBySource = ordersBySource

	// 6. Top 5 Products
	var topProducts []dashboard.TopProductStat
	if err := r.db.WithContext(ctx).Raw(`
		SELECT 
			it.product_name,
			SUM(it.quantity) AS total_qty,
			SUM(it.selling_price * it.quantity) AS total_revenue
		FROM order_items it
		JOIN orders o ON o.id = it.order_id
		WHERE o.status != 'CANCELLED'
		GROUP BY it.product_name
		ORDER BY total_qty DESC, total_revenue DESC
		LIMIT 5;
	`).Scan(&topProducts).Error; err != nil {
		return nil, fmt.Errorf("failed to query top products: %w", err)
	}
	overview.TopProducts = topProducts

	// 7. Recent 5 Orders
	type recentRow struct {
		ID           uuid.UUID `gorm:"column:id"`
		OrderNumber  string    `gorm:"column:order_number"`
		CustomerName string    `gorm:"column:customer_name"`
		Channel      string    `gorm:"column:channel"`
		Status       string    `gorm:"column:status"`
		GrossAmount  float64   `gorm:"column:gross_amount"`
		NetProfit    float64   `gorm:"column:net_profit"`
		CreatedAt    time.Time `gorm:"column:created_at"`
	}
	var recentRows []recentRow
	if err := r.db.WithContext(ctx).Raw(`
		SELECT id, order_number, customer_name, COALESCE(channel, 'MANUAL') as channel, status, gross_amount, net_profit, created_at
		FROM orders
		ORDER BY created_at DESC
		LIMIT 5;
	`).Scan(&recentRows).Error; err != nil {
		return nil, fmt.Errorf("failed to query recent orders: %w", err)
	}
	overview.RecentOrders = make([]dashboard.RecentOrderStat, len(recentRows))
	for i, r := range recentRows {
		overview.RecentOrders[i] = dashboard.RecentOrderStat{
			ID:           r.ID,
			OrderNumber:  r.OrderNumber,
			CustomerName: r.CustomerName,
			Source:       r.Channel,
			Status:       r.Status,
			GrossAmount:  r.GrossAmount,
			NetProfit:    r.NetProfit,
			CreatedAt:    r.CreatedAt,
		}
	}

	// 8. Multi-Period Daily Sales (7 Hari & 30 Hari)
	type dailyRow struct {
		DayKey string  `gorm:"column:day_key"`
		Omzet  float64 `gorm:"column:omzet"`
		Profit float64 `gorm:"column:profit"`
	}
	var dailyRows []dailyRow
	since30DaysWIB := startOfTodayWIB.AddDate(0, 0, -29)
	dailyQuery := `
		SELECT 
			TO_CHAR(created_at AT TIME ZONE 'Asia/Jakarta', 'YYYY-MM-DD') AS day_key,
			COALESCE(SUM(gross_amount), 0) AS omzet,
			COALESCE(SUM(net_profit), 0) AS profit
		FROM orders
		WHERE status != 'CANCELLED'
		  AND created_at >= ?
		GROUP BY day_key;
	`
	if err := r.db.WithContext(ctx).Raw(dailyQuery, since30DaysWIB).Scan(&dailyRows).Error; err != nil {
		return nil, fmt.Errorf("failed to query daily sales: %w", err)
	}

	dailyMap := make(map[string]dailyRow)
	for _, row := range dailyRows {
		dailyMap[row.DayKey] = row
	}

	// Generate 30D Array
	overview.ChartData30D = make([]dashboard.DailySalesData, 30)
	for i := 0; i < 30; i++ {
		curDate := since30DaysWIB.AddDate(0, 0, i)
		dayKey := curDate.Format("2006-01-02")
		dateLabel := formatIndoDate(curDate)
		data := dailyMap[dayKey]
		overview.ChartData30D[i] = dashboard.DailySalesData{
			Date:   dateLabel,
			Omzet:  data.Omzet,
			Profit: data.Profit,
		}
	}

	// Generate 7D Array (7 hari terakhir dari 30D)
	overview.ChartData7D = overview.ChartData30D[23:30]

	return overview, nil
}
