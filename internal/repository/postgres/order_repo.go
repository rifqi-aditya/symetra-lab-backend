package postgres

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"symetra-lab-backend-v2/internal/domain/order"
)

type orderGORM struct {
	ID                 uuid.UUID                   `gorm:"column:id;primaryKey;type:uuid"`
	UserID             uuid.UUID                   `gorm:"column:user_id;type:uuid"`
	OrderNumber        string                      `gorm:"column:order_number"`
	CustomerName       string                      `gorm:"column:customer_name"`
	CustomerContact    string                      `gorm:"column:customer_contact"`
	Channel            string                      `gorm:"column:channel"`
	GrossAmount        float64                     `gorm:"column:gross_amount"`
	ChannelFee         float64                     `gorm:"column:channel_fee"`
	NetAmount          float64                     `gorm:"column:net_amount"`
	CogsAmount         float64                     `gorm:"column:cogs_amount"`
	NetProfit          float64                     `gorm:"column:net_profit"`
	Status             string                      `gorm:"column:status"`
	Notes              string                      `gorm:"column:notes"`
	PaymentStatus      string                      `gorm:"column:payment_status"`
	PaidAt             *time.Time                  `gorm:"column:paid_at"`
	StartedAt          *time.Time                  `gorm:"column:started_at"`
	CompletedAt        *time.Time                  `gorm:"column:completed_at"`
	CreatedAt          time.Time                   `gorm:"column:created_at"`
	UpdatedAt          time.Time                   `gorm:"column:updated_at"`
	Items              []orderItemGORM             `gorm:"foreignKey:OrderID;references:ID;constraint:OnDelete:CASCADE"`
	MarketplaceDetails *orderMarketplaceDetailGORM `gorm:"foreignKey:OrderID;references:ID"`
	Allocations        []orderAllocationGORM       `gorm:"foreignKey:OrderID;references:ID;constraint:OnDelete:CASCADE"`
}

func (orderGORM) TableName() string {
	return "orders"
}

type orderAllocationGORM struct {
	ID            uuid.UUID `gorm:"column:id;primaryKey;type:uuid"`
	OrderID       uuid.UUID `gorm:"column:order_id;type:uuid"`
	CashAccountID uuid.UUID `gorm:"column:cash_account_id;type:uuid"`
	AccountCode   string    `gorm:"column:account_code"`
	Amount        float64   `gorm:"column:amount"`
	CreatedAt     time.Time `gorm:"column:created_at"`
}

func (orderAllocationGORM) TableName() string {
	return "order_allocations"
}

type orderMarketplaceDetailGORM struct {
	OrderID         uuid.UUID  `gorm:"column:order_id;primaryKey;type:uuid"`
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
	SKU              string       `gorm:"column:sku"`
	Quantity         int          `gorm:"column:quantity"`
	SellingPrice     float64      `gorm:"column:selling_price"`
	HPP              float64      `gorm:"column:hpp"`
	WeightGrams      float64      `gorm:"column:weight_grams"`
	PrintTimeHours   float64      `gorm:"column:print_time_hours"`
	FilamentCost     float64      `gorm:"column:filament_cost"`
	ComponentCost    float64      `gorm:"column:component_cost"`
	PackagingCost    float64      `gorm:"column:packaging_cost"`
	EnergyCost       float64      `gorm:"column:energy_cost"`
	MaintenanceCost  float64      `gorm:"column:maintenance_cost"`
	DepreciationCost float64      `gorm:"column:depreciation_cost"`
	ChannelItemID    int64        `gorm:"column:channel_item_id"`
	ChannelModelID   int64        `gorm:"column:channel_model_id"`
	MappingStatus    string       `gorm:"column:mapping_status"`
	CreatedAt        time.Time    `gorm:"column:created_at"`
	UpdatedAt        time.Time    `gorm:"column:updated_at"`
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
		if it.Product != nil && it.Product.Parent != nil {
			if it.Product.Parent.Name != "" {
				prodName = it.Product.Parent.Name
			}
			thumbURL = it.Product.Parent.ThumbnailURL
		}

		itemSKU := it.SKU

		itemObj := order.ReconstructOrderItemFull(
			it.ID,
			it.OrderID,
			it.ProductID,
			prodName,
			itemSKU,
			it.Quantity,
			it.SellingPrice,
			it.HPP,
			it.WeightGrams,
			it.PrintTimeHours,
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

	var fundFilament, fundEmergency, fundComponent, fundPackaging, fundElectricity, fundMaintenance, fundDepreciation, fundNetProfit float64
	for _, a := range g.Allocations {
		switch a.AccountCode {
		case "FILAMENT":
			fundFilament = a.Amount
		case "EMERGENCY":
			fundEmergency = a.Amount
		case "COMPONENT":
			fundComponent = a.Amount
		case "PACKAGING":
			fundPackaging = a.Amount
		case "ELECTRICITY":
			fundElectricity = a.Amount
		case "MAINTENANCE":
			fundMaintenance = a.Amount
		case "DEPRECIATION":
			fundDepreciation = a.Amount
		case "NET_PROFIT":
			fundNetProfit = a.Amount
		}
	}

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
		fundFilament,
		fundEmergency,
		fundComponent,
		fundPackaging,
		fundElectricity,
		fundMaintenance,
		fundDepreciation,
		fundNetProfit,
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

	if g.PaidAt != nil {
		domainOrder.SetPaidAt(g.PaidAt)
	}

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
		Preload("Items.Product.Parent").
		Preload("MarketplaceDetails").
		Preload("Allocations").
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
		Preload("Items.Product.Parent").
		Preload("MarketplaceDetails").
		Preload("Allocations").
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
		Preload("Items.Product.Parent").
		Preload("MarketplaceDetails").
		Preload("Allocations").
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

func upsertOrderAllocations(tx *gorm.DB, orderID uuid.UUID, createdAt time.Time, allocs map[string]float64) error {
	var accounts []cashAccountGORM
	if err := tx.Select("id, code").Find(&accounts).Error; err != nil {
		return err
	}
	accMap := make(map[string]uuid.UUID)
	for _, a := range accounts {
		accMap[a.Code] = a.ID
	}

	for code, amt := range allocs {
		accID, ok := accMap[code]
		if !ok {
			continue
		}
		record := orderAllocationGORM{
			ID:            uuid.New(),
			OrderID:       orderID,
			CashAccountID: accID,
			AccountCode:   code,
			Amount:        amt,
			CreatedAt:     createdAt,
		}
		err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "order_id"}, {Name: "cash_account_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"amount"}),
		}).Create(&record).Error
		if err != nil {
			return err
		}
	}
	return nil
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
			Status:           o.Status(),
			Notes:            o.Notes(),
			PaymentStatus:    o.PaymentStatus(),
			PaidAt:           o.PaidAt(),
			StartedAt:        o.StartedAt(),
			CompletedAt:      o.CompletedAt(),
			CreatedAt:        o.CreatedAt(),
			UpdatedAt:        o.UpdatedAt(),
		}

		if err := tx.Create(&g).Error; err != nil {
			return err
		}

		for _, it := range o.Items() {
			itemSKU := it.SKU()
			itemGORM := orderItemGORM{
				ID:               it.ID(),
				OrderID:          o.ID(),
				ProductID:        it.ProductID(),
				ProductName:      it.ProductName(),
				SKU:              itemSKU,
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
				CreatedAt:        it.CreatedAt(),
			}
			if err := tx.Create(&itemGORM).Error; err != nil {
				return err
			}
		}

		allocMap := map[string]float64{
			"FILAMENT":     o.FundFilament(),
			"EMERGENCY":    o.FundEmergency(),
			"COMPONENT":    o.FundComponent(),
			"PACKAGING":    o.FundPackaging(),
			"ELECTRICITY":  o.FundElectricity(),
			"MAINTENANCE":  o.FundMaintenance(),
			"DEPRECIATION": o.FundDepreciation(),
			"NET_PROFIT":   o.FundNetProfit(),
		}
		if err := upsertOrderAllocations(tx, o.ID(), o.CreatedAt(), allocMap); err != nil {
			return err
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
	_ = SyncCashAccountLedger(ctx, r.db)
	return nil
}

func (r *OrderRepository) UpdatePaymentStatus(ctx context.Context, userID, id uuid.UUID, paymentStatus string) error {
	now := time.Now()
	updates := map[string]interface{}{
		"payment_status": paymentStatus,
		"updated_at":     now,
	}
	if paymentStatus == "PAID" {
		updates["paid_at"] = &now
	} else if paymentStatus == "UNPAID" {
		updates["paid_at"] = nil
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
) (*order.OrderAllocationSummary, error) {
	type aggResult struct {
		TotalOrders         int64   `gorm:"column:total_orders"`
		TotalGrossSales     float64 `gorm:"column:total_gross_sales"`
		TotalChannelFees    float64 `gorm:"column:total_channel_fees"`
		TotalNetRevenue     float64 `gorm:"column:total_net_revenue"`
		TotalCOGS           float64 `gorm:"column:total_cogs"`
		TotalNetProfit      float64 `gorm:"column:total_net_profit"`
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
			COALESCE((
				SELECT COUNT(*) 
				FROM order_items oi
				JOIN orders o2 ON o2.id = oi.order_id
				WHERE o2.user_id = ? AND o2.status NOT IN ('CANCELLED', 'UNPAID') AND oi.mapping_status = 'UNMAPPED'
			), 0) AS unmapped_items_count,
			COALESCE(COUNT(CASE WHEN payment_status != 'PAID' AND channel = 'SHOPEE' THEN 1 END), 0) AS pending_orders_count,
			COALESCE(SUM(CASE WHEN payment_status != 'PAID' AND channel = 'SHOPEE' THEN (CASE WHEN gross_amount > 0 THEN gross_amount ELSE (net_amount + channel_fee) END) ELSE 0 END), 0) AS pending_gross_sales,
			COALESCE(SUM(CASE WHEN payment_status != 'PAID' AND channel = 'SHOPEE' THEN net_amount ELSE 0 END), 0) AS pending_escrow_amount,
			COALESCE(SUM(CASE WHEN payment_status != 'PAID' AND channel = 'SHOPEE' THEN net_profit ELSE 0 END), 0) AS pending_net_profit
		FROM orders
		WHERE user_id = ? 
		  AND status NOT IN ('CANCELLED', 'UNPAID')
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
	var fundFilament, fundEmergency, fundComponent, fundPackaging, fundElectricity, fundMaintenance, fundDepreciation, fundNetProfit float64
	var spentFilament, spentEmergency, spentComponent, spentPackaging, spentElectricity, spentMaintenance, spentDepreciation, spentNetProfit float64
	var allocFilament, allocEmergency, allocComponent, allocPackaging, allocElectricity, allocMaintenance, allocDepreciation, allocNetProfit float64

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
				case "EMERGENCY":
					fundEmergency = acc.CurrentBalance
					spentEmergency = acc.SpentAmount
					allocEmergency = acc.AllocatedAmount
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
			case "EMERGENCY_FUND":
				spentEmergency += e.TotalSpent
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
			case "REIMBURSEMENT":
				// Pelunasan talangan tidak memotong kas lagi karena anggaran belanja sudah dipotong saat transaksi dicatat
				continue
			default:
				spentNetProfit += e.TotalSpent
			}
		}

		type allocRow struct {
			AccountCode string  `gorm:"column:account_code"`
			Total       float64 `gorm:"column:total"`
		}
		var allocRows []allocRow
		allocQuery := `
			SELECT 
				oa.account_code, 
				COALESCE(SUM(oa.amount), 0) AS total
			FROM order_allocations oa
			JOIN orders o ON o.id = oa.order_id
			WHERE o.user_id = ? 
			  AND o.status NOT IN ('CANCELLED', 'UNPAID')
			  AND o.payment_status = 'PAID'
			  AND (? = 'ALL' OR o.channel = ?)
			  AND (?::timestamptz IS NULL OR o.created_at >= ?)
			  AND (?::timestamptz IS NULL OR o.created_at <= ?)
			GROUP BY oa.account_code
		`
		_ = r.db.WithContext(ctx).Raw(
			allocQuery,
			userID,
			ch, ch,
			dateFrom, dateFrom,
			dateTo, dateTo,
		).Scan(&allocRows).Error

		for _, ar := range allocRows {
			switch ar.AccountCode {
			case "FILAMENT":
				allocFilament = ar.Total
			case "EMERGENCY":
				allocEmergency = ar.Total
			case "COMPONENT":
				allocComponent = ar.Total
			case "PACKAGING":
				allocPackaging = ar.Total
			case "ELECTRICITY":
				allocElectricity = ar.Total
			case "MAINTENANCE":
				allocMaintenance = ar.Total
			case "DEPRECIATION":
				allocDepreciation = ar.Total
			case "NET_PROFIT":
				allocNetProfit = ar.Total
			}
		}

		fundFilament = allocFilament - spentFilament
		fundEmergency = allocEmergency - spentEmergency
		fundComponent = allocComponent - spentComponent
		fundPackaging = allocPackaging - spentPackaging
		fundElectricity = allocElectricity - spentElectricity
		fundMaintenance = allocMaintenance - spentMaintenance
		fundDepreciation = allocDepreciation - spentDepreciation
		fundNetProfit = allocNetProfit - spentNetProfit
	}

	return &order.OrderAllocationSummary{
		TotalOrders:           agg.TotalOrders,
		TotalGrossSales:       agg.TotalGrossSales,
		TotalChannelFees:      agg.TotalChannelFees,
		TotalNetRevenue:       agg.TotalNetRevenue,
		TotalCOGS:             agg.TotalCOGS,
		TotalNetProfit:        agg.TotalNetProfit,

		FundFilament:          fundFilament,
		FundEmergency:         fundEmergency,
		FundComponent:         fundComponent,
		FundPackaging:         fundPackaging,
		FundElectricity:       fundElectricity,
		FundMaintenance:       fundMaintenance,
		FundDepreciation:      fundDepreciation,
		FundNetProfit:         fundNetProfit,

		AllocatedFilament:     allocFilament,
		AllocatedEmergency:    allocEmergency,
		AllocatedComponent:    allocComponent,
		AllocatedPackaging:    allocPackaging,
		AllocatedElectricity:  allocElectricity,
		AllocatedMaintenance:  allocMaintenance,
		AllocatedDepreciation: allocDepreciation,
		AllocatedNetProfit:    allocNetProfit,

		SpentFilament:         spentFilament,
		SpentEmergency:        spentEmergency,
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

func (r *OrderRepository) GetNextManualOrderNumber(ctx context.Context, userID uuid.UUID, t time.Time) (string, error) {
	prefix := fmt.Sprintf("ORD-%s-", t.Format("200601"))
	var lastOrderNumber string
	err := r.db.WithContext(ctx).
		Model(&orderGORM{}).
		Where("user_id = ? AND order_number LIKE ?", userID, prefix+"%").
		Order("order_number DESC").
		Limit(1).
		Pluck("order_number", &lastOrderNumber).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return "", err
	}

	nextSeq := 1
	if lastOrderNumber != "" {
		parts := strings.Split(lastOrderNumber, "-")
		if len(parts) == 3 {
			if seq, parseErr := strconv.Atoi(parts[2]); parseErr == nil {
				nextSeq = seq + 1
			}
		}
	}
	return fmt.Sprintf("%s%04d", prefix, nextSeq), nil
}
