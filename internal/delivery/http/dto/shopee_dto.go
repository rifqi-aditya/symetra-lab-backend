package dto

import (
	"time"

	"github.com/google/uuid"

	"symetra-lab-backend-v2/internal/domain/shopee"
)

type ShopeeShopResponse struct {
	ID                   uint64    `json:"id"`
	ShopID               uint64    `json:"shop_id"`
	ShopName             string    `json:"shop_name"`
	Region               string    `json:"region"`
	AccessTokenExpiresAt time.Time `json:"access_token_expires_at"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

func ToShopResponse(s *shopee.ShopeeShop) ShopeeShopResponse {
	return ShopeeShopResponse{
		ID:                   s.ID(),
		ShopID:               s.ShopID(),
		ShopName:             s.ShopName(),
		Region:               s.Region(),
		AccessTokenExpiresAt: s.AccessTokenExpiresAt(),
		CreatedAt:            s.CreatedAt(),
		UpdatedAt:            s.UpdatedAt(),
	}
}

type ShopeeOrderItemResponse struct {
	ID               uint64     `json:"id"`
	OrderSN          string     `json:"order_sn"`
	ItemID           uint64     `json:"item_id"`
	ItemName         string     `json:"item_name"`
	ItemSKU          string     `json:"item_sku"`
	ModelID          uint64     `json:"model_id"`
	ModelName        string     `json:"model_name"`
	ModelSKU         string     `json:"model_sku"`
	Quantity         int        `json:"quantity"`
	OriginalPrice    float64    `json:"original_price"`
	DiscountedPrice  float64    `json:"discounted_price"`
	ProductID        *uuid.UUID `json:"product_id,omitempty"`
	MatchedSKU       string     `json:"matched_sku,omitempty"`
	MappingStatus    string     `json:"mapping_status"`
	FilamentCost     float64    `json:"filament_cost"`
	HardwareCost     float64    `json:"hardware_cost"`
	PackagingCost    float64    `json:"packaging_cost"`
	MachineCost      float64    `json:"machine_cost"`
	BaseHPP          float64    `json:"base_hpp"`
	NetProfit        float64    `json:"net_profit"`
	ElectricityCost  float64    `json:"electricity_cost"`
	MaintenanceCost  float64    `json:"maintenance_cost"`
	DepreciationCost float64    `json:"depreciation_cost"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

func ToShopeeOrderItemResponse(i *shopee.ShopeeOrderItem) ShopeeOrderItemResponse {
	return ShopeeOrderItemResponse{
		ID:               i.ID(),
		OrderSN:          i.OrderSN(),
		ItemID:           i.ItemID(),
		ItemName:         i.ItemName(),
		ItemSKU:          i.ItemSKU(),
		ModelID:          i.ModelID(),
		ModelName:        i.ModelName(),
		ModelSKU:         i.ModelSKU(),
		Quantity:         i.Quantity(),
		OriginalPrice:    i.OriginalPrice(),
		DiscountedPrice:  i.DiscountedPrice(),
		ProductID:        i.ProductID(),
		MatchedSKU:       i.MatchedSKU(),
		MappingStatus:    i.MappingStatus(),
		FilamentCost:     i.FilamentCost(),
		HardwareCost:     i.HardwareCost(),
		PackagingCost:    i.PackagingCost(),
		MachineCost:      i.MachineCost(),
		BaseHPP:          i.BaseHPP(),
		NetProfit:        i.NetProfit(),
		ElectricityCost:  i.ElectricityCost(),
		MaintenanceCost:  i.MaintenanceCost(),
		DepreciationCost: i.DepreciationCost(),
		CreatedAt:        i.CreatedAt(),
		UpdatedAt:        i.UpdatedAt(),
	}
}

type ShopeeOrderEscrowResponse struct {
	OrderSN                  string    `json:"order_sn"`
	EscrowAmount             float64   `json:"escrow_amount"`
	SellingPrice             float64   `json:"selling_price"`
	CommissionFee            float64   `json:"commission_fee"`
	CommissionRuleName       string    `json:"commission_rule_name,omitempty"`
	CommissionPercentage     float64   `json:"commission_percentage"`
	ServiceFee               float64   `json:"service_fee"`
	ServiceRuleName          string    `json:"service_rule_name,omitempty"`
	ServicePercentage        float64   `json:"service_percentage"`
	SellerTransactionFee     float64   `json:"seller_transaction_fee"`
	SellerOrderProcessingFee float64   `json:"seller_order_processing_fee"`
	SellerVoucherDiscount    float64   `json:"seller_voucher_discount"`
	TotalMarketplaceFee      float64   `json:"total_marketplace_fee"`
	TotalHPP                 float64   `json:"total_hpp"`
	KasFilamen               float64   `json:"kas_filamen"`
	KasKomponen              float64   `json:"kas_komponen"`
	KasPacking               float64   `json:"kas_packing"`
	KasListrik               float64   `json:"kas_listrik"`
	KasMaintenance           float64   `json:"kas_maintenance"`
	KasDepresiasi            float64   `json:"kas_depresiasi"`
	KasLabaBersih            float64   `json:"kas_laba_bersih"`
	FinancialStatus          string    `json:"financial_status"`
	CreatedAt                time.Time `json:"created_at"`
	UpdatedAt                time.Time `json:"updated_at"`
}

func ToShopeeOrderEscrowResponse(e *shopee.ShopeeOrderEscrow) *ShopeeOrderEscrowResponse {
	if e == nil {
		return nil
	}
	return &ShopeeOrderEscrowResponse{
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
}

type ShopeeOrderResponse struct {
	OrderSN           string                      `json:"order_sn"`
	ShopID            uint64                      `json:"shop_id"`
	OrderStatus       string                      `json:"order_status"`
	BuyerUserID       uint64                      `json:"buyer_user_id"`
	BuyerUsername     string                      `json:"buyer_username"`
	MessageToSeller   string                      `json:"message_to_seller"`
	ShipByDate        int64                       `json:"ship_by_date"`
	ShipByDateTime    *time.Time                  `json:"ship_by_date_time,omitempty"`
	ShippingCarrier   string                      `json:"shipping_carrier"`
	TrackingNumber    string                      `json:"tracking_number"`
	TotalAmount       float64                     `json:"total_amount"`
	BuyerCancelReason string                      `json:"buyer_cancel_reason,omitempty"`
	CreateTimeShopee  int64                       `json:"create_time_shopee"`
	UpdateTimeShopee  int64                       `json:"update_time_shopee"`
	Items             []ShopeeOrderItemResponse   `json:"items,omitempty"`
	Escrow            *ShopeeOrderEscrowResponse  `json:"escrow,omitempty"`
	CreatedAt         time.Time                   `json:"created_at"`
	UpdatedAt         time.Time                   `json:"updated_at"`
}

func ToShopeeOrderResponse(o *shopee.ShopeeOrder) ShopeeOrderResponse {
	items := make([]ShopeeOrderItemResponse, len(o.Items()))
	for i, it := range o.Items() {
		itCopy := it
		items[i] = ToShopeeOrderItemResponse(&itCopy)
	}

	return ShopeeOrderResponse{
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
		Items:             items,
		Escrow:            ToShopeeOrderEscrowResponse(o.Escrow()),
		CreatedAt:         o.CreatedAt(),
		UpdatedAt:         o.UpdatedAt(),
	}
}

type CashflowSummaryResponse struct {
	TotalCompletedOrders int64   `json:"total_completed_orders"`
	TotalOrders          int64   `json:"total_orders"`
	TotalGrossSales      float64 `json:"total_gross_sales"`
	TotalMarketplaceFees float64 `json:"total_marketplace_fees"`
	TotalEscrowNetIn     float64 `json:"total_escrow_net_in"`
	TotalHPP             float64 `json:"total_hpp"`
	
	// 7 Pos Kas Keuangan Real-time
	KasFilamen     float64 `json:"kas_filamen"`
	KasKomponen    float64 `json:"kas_komponen"`
	KasPacking     float64 `json:"kas_packing"`
	KasListrik     float64 `json:"kas_listrik"`
	KasMaintenance float64 `json:"kas_maintenance"`
	KasDepresiasi  float64 `json:"kas_depresiasi"`
	KasLabaBersih  float64 `json:"kas_laba_bersih"`

	// Frontend aliases
	FundFilament     float64 `json:"fund_filament"`
	FundHardware     float64 `json:"fund_hardware"`
	FundPackaging    float64 `json:"fund_packaging"`
	FundElectricity  float64 `json:"fund_electricity"`
	FundMaintenance  float64 `json:"fund_maintenance"`
	FundDepreciation float64 `json:"fund_depreciation"`
	FundNetProfit    float64 `json:"fund_net_profit"`

	AverageProfitMargin float64 `json:"average_profit_margin"`
	UnmappedItemsCount  int64   `json:"unmapped_items_count"`
}

func ToCashflowSummaryResponse(s *shopee.CashflowSummary) CashflowSummaryResponse {
	return CashflowSummaryResponse{
		TotalCompletedOrders: s.TotalCompletedOrders,
		TotalOrders:          s.TotalOrders,
		TotalGrossSales:      s.TotalGrossSales,
		TotalMarketplaceFees: s.TotalMarketplaceFees,
		TotalEscrowNetIn:     s.TotalEscrowNetIn,
		TotalHPP:             s.TotalHPP,
		KasFilamen:           s.KasFilamen,
		KasKomponen:          s.KasKomponen,
		KasPacking:           s.KasPacking,
		KasListrik:           s.KasListrik,
		KasMaintenance:       s.KasMaintenance,
		KasDepresiasi:        s.KasDepresiasi,
		KasLabaBersih:        s.KasLabaBersih,
		FundFilament:         s.KasFilamen,
		FundHardware:         s.KasKomponen,
		FundPackaging:        s.KasPacking,
		FundElectricity:      s.KasListrik,
		FundMaintenance:      s.KasMaintenance,
		FundDepreciation:     s.KasDepresiasi,
		FundNetProfit:        s.KasLabaBersih,
		AverageProfitMargin:  s.AverageProfitMargin,
		UnmappedItemsCount:   s.UnmappedItemsCount,
	}
}

type LinkSKURequest struct {
	ShopeeItemID   uint64    `json:"shopee_item_id"`
	ShopeeModelID  uint64    `json:"shopee_model_id"`
	ShopeeModelSKU string    `json:"shopee_model_sku"`
	ProductID      uuid.UUID `json:"product_id"`
}
