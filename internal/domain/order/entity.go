package order

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
)

type Order struct {
	id               uuid.UUID
	userID           uuid.UUID
	orderNumber      string
	customerName     string
	customerContact  string
	channel          string // MANUAL, SHOPEE, TOKOPEDIA
	grossAmount      float64
	channelFee       float64
	netAmount        float64
	cogsAmount       float64
	netProfit        float64
	fundFilament     float64
	fundComponent    float64
	fundPackaging    float64
	fundElectricity  float64
	fundMaintenance  float64
	fundDepreciation float64
	fundNetProfit    float64
	totalRevenue     float64
	totalHPP         float64
	totalProfit      float64
	status           string // PENDING, IN_PRODUCTION, COMPLETED, DELIVERED, CANCELLED
	notes            string
	source           string // MANUAL, DIRECT_WHATSAPP, OFFLINE, SHOPEE
	paymentStatus    string // PAID, UNPAID
	startedAt        *time.Time
	completedAt      *time.Time
	shippingCarrier  string
	trackingNumber   string
	shipByDateTime   *time.Time
	financialStatus  string
	items            []OrderItem
	createdAt        time.Time
	updatedAt        time.Time
}

type OrderItem struct {
	id               uuid.UUID
	orderID          uuid.UUID
	productID        *uuid.UUID
	productName      string
	thumbnailURL     *string
	itemSKU          string
	quantity         int
	sellingPrice     float64
	hpp              float64
	weightGrams      float64
	printTimeHours   float64
	machineID        *uuid.UUID
	filamentCost     float64
	componentCost    float64
	packagingCost    float64
	electricityCost  float64
	maintenanceCost  float64
	depreciationCost float64
	totalCogs        float64
	netProfit        float64
	energyCost       float64
	packingFee       float64
	channelItemID    int64
	channelModelID   int64
	mappingStatus    string
	matchedSKU       string
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
	var totalFilament, totalComponent, totalPackaging, totalElectricity, totalMaintenance, totalDepreciation float64
	assignedItems := make([]OrderItem, len(items))
	for i, item := range items {
		itemID := item.id
		if itemID == uuid.Nil {
			itemID = uuid.New()
		}

		fil := item.filamentCost
		comp := item.componentCost
		pack := item.packagingCost
		if pack == 0 && item.packingFee > 0 {
			pack = item.packingFee
		}
		elec := item.electricityCost
		if elec == 0 && item.energyCost > 0 {
			elec = item.energyCost
		}
		maint := item.maintenanceCost
		depr := item.depreciationCost

		if fil == 0 && item.hpp > 0 {
			fil = item.hpp - (pack + elec + maint + depr + comp)
			if fil < 0 {
				fil = 0
			}
		}

		assignedItems[i] = OrderItem{
			id:               itemID,
			orderID:          orderID,
			productID:        item.productID,
			productName:      item.productName,
			itemSKU:          item.itemSKU,
			quantity:         item.quantity,
			sellingPrice:     item.sellingPrice,
			hpp:              item.hpp,
			weightGrams:      item.weightGrams,
			printTimeHours:   item.printTimeHours,
			machineID:        item.machineID,
			filamentCost:     fil,
			componentCost:    comp,
			packagingCost:    pack,
			electricityCost:  elec,
			energyCost:       elec,
			depreciationCost: depr,
			maintenanceCost:  maint,
			packingFee:       pack,
			totalCogs:        item.hpp,
			netProfit:        item.sellingPrice - item.hpp,
			mappingStatus:    "MATCHED",
			createdAt:        time.Now(),
		}
		q := float64(item.quantity)
		totalRev += item.sellingPrice * q
		totalHPP += item.hpp * q
		totalFilament += fil * q
		totalComponent += comp * q
		totalPackaging += pack * q
		totalElectricity += elec * q
		totalMaintenance += maint * q
		totalDepreciation += depr * q
	}

	netProfit := totalRev - totalHPP
	now := time.Now()
	return &Order{
		id:               orderID,
		userID:           userID,
		orderNumber:      orderNumber,
		customerName:     customerName,
		customerContact:  customerContact,
		channel:          source,
		grossAmount:      totalRev,
		channelFee:       0,
		netAmount:        totalRev,
		cogsAmount:       totalHPP,
		netProfit:        netProfit,
		fundFilament:     totalFilament,
		fundComponent:    totalComponent,
		fundPackaging:    totalPackaging,
		fundElectricity:  totalElectricity,
		fundMaintenance:  totalMaintenance,
		fundDepreciation: totalDepreciation,
		fundNetProfit:    netProfit,
		totalRevenue:     totalRev,
		totalHPP:         totalHPP,
		totalProfit:      netProfit,
		status:           "PENDING",
		notes:            notes,
		source:           source,
		paymentStatus:    paymentStatus,
		items:            assignedItems,
		createdAt:        now,
		updatedAt:        now,
	}, nil
}

func ReconstructOrder(
	id, userID uuid.UUID,
	orderNumber, customerName, customerContact, channel string,
	grossAmount, channelFee, netAmount, cogsAmount, netProfit float64,
	fundFilament, fundComponent, fundPackaging, fundElectricity, fundMaintenance, fundDepreciation, fundNetProfit float64,
	totalRevenue, totalHPP, totalProfit float64,
	status, notes, source, paymentStatus string,
	startedAt, completedAt *time.Time,
	items []OrderItem,
	createdAt, updatedAt time.Time,
) *Order {
	return &Order{
		id:               id,
		userID:           userID,
		orderNumber:      orderNumber,
		customerName:     customerName,
		customerContact:  customerContact,
		channel:          channel,
		grossAmount:      grossAmount,
		channelFee:       channelFee,
		netAmount:        netAmount,
		cogsAmount:       cogsAmount,
		netProfit:        netProfit,
		fundFilament:     fundFilament,
		fundComponent:    fundComponent,
		fundPackaging:    fundPackaging,
		fundElectricity:  fundElectricity,
		fundMaintenance:  fundMaintenance,
		fundDepreciation: fundDepreciation,
		fundNetProfit:    fundNetProfit,
		totalRevenue:     totalRevenue,
		totalHPP:         totalHPP,
		totalProfit:      totalProfit,
		status:           status,
		notes:            notes,
		source:           source,
		paymentStatus:    paymentStatus,
		startedAt:        startedAt,
		completedAt:      completedAt,
		items:            items,
		createdAt:        createdAt,
		updatedAt:        updatedAt,
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
		electricityCost:  energyCost,
		depreciationCost: depreciationCost,
		maintenanceCost:  maintenanceCost,
		packingFee:       packingFee,
		packagingCost:    packingFee,
		totalCogs:        hpp,
		netProfit:        sellingPrice - hpp,
		mappingStatus:    "MATCHED",
		createdAt:        createdAt,
	}
}

func ReconstructOrderItemFull(
	id, orderID uuid.UUID,
	productID *uuid.UUID,
	productName, itemSKU string,
	quantity int,
	sellingPrice, hpp, weightGrams, printTimeHours float64,
	machineID *uuid.UUID,
	filamentCost, componentCost, packagingCost, electricityCost, maintenanceCost, depreciationCost, totalCogs, netProfit float64,
	channelItemID, channelModelID int64,
	mappingStatus, matchedSKU string,
	createdAt time.Time,
) OrderItem {
	return OrderItem{
		id:               id,
		orderID:          orderID,
		productID:        productID,
		productName:      productName,
		itemSKU:          itemSKU,
		quantity:         quantity,
		sellingPrice:     sellingPrice,
		hpp:              hpp,
		weightGrams:      weightGrams,
		printTimeHours:   printTimeHours,
		machineID:        machineID,
		filamentCost:     filamentCost,
		componentCost:    componentCost,
		packagingCost:    packagingCost,
		electricityCost:  electricityCost,
		maintenanceCost:  maintenanceCost,
		depreciationCost: depreciationCost,
		totalCogs:        totalCogs,
		netProfit:        netProfit,
		energyCost:       electricityCost,
		packingFee:       packagingCost,
		channelItemID:    channelItemID,
		channelModelID:   channelModelID,
		mappingStatus:    mappingStatus,
		matchedSKU:       matchedSKU,
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
func (o *Order) Channel() string            { return o.channel }
func (o *Order) GrossAmount() float64       { return o.grossAmount }
func (o *Order) ChannelFee() float64        { return o.channelFee }
func (o *Order) NetAmount() float64         { return o.netAmount }
func (o *Order) CogsAmount() float64        { return o.cogsAmount }
func (o *Order) NetProfit() float64         { return o.netProfit }
func (o *Order) FundFilament() float64      { return o.fundFilament }
func (o *Order) FundComponent() float64     { return o.fundComponent }
func (o *Order) FundPackaging() float64     { return o.fundPackaging }
func (o *Order) FundElectricity() float64   { return o.fundElectricity }
func (o *Order) FundMaintenance() float64   { return o.fundMaintenance }
func (o *Order) FundDepreciation() float64  { return o.fundDepreciation }
func (o *Order) FundNetProfit() float64     { return o.fundNetProfit }
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

func (o *Order) ShippingCarrier() string    { return o.shippingCarrier }
func (o *Order) TrackingNumber() string     { return o.trackingNumber }
func (o *Order) ShipByDateTime() *time.Time { return o.shipByDateTime }
func (o *Order) FinancialStatus() string    { return o.financialStatus }
func (o *Order) SetLogistics(carrier, tracking string, shipBy *time.Time) {
	o.shippingCarrier = carrier
	o.trackingNumber = tracking
	o.shipByDateTime = shipBy
}
func (o *Order) SetFinancialStatus(status string) {
	o.financialStatus = status
}

func (item *OrderItem) ID() uuid.UUID               { return item.id }
func (item *OrderItem) OrderID() uuid.UUID          { return item.orderID }
func (item *OrderItem) ProductID() *uuid.UUID       { return item.productID }
func (item *OrderItem) ProductName() string         { return item.productName }
func (item *OrderItem) ThumbnailURL() *string       { return item.thumbnailURL }
func (item *OrderItem) SetThumbnailURL(url *string) { item.thumbnailURL = url }
func (item *OrderItem) SetProductName(name string)  { item.productName = name }
func (item *OrderItem) HardwareCost() float64       { return item.componentCost }
func (item *OrderItem) MachineCost() float64 {
	return item.electricityCost + item.maintenanceCost + item.depreciationCost
}
func (item *OrderItem) ItemSKU() string             { return item.itemSKU }
func (item *OrderItem) Quantity() int               { return item.quantity }
func (item *OrderItem) SellingPrice() float64       { return item.sellingPrice }
func (item *OrderItem) HPP() float64                { return item.hpp }
func (item *OrderItem) WeightGrams() float64        { return item.weightGrams }
func (item *OrderItem) PrintTimeHours() float64     { return item.printTimeHours }
func (item *OrderItem) MachineID() *uuid.UUID       { return item.machineID }
func (item *OrderItem) FilamentCost() float64       { return item.filamentCost }
func (item *OrderItem) ComponentCost() float64      { return item.componentCost }
func (item *OrderItem) PackagingCost() float64      { return item.packagingCost }
func (item *OrderItem) ElectricityCost() float64    { return item.electricityCost }
func (item *OrderItem) MaintenanceCost() float64    { return item.maintenanceCost }
func (item *OrderItem) DepreciationCost() float64   { return item.depreciationCost }
func (item *OrderItem) TotalCogs() float64          { return item.totalCogs }
func (item *OrderItem) NetProfit() float64          { return item.netProfit }
func (item *OrderItem) EnergyCost() float64         { return item.energyCost }
func (item *OrderItem) PackingFee() float64         { return item.packingFee }
func (item *OrderItem) ChannelItemID() int64        { return item.channelItemID }
func (item *OrderItem) ChannelModelID() int64       { return item.channelModelID }
func (item *OrderItem) MappingStatus() string       { return item.mappingStatus }
func (item *OrderItem) MatchedSKU() string          { return item.matchedSKU }
func (item *OrderItem) CreatedAt() time.Time        { return item.createdAt }
