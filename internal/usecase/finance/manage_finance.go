package finance

import (
	"context"
	"time"

	"github.com/google/uuid"

	"symetra-lab-backend-v2/internal/domain/finance"
	"symetra-lab-backend-v2/internal/domain/order"
)

// ─── Finance Transaction Use Cases ───────────────────────────────────────────

type CreateTransactionInput struct {
	Type            finance.TransactionType
	Category        finance.TransactionCategory
	Amount          float64
	Description     string
	TransactionDate time.Time
	ReferenceType   string
	ReferenceID     string
	Notes           string
}

type ListTransactionsUseCase struct {
	repo finance.Repository
}

func NewListTransactionsUseCase(repo finance.Repository) *ListTransactionsUseCase {
	return &ListTransactionsUseCase{repo: repo}
}

func (uc *ListTransactionsUseCase) Execute(ctx context.Context, filter finance.TransactionFilter) ([]*finance.FinanceTransaction, error) {
	return uc.repo.FindTransactions(ctx, filter)
}

type CreateTransactionUseCase struct {
	repo finance.Repository
}

func NewCreateTransactionUseCase(repo finance.Repository) *CreateTransactionUseCase {
	return &CreateTransactionUseCase{repo: repo}
}

func (uc *CreateTransactionUseCase) Execute(ctx context.Context, input CreateTransactionInput) (*finance.FinanceTransaction, error) {
	tx, err := finance.NewFinanceTransaction(
		input.Type,
		input.Category,
		input.Amount,
		input.Description,
		input.TransactionDate,
		input.ReferenceType,
		input.ReferenceID,
		input.Notes,
	)
	if err != nil {
		return nil, err
	}
	if err := uc.repo.CreateTransaction(ctx, tx); err != nil {
		return nil, err
	}
	return tx, nil
}

type DeleteTransactionUseCase struct {
	repo finance.Repository
}

func NewDeleteTransactionUseCase(repo finance.Repository) *DeleteTransactionUseCase {
	return &DeleteTransactionUseCase{repo: repo}
}

func (uc *DeleteTransactionUseCase) Execute(ctx context.Context, id uuid.UUID) error {
	return uc.repo.DeleteTransaction(ctx, id)
}

type GetSummaryUseCase struct {
	repo finance.Repository
}

func NewGetSummaryUseCase(repo finance.Repository) *GetSummaryUseCase {
	return &GetSummaryUseCase{repo: repo}
}

func (uc *GetSummaryUseCase) Execute(ctx context.Context, from, to time.Time) (*finance.FinanceSummary, error) {
	return uc.repo.GetSummary(ctx, from, to)
}

// ─── Purchase Order Use Cases ─────────────────────────────────────────────────

type CreatePurchaseOrderItemInput struct {
	ItemType  finance.PurchaseItemType
	ItemName  string
	Quantity  float64
	Unit      string
	UnitPrice float64
}

type CreatePurchaseOrderInput struct {
	SupplierName string
	PurchaseDate time.Time
	Notes        string
	Items        []CreatePurchaseOrderItemInput
}

type CreatePurchaseOrderUseCase struct {
	repo finance.Repository
}

func NewCreatePurchaseOrderUseCase(repo finance.Repository) *CreatePurchaseOrderUseCase {
	return &CreatePurchaseOrderUseCase{repo: repo}
}

// Execute: Buat purchase order DAN otomatis buat finance_transaction EXPENSE yang ter-link
func (uc *CreatePurchaseOrderUseCase) Execute(ctx context.Context, input CreatePurchaseOrderInput) (*finance.PurchaseOrder, error) {
	items := make([]finance.PurchaseOrderItem, len(input.Items))
	for i, it := range input.Items {
		items[i] = finance.NewPurchaseOrderItem(
			it.ItemType,
			it.ItemName,
			it.Quantity,
			it.Unit,
			it.UnitPrice,
		)
	}

	po, err := finance.NewPurchaseOrder(input.SupplierName, input.PurchaseDate, input.Notes, items)
	if err != nil {
		return nil, err
	}

	// 1. Simpan purchase order dulu
	if err := uc.repo.CreatePurchaseOrder(ctx, po); err != nil {
		return nil, err
	}

	// 2. Tentukan kategori expense berdasarkan jenis item pertama / mayoritas
	category := finance.CategoryOtherExpense
	if len(input.Items) > 0 {
		switch input.Items[0].ItemType {
		case finance.PurchaseItemFilament:
			category = finance.CategoryFilament
		case finance.PurchaseItemHardware:
			category = finance.CategoryHardware
		case finance.PurchaseItemPackaging:
			category = finance.CategoryPackaging
		}
	}

	// 3. Otomatis buat finance_transaction EXPENSE yang ter-link ke PO ini
	tx, err := finance.NewFinanceTransaction(
		finance.TypeExpense,
		category,
		po.TotalAmount(),
		"Pembelian dari "+input.SupplierName,
		input.PurchaseDate,
		"PURCHASE_ORDER",
		po.ID().String(),
		input.Notes,
	)
	if err != nil {
		return nil, err
	}
	if err := uc.repo.CreateTransaction(ctx, tx); err != nil {
		return nil, err
	}

	// 4. Update link di purchase order
	if err := uc.repo.UpdatePurchaseOrderTransaction(ctx, po.ID(), tx.ID()); err != nil {
		return nil, err
	}

	po.SetLinkedTransaction(tx.ID())
	return po, nil
}

type ListPurchaseOrdersUseCase struct {
	repo finance.Repository
}

func NewListPurchaseOrdersUseCase(repo finance.Repository) *ListPurchaseOrdersUseCase {
	return &ListPurchaseOrdersUseCase{repo: repo}
}

func (uc *ListPurchaseOrdersUseCase) Execute(ctx context.Context) ([]*finance.PurchaseOrder, error) {
	return uc.repo.FindPurchaseOrders(ctx)
}

type GetPurchaseOrderUseCase struct {
	repo finance.Repository
}

func NewGetPurchaseOrderUseCase(repo finance.Repository) *GetPurchaseOrderUseCase {
	return &GetPurchaseOrderUseCase{repo: repo}
}

func (uc *GetPurchaseOrderUseCase) Execute(ctx context.Context, id uuid.UUID) (*finance.PurchaseOrder, error) {
	return uc.repo.FindPurchaseOrderByID(ctx, id)
}

type DeletePurchaseOrderUseCase struct {
	repo finance.Repository
}

func NewDeletePurchaseOrderUseCase(repo finance.Repository) *DeletePurchaseOrderUseCase {
	return &DeletePurchaseOrderUseCase{repo: repo}
}

func (uc *DeletePurchaseOrderUseCase) Execute(ctx context.Context, id uuid.UUID) error {
	return uc.repo.DeletePurchaseOrder(ctx, id)
}

// ─── Capital Record Use Cases ─────────────────────────────────────────────────

type CreateCapitalRecordInput struct {
	Type        finance.CapitalRecordType
	Amount      float64
	Description string
	RecordDate  time.Time
	Notes       string
}

type CreateCapitalRecordUseCase struct {
	repo finance.Repository
}

func NewCreateCapitalRecordUseCase(repo finance.Repository) *CreateCapitalRecordUseCase {
	return &CreateCapitalRecordUseCase{repo: repo}
}

// Execute: Buat capital record DAN otomatis buat finance_transaction CAPITAL_IN
func (uc *CreateCapitalRecordUseCase) Execute(ctx context.Context, input CreateCapitalRecordInput) (*finance.CapitalRecord, error) {
	cr, err := finance.NewCapitalRecord(input.Type, input.Amount, input.Description, input.RecordDate, input.Notes)
	if err != nil {
		return nil, err
	}

	// 1. Simpan capital record
	if err := uc.repo.CreateCapitalRecord(ctx, cr); err != nil {
		return nil, err
	}

	// 2. Tentukan kategori
	category := finance.CategoryAdditionalCapital
	if input.Type == finance.CapitalInitial {
		category = finance.CategoryInitialCapital
	}

	// 3. Otomatis buat finance_transaction CAPITAL_IN
	tx, err := finance.NewFinanceTransaction(
		finance.TypeCapitalIn,
		category,
		input.Amount,
		input.Description,
		input.RecordDate,
		"CAPITAL_RECORD",
		cr.ID().String(),
		input.Notes,
	)
	if err != nil {
		return nil, err
	}
	if err := uc.repo.CreateTransaction(ctx, tx); err != nil {
		return nil, err
	}

	cr.SetLinkedTransaction(tx.ID())
	return cr, nil
}

type ListCapitalRecordsUseCase struct {
	repo finance.Repository
}

func NewListCapitalRecordsUseCase(repo finance.Repository) *ListCapitalRecordsUseCase {
	return &ListCapitalRecordsUseCase{repo: repo}
}

func (uc *ListCapitalRecordsUseCase) Execute(ctx context.Context) ([]*finance.CapitalRecord, error) {
	return uc.repo.FindCapitalRecords(ctx)
}

type GetTotalCapitalUseCase struct {
	repo finance.Repository
}

func NewGetTotalCapitalUseCase(repo finance.Repository) *GetTotalCapitalUseCase {
	return &GetTotalCapitalUseCase{repo: repo}
}

func (uc *GetTotalCapitalUseCase) Execute(ctx context.Context) (float64, error) {
	return uc.repo.GetTotalCapital(ctx)
}

type GetOrderAllocationUseCase struct {
	orderRepo order.Repository
}

func NewGetOrderAllocationUseCase(orderRepo order.Repository) *GetOrderAllocationUseCase {
	return &GetOrderAllocationUseCase{orderRepo: orderRepo}
}

func (uc *GetOrderAllocationUseCase) Execute(ctx context.Context, userID uuid.UUID, channel string, dateFrom, dateTo *time.Time) (*finance.OrderAllocationSummary, error) {
	return uc.orderRepo.GetOrderAllocationSummary(ctx, userID, channel, dateFrom, dateTo)
}
