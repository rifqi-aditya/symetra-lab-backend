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

type shopeeOrderEscrowGORM struct {
	OrderSN                  string    `gorm:"column:order_sn;primaryKey"`
	EscrowAmount             float64   `gorm:"column:escrow_amount"`
	SellingPrice             float64   `gorm:"column:selling_price"`
	CommissionFee            float64   `gorm:"column:commission_fee"`
	CommissionRuleName       string    `gorm:"column:commission_rule_name"`
	CommissionPercentage     float64   `gorm:"column:commission_percentage"`
	ServiceFee               float64   `gorm:"column:service_fee"`
	ServiceRuleName          string    `gorm:"column:service_rule_name"`
	ServicePercentage        float64   `gorm:"column:service_percentage"`
	SellerTransactionFee     float64   `gorm:"column:seller_transaction_fee"`
	SellerOrderProcessingFee float64   `gorm:"column:seller_order_processing_fee"`
	SellerVoucherDiscount    float64   `gorm:"column:seller_voucher_discount"`
	TotalMarketplaceFee      float64   `gorm:"column:total_marketplace_fee"`
	TotalHPP                 float64   `gorm:"column:total_hpp"`
	KasFilamen               float64   `gorm:"column:kas_filamen"`
	KasKomponen              float64   `gorm:"column:kas_komponen"`
	KasPacking               float64   `gorm:"column:kas_packing"`
	KasListrik               float64   `gorm:"column:kas_listrik"`
	KasMaintenance           float64   `gorm:"column:kas_maintenance"`
	KasDepresiasi            float64   `gorm:"column:kas_depresiasi"`
	KasLabaBersih            float64   `gorm:"column:kas_laba_bersih"`
	FinancialStatus          string    `gorm:"column:financial_status"`
	CreatedAt                time.Time `gorm:"column:created_at"`
	UpdatedAt                time.Time `gorm:"column:updated_at"`
}

func (shopeeOrderEscrowGORM) TableName() string {
	return "shopee_order_escrows"
}

type shopeeOrderItemTableGORM struct {
	ID               uint64     `gorm:"column:id;primaryKey;autoIncrement"`
	OrderSN          string     `gorm:"column:order_sn;not null;index"`
	ItemID           uint64     `gorm:"column:item_id"`
	ItemName         string     `gorm:"column:item_name"`
	ItemSKU          string     `gorm:"column:item_sku"`
	ModelID          uint64     `gorm:"column:model_id"`
	ModelName        string     `gorm:"column:model_name"`
	ModelSKU         string     `gorm:"column:model_sku"`
	Quantity         int        `gorm:"column:quantity"`
	OriginalPrice    float64    `gorm:"column:original_price"`
	DiscountedPrice  float64    `gorm:"column:discounted_price"`
	ProductID        *uuid.UUID `gorm:"column:product_id;type:uuid"`
	MatchedSKU       string     `gorm:"column:matched_sku"`
	MappingStatus    string     `gorm:"column:mapping_status"`
	FilamentCost     float64    `gorm:"column:filament_cost"`
	HardwareCost     float64    `gorm:"column:hardware_cost"`
	PackagingCost    float64    `gorm:"column:packaging_cost"`
	MachineCost      float64    `gorm:"column:machine_cost"`
	BaseHPP          float64    `gorm:"column:base_hpp"`
	NetProfit        float64    `gorm:"column:net_profit"`
	ElectricityCost  float64    `gorm:"column:electricity_cost"`
	MaintenanceCost  float64    `gorm:"column:maintenance_cost"`
	DepreciationCost float64    `gorm:"column:depreciation_cost"`
	CreatedAt        time.Time  `gorm:"column:created_at"`
	UpdatedAt        time.Time  `gorm:"column:updated_at"`
}

func (shopeeOrderItemTableGORM) TableName() string {
	return "shopee_order_items"
}

type shopeeOrderTableGORM struct {
	OrderSN           string                      `gorm:"column:order_sn;primaryKey"`
	ShopID            uint64                      `gorm:"column:shop_id;not null"`
	OrderStatus       string                      `gorm:"column:order_status;not null"`
	BuyerUserID       uint64                      `gorm:"column:buyer_user_id"`
	BuyerUsername     string                      `gorm:"column:buyer_username"`
	MessageToSeller   string                      `gorm:"column:message_to_seller"`
	ShipByDate        int64                       `gorm:"column:ship_by_date"`
	ShipByDateTime    *time.Time                  `gorm:"column:ship_by_date_time"`
	ShippingCarrier   string                      `gorm:"column:shipping_carrier"`
	TrackingNumber    string                      `gorm:"column:tracking_number"`
	TotalAmount       float64                     `gorm:"column:total_amount"`
	BuyerCancelReason string                      `gorm:"column:buyer_cancel_reason"`
	CreateTimeShopee  int64                       `gorm:"column:create_time_shopee"`
	UpdateTimeShopee  int64                       `gorm:"column:update_time_shopee"`
	CreatedAt         time.Time                   `gorm:"column:created_at"`
	UpdatedAt         time.Time                   `gorm:"column:updated_at"`
	Items             []shopeeOrderItemTableGORM  `gorm:"foreignKey:OrderSN;references:OrderSN;constraint:OnDelete:CASCADE"`
	Escrow            *shopeeOrderEscrowGORM      `gorm:"foreignKey:OrderSN;references:OrderSN;constraint:OnDelete:CASCADE"`
}

func (shopeeOrderTableGORM) TableName() string {
	return "shopee_orders"
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

func mapShopeeOrderGORMToDomain(o *shopeeOrderTableGORM) *shopee.ShopeeOrder {
	if o == nil {
		return nil
	}

	items := make([]shopee.ShopeeOrderItem, len(o.Items))
	for i, it := range o.Items {
		items[i] = shopee.ReconstructOrderItem(
			it.ID,
			it.OrderSN,
			it.ItemID,
			it.ItemName,
			it.ItemSKU,
			it.ModelID,
			it.ModelName,
			it.ModelSKU,
			it.Quantity,
			it.OriginalPrice,
			it.DiscountedPrice,
			it.ProductID,
			it.MatchedSKU,
			it.MappingStatus,
			it.FilamentCost,
			it.HardwareCost,
			it.PackagingCost,
			it.MachineCost,
			it.BaseHPP,
			it.NetProfit,
			it.ElectricityCost,
			it.MaintenanceCost,
			it.DepreciationCost,
			it.CreatedAt,
			it.UpdatedAt,
		)
	}

	var escrow *shopee.ShopeeOrderEscrow
	if o.Escrow != nil {
		e := o.Escrow
		escrow = shopee.ReconstructOrderEscrow(
			e.OrderSN,
			e.EscrowAmount,
			e.SellingPrice,
			e.CommissionFee,
			e.CommissionRuleName,
			e.CommissionPercentage,
			e.ServiceFee,
			e.ServiceRuleName,
			e.ServicePercentage,
			e.SellerTransactionFee,
			e.SellerOrderProcessingFee,
			e.SellerVoucherDiscount,
			e.TotalMarketplaceFee,
			e.TotalHPP,
			e.KasFilamen,
			e.KasKomponen,
			e.KasPacking,
			e.KasListrik,
			e.KasMaintenance,
			e.KasDepresiasi,
			e.KasLabaBersih,
			e.FinancialStatus,
			e.CreatedAt,
			e.UpdatedAt,
		)
	}

	return shopee.ReconstructOrder(
		o.OrderSN,
		o.ShopID,
		o.OrderStatus,
		o.BuyerUserID,
		o.BuyerUsername,
		o.MessageToSeller,
		o.ShipByDate,
		o.ShipByDateTime,
		o.ShippingCarrier,
		o.TrackingNumber,
		o.TotalAmount,
		o.BuyerCancelReason,
		o.CreateTimeShopee,
		o.UpdateTimeShopee,
		items,
		escrow,
		o.CreatedAt,
		o.UpdatedAt,
	)
}

func (r *ShopeeRepository) FindShop(ctx context.Context, shopID uint64) (*shopee.ShopeeShop, error) {
	var s shopGORM
	err := r.db.WithContext(ctx).Where("shop_id = ?", shopID).First(&s).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, shopee.ErrShopNotFound
		}
		return nil, err
	}
	return mapShopGORMToDomain(&s), nil
}

func (r *ShopeeRepository) FindDefaultShop(ctx context.Context) (*shopee.ShopeeShop, error) {
	var s shopGORM
	err := r.db.WithContext(ctx).Order("created_at ASC").First(&s).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, shopee.ErrShopNotFound
		}
		return nil, err
	}
	return mapShopGORMToDomain(&s), nil
}

func (r *ShopeeRepository) FindAllShops(ctx context.Context) ([]*shopee.ShopeeShop, error) {
	var list []shopGORM
	err := r.db.WithContext(ctx).Order("created_at ASC").Find(&list).Error
	if err != nil {
		return nil, err
	}
	results := make([]*shopee.ShopeeShop, len(list))
	for i := range list {
		results[i] = mapShopGORMToDomain(&list[i])
	}
	return results, nil
}

func (r *ShopeeRepository) SaveShop(ctx context.Context, s *shopee.ShopeeShop) error {
	m := shopGORM{
		ID:                    s.ID(),
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
		DoUpdates: clause.AssignmentColumns([]string{"shop_name", "region", "access_token", "refresh_token", "access_token_expires_at", "refresh_token_expires_at", "updated_at"}),
	}).Create(&m).Error
}

func (r *ShopeeRepository) FindOrder(ctx context.Context, orderSN string) (*shopee.ShopeeOrder, error) {
	var o shopeeOrderTableGORM
	err := r.db.WithContext(ctx).
		Preload("Items").
		Preload("Escrow").
		Where("order_sn = ?", orderSN).
		First(&o).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, shopee.ErrOrderNotFound
		}
		return nil, err
	}
	return mapShopeeOrderGORMToDomain(&o), nil
}

func (r *ShopeeRepository) FindOrders(ctx context.Context, status, carrier, search string) ([]*shopee.ShopeeOrder, error) {
	query := r.db.WithContext(ctx).Model(&shopeeOrderTableGORM{}).
		Preload("Items").
		Preload("Escrow")

	if status != "" {
		query = query.Where("order_status = ?", status)
	}
	if carrier != "" {
		query = query.Where("shipping_carrier ILIKE ?", "%"+carrier+"%")
	}
	if search != "" {
		query = query.Where("order_sn ILIKE ? OR buyer_username ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	var list []shopeeOrderTableGORM
	if err := query.Order("created_at DESC").Find(&list).Error; err != nil {
		return nil, err
	}

	results := make([]*shopee.ShopeeOrder, len(list))
	for i := range list {
		results[i] = mapShopeeOrderGORMToDomain(&list[i])
	}
	return results, nil
}

func (r *ShopeeRepository) SaveOrder(ctx context.Context, o *shopee.ShopeeOrder) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		orderG := shopeeOrderTableGORM{
			OrderSN:           o.OrderSN(),
			ShopID:            o.ShopID(),
			OrderStatus:       o.OrderStatus(),
			BuyerUserID:       o.BuyerUserID(),
			BuyerUsername:     o.BuyerUsername(),
			MessageToSeller:   o.MessageToSeller(),
			ShipByDate:        o.ShipByDate(),
			ShipByDateTime:    o.ShipByDateTime(),
			ShippingCarrier:   o.ShippingCarrier(),
			TrackingNumber:    o.TrackingNumber(),
			TotalAmount:       o.TotalAmount(),
			BuyerCancelReason: o.BuyerCancelReason(),
			CreateTimeShopee:  o.CreateTimeShopee(),
			UpdateTimeShopee:  o.UpdateTimeShopee(),
			CreatedAt:         o.CreatedAt(),
			UpdatedAt:         o.UpdatedAt(),
		}

		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "order_sn"}},
			DoUpdates: clause.AssignmentColumns([]string{"order_status", "ship_by_date", "ship_by_date_time", "shipping_carrier", "tracking_number", "total_amount", "buyer_cancel_reason", "update_time_shopee", "updated_at"}),
		}).Create(&orderG).Error; err != nil {
			return err
		}

		for _, item := range o.Items() {
			itemG := shopeeOrderItemTableGORM{
				ID:               item.ID(),
				OrderSN:          item.OrderSN(),
				ItemID:           item.ItemID(),
				ItemName:         item.ItemName(),
				ItemSKU:          item.ItemSKU(),
				ModelID:          item.ModelID(),
				ModelName:        item.ModelName(),
				ModelSKU:         item.ModelSKU(),
				Quantity:         item.Quantity(),
				OriginalPrice:    item.OriginalPrice(),
				DiscountedPrice:  item.DiscountedPrice(),
				ProductID:        item.ProductID(),
				MatchedSKU:       item.MatchedSKU(),
				MappingStatus:    item.MappingStatus(),
				FilamentCost:     item.FilamentCost(),
				HardwareCost:     item.HardwareCost(),
				PackagingCost:    item.PackagingCost(),
				MachineCost:      item.MachineCost(),
				BaseHPP:          item.BaseHPP(),
				NetProfit:        item.NetProfit(),
				ElectricityCost:  item.ElectricityCost(),
				MaintenanceCost:  item.MaintenanceCost(),
				DepreciationCost: item.DepreciationCost(),
				CreatedAt:        item.CreatedAt(),
				UpdatedAt:        item.UpdatedAt(),
			}
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "id"}},
				DoUpdates: clause.AssignmentColumns([]string{"product_id", "matched_sku", "mapping_status", "filament_cost", "hardware_cost", "packaging_cost", "machine_cost", "base_hpp", "net_profit", "electricity_cost", "maintenance_cost", "depreciation_cost", "updated_at"}),
			}).Create(&itemG).Error; err != nil {
				return err
			}
		}

		if o.Escrow() != nil {
			e := o.Escrow()
			escrowG := shopeeOrderEscrowGORM{
				OrderSN:                  e.OrderSN(),
				EscrowAmount:             e.EscrowAmount(),
				SellingPrice:             e.SellingPrice(),
				CommissionFee:            e.CommissionFee(),
				CommissionRuleName:       e.CommissionRuleName(),
				CommissionPercentage:     e.CommissionPercentage(),
				ServiceFee:               e.ServiceFee(),
				ServiceRuleName:          e.ServiceRuleName(),
				ServicePercentage:        e.ServicePercentage(),
				SellerTransactionFee:     e.SellerTransactionFee(),
				SellerOrderProcessingFee: e.SellerOrderProcessingFee(),
				SellerVoucherDiscount:    e.SellerVoucherDiscount(),
				TotalMarketplaceFee:      e.TotalMarketplaceFee(),
				TotalHPP:                 e.TotalHPP(),
				KasFilamen:               e.KasFilamen(),
				KasKomponen:              e.KasKomponen(),
				KasPacking:               e.KasPacking(),
				KasListrik:               e.KasListrik(),
				KasMaintenance:           e.KasMaintenance(),
				KasDepresiasi:            e.KasDepresiasi(),
				KasLabaBersih:            e.KasLabaBersih(),
				FinancialStatus:          e.FinancialStatus(),
				CreatedAt:                e.CreatedAt(),
				UpdatedAt:                e.UpdatedAt(),
			}
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "order_sn"}},
				DoUpdates: clause.AssignmentColumns([]string{"escrow_amount", "selling_price", "commission_fee", "service_fee", "seller_transaction_fee", "seller_order_processing_fee", "total_marketplace_fee", "total_hpp", "kas_filamen", "kas_komponen", "kas_packing", "kas_listrik", "kas_maintenance", "kas_depresiasi", "kas_laba_bersih", "financial_status", "updated_at"}),
			}).Create(&escrowG).Error; err != nil {
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

	query := r.db.WithContext(ctx).Model(&shopeeOrderItemTableGORM{}).Where("item_id = ?", itemID)
	if modelID > 0 {
		query = query.Where("model_id = ?", modelID)
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
		KasFilamen           float64 `gorm:"column:kas_filamen"`
		KasKomponen          float64 `gorm:"column:kas_komponen"`
		KasPacking           float64 `gorm:"column:kas_packing"`
		KasListrik           float64 `gorm:"column:kas_listrik"`
		KasMaintenance       float64 `gorm:"column:kas_maintenance"`
		KasDepresiasi        float64 `gorm:"column:kas_depresiasi"`
		KasLabaBersih        float64 `gorm:"column:kas_laba_bersih"`
		UnmappedCount        int64   `gorm:"column:unmapped_count"`
	}

	var agg AggResult
	sql := `
		SELECT 
			COALESCE(COUNT(o.order_sn), 0) AS total_orders,
			COALESCE(SUM(o.total_amount), 0) AS total_gross_sales,
			COALESCE(SUM(e.total_marketplace_fee), 0) AS total_marketplace_fees,
			COALESCE(SUM(e.escrow_amount), 0) AS total_escrow_net_in,
			COALESCE(SUM(e.total_hpp), 0) AS total_hpp,
			COALESCE(SUM(e.kas_filamen), 0) AS kas_filamen,
			COALESCE(SUM(e.kas_komponen), 0) AS kas_komponen,
			COALESCE(SUM(e.kas_packing), 0) AS kas_packing,
			COALESCE(SUM(e.kas_listrik), 0) AS kas_listrik,
			COALESCE(SUM(e.kas_maintenance), 0) AS kas_maintenance,
			COALESCE(SUM(e.kas_depresiasi), 0) AS kas_depresiasi,
			COALESCE(SUM(e.kas_laba_bersih), 0) AS kas_laba_bersih,
			(SELECT COUNT(*) FROM shopee_order_items WHERE mapping_status = 'UNMAPPED') AS unmapped_count
		FROM shopee_orders o
		JOIN shopee_order_escrows e ON e.order_sn = o.order_sn
		WHERE o.order_status = 'COMPLETED' AND e.financial_status = 'RELEASED'
	`

	if err := r.db.WithContext(ctx).Raw(sql).Scan(&agg).Error; err != nil {
		return nil, err
	}

	avgMargin := 0.0
	if agg.TotalGrossSales > 0 {
		avgMargin = (agg.KasLabaBersih / agg.TotalGrossSales) * 100.0
	}

	return &shopee.CashflowSummary{
		TotalCompletedOrders: agg.TotalOrders,
		TotalOrders:          agg.TotalOrders,
		TotalGrossSales:      agg.TotalGrossSales,
		TotalMarketplaceFees: agg.TotalMarketplaceFees,
		TotalEscrowNetIn:     agg.TotalEscrowNetIn,
		TotalHPP:             agg.TotalHPP,
		KasFilamen:           agg.KasFilamen,
		KasKomponen:          agg.KasKomponen,
		KasPacking:           agg.KasPacking,
		KasListrik:           agg.KasListrik,
		KasMaintenance:       agg.KasMaintenance,
		KasDepresiasi:        agg.KasDepresiasi,
		KasLabaBersih:        agg.KasLabaBersih,
		AverageProfitMargin:  avgMargin,
		UnmappedItemsCount:   agg.UnmappedCount,
	}, nil
}

func (r *ShopeeRepository) RecalculateFinances(ctx context.Context) (int, error) {
	res := r.db.WithContext(ctx).Exec("UPDATE shopee_order_escrows SET kas_laba_bersih = GREATEST(escrow_amount - total_hpp, 0), updated_at = NOW()")
	if res.Error != nil {
		return 0, res.Error
	}
	return int(res.RowsAffected), nil
}
