package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"symetra-lab-backend-v2/internal/domain/finance"
	"symetra-lab-backend-v2/internal/domain/order"
)

type orderGORM struct {
	ID               uuid.UUID       `gorm:"column:id;primaryKey;type:uuid"`
	UserID           uuid.UUID       `gorm:"column:user_id;type:uuid"`
	OrderNumber      string          `gorm:"column:order_number"`
	CustomerName     string          `gorm:"column:customer_name"`
	CustomerContact  string          `gorm:"column:customer_contact"`
	Channel          string          `gorm:"column:channel"`
	GrossAmount      float64         `gorm:"column:gross_amount"`
	ChannelFee       float64         `gorm:"column:channel_fee"`
	NetAmount        float64         `gorm:"column:net_amount"`
	CogsAmount       float64         `gorm:"column:cogs_amount"`
	NetProfit        float64         `gorm:"column:net_profit"`
	FundFilament     float64         `gorm:"column:fund_filament"`
	FundComponent    float64         `gorm:"column:fund_component"`
	FundPackaging    float64         `gorm:"column:fund_packaging"`
	FundElectricity  float64         `gorm:"column:fund_electricity"`
	FundMaintenance  float64         `gorm:"column:fund_maintenance"`
	FundDepreciation float64         `gorm:"column:fund_depreciation"`
	FundNetProfit    float64         `gorm:"column:fund_net_profit"`
	Status           string          `gorm:"column:status"`
	Notes              string                      `gorm:"column:notes"`
	PaymentStatus      string                      `gorm:"column:payment_status"`
	StartedAt          *time.Time                  `gorm:"column:started_at"`
	CompletedAt        *time.Time                  `gorm:"column:completed_at"`
	CreatedAt          time.Time                   `gorm:"column:created_at"`
	UpdatedAt          time.Time                   `gorm:"column:updated_at"`
	Items              []orderItemGORM             `gorm:"foreignKey:OrderID;references:ID;constraint:OnDelete:CASCADE"`
	MarketplaceDetails *orderMarketplaceDetailGORM `gorm:"foreignKey:OrderID;references:ID"`
}

func (orderGORM) TableName() string {
	return "orders"
}

type orderMarketplaceDetailGORM struct {
	ID              uuid.UUID  `gorm:"column:id;primaryKey;type:uuid"`
	OrderID         uuid.UUID  `gorm:"column:order_id;type:uuid"`
	Channel         string     `gorm:"column:channel"`
	OrderSN         string     `gorm:"column:order_sn"`
	ShippingCarrier string     `gorm:"column:shipping_carrier"`
	TrackingNumber  string     `gorm:"column:tracking_number"`
	ShipByDate      int64      `gorm:"column:ship_by_date"`
	ShipByDateTime  *time.Time `gorm:"column:ship_by_date_time"`
	EscrowAmount    float64    `gorm:"column:escrow_amount"`
	FinancialStatus string     `gorm:"column:financial_status"`
}

func (orderMarketplaceDetailGORM) TableName() string {
	return "order_marketplace_details"
}

type orderItemGORM struct {
	ID               uuid.UUID    `gorm:"column:id;primaryKey;type:uuid"`
	OrderID          uuid.UUID    `gorm:"column:order_id;type:uuid"`
	ProductID        *uuid.UUID   `gorm:"column:product_id;type:uuid"`
	Product          *productGORM `gorm:"foreignKey:ProductID;references:ID"`
	ProductName      string       `gorm:"column:product_name"`
	ItemSKU          string     `gorm:"column:item_sku"`
	Quantity         int        `gorm:"column:quantity"`
	SellingPrice     float64    `gorm:"column:selling_price"`
	HPP              float64    `gorm:"column:hpp"`
	WeightGrams      float64    `gorm:"column:weight_grams"`
	PrintTimeHours   float64    `gorm:"column:print_time_hours"`
	FilamentCost     float64    `gorm:"column:filament_cost"`
	ComponentCost    float64    `gorm:"column:component_cost"`
	PackagingCost    float64    `gorm:"column:packaging_cost"`
	EnergyCost       float64    `gorm:"column:energy_cost"`
	MaintenanceCost  float64    `gorm:"column:maintenance_cost"`
	DepreciationCost float64    `gorm:"column:depreciation_cost"`
	ChannelItemID    int64      `gorm:"column:channel_item_id"`
	ChannelModelID   int64      `gorm:"column:channel_model_id"`
	MappingStatus    string     `gorm:"column:mapping_status"`
	MatchedSKU       string     `gorm:"column:matched_sku"`
	CreatedAt        time.Time  `gorm:"column:created_at"`
	UpdatedAt        time.Time  `gorm:"column:updated_at"`
}

func (orderItemGORM) TableName() string {
	return "order_items"
}

type OrderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func mapOrderGORMToDomain(g *orderGORM) *order.Order {
	if g == nil {
		return nil
	}

	items := make([]order.OrderItem, len(g.Items))
	for i, it := range g.Items {
		prodName := it.ProductName
		var thumbURL *string
		if it.Product != nil {
			if it.Product.Name != "" {
				prodName = it.Product.Name
			}
			thumbURL = it.Product.ThumbnailURL
		}

		var machineID *uuid.UUID
		if it.Product != nil && it.Product.DefaultMachineID != nil {
			machineID = it.Product.DefaultMachineID
		}

		itemObj := order.ReconstructOrderItemFull(
			it.ID,
			it.OrderID,
			it.ProductID,
			prodName,
			it.ItemSKU,
			it.Quantity,
			it.SellingPrice,
			it.HPP,
			it.WeightGrams,
			it.PrintTimeHours,
			machineID,
			it.FilamentCost,
			it.ComponentCost,
			it.PackagingCost,
			it.EnergyCost,
			it.MaintenanceCost,
			it.DepreciationCost,
			it.HPP,
			it.SellingPrice-it.HPP,
			it.ChannelItemID,
			it.ChannelModelID,
			it.MappingStatus,
			it.MatchedSKU,
			it.CreatedAt,
		)
		itemObj.SetThumbnailURL(thumbURL)
		items[i] = itemObj
	}

	ch := g.Channel
	if ch == "" {
		ch = "MANUAL"
	}

	gross := g.GrossAmount
	net := g.NetAmount
	if net == 0 {
		net = gross
	}
	cogs := g.CogsAmount
	profit := g.NetProfit
	if profit <= 0 && net > cogs && g.NetProfit == -cogs {
		profit = net - cogs
	}

	totRev := gross
	totHpp := cogs
	totProf := profit

	domainOrder := order.ReconstructOrder(
		g.ID,
		g.UserID,
		g.OrderNumber,
		g.CustomerName,
		g.CustomerContact,
		ch,
		gross,
		g.ChannelFee,
		net,
		cogs,
		profit,
		g.FundFilament,
		g.FundComponent,
		g.FundPackaging,
		g.FundElectricity,
		g.FundMaintenance,
		g.FundDepreciation,
		g.FundNetProfit,
		totRev,
		totHpp,
		totProf,
		g.Status,
		g.Notes,
		ch,
		g.PaymentStatus,
		g.StartedAt,
		g.CompletedAt,
		items,
		g.CreatedAt,
		g.UpdatedAt,
	)

	if g.MarketplaceDetails != nil {
		domainOrder.SetLogistics(
			g.MarketplaceDetails.ShippingCarrier,
			g.MarketplaceDetails.TrackingNumber,
			g.MarketplaceDetails.ShipByDateTime,
		)
		if g.MarketplaceDetails.FinancialStatus != "" {
			domainOrder.SetFinancialStatus(g.MarketplaceDetails.FinancialStatus)
		}
	}
	if domainOrder.FinancialStatus() == "" {
		if g.PaymentStatus == "PAID" || g.Status == "COMPLETED" {
			domainOrder.SetFinancialStatus("RELEASED")
		} else {
			domainOrder.SetFinancialStatus("PENDING_RELEASE")
		}
	}

	return domainOrder
}

func (r *OrderRepository) FindAll(ctx context.Context, userID uuid.UUID) ([]*order.Order, error) {
	var gormOrders []orderGORM
	err := r.db.WithContext(ctx).
		Preload("Items.Product").
		Preload("MarketplaceDetails").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&gormOrders).Error
	if err != nil {
		return nil, err
	}

	results := make([]*order.Order, len(gormOrders))
	for i := range gormOrders {
		results[i] = mapOrderGORMToDomain(&gormOrders[i])
	}
	return results, nil
}

func (r *OrderRepository) FindByID(ctx context.Context, userID, id uuid.UUID) (*order.Order, error) {
	var g orderGORM
	err := r.db.WithContext(ctx).
		Preload("Items.Product").
		Preload("MarketplaceDetails").
		Where("id = ? AND user_id = ?", id, userID).
		First(&g).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, order.ErrOrderNotFound
		}
		return nil, err
	}
	return mapOrderGORMToDomain(&g), nil
}

func (r *OrderRepository) FindByOrderNumber(ctx context.Context, userID uuid.UUID, orderNumber string) (*order.Order, error) {
	var g orderGORM
	err := r.db.WithContext(ctx).
		Preload("Items.Product").
		Preload("MarketplaceDetails").
		Where("(order_number = ? OR id IN (SELECT order_id FROM order_marketplace_details WHERE order_sn = ?)) AND user_id = ?", orderNumber, orderNumber, userID).
		First(&g).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, order.ErrOrderNotFound
		}
		return nil, err
	}
	return mapOrderGORMToDomain(&g), nil
}

func (r *OrderRepository) Create(ctx context.Context, o *order.Order) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		g := orderGORM{
			ID:               o.ID(),
			UserID:           o.UserID(),
			OrderNumber:      o.OrderNumber(),
			CustomerName:     o.CustomerName(),
			CustomerContact:  o.CustomerContact(),
			Channel:          o.Channel(),
			GrossAmount:      o.GrossAmount(),
			ChannelFee:       o.ChannelFee(),
			NetAmount:        o.NetAmount(),
			CogsAmount:       o.CogsAmount(),
			NetProfit:        o.NetProfit(),
			FundFilament:     o.FundFilament(),
			FundComponent:    o.FundComponent(),
			FundPackaging:    o.FundPackaging(),
			FundElectricity:  o.FundElectricity(),
			FundMaintenance:  o.FundMaintenance(),
			FundDepreciation: o.FundDepreciation(),
			FundNetProfit:    o.FundNetProfit(),
			Status:           o.Status(),
			Notes:            o.Notes(),
			PaymentStatus:    o.PaymentStatus(),
			StartedAt:        o.StartedAt(),
			CompletedAt:      o.CompletedAt(),
			CreatedAt:        o.CreatedAt(),
			UpdatedAt:        o.UpdatedAt(),
		}

		if err := tx.Create(&g).Error; err != nil {
			return err
		}

		for _, it := range o.Items() {
			itemGORM := orderItemGORM{
				ID:               it.ID(),
				OrderID:          o.ID(),
				ProductID:        it.ProductID(),
				ProductName:      it.ProductName(),
				ItemSKU:          it.ItemSKU(),
				Quantity:         it.Quantity(),
				SellingPrice:     it.SellingPrice(),
				HPP:              it.HPP(),
				WeightGrams:      it.WeightGrams(),
				PrintTimeHours:   it.PrintTimeHours(),
				FilamentCost:     it.FilamentCost(),
				ComponentCost:    it.ComponentCost(),
				PackagingCost:    it.PackagingCost(),
				EnergyCost:       it.EnergyCost(),
				MaintenanceCost:  it.MaintenanceCost(),
				DepreciationCost: it.DepreciationCost(),
				ChannelItemID:    it.ChannelItemID(),
				ChannelModelID:   it.ChannelModelID(),
				MappingStatus:    it.MappingStatus(),
				MatchedSKU:       it.MatchedSKU(),
				CreatedAt:        it.CreatedAt(),
			}
			if err := tx.Create(&itemGORM).Error; err != nil {
				return err
			}
		}

		return nil
	})
	if err == nil && o.PaymentStatus() == "PAID" {
		_ = SyncCashAccountLedger(ctx, r.db)
	}
	return err
}

func (r *OrderRepository) UpdateStatus(ctx context.Context, userID, id uuid.UUID, status string) error {
	updates := map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}
	now := time.Now()
	if status == "IN_PRODUCTION" {
		updates["started_at"] = &now
	}
	if status == "COMPLETED" {
		updates["completed_at"] = &now
	}

	res := r.db.WithContext(ctx).
		Model(&orderGORM{}).
		Where("id = ? AND user_id = ?", id, userID).
		Updates(updates)

	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return order.ErrOrderNotFound
	}
	return nil
}

func (r *OrderRepository) UpdatePaymentStatus(ctx context.Context, userID, id uuid.UUID, paymentStatus string) error {
	res := r.db.WithContext(ctx).
		Model(&orderGORM{}).
		Where("id = ? AND user_id = ?", id, userID).
		Updates(map[string]interface{}{
			"payment_status": paymentStatus,
			"updated_at":     time.Now(),
		})

	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return order.ErrOrderNotFound
	}
	_ = SyncCashAccountLedger(ctx, r.db)
	return nil
}

func (r *OrderRepository) Delete(ctx context.Context, userID, id uuid.UUID) error {
	res := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", id, userID).
		Delete(&orderGORM{})

	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return order.ErrOrderNotFound
	}
	_ = SyncCashAccountLedger(ctx, r.db)
	return nil
}

// GetOrderAllocationSummary executes a fast unified aggregation across all channels
func (r *OrderRepository) GetOrderAllocationSummary(
	ctx context.Context,
	userID uuid.UUID,
	channel string,
	dateFrom, dateTo *time.Time,
) (*finance.OrderAllocationSummary, error) {
	type aggResult struct {
		TotalOrders         int64   `gorm:"column:total_orders"`
		TotalGrossSales     float64 `gorm:"column:total_gross_sales"`
		TotalChannelFees    float64 `gorm:"column:total_channel_fees"`
		TotalNetRevenue     float64 `gorm:"column:total_net_revenue"`
		TotalCOGS           float64 `gorm:"column:total_cogs"`
		TotalNetProfit      float64 `gorm:"column:total_net_profit"`
		FundFilament        float64 `gorm:"column:fund_filament"`
		FundComponent       float64 `gorm:"column:fund_component"`
		FundPackaging       float64 `gorm:"column:fund_packaging"`
		FundElectricity     float64 `gorm:"column:fund_electricity"`
		FundMaintenance     float64 `gorm:"column:fund_maintenance"`
		FundDepreciation    float64 `gorm:"column:fund_depreciation"`
		FundNetProfit       float64 `gorm:"column:fund_net_profit"`
		UnmappedItemsCount  int64   `gorm:"column:unmapped_items_count"`
		PendingOrdersCount  int64   `gorm:"column:pending_orders_count"`
		PendingGrossSales   float64 `gorm:"column:pending_gross_sales"`
		PendingEscrowAmount float64 `gorm:"column:pending_escrow_amount"`
		PendingNetProfit    float64 `gorm:"column:pending_net_profit"`
	}

	query := `
		SELECT 
			COUNT(*) AS total_orders,
			COALESCE(SUM(CASE WHEN gross_amount > 0 THEN gross_amount ELSE (net_amount + channel_fee) END), 0) AS total_gross_sales,
			COALESCE(SUM(channel_fee), 0) AS total_channel_fees,
			COALESCE(SUM(net_amount), 0) AS total_net_revenue,
			COALESCE(SUM(cogs_amount), 0) AS total_cogs,
			COALESCE(SUM(net_profit), 0) AS total_net_profit,
			COALESCE(SUM(CASE WHEN payment_status = 'PAID' THEN fund_filament ELSE 0 END), 0) AS fund_filament,
			COALESCE(SUM(CASE WHEN payment_status = 'PAID' THEN fund_component ELSE 0 END), 0) AS fund_component,
			COALESCE(SUM(CASE WHEN payment_status = 'PAID' THEN fund_packaging ELSE 0 END), 0) AS fund_packaging,
			COALESCE(SUM(CASE WHEN payment_status = 'PAID' THEN fund_electricity ELSE 0 END), 0) AS fund_electricity,
			COALESCE(SUM(CASE WHEN payment_status = 'PAID' THEN fund_maintenance ELSE 0 END), 0) AS fund_maintenance,
			COALESCE(SUM(CASE WHEN payment_status = 'PAID' THEN fund_depreciation ELSE 0 END), 0) AS fund_depreciation,
			COALESCE(SUM(CASE WHEN payment_status = 'PAID' THEN fund_net_profit ELSE 0 END), 0) AS fund_net_profit,
			COALESCE((
				SELECT COUNT(*) 
				FROM order_items oi
				JOIN orders o2 ON o2.id = oi.order_id
				WHERE o2.user_id = ? AND oi.mapping_status = 'UNMAPPED'
			), 0) AS unmapped_items_count,
			COALESCE(COUNT(CASE WHEN payment_status != 'PAID' AND channel = 'SHOPEE' THEN 1 END), 0) AS pending_orders_count,
			COALESCE(SUM(CASE WHEN payment_status != 'PAID' AND channel = 'SHOPEE' THEN (CASE WHEN gross_amount > 0 THEN gross_amount ELSE (net_amount + channel_fee) END) ELSE 0 END), 0) AS pending_gross_sales,
			COALESCE(SUM(CASE WHEN payment_status != 'PAID' AND channel = 'SHOPEE' THEN net_amount ELSE 0 END), 0) AS pending_escrow_amount,
			COALESCE(SUM(CASE WHEN payment_status != 'PAID' AND channel = 'SHOPEE' THEN net_profit ELSE 0 END), 0) AS pending_net_profit
		FROM orders
		WHERE user_id = ? 
		  AND status != 'CANCELLED'
		  AND (? = 'ALL' OR channel = ?)
		  AND (?::timestamptz IS NULL OR created_at >= ?)
		  AND (?::timestamptz IS NULL OR created_at <= ?)
	`

	ch := channel
	if ch == "" {
		ch = "ALL"
	}

	var agg aggResult
	err := r.db.WithContext(ctx).Raw(
		query,
		userID,
		userID,
		ch, ch,
		dateFrom, dateFrom,
		dateTo, dateTo,
	).Scan(&agg).Error

	if err != nil {
		return nil, err
	}

	avgMargin := 0.0
	if agg.TotalGrossSales > 0 {
		avgMargin = (agg.TotalNetProfit / agg.TotalGrossSales) * 100.0
	}

	// 2. Query cash_accounts running ledger (O(1) direct read) for All-Time without custom filters
	var fundFilament, fundComponent, fundPackaging, fundElectricity, fundMaintenance, fundDepreciation, fundNetProfit float64
	var spentFilament, spentComponent, spentPackaging, spentElectricity, spentMaintenance, spentDepreciation, spentNetProfit float64
	var allocFilament, allocComponent, allocPackaging, allocElectricity, allocMaintenance, allocDepreciation, allocNetProfit float64

	isAllTime := dateFrom == nil && dateTo == nil && (channel == "" || channel == "ALL")
	useLedger := false

	if isAllTime {
		type accountRow struct {
			Code            string  `gorm:"column:code"`
			AllocatedAmount float64 `gorm:"column:allocated_amount"`
			SpentAmount     float64 `gorm:"column:spent_amount"`
			CurrentBalance  float64 `gorm:"column:current_balance"`
		}
		var accRows []accountRow
		err := r.db.WithContext(ctx).Table("cash_accounts").
			Select("code, allocated_amount, spent_amount, current_balance").
			Where("is_active = ?", true).
			Scan(&accRows).Error

		if err == nil && len(accRows) > 0 {
			useLedger = true
			for _, acc := range accRows {
				switch acc.Code {
				case "FILAMENT":
					fundFilament = acc.CurrentBalance
					spentFilament = acc.SpentAmount
					allocFilament = acc.AllocatedAmount
				case "COMPONENT":
					fundComponent = acc.CurrentBalance
					spentComponent = acc.SpentAmount
					allocComponent = acc.AllocatedAmount
				case "PACKAGING":
					fundPackaging = acc.CurrentBalance
					spentPackaging = acc.SpentAmount
					allocPackaging = acc.AllocatedAmount
				case "ELECTRICITY":
					fundElectricity = acc.CurrentBalance
					spentElectricity = acc.SpentAmount
					allocElectricity = acc.AllocatedAmount
				case "MAINTENANCE":
					fundMaintenance = acc.CurrentBalance
					spentMaintenance = acc.SpentAmount
					allocMaintenance = acc.AllocatedAmount
				case "DEPRECIATION":
					fundDepreciation = acc.CurrentBalance
					spentDepreciation = acc.SpentAmount
					allocDepreciation = acc.AllocatedAmount
				case "NET_PROFIT":
					fundNetProfit = acc.CurrentBalance
					spentNetProfit = acc.SpentAmount
					allocNetProfit = acc.AllocatedAmount
				}
			}
		}
	}

	if !useLedger {
		type expenseResult struct {
			Category   string  `gorm:"column:category"`
			TotalSpent float64 `gorm:"column:total_spent"`
		}
		var expenses []expenseResult
		expQuery := `
			SELECT category, COALESCE(SUM(amount), 0) AS total_spent
			FROM finance_transactions
			WHERE type = 'EXPENSE'
			  AND (?::timestamptz IS NULL OR transaction_date >= ?)
			  AND (?::timestamptz IS NULL OR transaction_date <= ?)
			GROUP BY category
		`
		_ = r.db.WithContext(ctx).Raw(expQuery, dateFrom, dateFrom, dateTo, dateTo).Scan(&expenses).Error

		spentFilament = 0
		spentComponent = 0
		spentPackaging = 0
		spentElectricity = 0
		spentMaintenance = 0
		spentDepreciation = 0
		spentNetProfit = 0

		for _, e := range expenses {
			switch e.Category {
			case "FILAMENT":
				spentFilament += e.TotalSpent
			case "HARDWARE":
				spentComponent += e.TotalSpent
			case "PACKAGING":
				spentPackaging += e.TotalSpent
			case "ELECTRICITY":
				spentElectricity += e.TotalSpent
			case "MACHINE_MAINTENANCE":
				spentMaintenance += e.TotalSpent
			case "MACHINE_PURCHASE":
				spentDepreciation += e.TotalSpent
			default:
				spentNetProfit += e.TotalSpent
			}
		}

		allocFilament = agg.FundFilament
		allocComponent = agg.FundComponent
		allocPackaging = agg.FundPackaging
		allocElectricity = agg.FundElectricity
		allocMaintenance = agg.FundMaintenance
		allocDepreciation = agg.FundDepreciation
		allocNetProfit = agg.FundNetProfit

		fundFilament = allocFilament - spentFilament
		fundComponent = allocComponent - spentComponent
		fundPackaging = allocPackaging - spentPackaging
		fundElectricity = allocElectricity - spentElectricity
		fundMaintenance = allocMaintenance - spentMaintenance
		fundDepreciation = allocDepreciation - spentDepreciation
		fundNetProfit = allocNetProfit - spentNetProfit
	}

	return &finance.OrderAllocationSummary{
		TotalOrders:           agg.TotalOrders,
		TotalGrossSales:       agg.TotalGrossSales,
		TotalChannelFees:      agg.TotalChannelFees,
		TotalNetRevenue:       agg.TotalNetRevenue,
		TotalCOGS:             agg.TotalCOGS,
		TotalNetProfit:        agg.TotalNetProfit,

		FundFilament:          fundFilament,
		FundComponent:         fundComponent,
		FundPackaging:         fundPackaging,
		FundElectricity:       fundElectricity,
		FundMaintenance:       fundMaintenance,
		FundDepreciation:      fundDepreciation,
		FundNetProfit:         fundNetProfit,

		AllocatedFilament:     allocFilament,
		AllocatedComponent:    allocComponent,
		AllocatedPackaging:    allocPackaging,
		AllocatedElectricity:  allocElectricity,
		AllocatedMaintenance:  allocMaintenance,
		AllocatedDepreciation: allocDepreciation,
		AllocatedNetProfit:    allocNetProfit,

		SpentFilament:         spentFilament,
		SpentComponent:        spentComponent,
		SpentPackaging:        spentPackaging,
		SpentElectricity:      spentElectricity,
		SpentMaintenance:      spentMaintenance,
		SpentDepreciation:     spentDepreciation,
		SpentNetProfit:        spentNetProfit,

		AverageProfitMargin:   avgMargin,
		UnmappedItemsCount:    agg.UnmappedItemsCount,
		PendingOrdersCount:    agg.PendingOrdersCount,
		PendingGrossSales:     agg.PendingGrossSales,
		PendingEscrowAmount:   agg.PendingEscrowAmount,
		PendingNetProfit:      agg.PendingNetProfit,
	}, nil
}
