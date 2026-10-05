package dto

import (
	"time"

	"github.com/google/uuid"

	"symetra-lab-backend-v2/internal/domain/order"
)

type CreateOrderItemRequest struct {
	ProductID      *string  `json:"product_id"`
	ProductName    string   `json:"product_name"`
	Quantity       int      `json:"quantity"`
	SellingPrice   float64  `json:"selling_price"`
	HPP            *float64 `json:"hpp"`
	WeightGrams    *float64 `json:"weight_grams"`
	PrintTimeHours *float64 `json:"print_time_hours"`
	MachineID      *string  `json:"machine_id"`
}

type CreateOrderRequest struct {
	CustomerName    string                   `json:"customer_name"`
	CustomerContact string                   `json:"customer_contact"`
	Notes           string                   `json:"notes"`
	Source          string                   `json:"source"`
	PaymentStatus   string                   `json:"payment_status"`
	Items           []CreateOrderItemRequest `json:"items"`
}

type UpdateOrderStatusRequest struct {
	Status      string `json:"status"`
	OrderStatus string `json:"order_status"` // support v1 naming
}

type UpdatePaymentStatusRequest struct {
	PaymentStatus string `json:"payment_status"`
}

type OrderItemResponse struct {
	ID               uuid.UUID  `json:"id"`
	OrderID          uuid.UUID  `json:"order_id"`
	ProductID        *uuid.UUID `json:"product_id,omitempty"`
	ProductName      string     `json:"product_name"`
	ItemSKU          string     `json:"item_sku,omitempty"`
	Quantity         int        `json:"quantity"`
	SellingPrice     float64    `json:"selling_price"`
	HPP              float64    `json:"hpp"`
	WeightGrams      float64    `json:"weight_grams"`
	PrintTimeHours   float64    `json:"print_time_hours"`
	MachineID        *uuid.UUID `json:"machine_id,omitempty"`
	FilamentCost     float64    `json:"filament_cost"`
	ComponentCost    float64    `json:"component_cost"`
	PackagingCost    float64    `json:"packaging_cost"`
	ElectricityCost  float64    `json:"electricity_cost"`
	MaintenanceCost  float64    `json:"maintenance_cost"`
	DepreciationCost float64    `json:"depreciation_cost"`
	TotalCogs        float64    `json:"total_cogs"`
	NetProfit        float64    `json:"net_profit"`
	EnergyCost       float64    `json:"energy_cost"`
	PackingFee       float64    `json:"packing_fee"`
	ChannelItemID    int64      `json:"channel_item_id,omitempty"`
	ChannelModelID   int64      `json:"channel_model_id,omitempty"`
	MappingStatus    string     `json:"mapping_status,omitempty"`
	MatchedSKU       string     `json:"matched_sku,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
}

type OrderResponse struct {
	ID               uuid.UUID           `json:"id"`
	UserID           uuid.UUID           `json:"user_id"`
	OrderNumber      string              `json:"order_number"`
	CustomerName     string              `json:"customer_name"`
	CustomerContact  string              `json:"customer_contact"`
	Channel          string              `json:"channel"`
	GrossAmount      float64             `json:"gross_amount"`
	ChannelFee       float64             `json:"channel_fee"`
	NetAmount        float64             `json:"net_amount"`
	CogsAmount       float64             `json:"cogs_amount"`
	NetProfit        float64             `json:"net_profit"`
	FundFilament     float64             `json:"fund_filament"`
	FundComponent    float64             `json:"fund_component"`
	FundPackaging    float64             `json:"fund_packaging"`
	FundElectricity  float64             `json:"fund_electricity"`
	FundMaintenance  float64             `json:"fund_maintenance"`
	FundDepreciation float64             `json:"fund_depreciation"`
	FundNetProfit    float64             `json:"fund_net_profit"`
	TotalRevenue     float64             `json:"total_revenue"`
	TotalHPP         float64             `json:"total_hpp"`
	TotalProfit      float64             `json:"total_profit"`
	Status           string              `json:"status"`
	Notes            string              `json:"notes"`
	Source           string              `json:"source"`
	PaymentStatus    string              `json:"payment_status"`
	StartedAt        *time.Time          `json:"started_at,omitempty"`
	CompletedAt      *time.Time          `json:"completed_at,omitempty"`
	Items            []OrderItemResponse `json:"items,omitempty"`
	CreatedAt        time.Time           `json:"created_at"`
	UpdatedAt        time.Time           `json:"updated_at"`
}

type OrderAllocationSummaryResponse struct {
	TotalOrders         int64   `json:"total_orders"`
	TotalGrossSales     float64 `json:"total_gross_sales"`
	TotalChannelFees    float64 `json:"total_channel_fees"`
	TotalNetRevenue     float64 `json:"total_net_revenue"`
	TotalCOGS           float64 `json:"total_cogs"`
	TotalNetProfit      float64 `json:"total_net_profit"`
	AverageProfitMargin float64 `json:"average_profit_margin"`
	FundFilament        float64 `json:"fund_filament"`
	FundComponent       float64 `json:"fund_component"`
	FundPackaging       float64 `json:"fund_packaging"`
	FundElectricity     float64 `json:"fund_electricity"`
	FundMaintenance     float64 `json:"fund_maintenance"`
	FundDepreciation    float64 `json:"fund_depreciation"`
	FundNetProfit       float64 `json:"fund_net_profit"`
	UnmappedItemsCount  int64   `json:"unmapped_items_count"`

	// Inflow Alokasi dari Penjualan Pesanan
	AllocatedFilament    float64 `json:"allocated_filament"`
	AllocatedComponent   float64 `json:"allocated_component"`
	AllocatedPackaging   float64 `json:"allocated_packaging"`
	AllocatedElectricity float64 `json:"allocated_electricity"`
	AllocatedMaintenance float64 `json:"allocated_maintenance"`
	AllocatedDepreciation float64 `json:"allocated_depreciation"`
	AllocatedNetProfit   float64 `json:"allocated_net_profit"`

	// Outflow Pengeluaran Belanja (PO & Expenses)
	SpentFilament        float64 `json:"spent_filament"`
	SpentComponent       float64 `json:"spent_component"`
	SpentPackaging       float64 `json:"spent_packaging"`
	SpentElectricity     float64 `json:"spent_electricity"`
	SpentMaintenance     float64 `json:"spent_maintenance"`
	SpentDepreciation    float64 `json:"spent_depreciation"`
	SpentNetProfit       float64 `json:"spent_net_profit"`

	// Legacy aliases for backwards compatibility
	TotalCompletedOrders int64   `json:"total_completed_orders"`
	TotalEscrowNetIn     float64 `json:"total_escrow_net_in"`
	TotalMarketplaceFees float64 `json:"total_marketplace_fees"`
	TotalHPP             float64 `json:"total_hpp"`
	KasFilamen           float64 `json:"kas_filamen"`
	KasKomponen          float64 `json:"kas_komponen"`
	KasPacking           float64 `json:"kas_packing"`
	KasListrik           float64 `json:"kas_listrik"`
	KasMaintenance       float64 `json:"kas_maintenance"`
	KasDepresiasi        float64 `json:"kas_depresiasi"`
	KasLabaBersih        float64 `json:"kas_laba_bersih"`
}

func ToOrderItemResponse(it *order.OrderItem) OrderItemResponse {
	return OrderItemResponse{
		ID:               it.ID(),
		OrderID:          it.OrderID(),
		ProductID:        it.ProductID(),
		ProductName:      it.ProductName(),
		ItemSKU:          it.ItemSKU(),
		Quantity:         it.Quantity(),
		SellingPrice:     it.SellingPrice(),
		HPP:              it.HPP(),
		WeightGrams:      it.WeightGrams(),
		PrintTimeHours:   it.PrintTimeHours(),
		MachineID:        it.MachineID(),
		FilamentCost:     it.FilamentCost(),
		ComponentCost:    it.ComponentCost(),
		PackagingCost:    it.PackagingCost(),
		ElectricityCost:  it.ElectricityCost(),
		MaintenanceCost:  it.MaintenanceCost(),
		DepreciationCost: it.DepreciationCost(),
		TotalCogs:        it.TotalCogs(),
		NetProfit:        it.NetProfit(),
		EnergyCost:       it.EnergyCost(),
		PackingFee:       it.PackingFee(),
		ChannelItemID:    it.ChannelItemID(),
		ChannelModelID:   it.ChannelModelID(),
		MappingStatus:    it.MappingStatus(),
		MatchedSKU:       it.MatchedSKU(),
		CreatedAt:        it.CreatedAt(),
	}
}

func ToOrderResponse(o *order.Order) OrderResponse {
	items := make([]OrderItemResponse, len(o.Items()))
	for i, it := range o.Items() {
		itCopy := it
		items[i] = ToOrderItemResponse(&itCopy)
	}

	gross := o.GrossAmount()
	if gross == 0 {
		gross = o.TotalRevenue()
	}
	net := o.NetAmount()
	if net == 0 {
		net = gross
	}
	totRev := o.TotalRevenue()
	if totRev == 0 {
		totRev = gross
	}
	if totRev == 0 {
		totRev = net
	}
	cogs := o.CogsAmount()
	if cogs == 0 {
		cogs = o.TotalHPP()
	}
	totHpp := o.TotalHPP()
	if totHpp == 0 {
		totHpp = cogs
	}
	profit := o.NetProfit()
	if profit == 0 {
		profit = o.TotalProfit()
	}
	totProfit := o.TotalProfit()
	if totProfit == 0 {
		totProfit = profit
	}

	return OrderResponse{
		ID:               o.ID(),
		UserID:           o.UserID(),
		OrderNumber:      o.OrderNumber(),
		CustomerName:     o.CustomerName(),
		CustomerContact:  o.CustomerContact(),
		Channel:          o.Channel(),
		GrossAmount:      gross,
		ChannelFee:       o.ChannelFee(),
		NetAmount:        net,
		CogsAmount:       cogs,
		NetProfit:        profit,
		FundFilament:     o.FundFilament(),
		FundComponent:    o.FundComponent(),
		FundPackaging:    o.FundPackaging(),
		FundElectricity:  o.FundElectricity(),
		FundMaintenance:  o.FundMaintenance(),
		FundDepreciation: o.FundDepreciation(),
		FundNetProfit:    o.FundNetProfit(),
		TotalRevenue:     totRev,
		TotalHPP:         totHpp,
		TotalProfit:      totProfit,
		Status:           o.Status(),
		Notes:            o.Notes(),
		Source:           o.Source(),
		PaymentStatus:    o.PaymentStatus(),
		StartedAt:        o.StartedAt(),
		CompletedAt:      o.CompletedAt(),
		Items:            items,
		CreatedAt:        o.CreatedAt(),
		UpdatedAt:        o.UpdatedAt(),
	}
}
