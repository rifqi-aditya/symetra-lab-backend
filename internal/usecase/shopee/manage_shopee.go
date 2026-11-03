package shopee

import (
	usecaseProduct "symetra-lab-backend-v2/internal/usecase/product"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"symetra-lab-backend-v2/config"
	"symetra-lab-backend-v2/internal/domain/finance"
	"symetra-lab-backend-v2/internal/domain/shopee"
	credUC "symetra-lab-backend-v2/internal/usecase/marketplacecred"
	pkgShopee "symetra-lab-backend-v2/pkg/shopee"
)

type ShopeeUseCases struct {
	cfg            *config.Config
	repo           shopee.Repository
	shopeeClient   *pkgShopee.Client
	financeRepo    finance.Repository
	processOrderUC *ProcessShopeeOrderUseCase
	linkSKUUC      *LinkShopeeSKUUseCase
	recalcUC       *RecalculateFinancesUseCase
	credUC         *credUC.UseCase
}

func (uc *ShopeeUseCases) WithCredUseCase(c *credUC.UseCase) *ShopeeUseCases {
	uc.credUC = c
	return uc
}

func (uc *ShopeeUseCases) getClient(ctx context.Context) (*pkgShopee.Client, error) {
	if uc.credUC != nil {
		if cl, err := uc.credUC.GetShopeeClient(ctx); err == nil && cl != nil {
			return cl, nil
		}
	}
	if uc.shopeeClient != nil && uc.shopeeClient.PartnerID > 0 {
		return uc.shopeeClient, nil
	}
	return nil, shopee.ErrPartnerNotConfig
}

func NewShopeeUseCases(
	cfg *config.Config,
	repo shopee.Repository,
	client *pkgShopee.Client,
	financeRepo finance.Repository,
	processOrderUC *ProcessShopeeOrderUseCase,
	linkSKUUC *LinkShopeeSKUUseCase,
	recalcUC *RecalculateFinancesUseCase,
) *ShopeeUseCases {
	return &ShopeeUseCases{
		cfg:            cfg,
		repo:           repo,
		shopeeClient:   client,
		financeRepo:    financeRepo,
		processOrderUC: processOrderUC,
		linkSKUUC:      linkSKUUC,
		recalcUC:       recalcUC,
	}
}

// 1. Auth & Shops
func (uc *ShopeeUseCases) GetAuthURL(ctx context.Context) (string, error) {
	client, err := uc.getClient(ctx)
	if err != nil {
		return "", err
	}
	return client.BuildAuthURL()
}

func (uc *ShopeeUseCases) HandleCallback(ctx context.Context, code string, shopID uint64) (*shopee.ShopeeShop, error) {
	client, err := uc.getClient(ctx)
	if err != nil {
		return nil, err
	}

	tokenResp, err := client.GetAccessToken(code, shopID)
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

	client, err := uc.getClient(ctx)
	if err != nil {
		return nil, err
	}

	resp, err := client.RefreshAccessToken(shop.RefreshToken(), shop.ShopID())
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
		if err != nil && errors.Is(err, shopee.ErrShopNotFound) {
			shop, err = uc.repo.FindDefaultShop(ctx)
		}
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
	if uc.linkSKUUC != nil {
		return uc.linkSKUUC.Execute(ctx, itemID, modelID, productID)
	}
	return uc.repo.LinkSKU(ctx, itemID, modelID, productID)
}

func (uc *ShopeeUseCases) SyncShopeeOrders(ctx context.Context) (int, error) {
	shop, err := uc.GetValidShop(ctx, 0)
	if err != nil {
		return 0, err
	}

	client, err := uc.getClient(ctx)
	if err != nil {
		return 0, err
	}

	now := time.Now().Unix()
	timeFrom := now - (14 * 86400) // last 14 days

	seenSNs := make(map[string]bool)
	var orderSNs []string

	// 1. Ambil berdasarkan update_time agar status terkini (SHIPPED, COMPLETED) tertangkap
	cursor := ""
	for {
		listResp, err := client.GetOrderListWithRangeField(shop.AccessToken(), shop.ShopID(), timeFrom, now, "", cursor, "update_time")
		if err != nil {
			return 0, err
		}
		for _, o := range listResp.Response.OrderList {
			if !seenSNs[o.OrderSN] {
				seenSNs[o.OrderSN] = true
				orderSNs = append(orderSNs, o.OrderSN)
			}
		}
		if !listResp.Response.More || listResp.Response.NextCursor == "" {
			break
		}
		cursor = listResp.Response.NextCursor
	}

	// 2. Ambil juga berdasarkan create_time untuk memastikan order baru tidak ada yang terlewat
	cursor = ""
	for {
		listResp, err := client.GetOrderListWithRangeField(shop.AccessToken(), shop.ShopID(), timeFrom, now, "", cursor, "create_time")
		if err != nil {
			break
		}
		for _, o := range listResp.Response.OrderList {
			if !seenSNs[o.OrderSN] {
				seenSNs[o.OrderSN] = true
				orderSNs = append(orderSNs, o.OrderSN)
			}
		}
		if !listResp.Response.More || listResp.Response.NextCursor == "" {
			break
		}
		cursor = listResp.Response.NextCursor
	}

	if len(orderSNs) == 0 {
		return 0, nil
	}

	savedCount := 0
	var costingCtx *usecaseProduct.CostingContext
	if uc.processOrderUC != nil && uc.processOrderUC.SKUMatcher() != nil && uc.processOrderUC.SKUMatcher().CostCalculator() != nil {
		costingCtx = uc.processOrderUC.SKUMatcher().CostCalculator().LoadContext(ctx, uuid.Nil)
	}

	// Batch detail fetch in chunks of 50 (Shopee API limit per request)
	chunkSize := 50
	for i := 0; i < len(orderSNs); i += chunkSize {
		end := i + chunkSize
		if end > len(orderSNs) {
			end = len(orderSNs)
		}
		batchSNs := orderSNs[i:end]

		detailResp, err := client.GetOrderDetail(shop.AccessToken(), shop.ShopID(), batchSNs)
		if err != nil {
			return savedCount, err
		}

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

			carrier := o.ShippingCarrier
			if carrier == "" {
				carrier = o.CheckoutShippingCarrier
			}
			var trackingNum string
			if len(o.PackageList) > 0 {
				if carrier == "" {
					carrier = o.PackageList[0].ShippingCarrier
				}
				trackingNum = o.PackageList[0].PackageNumber
			}

			var itemsTotal float64
			for _, it := range items {
				p := it.DiscountedPrice()
				if p <= 0 {
					p = it.OriginalPrice()
				}
				itemsTotal += p * float64(it.Quantity())
			}

			var escrow *shopee.ShopeeOrderEscrow
			if o.OrderStatus != "UNPAID" && o.OrderStatus != "CANCELLED" {
				escrow = uc.buildEscrowForOrder(shop, o.OrderSN, o.OrderStatus, itemsTotal)
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
				carrier,
				trackingNum,
				o.TotalAmount,
				o.BuyerCancelReason,
				o.CreateTime,
				o.UpdateTime,
				o.PayTime,
				items,
				escrow,
				time.Unix(o.CreateTime, 0),
				time.Unix(o.UpdateTime, 0),
			)

			saveErr := error(nil)
			if uc.processOrderUC != nil {
				saveErr = uc.processOrderUC.Execute(ctx, domainOrder, costingCtx)
			} else {
				saveErr = uc.repo.SaveOrder(ctx, domainOrder)
			}
			if saveErr == nil {
				savedCount++
			}
		}
	}

	// Sinkronisasi otomatis pencairan dana (escrow release) ke finance_transactions
	_, _ = uc.SyncReleasedEscrows(ctx, shop.ShopID())

	return savedCount, nil
}

// SyncReleasedEscrows menyinkronkan daftar pencairan dana (escrow released) dari Shopee ke finance_transactions
func (uc *ShopeeUseCases) SyncReleasedEscrows(ctx context.Context, shopID uint64) (int, error) {
	shop, err := uc.GetValidShop(ctx, shopID)
	if err != nil {
		return 0, err
	}
	client, err := uc.getClient(ctx)
	if err != nil {
		return 0, err
	}

	now := time.Now().Unix()
	timeFrom := now - (14 * 86400) // 14 hari terakhir

	escrowResp, err := client.GetEscrowList(shop.AccessToken(), shop.ShopID(), timeFrom, now, 50)
	if err != nil {
		return 0, err
	}

	wibLocation := time.FixedZone("WIB", 7*3600)
	syncedCount := 0
	for _, item := range escrowResp.Response.EscrowList {
		if item.EscrowReleaseTime <= 0 || item.PayoutAmount <= 0 {
			continue
		}

		// Pastikan status order & order_marketplace_details juga tersinkronisasi menjadi COMPLETED & RELEASED
		_, _ = uc.SyncSingleOrder(ctx, shop.ShopID(), item.OrderSN)

		releaseDate := time.Unix(item.EscrowReleaseTime, 0).In(wibLocation)
		tx, txErr := finance.NewFinanceTransaction(
			finance.TypeIncome,
			finance.CategorySalesShopee,
			nil,
			item.PayoutAmount,
			1,
			"Escrow Shopee - Order "+item.OrderSN,
			releaseDate,
			"SHOPEE_ESCROW",
			item.OrderSN,
			"",
			"",
		)
		if txErr == nil && uc.financeRepo != nil {
			if err := uc.financeRepo.UpsertShopeeEscrowTransaction(ctx, tx); err == nil {
				syncedCount++
			}
		}
	}

	return syncedCount, nil
}

// SyncSingleOrder mengambil detail dan escrow untuk 1 order_sn spesifik dan menyimpannya ke database
func (uc *ShopeeUseCases) SyncSingleOrder(ctx context.Context, shopID uint64, orderSN string) (*shopee.ShopeeOrder, error) {
	shop, err := uc.GetValidShop(ctx, shopID)
	if err != nil {
		return nil, fmt.Errorf("failed to get valid shop for id %d: %w", shopID, err)
	}

	client, err := uc.getClient(ctx)
	if err != nil {
		return nil, err
	}

	detailResp, err := client.GetOrderDetail(shop.AccessToken(), shop.ShopID(), []string{orderSN})
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

	var itemsTotal float64
	for _, it := range items {
		p := it.DiscountedPrice()
		if p <= 0 {
			p = it.OriginalPrice()
		}
		itemsTotal += p * float64(it.Quantity())
	}

	var escrow *shopee.ShopeeOrderEscrow
	if o.OrderStatus != "UNPAID" && o.OrderStatus != "CANCELLED" {
		escrow = uc.buildEscrowForOrder(shop, orderSN, o.OrderStatus, itemsTotal)
	}

	carrier := o.ShippingCarrier
	if carrier == "" {
		carrier = o.CheckoutShippingCarrier
	}
	if carrier == "" && len(o.PackageList) > 0 {
		carrier = o.PackageList[0].ShippingCarrier
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
		carrier,
		"",
		o.TotalAmount,
		o.BuyerCancelReason,
		o.CreateTime,
		o.UpdateTime,
		o.PayTime,
		items,
		escrow,
		time.Unix(o.CreateTime, 0),
		time.Unix(o.UpdateTime, 0),
	)

	var costingCtx *usecaseProduct.CostingContext
	if uc.processOrderUC != nil && uc.processOrderUC.SKUMatcher() != nil && uc.processOrderUC.SKUMatcher().CostCalculator() != nil {
		costingCtx = uc.processOrderUC.SKUMatcher().CostCalculator().LoadContext(ctx, uuid.Nil)
	}

	if uc.processOrderUC != nil {
		if err := uc.processOrderUC.Execute(ctx, domainOrder, costingCtx); err != nil {
			return nil, fmt.Errorf("failed to process and save order: %w", err)
		}
	} else {
		if err := uc.repo.SaveOrder(ctx, domainOrder); err != nil {
			return nil, fmt.Errorf("failed to save order to database: %w", err)
		}
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

	// Auto-create/update/void finance_transaction saat pesanan selesai atau di-refund
	if ord != nil && ord.Escrow() != nil && uc.financeRepo != nil {
		if ord.Escrow().FinancialStatus() == "REFUNDED" || ord.Escrow().EscrowAmount() <= 0 {
			tx, txErr := finance.NewFinanceTransaction(
				finance.TypeIncome,
				finance.CategorySalesShopee,
				nil,
				0,
				1,
				"Escrow Shopee - Order "+orderSN+" [REFUNDED]",
				time.Now(),
				"SHOPEE_ESCROW",
				orderSN,
				"",
				"",
			)
			if txErr == nil {
				_ = uc.financeRepo.UpsertShopeeEscrowTransaction(ctx, tx)
			}
		} else if ord.OrderStatus() == "COMPLETED" && ord.Escrow().EscrowAmount() > 0 {
			releaseDate := time.Now()
			tx, txErr := finance.NewFinanceTransaction(
				finance.TypeIncome,
				finance.CategorySalesShopee,
				nil,
				ord.Escrow().EscrowAmount(),
				1,
				"Escrow Shopee - Order "+orderSN,
				releaseDate,
				"SHOPEE_ESCROW",
				orderSN,
				"",
				"",
			)
			if txErr == nil {
				_ = uc.financeRepo.UpsertShopeeEscrowTransaction(ctx, tx)
			}
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

	client, err := uc.getClient(ctx)
	if err != nil {
		return nil, err
	}

	paramResp, err := client.GetShippingParameter(shop.AccessToken(), shop.ShopID(), orderSN)
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

	shipResp, err := client.ShipOrder(shop.AccessToken(), shop.ShopID(), shipReq)
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

	client, err := uc.getClient(ctx)
	if err != nil {
		return nil, err
	}
	return client.DownloadShippingDocument(shop.AccessToken(), shop.ShopID(), orderSN)
}

// 4. Financial & 7 Pos Kas Escrow
func (uc *ShopeeUseCases) GetCashflowSummary(ctx context.Context) (*shopee.CashflowSummary, error) {
	return uc.repo.GetCashflowSummary(ctx)
}

func (uc *ShopeeUseCases) RecalculateFinances(ctx context.Context) (int, error) {
	if uc.recalcUC != nil {
		return uc.recalcUC.Execute(ctx)
	}
	return uc.repo.RecalculateFinances(ctx)
}

func (uc *ShopeeUseCases) buildEscrowForOrder(shop *shopee.ShopeeShop, orderSN string, orderStatus string, itemsTotal float64) *shopee.ShopeeOrderEscrow {
	client, err := uc.getClient(context.Background())
	if err != nil || client == nil || shop == nil {
		return nil
	}
	escrowResp, err := client.GetEscrowDetail(shop.AccessToken(), shop.ShopID(), orderSN)
	if err != nil || escrowResp == nil {
		return nil
	}
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
	if orderStatus == "COMPLETED" || orderStatus == "SHIPPED" || orderStatus == "TO_CONFIRM_RECEIVE" {
		if inc.EscrowAmount == 0 && (inc.BuyerTotalAmount > 0 || itemsTotal > 0) {
			finStatus = "REFUNDED"
		} else if orderStatus == "COMPLETED" {
			finStatus = "RELEASED"
		}
	} else if orderStatus == "CANCELLED" || orderStatus == "IN_CANCEL" {
		finStatus = "CANCELLED"
	}

	sellingPrice := inc.OrderSellingPrice
	if sellingPrice <= 0 {
		sellingPrice = inc.SellingPrice
	}
	if sellingPrice <= 0 && itemsTotal > 0 {
		sellingPrice = itemsTotal
	}
	if sellingPrice <= 0 && inc.EscrowAmount > 0 {
		sellingPrice = inc.EscrowAmount + totalMarketFee
	}

	return shopee.ReconstructOrderEscrow(
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
}
