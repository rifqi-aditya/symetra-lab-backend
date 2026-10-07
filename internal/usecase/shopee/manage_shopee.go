package shopee

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"symetra-lab-backend-v2/config"
	"symetra-lab-backend-v2/internal/domain/finance"
	"symetra-lab-backend-v2/internal/domain/shopee"
	pkgShopee "symetra-lab-backend-v2/pkg/shopee"
)

type ShopeeUseCases struct {
	cfg          *config.Config
	repo         shopee.Repository
	shopeeClient *pkgShopee.Client
	financeRepo  finance.Repository
}

func NewShopeeUseCases(
	cfg *config.Config,
	repo shopee.Repository,
	client *pkgShopee.Client,
	financeRepo finance.Repository,
) *ShopeeUseCases {
	return &ShopeeUseCases{
		cfg:          cfg,
		repo:         repo,
		shopeeClient: client,
		financeRepo:  financeRepo,
	}
}

// 1. Auth & Shops
func (uc *ShopeeUseCases) GetAuthURL() (string, error) {
	if uc.shopeeClient == nil || uc.shopeeClient.PartnerID == 0 || uc.shopeeClient.PartnerKey == "" {
		return "", shopee.ErrPartnerNotConfig
	}
	return uc.shopeeClient.BuildAuthURL()
}

func (uc *ShopeeUseCases) HandleCallback(ctx context.Context, code string, shopID uint64) (*shopee.ShopeeShop, error) {
	if uc.shopeeClient == nil {
		return nil, shopee.ErrPartnerNotConfig
	}

	tokenResp, err := uc.shopeeClient.GetAccessToken(code, shopID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	accessExpires := now.Add(time.Duration(tokenResp.ExpireIn) * time.Second)
	refreshExpires := now.Add(30 * 24 * time.Hour)

	shop := shopee.ReconstructShop(
		0,
		shopID,
		fmt.Sprintf("Shopee Store (%d)", shopID),
		"ID",
		tokenResp.AccessToken,
		tokenResp.RefreshToken,
		accessExpires,
		refreshExpires,
		now,
		now,
	)

	if err := uc.repo.SaveShop(ctx, shop); err != nil {
		return nil, err
	}

	return uc.repo.FindShop(ctx, shopID)
}

func (uc *ShopeeUseCases) ListShops(ctx context.Context) ([]*shopee.ShopeeShop, error) {
	return uc.repo.FindAllShops(ctx)
}

func (uc *ShopeeUseCases) RefreshToken(ctx context.Context, shopID uint64) (*shopee.ShopeeShop, error) {
	shop, err := uc.repo.FindShop(ctx, shopID)
	if err != nil {
		return nil, err
	}

	if uc.shopeeClient == nil {
		return nil, shopee.ErrPartnerNotConfig
	}

	resp, err := uc.shopeeClient.RefreshAccessToken(shop.RefreshToken(), shop.ShopID())
	if err != nil {
		return nil, err
	}

	now := time.Now()
	shop.UpdateTokens(
		resp.AccessToken,
		resp.RefreshToken,
		now.Add(time.Duration(resp.ExpireIn)*time.Second),
		now.Add(30*24*time.Hour),
	)

	if err := uc.repo.SaveShop(ctx, shop); err != nil {
		return nil, err
	}

	return shop, nil
}

// Helper to get valid shop token
func (uc *ShopeeUseCases) GetValidShop(ctx context.Context, shopID uint64) (*shopee.ShopeeShop, error) {
	var shop *shopee.ShopeeShop
	var err error

	if shopID > 0 {
		shop, err = uc.repo.FindShop(ctx, shopID)
	} else {
		shop, err = uc.repo.FindDefaultShop(ctx)
	}

	if err != nil {
		return nil, err
	}

	if shop.IsTokenExpired() {
		return uc.RefreshToken(ctx, shop.ShopID())
	}
	return shop, nil
}

// 2. Orders & Sync
func (uc *ShopeeUseCases) ListOrders(ctx context.Context, status, carrier, search string) ([]*shopee.ShopeeOrder, error) {
	return uc.repo.FindOrders(ctx, status, carrier, search)
}

func (uc *ShopeeUseCases) GetOrderDetail(ctx context.Context, orderSN string) (*shopee.ShopeeOrder, error) {
	return uc.repo.FindOrder(ctx, orderSN)
}

func (uc *ShopeeUseCases) LinkSKU(ctx context.Context, itemID, modelID uint64, productID uuid.UUID) error {
	return uc.repo.LinkSKU(ctx, itemID, modelID, productID)
}

func (uc *ShopeeUseCases) SyncShopeeOrders(ctx context.Context) (int, error) {
	shop, err := uc.GetValidShop(ctx, 0)
	if err != nil {
		return 0, err
	}

	if uc.shopeeClient == nil {
		return 0, shopee.ErrPartnerNotConfig
	}

	now := time.Now().Unix()
	timeFrom := now - (14 * 86400) // last 14 days

	listResp, err := uc.shopeeClient.GetOrderList(shop.AccessToken(), shop.ShopID(), timeFrom, now, "", "")
	if err != nil {
		return 0, err
	}

	orderSNs := make([]string, len(listResp.Response.OrderList))
	for i, o := range listResp.Response.OrderList {
		orderSNs[i] = o.OrderSN
	}

	if len(orderSNs) == 0 {
		return 0, nil
	}

	detailResp, err := uc.shopeeClient.GetOrderDetail(shop.AccessToken(), shop.ShopID(), orderSNs)
	if err != nil {
		return 0, err
	}

	savedCount := 0
	for _, o := range detailResp.Response.OrderList {
		items := make([]shopee.ShopeeOrderItem, len(o.ItemList))
		for j, it := range o.ItemList {
			items[j] = shopee.ReconstructOrderItem(
				0,
				o.OrderSN,
				it.ItemID,
				it.ItemName,
				it.ItemSKU,
				it.ModelID,
				it.ModelName,
				it.ModelSKU,
				it.ModelQuantityPurchased,
				it.ModelOriginalPrice,
				it.ModelDiscountedPrice,
				nil,
				"",
				"UNMAPPED",
				0, 0, 0, 0, 0, 0, 0, 0, 0,
				time.Now(),
				time.Now(),
			)
		}

		var shipByDateTime *time.Time
		if o.ShipByDate > 0 {
			t := time.Unix(o.ShipByDate, 0)
			shipByDateTime = &t
		}

		domainOrder := shopee.ReconstructOrder(
			o.OrderSN,
			shop.ShopID(),
			o.OrderStatus,
			o.BuyerUserID,
			o.BuyerUsername,
			o.MessageToSeller,
			o.ShipByDate,
			shipByDateTime,
			o.ShippingCarrier,
			"",
			o.TotalAmount,
			o.BuyerCancelReason,
			o.CreateTime,
			o.UpdateTime,
			items,
			nil,
			time.Unix(o.CreateTime, 0),
			time.Unix(o.UpdateTime, 0),
		)

		if err := uc.repo.SaveOrder(ctx, domainOrder); err == nil {
			savedCount++
		}
	}

	return savedCount, nil
}

// SyncSingleOrder mengambil detail dan escrow untuk 1 order_sn spesifik dan menyimpannya ke database
func (uc *ShopeeUseCases) SyncSingleOrder(ctx context.Context, shopID uint64, orderSN string) (*shopee.ShopeeOrder, error) {
	shop, err := uc.GetValidShop(ctx, shopID)
	if err != nil {
		return nil, fmt.Errorf("failed to get valid shop for id %d: %w", shopID, err)
	}

	if uc.shopeeClient == nil {
		return nil, shopee.ErrPartnerNotConfig
	}

	detailResp, err := uc.shopeeClient.GetOrderDetail(shop.AccessToken(), shop.ShopID(), []string{orderSN})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch order detail from shopee: %w", err)
	}

	if len(detailResp.Response.OrderList) == 0 {
		return nil, fmt.Errorf("order %s not found in shopee response", orderSN)
	}

	o := detailResp.Response.OrderList[0]
	items := make([]shopee.ShopeeOrderItem, len(o.ItemList))
	for j, it := range o.ItemList {
		items[j] = shopee.ReconstructOrderItem(
			0,
			o.OrderSN,
			it.ItemID,
			it.ItemName,
			it.ItemSKU,
			it.ModelID,
			it.ModelName,
			it.ModelSKU,
			it.ModelQuantityPurchased,
			it.ModelOriginalPrice,
			it.ModelDiscountedPrice,
			nil,
			"",
			"UNMAPPED",
			0, 0, 0, 0, 0, 0, 0, 0, 0,
			time.Now(),
			time.Now(),
		)
	}

	var shipByDateTime *time.Time
	if o.ShipByDate > 0 {
		t := time.Unix(o.ShipByDate, 0)
		shipByDateTime = &t
	}

	var escrow *shopee.ShopeeOrderEscrow
	// Coba ambil escrow jika pesanan sudah diproses atau selesai
	if o.OrderStatus == "COMPLETED" || o.OrderStatus == "SHIPPED" || o.OrderStatus == "PROCESSED" {
		escrowResp, err := uc.shopeeClient.GetEscrowDetail(shop.AccessToken(), shop.ShopID(), orderSN)
		if err == nil && escrowResp != nil {
			inc := escrowResp.Response.OrderIncome
			commRule := ""
			commPct := 0.0
			if len(inc.NetCommissionFeeInfo) > 0 {
				commRule = inc.NetCommissionFeeInfo[0].RuleDisplayName
				commPct = inc.NetCommissionFeeInfo[0].FeeRate
			}
			svcRule := ""
			svcPct := 0.0
			if len(inc.NetServiceFeeInfo) > 0 {
				svcRule = inc.NetServiceFeeInfo[0].RuleDisplayName
				svcPct = inc.NetServiceFeeInfo[0].FeeRate
			}

			totalMarketFee := inc.CommissionFee + inc.ServiceFee + inc.SellerTransactionFee + inc.SellerOrderProcessingFee
			finStatus := "PENDING"
			if o.OrderStatus == "COMPLETED" {
				finStatus = "RELEASED"
			}

			sellingPrice := inc.OrderSellingPrice
			if sellingPrice <= 0 {
				sellingPrice = inc.SellingPrice
			}
			if sellingPrice <= 0 && o.TotalAmount > 0 {
				sellingPrice = o.TotalAmount
			}
			if sellingPrice <= 0 && inc.EscrowAmount > 0 {
				sellingPrice = inc.EscrowAmount + totalMarketFee
			}

			eObj := shopee.ReconstructOrderEscrow(
				orderSN,
				inc.EscrowAmount,
				sellingPrice,
				inc.CommissionFee,
				commRule,
				commPct,
				inc.ServiceFee,
				svcRule,
				svcPct,
				inc.SellerTransactionFee,
				inc.SellerOrderProcessingFee,
				inc.VoucherFromSeller,
				totalMarketFee,
				0, // TotalHPP (dihitung oleh repository/alokasi)
				0, 0, 0, 0, 0, 0,
				inc.EscrowAmount,
				finStatus,
				time.Now(),
				time.Now(),
			)
			escrow = eObj
		}
	}

	domainOrder := shopee.ReconstructOrder(
		o.OrderSN,
		shop.ShopID(),
		o.OrderStatus,
		o.BuyerUserID,
		o.BuyerUsername,
		o.MessageToSeller,
		o.ShipByDate,
		shipByDateTime,
		o.ShippingCarrier,
		"",
		o.TotalAmount,
		o.BuyerCancelReason,
		o.CreateTime,
		o.UpdateTime,
		items,
		escrow,
		time.Unix(o.CreateTime, 0),
		time.Unix(o.UpdateTime, 0),
	)

	if err := uc.repo.SaveOrder(ctx, domainOrder); err != nil {
		return nil, fmt.Errorf("failed to save order to database: %w", err)
	}

	return uc.repo.FindOrder(ctx, orderSN)
}

// HandleOrderStatusPush memproses payload webhook order_status_push dari Shopee
func (uc *ShopeeUseCases) HandleOrderStatusPush(ctx context.Context, shopID uint64, orderSN, status string) (*shopee.ShopeeOrder, error) {
	// Sync data order terbaru secara realtime ke DB
	ord, err := uc.SyncSingleOrder(ctx, shopID, orderSN)
	if err != nil {
		return nil, err
	}

	// Auto-create finance_transaction saat pesanan COMPLETED
	if ord != nil && ord.OrderStatus() == "COMPLETED" && ord.Escrow() != nil && uc.financeRepo != nil {
		orderDate := ord.CreatedAt()
		if ord.CreateTimeShopee() > 0 {
			orderDate = time.Unix(ord.CreateTimeShopee(), 0)
		}
		tx, txErr := finance.NewFinanceTransaction(
			finance.TypeIncome,
			finance.CategorySalesShopee,
			nil,
			ord.Escrow().EscrowAmount(),
			"Escrow Shopee - Order "+orderSN,
			orderDate,
			"SHOPEE_ESCROW",
			orderSN,
			"",
		)
		if txErr == nil {
			// CreateTransactionIfNotExists: skip jika order_sn sudah ada (idempotent)
			_ = uc.financeRepo.CreateTransactionIfNotExists(ctx, tx)
		}
	}

	return ord, nil
}

// 3. Logistics
func (uc *ShopeeUseCases) ShipOrder(ctx context.Context, orderSN string) (map[string]interface{}, error) {
	ord, err := uc.repo.FindOrder(ctx, orderSN)
	if err != nil {
		return nil, err
	}

	shop, err := uc.GetValidShop(ctx, ord.ShopID())
	if err != nil {
		return nil, err
	}

	paramResp, err := uc.shopeeClient.GetShippingParameter(shop.AccessToken(), shop.ShopID(), orderSN)
	if err != nil {
		return nil, err
	}

	shipReq := pkgShopee.ShopeeShipOrderRequest{OrderSN: orderSN}
	if len(paramResp.Response.Dropoff.BranchList) > 0 {
		shipReq.Dropoff = &pkgShopee.ShopeeShipDropoff{
			BranchID: paramResp.Response.Dropoff.BranchList[0].BranchID,
		}
	} else if len(paramResp.Response.Pickup.AddressList) > 0 {
		shipReq.Pickup = &pkgShopee.ShopeeShipPickup{
			AddressID: paramResp.Response.Pickup.AddressList[0].AddressID,
		}
	} else {
		shipReq.Dropoff = &pkgShopee.ShopeeShipDropoff{}
	}

	shipResp, err := uc.shopeeClient.ShipOrder(shop.AccessToken(), shop.ShopID(), shipReq)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"status":   "success",
		"message":  "Pesanan berhasil diproses untuk pengiriman",
		"order_sn": orderSN,
		"response": shipResp,
	}, nil
}

func (uc *ShopeeUseCases) DownloadShippingLabel(ctx context.Context, orderSN string) ([]byte, error) {
	ord, err := uc.repo.FindOrder(ctx, orderSN)
	if err != nil {
		return nil, err
	}

	shop, err := uc.GetValidShop(ctx, ord.ShopID())
	if err != nil {
		return nil, err
	}

	return uc.shopeeClient.DownloadShippingDocument(shop.AccessToken(), shop.ShopID(), orderSN)
}

// 4. Financial & 7 Pos Kas Escrow
func (uc *ShopeeUseCases) GetCashflowSummary(ctx context.Context) (*shopee.CashflowSummary, error) {
	return uc.repo.GetCashflowSummary(ctx)
}

func (uc *ShopeeUseCases) RecalculateFinances(ctx context.Context) (int, error) {
	return uc.repo.RecalculateFinances(ctx)
}
