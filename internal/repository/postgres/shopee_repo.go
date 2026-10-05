package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"symetra-lab-backend-v2/internal/domain/shopee"
)

type shopGORM struct {
	ID                    uint64    `gorm:"column:id;primaryKey;autoIncrement"`
	ShopID                uint64    `gorm:"column:shop_id;uniqueIndex;not null"`
	ShopName              string    `gorm:"column:shop_name"`
	Region                string    `gorm:"column:region"`
	AccessToken           string    `gorm:"column:access_token;not null"`
	RefreshToken          string    `gorm:"column:refresh_token;not null"`
	AccessTokenExpiresAt  time.Time `gorm:"column:access_token_expires_at"`
	RefreshTokenExpiresAt time.Time `gorm:"column:refresh_token_expires_at"`
	CreatedAt             time.Time `gorm:"column:created_at"`
	UpdatedAt             time.Time `gorm:"column:updated_at"`
}

func (shopGORM) TableName() string {
	return "shops"
}

type shopeeOrderMarketplaceGORM struct {
	OrderID                  uuid.UUID  `gorm:"column:order_id;primaryKey;type:uuid"`
	Channel                  string     `gorm:"column:channel"`
	ShopID                   uint64     `gorm:"column:shop_id"`
	OrderSN                  string     `gorm:"column:order_sn;uniqueIndex"`
	BuyerUserID              uint64     `gorm:"column:buyer_user_id"`
	BuyerUsername            string     `gorm:"column:buyer_username"`
	MessageToSeller          string     `gorm:"column:message_to_seller"`
	ShippingCarrier          string     `gorm:"column:shipping_carrier"`
	TrackingNumber           string     `gorm:"column:tracking_number"`
	ShipByDate               int64      `gorm:"column:ship_by_date"`
	ShipByDateTime           *time.Time `gorm:"column:ship_by_date_time"`
	CommissionFee            float64    `gorm:"column:commission_fee"`
	ServiceFee               float64    `gorm:"column:service_fee"`
	SellerTransactionFee     float64    `gorm:"column:seller_transaction_fee"`
	SellerOrderProcessingFee float64    `gorm:"column:seller_order_processing_fee"`
	SellerVoucherDiscount    float64    `gorm:"column:seller_voucher_discount"`
	TotalMarketplaceFee      float64    `gorm:"column:total_marketplace_fee"`
	EscrowAmount             float64    `gorm:"column:escrow_amount"`
	FinancialStatus          string     `gorm:"column:financial_status"`
	CreatedAt                time.Time  `gorm:"column:created_at"`
	UpdatedAt                time.Time  `gorm:"column:updated_at"`
}

func (shopeeOrderMarketplaceGORM) TableName() string {
	return "order_marketplace_details"
}

type shopeeUnifiedItemGORM struct {
	ID               uuid.UUID  `gorm:"column:id;primaryKey;type:uuid"`
	OrderID          uuid.UUID  `gorm:"column:order_id;type:uuid"`
	ChannelItemID    uint64     `gorm:"column:channel_item_id"`
	ChannelModelID   uint64     `gorm:"column:channel_model_id"`
	ProductName      string     `gorm:"column:product_name"`
	ItemSKU          string     `gorm:"column:item_sku"`
	MatchedSKU       string     `gorm:"column:matched_sku"`
	MappingStatus    string     `gorm:"column:mapping_status"`
	Quantity         int        `gorm:"column:quantity"`
	SellingPrice     float64    `gorm:"column:selling_price"`
	HPP              float64    `gorm:"column:hpp"`
	FilamentCost     float64    `gorm:"column:filament_cost"`
	ComponentCost    float64    `gorm:"column:component_cost"`
	PackagingCost    float64    `gorm:"column:packaging_cost"`
	EnergyCost       float64    `gorm:"column:energy_cost"`
	MaintenanceCost  float64    `gorm:"column:maintenance_cost"`
	DepreciationCost float64    `gorm:"column:depreciation_cost"`
	TotalCOGS        float64    `gorm:"column:total_cogs"`
	NetProfit        float64    `gorm:"column:net_profit"`
	ProductID        *uuid.UUID `gorm:"column:product_id;type:uuid"`
	CreatedAt        time.Time  `gorm:"column:created_at"`
	UpdatedAt        time.Time  `gorm:"column:updated_at"`
}

func (shopeeUnifiedItemGORM) TableName() string {
	return "order_items"
}

type shopeeUnifiedOrderGORM struct {
	ID               uuid.UUID                   `gorm:"column:id;primaryKey;type:uuid"`
	OrderNumber      string                      `gorm:"column:order_number;index"`
	CustomerName     string                      `gorm:"column:customer_name"`
	Status           string                      `gorm:"column:status"`
	Channel          string                      `gorm:"column:channel"`
	PaymentStatus    string                      `gorm:"column:payment_status"`
	GrossAmount      float64                     `gorm:"column:gross_amount"`
	ChannelFee       float64                     `gorm:"column:channel_fee"`
	NetAmount        float64                     `gorm:"column:net_amount"`
	COGSAmount       float64                     `gorm:"column:cogs_amount"`
	NetProfit        float64                     `gorm:"column:net_profit"`
	FundFilament     float64                     `gorm:"column:fund_filament"`
	FundComponent    float64                     `gorm:"column:fund_component"`
	FundPackaging    float64                     `gorm:"column:fund_packaging"`
	FundElectricity  float64                     `gorm:"column:fund_electricity"`
	FundMaintenance  float64                     `gorm:"column:fund_maintenance"`
	FundDepreciation float64                     `gorm:"column:fund_depreciation"`
	FundNetProfit    float64                     `gorm:"column:fund_net_profit"`
	CreatedAt        time.Time                   `gorm:"column:created_at"`
	UpdatedAt        time.Time                   `gorm:"column:updated_at"`
	Marketplace      *shopeeOrderMarketplaceGORM `gorm:"foreignKey:OrderID;references:ID"`
	Items            []shopeeUnifiedItemGORM     `gorm:"foreignKey:OrderID;references:ID"`
}

func (shopeeUnifiedOrderGORM) TableName() string {
	return "orders"
}

type ShopeeRepository struct {
	db *gorm.DB
}

func NewShopeeRepository(db *gorm.DB) *ShopeeRepository {
	return &ShopeeRepository{db: db}
}

func mapShopGORMToDomain(s *shopGORM) *shopee.ShopeeShop {
	if s == nil {
		return nil
	}
	return shopee.ReconstructShop(
		s.ID,
		s.ShopID,
		s.ShopName,
		s.Region,
		s.AccessToken,
		s.RefreshToken,
		s.AccessTokenExpiresAt,
		s.RefreshTokenExpiresAt,
		s.CreatedAt,
		s.UpdatedAt,
	)
}

func mapShopeeUnifiedOrderGORMToDomain(o *shopeeUnifiedOrderGORM) *shopee.ShopeeOrder {
	if o == nil {
		return nil
	}

	items := make([]shopee.ShopeeOrderItem, len(o.Items))
	for i, it := range o.Items {
		var dummyID uint64 = uint64(i + 1)
		items[i] = shopee.ReconstructOrderItem(
			dummyID,
			o.OrderNumber,
			it.ChannelItemID,
			it.ProductName,
			it.ItemSKU,
			it.ChannelModelID,
			"",
			it.ItemSKU,
			it.Quantity,
			it.SellingPrice,
			it.SellingPrice,
			it.ProductID,
			it.MatchedSKU,
			it.MappingStatus,
			it.FilamentCost,
			it.ComponentCost,
			it.PackagingCost,
			0, // machineCost
			it.TotalCOGS,
			it.NetProfit,
			it.EnergyCost,
			it.MaintenanceCost,
			it.DepreciationCost,
			it.CreatedAt,
			it.UpdatedAt,
		)
	}

	var escrow *shopee.ShopeeOrderEscrow
	if o.Marketplace != nil {
		m := o.Marketplace
		escrow = shopee.ReconstructOrderEscrow(
			o.OrderNumber,
			o.NetAmount,
			o.GrossAmount,
			m.CommissionFee,
			"",
			0,
			m.ServiceFee,
			"",
			0,
			m.SellerTransactionFee,
			m.SellerOrderProcessingFee,
			m.SellerVoucherDiscount,
			o.ChannelFee,
			o.COGSAmount,
			o.FundFilament,
			o.FundComponent,
			o.FundPackaging,
			o.FundElectricity,
			o.FundMaintenance,
			o.FundDepreciation,
			o.FundNetProfit,
			m.FinancialStatus,
			m.CreatedAt,
			m.UpdatedAt,
		)
	}

	var shopID, buyerUserID uint64
	var msg, carrier, tracking string
	var shipByDate int64
	var shipByDateTime *time.Time
	if o.Marketplace != nil {
		m := o.Marketplace
		shopID = m.ShopID
		buyerUserID = m.BuyerUserID
		msg = m.MessageToSeller
		carrier = m.ShippingCarrier
		tracking = m.TrackingNumber
		shipByDate = m.ShipByDate
		shipByDateTime = m.ShipByDateTime
	}

	return shopee.ReconstructOrder(
		o.OrderNumber,
		shopID,
		o.Status,
		buyerUserID,
		o.CustomerName,
		msg,
		shipByDate,
		shipByDateTime,
		carrier,
		tracking,
		o.GrossAmount,
		"",
		0,
		0,
		items,
		escrow,
		o.CreatedAt,
		o.UpdatedAt,
	)
}

func (r *ShopeeRepository) FindShop(ctx context.Context, shopID uint64) (*shopee.ShopeeShop, error) {
	var s shopGORM
	if err := r.db.WithContext(ctx).Where("shop_id = ?", shopID).First(&s).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, shopee.ErrShopNotFound
		}
		return nil, err
	}
	return mapShopGORMToDomain(&s), nil
}

func (r *ShopeeRepository) FindDefaultShop(ctx context.Context) (*shopee.ShopeeShop, error) {
	var s shopGORM
	if err := r.db.WithContext(ctx).First(&s).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, shopee.ErrShopNotFound
		}
		return nil, err
	}
	return mapShopGORMToDomain(&s), nil
}

func (r *ShopeeRepository) FindAllShops(ctx context.Context) ([]*shopee.ShopeeShop, error) {
	var list []shopGORM
	if err := r.db.WithContext(ctx).Find(&list).Error; err != nil {
		return nil, err
	}
	results := make([]*shopee.ShopeeShop, len(list))
	for i := range list {
		results[i] = mapShopGORMToDomain(&list[i])
	}
	return results, nil
}

func (r *ShopeeRepository) SaveShop(ctx context.Context, s *shopee.ShopeeShop) error {
	record := shopGORM{
		ShopID:                s.ShopID(),
		ShopName:              s.ShopName(),
		Region:                s.Region(),
		AccessToken:           s.AccessToken(),
		RefreshToken:          s.RefreshToken(),
		AccessTokenExpiresAt:  s.AccessTokenExpiresAt(),
		RefreshTokenExpiresAt: s.RefreshTokenExpiresAt(),
		CreatedAt:             s.CreatedAt(),
		UpdatedAt:             s.UpdatedAt(),
	}

	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "shop_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"access_token", "refresh_token", "access_token_expires_at", "refresh_token_expires_at", "updated_at"}),
	}).Create(&record).Error
}

func (r *ShopeeRepository) FindOrder(ctx context.Context, orderSN string) (*shopee.ShopeeOrder, error) {
	var o shopeeUnifiedOrderGORM
	err := r.db.WithContext(ctx).
		Preload("Items").
		Preload("Marketplace").
		Where("order_number = ? AND channel = 'SHOPEE'", orderSN).
		First(&o).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, shopee.ErrOrderNotFound
		}
		return nil, err
	}
	return mapShopeeUnifiedOrderGORMToDomain(&o), nil
}

func (r *ShopeeRepository) FindOrders(ctx context.Context, status, carrier, search string) ([]*shopee.ShopeeOrder, error) {
	query := r.db.WithContext(ctx).Model(&shopeeUnifiedOrderGORM{}).
		Preload("Items").
		Preload("Marketplace").
		Where("channel = 'SHOPEE'")

	if status != "" {
		query = query.Where("status = ?", status)
	}
	if carrier != "" {
		query = query.Joins("JOIN order_marketplace_details ON order_marketplace_details.order_id = orders.id").
			Where("order_marketplace_details.shipping_carrier ILIKE ?", "%"+carrier+"%")
	}
	if search != "" {
		query = query.Where("orders.order_number ILIKE ? OR orders.customer_name ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	var list []shopeeUnifiedOrderGORM
	if err := query.Order("orders.created_at DESC").Find(&list).Error; err != nil {
		return nil, err
	}

	results := make([]*shopee.ShopeeOrder, len(list))
	for i := range list {
		results[i] = mapShopeeUnifiedOrderGORMToDomain(&list[i])
	}
	return results, nil
}

func (r *ShopeeRepository) SaveOrder(ctx context.Context, o *shopee.ShopeeOrder) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existingOrder shopeeUnifiedOrderGORM
		err := tx.Where("order_number = ?", o.OrderSN()).First(&existingOrder).Error
		orderID := existingOrder.ID
		if err != nil || orderID == uuid.Nil {
			orderID = uuid.New()
		}

		var grossAmount, netAmount, channelFee, cogsAmount, netProfit float64
		var fFilament, fComponent, fPackaging, fElectricity, fMaintenance, fDepreciation, fNetProfit float64
		paymentStatus := "UNPAID"

		if o.Escrow() != nil {
			e := o.Escrow()
			grossAmount = e.SellingPrice()
			netAmount = e.EscrowAmount()
			channelFee = e.TotalMarketplaceFee()
			fFilament = e.KasFilamen()
			fComponent = e.KasKomponen()
			fPackaging = e.KasPacking()
			fElectricity = e.KasListrik()
			fMaintenance = e.KasMaintenance()
			fDepreciation = e.KasDepresiasi()
			cogsAmount = fFilament + fComponent + fPackaging + fElectricity + fMaintenance + fDepreciation
			// Selisih antara pencairan escrow dengan pos biaya dialokasikan ke laba bersih:
			fNetProfit = netAmount - cogsAmount
			netProfit = fNetProfit
			if e.FinancialStatus() == "RELEASED" {
				paymentStatus = "PAID"
			}
		} else {
			grossAmount = o.TotalAmount()
		}

		orderCreatedAt := o.CreatedAt()
		if o.CreateTimeShopee() > 0 {
			orderCreatedAt = time.Unix(o.CreateTimeShopee(), 0)
		}
		orderUpdatedAt := o.UpdatedAt()
		if o.UpdateTimeShopee() > 0 {
			orderUpdatedAt = time.Unix(o.UpdateTimeShopee(), 0)
		}

		orderRecord := shopeeUnifiedOrderGORM{
			ID:               orderID,
			OrderNumber:      o.OrderSN(),
			CustomerName:     o.BuyerUsername(),
			Status:           o.OrderStatus(),
			Channel:          "SHOPEE",
			PaymentStatus:    paymentStatus,
			GrossAmount:      grossAmount,
			ChannelFee:       channelFee,
			NetAmount:        netAmount,
			COGSAmount:       cogsAmount,
			NetProfit:        netProfit,
			FundFilament:     fFilament,
			FundComponent:    fComponent,
			FundPackaging:    fPackaging,
			FundElectricity:  fElectricity,
			FundMaintenance:  fMaintenance,
			FundDepreciation: fDepreciation,
			FundNetProfit:    fNetProfit,
			CreatedAt:        orderCreatedAt,
			UpdatedAt:        orderUpdatedAt,
		}

		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},
			DoUpdates: clause.AssignmentColumns([]string{"customer_name", "status", "payment_status", "gross_amount", "channel_fee", "net_amount", "cogs_amount", "net_profit", "fund_filament", "fund_component", "fund_packaging", "fund_electricity", "fund_maintenance", "fund_depreciation", "fund_net_profit", "updated_at"}),
		}).Create(&orderRecord).Error; err != nil {
			return err
		}

		// Marketplace details
		mDetails := shopeeOrderMarketplaceGORM{
			OrderID:         orderID,
			Channel:         "SHOPEE",
			ShopID:          o.ShopID(),
			OrderSN:         o.OrderSN(),
			BuyerUserID:     o.BuyerUserID(),
			BuyerUsername:   o.BuyerUsername(),
			MessageToSeller: o.MessageToSeller(),
			ShippingCarrier: o.ShippingCarrier(),
			TrackingNumber:  o.TrackingNumber(),
			ShipByDate:      o.ShipByDate(),
			ShipByDateTime:  o.ShipByDateTime(),
			CreatedAt:       orderCreatedAt,
			UpdatedAt:       orderUpdatedAt,
		}
		if o.Escrow() != nil {
			e := o.Escrow()
			mDetails.CommissionFee = e.CommissionFee()
			mDetails.ServiceFee = e.ServiceFee()
			mDetails.SellerTransactionFee = e.SellerTransactionFee()
			mDetails.SellerOrderProcessingFee = e.SellerOrderProcessingFee()
			mDetails.SellerVoucherDiscount = e.SellerVoucherDiscount()
			mDetails.TotalMarketplaceFee = e.TotalMarketplaceFee()
			mDetails.EscrowAmount = e.EscrowAmount()
			mDetails.FinancialStatus = e.FinancialStatus()
		}

		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "order_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"shipping_carrier", "tracking_number", "ship_by_date", "ship_by_date_time", "commission_fee", "service_fee", "seller_transaction_fee", "seller_order_processing_fee", "seller_voucher_discount", "total_marketplace_fee", "escrow_amount", "financial_status", "updated_at"}),
		}).Create(&mDetails).Error; err != nil {
			return err
		}

		// Items
		for _, item := range o.Items() {
			var existingItem shopeeUnifiedItemGORM
			_ = tx.Where("order_id = ? AND channel_item_id = ?", orderID, item.ItemID()).First(&existingItem).Error
			itemID := existingItem.ID
			if itemID == uuid.Nil {
				itemID = uuid.New()
			}

			itemG := shopeeUnifiedItemGORM{
				ID:               itemID,
				OrderID:          orderID,
				ChannelItemID:    item.ItemID(),
				ChannelModelID:   item.ModelID(),
				ProductName:      item.ItemName(),
				ItemSKU:          item.ItemSKU(),
				MatchedSKU:       item.MatchedSKU(),
				MappingStatus:    item.MappingStatus(),
				Quantity:         item.Quantity(),
				SellingPrice:     item.DiscountedPrice(),
				HPP:              item.BaseHPP(),
				TotalCOGS:        item.BaseHPP(),
				FilamentCost:     item.FilamentCost(),
				ComponentCost:    item.HardwareCost(),
				PackagingCost:    item.PackagingCost(),
				EnergyCost:       item.ElectricityCost(),
				MaintenanceCost:  item.MaintenanceCost(),
				DepreciationCost: item.DepreciationCost(),
				NetProfit:        item.NetProfit(),
				ProductID:        item.ProductID(),
				CreatedAt:        orderCreatedAt,
				UpdatedAt:        orderUpdatedAt,
			}

			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "id"}},
				DoUpdates: clause.AssignmentColumns([]string{"product_id", "matched_sku", "mapping_status", "selling_price", "hpp", "total_cogs", "filament_cost", "component_cost", "packaging_cost", "energy_cost", "maintenance_cost", "depreciation_cost", "net_profit", "updated_at"}),
			}).Create(&itemG).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *ShopeeRepository) LinkSKU(ctx context.Context, itemID, modelID uint64, productID uuid.UUID) error {
	updates := map[string]interface{}{
		"product_id":     productID,
		"mapping_status": "MAPPED",
		"updated_at":     time.Now(),
	}

	query := r.db.WithContext(ctx).Table("order_items").Where("channel_item_id = ?", itemID)
	if modelID > 0 {
		query = query.Where("channel_model_id = ?", modelID)
	}

	return query.Updates(updates).Error
}

func (r *ShopeeRepository) GetCashflowSummary(ctx context.Context) (*shopee.CashflowSummary, error) {
	type AggResult struct {
		TotalOrders          int64   `gorm:"column:total_orders"`
		TotalGrossSales      float64 `gorm:"column:total_gross_sales"`
		TotalMarketplaceFees float64 `gorm:"column:total_marketplace_fees"`
		TotalEscrowNetIn     float64 `gorm:"column:total_escrow_net_in"`
		TotalHPP             float64 `gorm:"column:total_hpp"`
		FundFilament         float64 `gorm:"column:fund_filament"`
		FundComponent        float64 `gorm:"column:fund_component"`
		FundPackaging        float64 `gorm:"column:fund_packaging"`
		FundElectricity      float64 `gorm:"column:fund_electricity"`
		FundMaintenance      float64 `gorm:"column:fund_maintenance"`
		FundDepreciation     float64 `gorm:"column:fund_depreciation"`
		FundNetProfit        float64 `gorm:"column:fund_net_profit"`
		UnmappedCount        int64   `gorm:"column:unmapped_count"`
	}

	var agg AggResult
	sql := `
		SELECT 
			COALESCE(COUNT(id), 0) AS total_orders,
			COALESCE(SUM(gross_amount), 0) AS total_gross_sales,
			COALESCE(SUM(channel_fee), 0) AS total_marketplace_fees,
			COALESCE(SUM(net_amount), 0) AS total_escrow_net_in,
			COALESCE(SUM(cogs_amount), 0) AS total_hpp,
			COALESCE(SUM(fund_filament), 0) AS fund_filament,
			COALESCE(SUM(fund_component), 0) AS fund_component,
			COALESCE(SUM(fund_packaging), 0) AS fund_packaging,
			COALESCE(SUM(fund_electricity), 0) AS fund_electricity,
			COALESCE(SUM(fund_maintenance), 0) AS fund_maintenance,
			COALESCE(SUM(fund_depreciation), 0) AS fund_depreciation,
			COALESCE(SUM(fund_net_profit), 0) AS fund_net_profit,
			(SELECT COUNT(*) FROM order_items oi JOIN orders o ON o.id = oi.order_id WHERE o.channel = 'SHOPEE' AND oi.mapping_status = 'UNMAPPED') AS unmapped_count
		FROM orders
		WHERE channel = 'SHOPEE' AND status = 'COMPLETED'
	`

	if err := r.db.WithContext(ctx).Raw(sql).Scan(&agg).Error; err != nil {
		return nil, err
	}

	avgMargin := 0.0
	if agg.TotalGrossSales > 0 {
		avgMargin = (agg.FundNetProfit / agg.TotalGrossSales) * 100.0
	}

	return &shopee.CashflowSummary{
		TotalCompletedOrders: agg.TotalOrders,
		TotalOrders:          agg.TotalOrders,
		TotalGrossSales:      agg.TotalGrossSales,
		TotalMarketplaceFees: agg.TotalMarketplaceFees,
		TotalEscrowNetIn:     agg.TotalEscrowNetIn,
		TotalHPP:             agg.TotalHPP,
		KasFilamen:           agg.FundFilament,
		KasKomponen:          agg.FundComponent,
		KasPacking:           agg.FundPackaging,
		KasListrik:           agg.FundElectricity,
		KasMaintenance:       agg.FundMaintenance,
		KasDepresiasi:        agg.FundDepreciation,
		KasLabaBersih:        agg.FundNetProfit,
		AverageProfitMargin:  avgMargin,
		UnmappedItemsCount:   agg.UnmappedCount,
	}, nil
}

func (r *ShopeeRepository) RecalculateFinances(ctx context.Context) (int, error) {
	res := r.db.WithContext(ctx).Exec(`
		UPDATE orders 
		SET fund_net_profit = net_amount - (fund_filament + fund_component + fund_packaging + fund_electricity + fund_maintenance + fund_depreciation),
		    net_profit = net_amount - (fund_filament + fund_component + fund_packaging + fund_electricity + fund_maintenance + fund_depreciation),
		    cogs_amount = (fund_filament + fund_component + fund_packaging + fund_electricity + fund_maintenance + fund_depreciation),
		    updated_at = NOW() 
		WHERE payment_status = 'PAID'
	`)
	if res.Error != nil {
		return 0, res.Error
	}
	return int(res.RowsAffected), nil
}
