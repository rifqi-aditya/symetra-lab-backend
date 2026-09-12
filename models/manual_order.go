package models

import (
	"time"
)

// Order merepresentasikan pesanan manual / offline internal bengkel di tabel orders
type Order struct {
	ID              string       `gorm:"type:uuid;primaryKey" json:"id"`
	UserID          string       `gorm:"type:uuid;not null;index" json:"user_id"`
	OrderNumber     string       `gorm:"size:64;index" json:"order_number"`
	CustomerName    string       `gorm:"size:255" json:"customer_name"`
	CustomerContact string       `gorm:"size:255" json:"customer_contact"`
	TotalRevenue    float64      `gorm:"type:numeric;default:0" json:"total_revenue"`
	TotalHPP        float64      `gorm:"type:numeric;default:0" json:"total_hpp"`
	TotalProfit     float64      `gorm:"type:numeric;default:0" json:"total_profit"`
	Status          string       `gorm:"size:32;default:'PENDING'" json:"status"` // PENDING, IN_PRODUCTION, COMPLETED, DELIVERED, CANCELLED
	Notes           string       `gorm:"type:text" json:"notes"`
	Source          string       `gorm:"size:64;default:'MANUAL'" json:"source"` // MANUAL, DIRECT_WHATSAPP, OFFLINE
	PaymentStatus   string       `gorm:"size:32;default:'UNPAID'" json:"payment_status"` // PAID, UNPAID
	StartedAt       *time.Time   `json:"started_at,omitempty"`
	CompletedAt     *time.Time   `json:"completed_at,omitempty"`
	CreatedAt       time.Time    `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time    `gorm:"autoUpdateTime" json:"updated_at"`

	Items []OrderItem `gorm:"foreignKey:OrderID;references:ID;constraint:OnDelete:CASCADE" json:"items,omitempty"`
}

func (Order) TableName() string {
	return "orders"
}

// OrderItem merepresentasikan item fisik dalam pesanan manual di tabel order_items
type OrderItem struct {
	ID               string           `gorm:"type:uuid;primaryKey" json:"id"`
	OrderID          string           `gorm:"type:uuid;not null;index" json:"order_id"`
	ProductID        *string          `gorm:"type:uuid;index" json:"product_id,omitempty"`
	ProductName      string           `gorm:"size:255;not null" json:"product_name"`
	Quantity         int              `gorm:"default:1" json:"quantity"`
	SellingPrice     float64          `gorm:"type:numeric;default:0" json:"selling_price"`
	HPP              float64          `gorm:"type:numeric;default:0" json:"hpp"`
	WeightGrams      float64          `gorm:"type:numeric;default:0" json:"weight_grams"`
	PrintTimeHours   float64          `gorm:"type:numeric;default:0" json:"print_time_hours"`
	MachineID        *string          `gorm:"type:uuid;index" json:"machine_id,omitempty"`
	EnergyCost       float64          `gorm:"type:numeric;default:0" json:"energy_cost"`
	DepreciationCost float64          `gorm:"type:numeric;default:0" json:"depreciation_cost"`
	MaintenanceCost  float64          `gorm:"type:numeric;default:0" json:"maintenance_cost"`
	PackingFee       float64          `gorm:"type:numeric;default:0" json:"packing_fee"`
	CreatedAt        time.Time        `gorm:"autoCreateTime" json:"created_at"`

	Product    *Product         `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	Machine    *Machine         `gorm:"foreignKey:MachineID" json:"machine,omitempty"`
	Filaments  []OrderFilament  `gorm:"foreignKey:OrderItemID;references:ID;constraint:OnDelete:CASCADE" json:"filaments,omitempty"`
	Components []OrderComponent `gorm:"foreignKey:OrderItemID;references:ID;constraint:OnDelete:CASCADE" json:"components,omitempty"`
}

func (OrderItem) TableName() string {
	return "order_items"
}

// OrderFilament merepresentasikan snapshot konsumsi filamen per order item di order_filaments
type OrderFilament struct {
	ID                   string    `gorm:"type:uuid;primaryKey" json:"id"`
	OrderItemID          string    `gorm:"type:uuid;not null;index" json:"order_item_id"`
	FilamentID           *string   `gorm:"type:uuid;index" json:"filament_id,omitempty"`
	FilamentSnapshotName string    `gorm:"size:255" json:"filament_snapshot_name"`
	WeightUsedGrams      float64   `gorm:"type:numeric;default:0" json:"weight_used_grams"`
	CostContribution     float64   `gorm:"type:numeric;default:0" json:"cost_contribution"`
	CreatedAt            time.Time `gorm:"autoCreateTime" json:"created_at"`

	Filament *Filament `gorm:"foreignKey:FilamentID" json:"filament,omitempty"`
}

func (OrderFilament) TableName() string {
	return "order_filaments"
}

// OrderComponent merepresentasikan snapshot konsumsi komponen per order item di order_components
type OrderComponent struct {
	ID                    string    `gorm:"type:uuid;primaryKey" json:"id"`
	OrderItemID           string    `gorm:"type:uuid;not null;index" json:"order_item_id"`
	ComponentID           *string   `gorm:"type:uuid;index" json:"component_id,omitempty"`
	ComponentSnapshotName string    `gorm:"size:255" json:"component_snapshot_name"`
	Quantity              float64   `gorm:"type:numeric;default:1" json:"quantity"`
	CostContribution      float64   `gorm:"type:numeric;default:0" json:"cost_contribution"`
	CreatedAt             time.Time `gorm:"autoCreateTime" json:"created_at"`

	Component *Component `gorm:"foreignKey:ComponentID" json:"component,omitempty"`
}

func (OrderComponent) TableName() string {
	return "order_components"
}

// DTOs untuk Pembuatan dan Pengelolaan Pesanan Manual

type CreateManualOrderItemInput struct {
	ProductID      *string  `json:"product_id"`                  // Opsional jika custom order non-katalog
	ProductName    string   `json:"product_name" binding:"required"`
	Quantity       int      `json:"quantity" binding:"required,min=1"`
	SellingPrice   float64  `json:"selling_price" binding:"required,min=0"`
	CustomWeight   *float64 `json:"custom_weight_grams,omitempty"`    // Override berat jika diperlukan
	CustomPrintTime *float64 `json:"custom_print_time_hours,omitempty"` // Override jam cetak jika diperlukan
	MachineID      *string  `json:"machine_id,omitempty"`
}

type CreateManualOrderRequest struct {
	CustomerName    string                       `json:"customer_name" binding:"required"`
	CustomerContact string                       `json:"customer_contact"`
	Notes           string                       `json:"notes"`
	Source          string                       `json:"source"`         // MANUAL, DIRECT_WHATSAPP, OFFLINE
	PaymentStatus   string                       `json:"payment_status"`  // PAID, UNPAID
	Items           []CreateManualOrderItemInput `json:"items" binding:"required,min=1"`
}

type UpdateOrderStatusRequest struct {
	Status string `json:"status" binding:"required"` // PENDING, IN_PRODUCTION, COMPLETED, DELIVERED, CANCELLED
}

type UpdatePaymentStatusRequest struct {
	PaymentStatus string `json:"payment_status" binding:"required"` // PAID, UNPAID
}

// DTOs untuk Antrean Cetak Terpadu (Production Queue)

type ProductionQueueItem struct {
	JobID             string     `json:"job_id"`
	Source            string     `json:"source"`             // "SHOPEE" atau "MANUAL"
	OrderIdentifier   string     `json:"order_identifier"`   // OrderSN (Shopee) atau OrderNumber (Manual)
	CustomerName      string     `json:"customer_name"`
	ItemName          string     `json:"item_name"`
	VariantSKU        string     `json:"variant_sku,omitempty"`
	Quantity          int        `json:"quantity"`
	WeightGrams       float64    `json:"weight_grams"`
	PrintTimeHours    float64    `json:"print_time_hours"`
	TotalPrintHours   float64    `json:"total_print_hours"`  // PrintTimeHours * Quantity
	AssignedMachineID *string    `json:"assigned_machine_id,omitempty"`
	AssignedMachine   string     `json:"assigned_machine_name,omitempty"`
	Status            string     `json:"status"`
	Deadline          *time.Time `json:"deadline,omitempty"` // SLA Shopee ship_by_date atau created_at + 2 hari
	IsUrgent          bool       `json:"is_urgent"`
	CreatedAt         time.Time  `json:"created_at"`
}

type CompletePrintJobRequest struct {
	Source    string `json:"source" binding:"required"`    // "SHOPEE" atau "MANUAL"
	JobID     string `json:"job_id" binding:"required"`    // item ID
	MachineID string `json:"machine_id"`                   // ID printer yang digunakan (opsional jika sudah terisi)
}

type CompletePrintJobResponse struct {
	Status             string   `json:"status"`
	Message            string   `json:"message"`
	DeductedFilaments  []string `json:"deducted_filaments"`
	AddedMachineHours  float64  `json:"added_machine_hours"`
	LowStockWarning    bool     `json:"low_stock_warning"`
	MaintenanceWarning bool     `json:"maintenance_warning"`
}
