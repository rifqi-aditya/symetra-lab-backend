package finance

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Repository interface {
	// FinanceTransaction
	CreateTransaction(ctx context.Context, t *FinanceTransaction) error
	CreateTransactionIfNotExists(ctx context.Context, t *FinanceTransaction) error // idempotent, skip if reference already exists
	UpsertShopeeEscrowTransaction(ctx context.Context, t *FinanceTransaction) error // inserts or updates settlement date & amount
	FindTransactionByID(ctx context.Context, id uuid.UUID) (*FinanceTransaction, error)
	FindTransactions(ctx context.Context, filter TransactionFilter) ([]*FinanceTransaction, error)
	DeleteTransaction(ctx context.Context, id uuid.UUID) error
	GetSummary(ctx context.Context, from, to time.Time) (*FinanceSummary, error)

	// CashAccount (Pos Kas)
	ListCashAccounts(ctx context.Context) ([]*CashAccount, error)

	// PurchaseOrder
	CreatePurchaseOrder(ctx context.Context, po *PurchaseOrder) error
	FindPurchaseOrderByID(ctx context.Context, id uuid.UUID) (*PurchaseOrder, error)
	FindPurchaseOrders(ctx context.Context) ([]*PurchaseOrder, error)
	UpdatePurchaseOrderTransaction(ctx context.Context, poID, txID uuid.UUID) error
	DeletePurchaseOrder(ctx context.Context, id uuid.UUID) error

	// CapitalRecord
	CreateCapitalRecord(ctx context.Context, cr *CapitalRecord) error
	FindCapitalRecords(ctx context.Context) ([]*CapitalRecord, error)
	DeleteCapitalRecord(ctx context.Context, id uuid.UUID) error
	GetTotalCapital(ctx context.Context) (float64, error)
}

type TransactionFilter struct {
	Type     TransactionType
	Category TransactionCategory
	DateFrom *time.Time
	DateTo   *time.Time
}
