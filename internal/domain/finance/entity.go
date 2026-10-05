package finance

import (
	"time"

	"github.com/google/uuid"
)

// ─── Konstanta Tipe & Kategori ───────────────────────────────────────────────

type TransactionType string

const (
	TypeIncome    TransactionType = "INCOME"
	TypeExpense   TransactionType = "EXPENSE"
	TypeCapitalIn TransactionType = "CAPITAL_IN"
)

type TransactionCategory string

const (
	// INCOME categories
	CategorySalesShopee TransactionCategory = "SALES_SHOPEE"
	CategorySalesManual TransactionCategory = "SALES_MANUAL"
	CategoryOtherIncome TransactionCategory = "OTHER_INCOME"

	// EXPENSE categories
	CategoryFilament           TransactionCategory = "FILAMENT"
	CategoryHardware           TransactionCategory = "HARDWARE"
	CategoryPackaging          TransactionCategory = "PACKAGING"
	CategoryElectricity        TransactionCategory = "ELECTRICITY"
	CategoryInternet           TransactionCategory = "INTERNET"
	CategoryMachineMaintenance TransactionCategory = "MACHINE_MAINTENANCE"
	CategoryMachinePurchase    TransactionCategory = "MACHINE_PURCHASE"
	CategoryShipping           TransactionCategory = "SHIPPING_COST"
	CategoryMarketing          TransactionCategory = "MARKETING"
	CategoryRent               TransactionCategory = "RENT"
	CategoryOtherExpense       TransactionCategory = "OTHER_EXPENSE"

	// CAPITAL_IN categories
	CategoryInitialCapital    TransactionCategory = "INITIAL_CAPITAL"
	CategoryAdditionalCapital TransactionCategory = "ADDITIONAL_CAPITAL"
)

type PurchaseItemType string

const (
	PurchaseItemFilament  PurchaseItemType = "FILAMENT"
	PurchaseItemHardware  PurchaseItemType = "HARDWARE"
	PurchaseItemPackaging PurchaseItemType = "PACKAGING"
	PurchaseItemOther     PurchaseItemType = "OTHER"
)

type CapitalRecordType string

const (
	CapitalInitial  CapitalRecordType = "INITIAL"
	CapitalAddition CapitalRecordType = "ADDITION"
)

// ─── FinanceTransaction Entity ───────────────────────────────────────────────

type FinanceTransaction struct {
	id              uuid.UUID
	txType          TransactionType
	category        TransactionCategory
	amount          float64
	description     string
	transactionDate time.Time
	referenceType   string
	referenceID     string
	notes           string
	createdAt       time.Time
	updatedAt       time.Time
}

// Constructor: buat transaksi baru (untuk create)
func NewFinanceTransaction(
	txType TransactionType,
	category TransactionCategory,
	amount float64,
	description string,
	transactionDate time.Time,
	referenceType, referenceID, notes string,
) (*FinanceTransaction, error) {
	if amount <= 0 {
		return nil, ErrAmountMustBePositive
	}
	if description == "" {
		return nil, ErrDescriptionRequired
	}
	now := time.Now()
	return &FinanceTransaction{
		id:              uuid.New(),
		txType:          txType,
		category:        category,
		amount:          amount,
		description:     description,
		transactionDate: transactionDate,
		referenceType:   referenceType,
		referenceID:     referenceID,
		notes:           notes,
		createdAt:       now,
		updatedAt:       now,
	}, nil
}

// Constructor: rebuild dari database (untuk read)
func ReconstructFinanceTransaction(
	id uuid.UUID,
	txType TransactionType,
	category TransactionCategory,
	amount float64,
	description string,
	transactionDate time.Time,
	referenceType, referenceID, notes string,
	createdAt, updatedAt time.Time,
) *FinanceTransaction {
	return &FinanceTransaction{
		id:              id,
		txType:          txType,
		category:        category,
		amount:          amount,
		description:     description,
		transactionDate: transactionDate,
		referenceType:   referenceType,
		referenceID:     referenceID,
		notes:           notes,
		createdAt:       createdAt,
		updatedAt:       updatedAt,
	}
}

// Getters
func (t *FinanceTransaction) ID()              uuid.UUID           { return t.id }
func (t *FinanceTransaction) Type()            TransactionType     { return t.txType }
func (t *FinanceTransaction) Category()        TransactionCategory { return t.category }
func (t *FinanceTransaction) Amount()          float64             { return t.amount }
func (t *FinanceTransaction) Description()     string              { return t.description }
func (t *FinanceTransaction) TransactionDate() time.Time           { return t.transactionDate }
func (t *FinanceTransaction) ReferenceType()   string              { return t.referenceType }
func (t *FinanceTransaction) ReferenceID()     string              { return t.referenceID }
func (t *FinanceTransaction) Notes()           string              { return t.notes }
func (t *FinanceTransaction) CreatedAt()       time.Time           { return t.createdAt }
func (t *FinanceTransaction) UpdatedAt()       time.Time           { return t.updatedAt }

// ─── PurchaseOrderItem Entity ─────────────────────────────────────────────────

type PurchaseOrderItem struct {
	id              uuid.UUID
	purchaseOrderID uuid.UUID
	itemType        PurchaseItemType
	itemName        string
	quantity        float64
	unit            string
	unitPrice       float64
	totalPrice      float64
	createdAt       time.Time
}

func NewPurchaseOrderItem(
	itemType PurchaseItemType,
	itemName string,
	quantity float64,
	unit string,
	unitPrice float64,
) PurchaseOrderItem {
	return PurchaseOrderItem{
		id:         uuid.New(),
		itemType:   itemType,
		itemName:   itemName,
		quantity:   quantity,
		unit:       unit,
		unitPrice:  unitPrice,
		totalPrice: quantity * unitPrice,
		createdAt:  time.Now(),
	}
}

func ReconstructPurchaseOrderItem(
	id, purchaseOrderID uuid.UUID,
	itemType PurchaseItemType,
	itemName string,
	quantity float64,
	unit string,
	unitPrice, totalPrice float64,
	createdAt time.Time,
) PurchaseOrderItem {
	return PurchaseOrderItem{
		id:              id,
		purchaseOrderID: purchaseOrderID,
		itemType:        itemType,
		itemName:        itemName,
		quantity:        quantity,
		unit:            unit,
		unitPrice:       unitPrice,
		totalPrice:      totalPrice,
		createdAt:       createdAt,
	}
}

func (i *PurchaseOrderItem) ID()              uuid.UUID        { return i.id }
func (i *PurchaseOrderItem) PurchaseOrderID() uuid.UUID        { return i.purchaseOrderID }
func (i *PurchaseOrderItem) ItemType()        PurchaseItemType { return i.itemType }
func (i *PurchaseOrderItem) ItemName()        string           { return i.itemName }
func (i *PurchaseOrderItem) Quantity()        float64          { return i.quantity }
func (i *PurchaseOrderItem) Unit()            string           { return i.unit }
func (i *PurchaseOrderItem) UnitPrice()       float64          { return i.unitPrice }
func (i *PurchaseOrderItem) TotalPrice()      float64          { return i.totalPrice }
func (i *PurchaseOrderItem) CreatedAt()       time.Time        { return i.createdAt }

// ─── PurchaseOrder Entity ─────────────────────────────────────────────────────

type PurchaseOrder struct {
	id                  uuid.UUID
	supplierName        string
	purchaseDate        time.Time
	totalAmount         float64
	status              string
	notes               string
	linkedTransactionID *uuid.UUID
	items               []PurchaseOrderItem
	createdAt           time.Time
	updatedAt           time.Time
}

func NewPurchaseOrder(
	supplierName string,
	purchaseDate time.Time,
	notes string,
	items []PurchaseOrderItem,
) (*PurchaseOrder, error) {
	if supplierName == "" {
		return nil, ErrSupplierNameRequired
	}
	if len(items) == 0 {
		return nil, ErrPurchaseItemsRequired
	}
	var total float64
	for _, item := range items {
		total += item.TotalPrice()
	}
	now := time.Now()
	return &PurchaseOrder{
		id:           uuid.New(),
		supplierName: supplierName,
		purchaseDate: purchaseDate,
		totalAmount:  total,
		status:       "RECEIVED",
		notes:        notes,
		items:        items,
		createdAt:    now,
		updatedAt:    now,
	}, nil
}

func ReconstructPurchaseOrder(
	id uuid.UUID,
	supplierName string,
	purchaseDate time.Time,
	totalAmount float64,
	status, notes string,
	linkedTransactionID *uuid.UUID,
	items []PurchaseOrderItem,
	createdAt, updatedAt time.Time,
) *PurchaseOrder {
	return &PurchaseOrder{
		id:                  id,
		supplierName:        supplierName,
		purchaseDate:        purchaseDate,
		totalAmount:         totalAmount,
		status:              status,
		notes:               notes,
		linkedTransactionID: linkedTransactionID,
		items:               items,
		createdAt:           createdAt,
		updatedAt:           updatedAt,
	}
}

func (p *PurchaseOrder) ID()                  uuid.UUID           { return p.id }
func (p *PurchaseOrder) SupplierName()        string              { return p.supplierName }
func (p *PurchaseOrder) PurchaseDate()        time.Time           { return p.purchaseDate }
func (p *PurchaseOrder) TotalAmount()         float64             { return p.totalAmount }
func (p *PurchaseOrder) Status()              string              { return p.status }
func (p *PurchaseOrder) Notes()               string              { return p.notes }
func (p *PurchaseOrder) LinkedTransactionID() *uuid.UUID          { return p.linkedTransactionID }
func (p *PurchaseOrder) Items()               []PurchaseOrderItem { return p.items }
func (p *PurchaseOrder) CreatedAt()           time.Time           { return p.createdAt }
func (p *PurchaseOrder) UpdatedAt()           time.Time           { return p.updatedAt }

func (p *PurchaseOrder) SetLinkedTransaction(txID uuid.UUID) {
	p.linkedTransactionID = &txID
	p.updatedAt = time.Now()
}

// ─── CapitalRecord Entity ─────────────────────────────────────────────────────

type CapitalRecord struct {
	id                  uuid.UUID
	recordType          CapitalRecordType
	amount              float64
	description         string
	recordDate          time.Time
	linkedTransactionID *uuid.UUID
	notes               string
	createdAt           time.Time
	updatedAt           time.Time
}

func NewCapitalRecord(
	recordType CapitalRecordType,
	amount float64,
	description string,
	recordDate time.Time,
	notes string,
) (*CapitalRecord, error) {
	if amount <= 0 {
		return nil, ErrAmountMustBePositive
	}
	if description == "" {
		return nil, ErrDescriptionRequired
	}
	now := time.Now()
	return &CapitalRecord{
		id:          uuid.New(),
		recordType:  recordType,
		amount:      amount,
		description: description,
		recordDate:  recordDate,
		notes:       notes,
		createdAt:   now,
		updatedAt:   now,
	}, nil
}

func ReconstructCapitalRecord(
	id uuid.UUID,
	recordType CapitalRecordType,
	amount float64,
	description string,
	recordDate time.Time,
	linkedTransactionID *uuid.UUID,
	notes string,
	createdAt, updatedAt time.Time,
) *CapitalRecord {
	return &CapitalRecord{
		id:                  id,
		recordType:          recordType,
		amount:              amount,
		description:         description,
		recordDate:          recordDate,
		linkedTransactionID: linkedTransactionID,
		notes:               notes,
		createdAt:           createdAt,
		updatedAt:           updatedAt,
	}
}

func (c *CapitalRecord) ID()                  uuid.UUID         { return c.id }
func (c *CapitalRecord) RecordType()          CapitalRecordType { return c.recordType }
func (c *CapitalRecord) Amount()              float64           { return c.amount }
func (c *CapitalRecord) Description()         string            { return c.description }
func (c *CapitalRecord) RecordDate()          time.Time         { return c.recordDate }
func (c *CapitalRecord) LinkedTransactionID() *uuid.UUID        { return c.linkedTransactionID }
func (c *CapitalRecord) Notes()               string            { return c.notes }
func (c *CapitalRecord) CreatedAt()           time.Time         { return c.createdAt }
func (c *CapitalRecord) UpdatedAt()           time.Time         { return c.updatedAt }

func (c *CapitalRecord) SetLinkedTransaction(txID uuid.UUID) {
	c.linkedTransactionID = &txID
	c.updatedAt = time.Now()
}

// ─── FinanceSummary (untuk dashboard) ────────────────────────────────────────

type FinanceSummary struct {
	TotalIncome    float64
	TotalExpense   float64
	TotalCapitalIn float64
	NetCashFlow    float64 // TotalIncome - TotalExpense

	ByCategory map[TransactionCategory]float64
}

// ─── OrderAllocationSummary (7-Bucket Fund Allocation) ───────────────────────

type OrderAllocationSummary struct {
	TotalOrders         int64
	TotalGrossSales     float64
	TotalChannelFees    float64
	TotalNetRevenue     float64
	TotalCOGS           float64
	TotalNetProfit      float64

	// Net Available Fund Balances (Saldo Tersedia = Allocated - Spent)
	FundFilament        float64
	FundComponent       float64
	FundPackaging       float64
	FundElectricity     float64
	FundMaintenance     float64
	FundDepreciation    float64
	FundNetProfit       float64

	// Allocated Inflow from Orders
	AllocatedFilament    float64
	AllocatedComponent   float64
	AllocatedPackaging   float64
	AllocatedElectricity float64
	AllocatedMaintenance float64
	AllocatedDepreciation float64
	AllocatedNetProfit   float64

	// Actual Spent Outflow (PO & Expenses)
	SpentFilament        float64
	SpentComponent       float64
	SpentPackaging       float64
	SpentElectricity     float64
	SpentMaintenance     float64
	SpentDepreciation    float64
	SpentNetProfit       float64

	AverageProfitMargin float64
	UnmappedItemsCount  int64
}
