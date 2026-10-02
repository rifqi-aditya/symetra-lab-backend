package order

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
)

type Order struct {
	id              uuid.UUID
	userID          uuid.UUID
	orderNumber     string
	customerName    string
	customerContact string
	totalRevenue    float64
	totalHPP        float64
	totalProfit     float64
	status          string // PENDING, IN_PRODUCTION, COMPLETED, DELIVERED, CANCELLED
	notes           string
	source          string // MANUAL, DIRECT_WHATSAPP, OFFLINE
	paymentStatus   string // PAID, UNPAID
	startedAt       *time.Time
	completedAt     *time.Time
	items           []OrderItem
	createdAt       time.Time
	updatedAt       time.Time
}

type OrderItem struct {
	id               uuid.UUID
	orderID          uuid.UUID
	productID        *uuid.UUID
	productName      string
	quantity         int
	sellingPrice     float64
	hpp              float64
	weightGrams      float64
	printTimeHours   float64
	machineID        *uuid.UUID
	energyCost       float64
	depreciationCost float64
	maintenanceCost  float64
	packingFee       float64
	createdAt        time.Time
}

func NewOrder(
	userID uuid.UUID,
	customerName, customerContact, notes, source, paymentStatus string,
	items []OrderItem,
) (*Order, error) {
	if userID == uuid.Nil {
		return nil, ErrUserIDRequired
	}
	if customerName == "" {
		return nil, ErrCustomerNameRequired
	}
	if len(items) == 0 {
		return nil, ErrOrderItemsRequired
	}

	if source == "" {
		source = "MANUAL"
	}
	if paymentStatus == "" {
		paymentStatus = "UNPAID"
	}

	orderID := uuid.New()
	orderNumber := fmt.Sprintf("ORD-%s-%04d", time.Now().Format("20060102"), rand.Intn(10000))

	var totalRev, totalHPP float64
	assignedItems := make([]OrderItem, len(items))
	for i, item := range items {
		itemID := item.id
		if itemID == uuid.Nil {
			itemID = uuid.New()
		}
		assignedItems[i] = OrderItem{
			id:               itemID,
			orderID:          orderID,
			productID:        item.productID,
			productName:      item.productName,
			quantity:         item.quantity,
			sellingPrice:     item.sellingPrice,
			hpp:              item.hpp,
			weightGrams:      item.weightGrams,
			printTimeHours:   item.printTimeHours,
			machineID:        item.machineID,
			energyCost:       item.energyCost,
			depreciationCost: item.depreciationCost,
			maintenanceCost:  item.maintenanceCost,
			packingFee:       item.packingFee,
			createdAt:        time.Now(),
		}
		totalRev += item.sellingPrice * float64(item.quantity)
		totalHPP += item.hpp * float64(item.quantity)
	}

	now := time.Now()
	return &Order{
		id:              orderID,
		userID:          userID,
		orderNumber:     orderNumber,
		customerName:    customerName,
		customerContact: customerContact,
		totalRevenue:    totalRev,
		totalHPP:        totalHPP,
		totalProfit:     totalRev - totalHPP,
		status:          "PENDING",
		notes:           notes,
		source:          source,
		paymentStatus:   paymentStatus,
		items:           assignedItems,
		createdAt:       now,
		updatedAt:       now,
	}, nil
}

func ReconstructOrder(
	id, userID uuid.UUID,
	orderNumber, customerName, customerContact string,
	totalRevenue, totalHPP, totalProfit float64,
	status, notes, source, paymentStatus string,
	startedAt, completedAt *time.Time,
	items []OrderItem,
	createdAt, updatedAt time.Time,
) *Order {
	return &Order{
		id:              id,
		userID:          userID,
		orderNumber:     orderNumber,
		customerName:    customerName,
		customerContact: customerContact,
		totalRevenue:    totalRevenue,
		totalHPP:        totalHPP,
		totalProfit:     totalProfit,
		status:          status,
		notes:           notes,
		source:          source,
		paymentStatus:   paymentStatus,
		startedAt:       startedAt,
		completedAt:     completedAt,
		items:           items,
		createdAt:       createdAt,
		updatedAt:       updatedAt,
	}
}

func ReconstructOrderItem(
	id, orderID uuid.UUID,
	productID *uuid.UUID,
	productName string,
	quantity int,
	sellingPrice, hpp, weightGrams, printTimeHours float64,
	machineID *uuid.UUID,
	energyCost, depreciationCost, maintenanceCost, packingFee float64,
	createdAt time.Time,
) OrderItem {
	return OrderItem{
		id:               id,
		orderID:          orderID,
		productID:        productID,
		productName:      productName,
		quantity:         quantity,
		sellingPrice:     sellingPrice,
		hpp:              hpp,
		weightGrams:      weightGrams,
		printTimeHours:   printTimeHours,
		machineID:        machineID,
		energyCost:       energyCost,
		depreciationCost: depreciationCost,
		maintenanceCost:  maintenanceCost,
		packingFee:       packingFee,
		createdAt:        createdAt,
	}
}

// Business methods
func (o *Order) UpdateStatus(newStatus string) {
	o.status = newStatus
	now := time.Now()
	o.updatedAt = now
	if newStatus == "IN_PRODUCTION" && o.startedAt == nil {
		o.startedAt = &now
	}
	if newStatus == "COMPLETED" && o.completedAt == nil {
		o.completedAt = &now
	}
}

func (o *Order) UpdatePaymentStatus(newStatus string) {
	o.paymentStatus = newStatus
	o.updatedAt = time.Now()
}

// Getters
func (o *Order) ID() uuid.UUID              { return o.id }
func (o *Order) UserID() uuid.UUID          { return o.userID }
func (o *Order) OrderNumber() string        { return o.orderNumber }
func (o *Order) CustomerName() string       { return o.customerName }
func (o *Order) CustomerContact() string    { return o.customerContact }
func (o *Order) TotalRevenue() float64      { return o.totalRevenue }
func (o *Order) TotalHPP() float64          { return o.totalHPP }
func (o *Order) TotalProfit() float64       { return o.totalProfit }
func (o *Order) Status() string             { return o.status }
func (o *Order) Notes() string              { return o.notes }
func (o *Order) Source() string             { return o.source }
func (o *Order) PaymentStatus() string      { return o.paymentStatus }
func (o *Order) StartedAt() *time.Time      { return o.startedAt }
func (o *Order) CompletedAt() *time.Time    { return o.completedAt }
func (o *Order) Items() []OrderItem         { return o.items }
func (o *Order) CreatedAt() time.Time       { return o.createdAt }
func (o *Order) UpdatedAt() time.Time       { return o.updatedAt }

func (item *OrderItem) ID() uuid.UUID               { return item.id }
func (item *OrderItem) OrderID() uuid.UUID          { return item.orderID }
func (item *OrderItem) ProductID() *uuid.UUID       { return item.productID }
func (item *OrderItem) ProductName() string         { return item.productName }
func (item *OrderItem) Quantity() int               { return item.quantity }
func (item *OrderItem) SellingPrice() float64       { return item.sellingPrice }
func (item *OrderItem) HPP() float64                { return item.hpp }
func (item *OrderItem) WeightGrams() float64        { return item.weightGrams }
func (item *OrderItem) PrintTimeHours() float64     { return item.printTimeHours }
func (item *OrderItem) MachineID() *uuid.UUID       { return item.machineID }
func (item *OrderItem) EnergyCost() float64         { return item.energyCost }
func (item *OrderItem) DepreciationCost() float64   { return item.depreciationCost }
func (item *OrderItem) MaintenanceCost() float64    { return item.maintenanceCost }
func (item *OrderItem) PackingFee() float64         { return item.packingFee }
func (item *OrderItem) CreatedAt() time.Time        { return item.createdAt }
