package shopee

import (
	"time"

	"github.com/google/uuid"
)

type ShopeeShop struct {
	id                    uint64
	shopID                uint64
	shopName              string
	region                string
	accessToken           string
	refreshToken          string
	accessTokenExpiresAt  time.Time
	refreshTokenExpiresAt time.Time
	createdAt             time.Time
	updatedAt             time.Time
}

func ReconstructShop(
	id, shopID uint64,
	shopName, region, accessToken, refreshToken string,
	accessExpires, refreshExpires, createdAt, updatedAt time.Time,
) *ShopeeShop {
	return &ShopeeShop{
		id:                    id,
		shopID:                shopID,
		shopName:              shopName,
		region:                region,
		accessToken:           accessToken,
		refreshToken:          refreshToken,
		accessTokenExpiresAt:  accessExpires,
		refreshTokenExpiresAt: refreshExpires,
		createdAt:             createdAt,
		updatedAt:             updatedAt,
	}
}

func (s *ShopeeShop) ID() uint64                    { return s.id }
func (s *ShopeeShop) ShopID() uint64                { return s.shopID }
func (s *ShopeeShop) ShopName() string              { return s.shopName }
func (s *ShopeeShop) Region() string                { return s.region }
func (s *ShopeeShop) AccessToken() string           { return s.accessToken }
func (s *ShopeeShop) RefreshToken() string          { return s.refreshToken }
func (s *ShopeeShop) AccessTokenExpiresAt() time.Time  { return s.accessTokenExpiresAt }
func (s *ShopeeShop) RefreshTokenExpiresAt() time.Time { return s.refreshTokenExpiresAt }
func (s *ShopeeShop) CreatedAt() time.Time          { return s.createdAt }
func (s *ShopeeShop) UpdatedAt() time.Time          { return s.updatedAt }
func (s *ShopeeShop) IsTokenExpired() bool {
	return time.Now().Add(5 * time.Minute).After(s.accessTokenExpiresAt)
}

func (s *ShopeeShop) UpdateTokens(access, refresh string, accessExpires, refreshExpires time.Time) {
	s.accessToken = access
	s.refreshToken = refresh
	s.accessTokenExpiresAt = accessExpires
	s.refreshTokenExpiresAt = refreshExpires
	s.updatedAt = time.Now()
}

type ShopeeOrderItem struct {
	id               uint64
	orderSN          string
	itemID           uint64
	itemName         string
	itemSKU          string
	modelID          uint64
	modelName        string
	modelSKU         string
	quantity         int
	originalPrice    float64
	discountedPrice  float64
	productID        *uuid.UUID
	matchedSKU       string
	mappingStatus    string
	filamentCost     float64
	hardwareCost     float64
	packagingCost    float64
	machineCost      float64
	baseHPP          float64
	netProfit        float64
	electricityCost  float64
	maintenanceCost  float64
	depreciationCost float64
	createdAt        time.Time
	updatedAt        time.Time
}

func ReconstructOrderItem(
	id uint64, orderSN string, itemID uint64, itemName, itemSKU string,
	modelID uint64, modelName, modelSKU string, quantity int,
	origPrice, discPrice float64, productID *uuid.UUID, matchedSKU, mappingStatus string,
	filCost, hwCost, pkgCost, machCost, baseHPP, netProfit, elecCost, maintCost, depCost float64,
	createdAt, updatedAt time.Time,
) ShopeeOrderItem {
	return ShopeeOrderItem{
		id:               id,
		orderSN:          orderSN,
		itemID:           itemID,
		itemName:         itemName,
		itemSKU:          itemSKU,
		modelID:          modelID,
		modelName:        modelName,
		modelSKU:         modelSKU,
		quantity:         quantity,
		originalPrice:    origPrice,
		discountedPrice:  discPrice,
		productID:        productID,
		matchedSKU:       matchedSKU,
		mappingStatus:    mappingStatus,
		filamentCost:     filCost,
		hardwareCost:     hwCost,
		packagingCost:    pkgCost,
		machineCost:      machCost,
		baseHPP:          baseHPP,
		netProfit:        netProfit,
		electricityCost:  elecCost,
		maintenanceCost:  maintCost,
		depreciationCost: depCost,
		createdAt:        createdAt,
		updatedAt:        updatedAt,
	}
}

func (i *ShopeeOrderItem) ID() uint64               { return i.id }
func (i *ShopeeOrderItem) OrderSN() string          { return i.orderSN }
func (i *ShopeeOrderItem) ItemID() uint64           { return i.itemID }
func (i *ShopeeOrderItem) ItemName() string         { return i.itemName }
func (i *ShopeeOrderItem) ItemSKU() string          { return i.itemSKU }
func (i *ShopeeOrderItem) ModelID() uint64          { return i.modelID }
func (i *ShopeeOrderItem) ModelName() string        { return i.modelName }
func (i *ShopeeOrderItem) ModelSKU() string         { return i.modelSKU }
func (i *ShopeeOrderItem) Quantity() int            { return i.quantity }
func (i *ShopeeOrderItem) OriginalPrice() float64   { return i.originalPrice }
func (i *ShopeeOrderItem) DiscountedPrice() float64 { return i.discountedPrice }
func (i *ShopeeOrderItem) ProductID() *uuid.UUID    { return i.productID }
func (i *ShopeeOrderItem) MatchedSKU() string       { return i.matchedSKU }
func (i *ShopeeOrderItem) MappingStatus() string    { return i.mappingStatus }
func (i *ShopeeOrderItem) FilamentCost() float64    { return i.filamentCost }
func (i *ShopeeOrderItem) HardwareCost() float64    { return i.hardwareCost }
func (i *ShopeeOrderItem) PackagingCost() float64   { return i.packagingCost }
func (i *ShopeeOrderItem) MachineCost() float64     { return i.machineCost }
func (i *ShopeeOrderItem) BaseHPP() float64         { return i.baseHPP }
func (i *ShopeeOrderItem) NetProfit() float64       { return i.netProfit }
func (i *ShopeeOrderItem) ElectricityCost() float64 { return i.electricityCost }
func (i *ShopeeOrderItem) MaintenanceCost() float64 { return i.maintenanceCost }
func (i *ShopeeOrderItem) DepreciationCost() float64{ return i.depreciationCost }
func (i *ShopeeOrderItem) CreatedAt() time.Time     { return i.createdAt }
func (i *ShopeeOrderItem) UpdatedAt() time.Time     { return i.updatedAt }

type ShopeeOrderEscrow struct {
	orderSN                  string
	escrowAmount             float64
	sellingPrice             float64
	commissionFee            float64
	commissionRuleName       string
	commissionPercentage     float64
	serviceFee               float64
	serviceRuleName          string
	servicePercentage        float64
	sellerTransactionFee     float64
	sellerOrderProcessingFee float64
	sellerVoucherDiscount    float64
	totalMarketplaceFee      float64
	totalHPP                 float64
	kasFilamen               float64
	kasKomponen              float64
	kasPacking               float64
	kasListrik               float64
	kasMaintenance           float64
	kasDepresiasi            float64
	kasLabaBersih            float64
	financialStatus          string // PENDING, RELEASED
	createdAt                time.Time
	updatedAt                time.Time
}

func ReconstructOrderEscrow(
	orderSN string, escrowAmount, sellingPrice, commissionFee float64,
	commRule string, commPct, serviceFee float64, svcRule string, svcPct,
	sellerTxFee, sellerProcFee, sellerVoucher, totalMarketplaceFee, totalHPP,
	kasFil, kasKomp, kasPack, kasElec, kasMaint, kasDep, kasLaba float64,
	finStatus string, createdAt, updatedAt time.Time,
) *ShopeeOrderEscrow {
	return &ShopeeOrderEscrow{
		orderSN:                  orderSN,
		escrowAmount:             escrowAmount,
		sellingPrice:             sellingPrice,
		commissionFee:            commissionFee,
		commissionRuleName:       commRule,
		commissionPercentage:     commPct,
		serviceFee:               serviceFee,
		serviceRuleName:          svcRule,
		servicePercentage:        svcPct,
		sellerTransactionFee:     sellerTxFee,
		sellerOrderProcessingFee: sellerProcFee,
		sellerVoucherDiscount:    sellerVoucher,
		totalMarketplaceFee:      totalMarketplaceFee,
		totalHPP:                 totalHPP,
		kasFilamen:               kasFil,
		kasKomponen:              kasKomp,
		kasPacking:               kasPack,
		kasListrik:               kasElec,
		kasMaintenance:           kasMaint,
		kasDepresiasi:            kasDep,
		kasLabaBersih:            kasLaba,
		financialStatus:          finStatus,
		createdAt:                createdAt,
		updatedAt:                updatedAt,
	}
}

func (e *ShopeeOrderEscrow) OrderSN() string                  { return e.orderSN }
func (e *ShopeeOrderEscrow) EscrowAmount() float64             { return e.escrowAmount }
func (e *ShopeeOrderEscrow) SellingPrice() float64             { return e.sellingPrice }
func (e *ShopeeOrderEscrow) CommissionFee() float64            { return e.commissionFee }
func (e *ShopeeOrderEscrow) CommissionRuleName() string        { return e.commissionRuleName }
func (e *ShopeeOrderEscrow) CommissionPercentage() float64    { return e.commissionPercentage }
func (e *ShopeeOrderEscrow) ServiceFee() float64               { return e.serviceFee }
func (e *ShopeeOrderEscrow) ServiceRuleName() string           { return e.serviceRuleName }
func (e *ShopeeOrderEscrow) ServicePercentage() float64       { return e.servicePercentage }
func (e *ShopeeOrderEscrow) SellerTransactionFee() float64     { return e.sellerTransactionFee }
func (e *ShopeeOrderEscrow) SellerOrderProcessingFee() float64 { return e.sellerOrderProcessingFee }
func (e *ShopeeOrderEscrow) SellerVoucherDiscount() float64    { return e.sellerVoucherDiscount }
func (e *ShopeeOrderEscrow) TotalMarketplaceFee() float64      { return e.totalMarketplaceFee }
func (e *ShopeeOrderEscrow) TotalHPP() float64                 { return e.totalHPP }
func (e *ShopeeOrderEscrow) KasFilamen() float64               { return e.kasFilamen }
func (e *ShopeeOrderEscrow) KasKomponen() float64              { return e.kasKomponen }
func (e *ShopeeOrderEscrow) KasPacking() float64               { return e.kasPacking }
func (e *ShopeeOrderEscrow) KasListrik() float64               { return e.kasListrik }
func (e *ShopeeOrderEscrow) KasMaintenance() float64           { return e.kasMaintenance }
func (e *ShopeeOrderEscrow) KasDepresiasi() float64            { return e.kasDepresiasi }
func (e *ShopeeOrderEscrow) KasLabaBersih() float64            { return e.kasLabaBersih }
func (e *ShopeeOrderEscrow) FinancialStatus() string          { return e.financialStatus }
func (e *ShopeeOrderEscrow) CreatedAt() time.Time              { return e.createdAt }
func (e *ShopeeOrderEscrow) UpdatedAt() time.Time              { return e.updatedAt }

type ShopeeOrder struct {
	orderSN           string
	shopID            uint64
	orderStatus       string
	buyerUserID       uint64
	buyerUsername     string
	messageToSeller   string
	shipByDate        int64
	shipByDateTime    *time.Time
	shippingCarrier   string
	trackingNumber    string
	totalAmount       float64
	buyerCancelReason string
	createTimeShopee  int64
	updateTimeShopee  int64
	items             []ShopeeOrderItem
	escrow            *ShopeeOrderEscrow
	createdAt         time.Time
	updatedAt         time.Time
}

func ReconstructOrder(
	orderSN string, shopID uint64, orderStatus string, buyerUserID uint64, buyerUsername, msg string,
	shipByDate int64, shipByDateTime *time.Time, shippingCarrier, trackingNumber string,
	totalAmount float64, cancelReason string, createTimeShopee, updateTimeShopee int64,
	items []ShopeeOrderItem, escrow *ShopeeOrderEscrow, createdAt, updatedAt time.Time,
) *ShopeeOrder {
	return &ShopeeOrder{
		orderSN:           orderSN,
		shopID:            shopID,
		orderStatus:       orderStatus,
		buyerUserID:       buyerUserID,
		buyerUsername:     buyerUsername,
		messageToSeller:   msg,
		shipByDate:        shipByDate,
		shipByDateTime:    shipByDateTime,
		shippingCarrier:   shippingCarrier,
		trackingNumber:    trackingNumber,
		totalAmount:       totalAmount,
		buyerCancelReason: cancelReason,
		createTimeShopee:  createTimeShopee,
		updateTimeShopee:  updateTimeShopee,
		items:             items,
		escrow:            escrow,
		createdAt:         createdAt,
		updatedAt:         updatedAt,
	}
}

func (o *ShopeeOrder) OrderSN() string             { return o.orderSN }
func (o *ShopeeOrder) ShopID() uint64              { return o.shopID }
func (o *ShopeeOrder) OrderStatus() string         { return o.orderStatus }
func (o *ShopeeOrder) BuyerUserID() uint64         { return o.buyerUserID }
func (o *ShopeeOrder) BuyerUsername() string       { return o.buyerUsername }
func (o *ShopeeOrder) MessageToSeller() string     { return o.messageToSeller }
func (o *ShopeeOrder) ShipByDate() int64           { return o.shipByDate }
func (o *ShopeeOrder) ShipByDateTime() *time.Time  { return o.shipByDateTime }
func (o *ShopeeOrder) ShippingCarrier() string     { return o.shippingCarrier }
func (o *ShopeeOrder) TrackingNumber() string      { return o.trackingNumber }
func (o *ShopeeOrder) TotalAmount() float64        { return o.totalAmount }
func (o *ShopeeOrder) BuyerCancelReason() string   { return o.buyerCancelReason }
func (o *ShopeeOrder) CreateTimeShopee() int64     { return o.createTimeShopee }
func (o *ShopeeOrder) UpdateTimeShopee() int64     { return o.updateTimeShopee }
func (o *ShopeeOrder) Items() []ShopeeOrderItem    { return o.items }
func (o *ShopeeOrder) Escrow() *ShopeeOrderEscrow  { return o.escrow }
func (o *ShopeeOrder) CreatedAt() time.Time        { return o.createdAt }
func (o *ShopeeOrder) UpdatedAt() time.Time        { return o.updatedAt }

type CashflowSummary struct {
	TotalCompletedOrders int64
	TotalOrders          int64
	TotalGrossSales      float64
	TotalMarketplaceFees float64
	TotalEscrowNetIn     float64
	TotalHPP             float64
	KasFilamen           float64
	KasKomponen          float64
	KasPacking           float64
	KasListrik           float64
	KasMaintenance       float64
	KasDepresiasi        float64
	KasLabaBersih        float64
	AverageProfitMargin  float64
	UnmappedItemsCount   int64
}
