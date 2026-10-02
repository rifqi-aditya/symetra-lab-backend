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
	Quantity         int        `json:"quantity"`
	SellingPrice     float64    `json:"selling_price"`
	HPP              float64    `json:"hpp"`
	WeightGrams      float64    `json:"weight_grams"`
	PrintTimeHours   float64    `json:"print_time_hours"`
	MachineID        *uuid.UUID `json:"machine_id,omitempty"`
	EnergyCost       float64    `json:"energy_cost"`
	DepreciationCost float64    `json:"depreciation_cost"`
	MaintenanceCost  float64    `json:"maintenance_cost"`
	PackingFee       float64    `json:"packing_fee"`
	CreatedAt        time.Time  `json:"created_at"`
}

type OrderResponse struct {
	ID              uuid.UUID           `json:"id"`
	UserID          uuid.UUID           `json:"user_id"`
	OrderNumber     string              `json:"order_number"`
	CustomerName    string              `json:"customer_name"`
	CustomerContact string              `json:"customer_contact"`
	TotalRevenue    float64             `json:"total_revenue"`
	TotalHPP        float64             `json:"total_hpp"`
	TotalProfit     float64             `json:"total_profit"`
	Status          string              `json:"status"`
	Notes           string              `json:"notes"`
	Source          string              `json:"source"`
	PaymentStatus   string              `json:"payment_status"`
	StartedAt       *time.Time          `json:"started_at,omitempty"`
	CompletedAt     *time.Time          `json:"completed_at,omitempty"`
	Items           []OrderItemResponse `json:"items,omitempty"`
	CreatedAt       time.Time           `json:"created_at"`
	UpdatedAt       time.Time           `json:"updated_at"`
}

func ToOrderItemResponse(it *order.OrderItem) OrderItemResponse {
	return OrderItemResponse{
		ID:               it.ID(),
		OrderID:          it.OrderID(),
		ProductID:        it.ProductID(),
		ProductName:      it.ProductName(),
		Quantity:         it.Quantity(),
		SellingPrice:     it.SellingPrice(),
		HPP:              it.HPP(),
		WeightGrams:      it.WeightGrams(),
		PrintTimeHours:   it.PrintTimeHours(),
		MachineID:        it.MachineID(),
		EnergyCost:       it.EnergyCost(),
		DepreciationCost: it.DepreciationCost(),
		MaintenanceCost:  it.MaintenanceCost(),
		PackingFee:       it.PackingFee(),
		CreatedAt:        it.CreatedAt(),
	}
}

func ToOrderResponse(o *order.Order) OrderResponse {
	items := make([]OrderItemResponse, len(o.Items()))
	for i, it := range o.Items() {
		itCopy := it
		items[i] = ToOrderItemResponse(&itCopy)
	}

	return OrderResponse{
		ID:              o.ID(),
		UserID:          o.UserID(),
		OrderNumber:     o.OrderNumber(),
		CustomerName:    o.CustomerName(),
		CustomerContact: o.CustomerContact(),
		TotalRevenue:    o.TotalRevenue(),
		TotalHPP:        o.TotalHPP(),
		TotalProfit:     o.TotalProfit(),
		Status:          o.Status(),
		Notes:           o.Notes(),
		Source:          o.Source(),
		PaymentStatus:   o.PaymentStatus(),
		StartedAt:       o.StartedAt(),
		CompletedAt:     o.CompletedAt(),
		Items:           items,
		CreatedAt:       o.CreatedAt(),
		UpdatedAt:       o.UpdatedAt(),
	}
}
