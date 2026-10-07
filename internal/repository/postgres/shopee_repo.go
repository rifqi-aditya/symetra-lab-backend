package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"symetra-lab-backend-v2/internal/domain/costing"
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
	MachineID        *uuid.UUID `gorm:"column:machine_id;type:uuid"`
	WeightGrams      float64    `gorm:"column:weight_grams"`
	PrintTimeHours   float64    `gorm:"column:print_time_hours"`
	CreatedAt        time.Time  `gorm:"column:created_at"`
	UpdatedAt        time.Time  `gorm:"column:updated_at"`
}

func (shopeeUnifiedItemGORM) TableName() string {
	return "order_items"
}

type shopeeUnifiedOrderGORM struct {
	ID               uuid.UUID                   `gorm:"column:id;primaryKey;type:uuid"`
	UserID           uuid.UUID                   `gorm:"column:user_id;type:uuid"`
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
	var existing shopGORM
	err := r.db.WithContext(ctx).Where("shop_id = ?", s.ShopID()).First(&existing).Error
	if err == nil {
		return r.db.WithContext(ctx).Model(&existing).Updates(map[string]interface{}{
			"access_token":             s.AccessToken(),
			"refresh_token":            s.RefreshToken(),
			"access_token_expires_at":  s.AccessTokenExpiresAt(),
			"refresh_token_expires_at": s.RefreshTokenExpiresAt(),
			"updated_at":               time.Now(),
		}).Error
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	record := shopGORM{
		ShopID:                s.ShopID(),
		ShopName:              s.ShopName(),
		Region:                s.Region(),
		AccessToken:           s.AccessToken(),
		RefreshToken:          s.RefreshToken(),
		AccessTokenExpiresAt:  s.AccessTokenExpiresAt(),
		RefreshTokenExpiresAt: s.RefreshTokenExpiresAt(),
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	}

	return r.db.WithContext(ctx).Create(&record).Error
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

func computeItemCostsFromProduct(tx *gorm.DB, prod *productGORM, qty int, sellingPrice float64) (
	filCost, compCost, packCost, elecCost, maintCost, depCost, totalCOGS, netProf float64,
) {
	if qty <= 0 {
		qty = 1
	}
	q := float64(qty)

	var compInputs []costing.ComponentCostInput
	for _, c := range prod.Components {
		price := 0.0
		if c.ComponentID != nil {
			var comp componentGORM
			if err := tx.Where("id = ?", *c.ComponentID).First(&comp).Error; err == nil {
				price = comp.PricePerUnit
			}
		}
		compInputs = append(compInputs, costing.ComponentCostInput{
			PricePerUnit:  price,
			Quantity:      c.Quantity,
			MarkupPercent: c.MarkupPercent,
		})
	}

	var packInputs []costing.PackagingCostInput
	for _, p := range prod.PackagingItems {
		uCost := 0.0
		var pack packagingItemGORM
		if err := tx.Where("id = ?", p.PackagingItemID).First(&pack).Error; err == nil {
			uCost = pack.UnitCost
			if uCost <= 0 && pack.PurchaseQuantity > 0 {
				uCost = pack.PurchasePrice / pack.PurchaseQuantity
			}
		}
		packInputs = append(packInputs, costing.PackagingCostInput{
			UnitCost:     uCost,
			QuantityUsed: p.QuantityUsed,
		})
	}

	batchSize := prod.BatchSize
	if batchSize <= 0 {
		batchSize = 1
	}

	cb := costing.Calculate(costing.CalculationInput{
		MaterialType:          prod.MaterialType,
		DefaultWeightGrams:    prod.DefaultWeightGrams,
		DefaultPrintTimeHours: prod.DefaultPrintTimeHours,
		BatchSize:             batchSize,
		PackingFeeIDR:         prod.PackingFeeIDR,
		BaseHPP:               prod.BaseHPP,
		BaseSellingPrice:      prod.BaseSellingPrice,
		Components:            compInputs,
		PackagingItems:        packInputs,
	})

	filCost = cb.FilamentCost * q
	compCost = cb.HardwareCost * q
	packCost = cb.PackagingCost * q
	elecCost = cb.ElectricityCost * q
	maintCost = cb.MaintenanceCost * q
	depCost = cb.DepreciationCost * q
	totalCOGS = cb.BaseHPP * q

	if totalCOGS == 0 && prod.BaseHPP > 0 {
		totalCOGS = prod.BaseHPP * q
		filCost = totalCOGS
	}

	netProf = (sellingPrice * q) - totalCOGS
	return
}

func recalculateOrderFunds(tx *gorm.DB, orderID uuid.UUID) {
	type CostSums struct {
		Filament     float64 `gorm:"column:f_fil"`
		Component    float64 `gorm:"column:f_comp"`
		Packaging    float64 `gorm:"column:f_pack"`
		Electricity  float64 `gorm:"column:f_elec"`
		Maintenance  float64 `gorm:"column:f_maint"`
		Depreciation float64 `gorm:"column:f_dep"`
		TotalCOGS    float64 `gorm:"column:tot_cogs"`
	}
	var sums CostSums
	_ = tx.Raw(`
		SELECT 
			COALESCE(SUM(filament_cost), 0) AS f_fil,
			COALESCE(SUM(component_cost), 0) AS f_comp,
			COALESCE(SUM(packaging_cost), 0) AS f_pack,
			COALESCE(SUM(energy_cost), 0) AS f_elec,
			COALESCE(SUM(maintenance_cost), 0) AS f_maint,
			COALESCE(SUM(depreciation_cost), 0) AS f_dep,
			COALESCE(SUM(total_cogs), 0) AS tot_cogs
		FROM order_items
		WHERE order_id = ?
	`, orderID).Scan(&sums).Error

	var ord shopeeUnifiedOrderGORM
	if err := tx.Where("id = ?", orderID).First(&ord).Error; err == nil {
		netBase := ord.NetAmount
		if netBase <= 0 {
			netBase = ord.GrossAmount
		}
		netProfit := netBase - sums.TotalCOGS
		_ = tx.Model(&shopeeUnifiedOrderGORM{}).Where("id = ?", orderID).Updates(map[string]interface{}{
			"net_amount":        netBase,
			"fund_filament":     sums.Filament,
			"fund_component":    sums.Component,
			"fund_packaging":    sums.Packaging,
			"fund_electricity":  sums.Electricity,
			"fund_maintenance":  sums.Maintenance,
			"fund_depreciation": sums.Depreciation,
			"cogs_amount":       sums.TotalCOGS,
			"net_profit":        netProfit,
			"fund_net_profit":   netProfit,
			"updated_at":        time.Now(),
		}).Error
	}
}

func (r *ShopeeRepository) SaveOrder(ctx context.Context, o *shopee.ShopeeOrder) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existingOrder shopeeUnifiedOrderGORM
		err := tx.Where("order_number = ?", o.OrderSN()).First(&existingOrder).Error
		orderID := existingOrder.ID
		if err != nil || orderID == uuid.Nil {
			orderID = uuid.New()
		}

		var grossAmount, netAmount, channelFee float64
		paymentStatus := "UNPAID"

		if o.Escrow() != nil && o.Escrow().EscrowAmount() > 0 {
			e := o.Escrow()
			grossAmount = e.SellingPrice()
			netAmount = e.EscrowAmount()
			channelFee = e.TotalMarketplaceFee()
			if e.FinancialStatus() == "RELEASED" {
				paymentStatus = "PAID"
			}
		} else {
			grossAmount = o.TotalAmount()
			if channelFee > 0 {
				netAmount = grossAmount - channelFee
			} else {
				netAmount = grossAmount
			}
		}

		orderCreatedAt := o.CreatedAt()
		if o.CreateTimeShopee() > 0 {
			orderCreatedAt = time.Unix(o.CreateTimeShopee(), 0)
		}
		orderUpdatedAt := o.UpdatedAt()
		if o.UpdateTimeShopee() > 0 {
			orderUpdatedAt = time.Unix(o.UpdateTimeShopee(), 0)
		}

		// Dapatkan default admin user_id jika belum ada
		userID := existingOrder.UserID
		if userID == uuid.Nil {
			var firstUser struct{ ID uuid.UUID }
			if err := tx.Raw("SELECT id FROM auth.users ORDER BY created_at ASC LIMIT 1").Scan(&firstUser).Error; err == nil && firstUser.ID != uuid.Nil {
				userID = firstUser.ID
			}
		}

		var totFil, totComp, totPack, totElec, totMaint, totDep, totCOGS float64
		var itemsToInsert []shopeeUnifiedItemGORM

		// 1. Hitung biaya & siapkan item-item pesanan
		for _, item := range o.Items() {
			var existingItem shopeeUnifiedItemGORM
			_ = tx.Where("order_id = ? AND channel_item_id = ?", orderID, item.ItemID()).First(&existingItem).Error
			itemID := existingItem.ID
			if itemID == uuid.Nil {
				itemID = uuid.New()
			}

			productID := item.ProductID()
			matchedSKU := item.MatchedSKU()
			mappingStatus := item.MappingStatus()

			filCost := item.FilamentCost()
			compCost := item.HardwareCost()
			packCost := item.PackagingCost()
			elecCost := item.ElectricityCost()
			maintCost := item.MaintenanceCost()
			depCost := item.DepreciationCost()
			baseHPP := item.BaseHPP()
			netProf := item.NetProfit()

			if productID == nil && existingItem.ProductID != nil {
				productID = existingItem.ProductID
				matchedSKU = existingItem.MatchedSKU
				mappingStatus = existingItem.MappingStatus
				filCost = existingItem.FilamentCost
				compCost = existingItem.ComponentCost
				packCost = existingItem.PackagingCost
				elecCost = existingItem.EnergyCost
				maintCost = existingItem.MaintenanceCost
				depCost = existingItem.DepreciationCost
				baseHPP = existingItem.HPP
				netProf = existingItem.NetProfit
			}

			// Cek auto-match exact SKU dari Shopee ke tabel products jika belum mapped
			if productID == nil {
				targetSKU := item.ModelSKU()
				if targetSKU == "" {
					targetSKU = item.ItemSKU()
				}
				if targetSKU != "" {
					var p productGORM
					if err := tx.Preload("Components").Preload("PackagingItems").Where("UPPER(TRIM(sku)) = UPPER(TRIM(?))", targetSKU).First(&p).Error; err == nil {
						pID := p.ID
						productID = &pID
						if p.SKU != nil {
							matchedSKU = *p.SKU
						}
						mappingStatus = "MATCHED"
						filCost, compCost, packCost, elecCost, maintCost, depCost, baseHPP, netProf = computeItemCostsFromProduct(tx, &p, item.Quantity(), item.DiscountedPrice())
					}
				}

				// Jika belum cocok dengan SKU, cek riwayat penautan manual sebelumnya berdasarkan channel_item_id
				if productID == nil && item.ItemID() > 0 {
					var prevItem shopeeUnifiedItemGORM
					prevQ := tx.Where("channel_item_id = ? AND product_id IS NOT NULL", item.ItemID())
					if item.ModelID() > 0 {
						prevQ = prevQ.Where("channel_model_id = ?", item.ModelID())
					}
					if err := prevQ.Order("updated_at DESC").First(&prevItem).Error; err == nil && prevItem.ProductID != nil {
						var p productGORM
						if err := tx.Preload("Components").Preload("PackagingItems").Where("id = ?", *prevItem.ProductID).First(&p).Error; err == nil {
							pID := p.ID
							productID = &pID
							if p.SKU != nil {
								matchedSKU = *p.SKU
							}
							mappingStatus = "MATCHED"
							filCost, compCost, packCost, elecCost, maintCost, depCost, baseHPP, netProf = computeItemCostsFromProduct(tx, &p, item.Quantity(), item.DiscountedPrice())
						}
					}
				}
			}

			totFil += filCost
			totComp += compCost
			totPack += packCost
			totElec += elecCost
			totMaint += maintCost
			totDep += depCost
			totCOGS += baseHPP
			prodTitle := item.ItemName()
			var itemMachineID *uuid.UUID
			var itemWeight float64
			var itemPrintTime float64
			itemSKU := item.ItemSKU()

			if productID != nil {
				var pr productGORM
				if err := tx.Where("id = ?", *productID).First(&pr).Error; err == nil {
					if pr.Name != "" {
						prodTitle = pr.Name
					}
					itemMachineID = pr.DefaultMachineID
					itemWeight = pr.DefaultWeightGrams * float64(item.Quantity())
					itemPrintTime = pr.DefaultPrintTimeHours * float64(item.Quantity())
					if itemSKU == "" && pr.SKU != nil {
						itemSKU = *pr.SKU
					}
				}
			}

			itemsToInsert = append(itemsToInsert, shopeeUnifiedItemGORM{
				ID:               itemID,
				OrderID:          orderID,
				ChannelItemID:    item.ItemID(),
				ChannelModelID:   item.ModelID(),
				ProductName:      prodTitle,
				ItemSKU:          itemSKU,
				MatchedSKU:       matchedSKU,
				MappingStatus:    mappingStatus,
				Quantity:         item.Quantity(),
				SellingPrice:     item.DiscountedPrice(),
				HPP:              baseHPP,
				TotalCOGS:        baseHPP,
				FilamentCost:     filCost,
				ComponentCost:    compCost,
				PackagingCost:    packCost,
				EnergyCost:       elecCost,
				MaintenanceCost:  maintCost,
				DepreciationCost: depCost,
				NetProfit:        netProf,
				ProductID:        productID,
				MachineID:        itemMachineID,
				WeightGrams:      itemWeight,
				PrintTimeHours:   itemPrintTime,
				CreatedAt:        orderCreatedAt,
				UpdatedAt:        orderUpdatedAt,
			})
		}

		cogsAmount := totCOGS
		netProfit := netAmount - cogsAmount
		fNetProfit := netProfit

		// 2. Simpan induk tabel orders TERLEBIH DAHULU agar Foreign Key tidak error
		orderRecord := shopeeUnifiedOrderGORM{
			ID:               orderID,
			UserID:           userID,
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
			FundFilament:     totFil,
			FundComponent:    totComp,
			FundPackaging:    totPack,
			FundElectricity:  totElec,
			FundMaintenance:  totMaint,
			FundDepreciation: totDep,
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

		// 3. Simpan detail tabel anak: order_marketplace_details
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

		// 4. Simpan detail tabel anak: order_items
		for _, itemG := range itemsToInsert {
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
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Ambil data master produk beserta komponen & packaging
		var prod productGORM
		if err := tx.Preload("Components").Preload("PackagingItems").Where("id = ?", productID).First(&prod).Error; err != nil {
			return fmt.Errorf("product not found: %w", err)
		}

		matchedSKU := ""
		if prod.SKU != nil {
			matchedSKU = *prod.SKU
		}

		// 2. Ambil semua baris order_items yang channel_item_id / channel_model_id nya cocok
		itemQuery := tx.Model(&shopeeUnifiedItemGORM{}).Where("channel_item_id = ?", itemID)
		if modelID > 0 {
			itemQuery = itemQuery.Where("channel_model_id = ?", modelID)
		}

		var items []shopeeUnifiedItemGORM
		if err := itemQuery.Find(&items).Error; err != nil {
			return err
		}

		orderIDsMap := make(map[uuid.UUID]bool)

		// 3. Update kolom matched_sku, product_id, dan kalkulasi 6 pos biaya untuk setiap item
		for _, it := range items {
			orderIDsMap[it.OrderID] = true
			filCost, compCost, packCost, elecCost, maintCost, depCost, totalCOGS, netProf := computeItemCostsFromProduct(tx, &prod, it.Quantity, it.SellingPrice)

			updates := map[string]interface{}{
				"product_id":        productID,
				"product_name":      prod.Name,
				"matched_sku":       matchedSKU,
				"mapping_status":    "MATCHED",
				"filament_cost":     filCost,
				"component_cost":    compCost,
				"packaging_cost":    packCost,
				"energy_cost":       elecCost,
				"maintenance_cost":  maintCost,
				"depreciation_cost": depCost,
				"hpp":               totalCOGS,
				"total_cogs":        totalCOGS,
				"net_profit":        netProf,
				"updated_at":        time.Now(),
			}

			if prod.DefaultMachineID != nil {
				updates["machine_id"] = prod.DefaultMachineID
			}
			if prod.DefaultWeightGrams > 0 {
				updates["weight_grams"] = prod.DefaultWeightGrams * float64(it.Quantity)
			}
			if prod.DefaultPrintTimeHours > 0 {
				updates["print_time_hours"] = prod.DefaultPrintTimeHours * float64(it.Quantity)
			}
			if it.ItemSKU == "" && matchedSKU != "" {
				updates["item_sku"] = matchedSKU
			}

			if err := tx.Model(&shopeeUnifiedItemGORM{}).Where("id = ?", it.ID).Updates(updates).Error; err != nil {
				return err
			}
		}

		// 4. Trigger kalkulasi ulang 7 pos kas pada setiap orders induk yang terpengaruh
		for oID := range orderIDsMap {
			recalculateOrderFunds(tx, oID)
		}

		return nil
	})
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
