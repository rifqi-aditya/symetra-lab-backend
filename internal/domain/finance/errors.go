package finance

import "errors"

var (
	ErrTransactionNotFound    = errors.New("finance transaction not found")
	ErrPurchaseOrderNotFound  = errors.New("purchase order not found")
	ErrCapitalRecordNotFound  = errors.New("capital record not found")
	ErrInvalidTransactionType = errors.New("invalid transaction type")
	ErrInvalidCategory        = errors.New("invalid transaction category")
	ErrAmountMustBePositive   = errors.New("amount must be positive")
	ErrDescriptionRequired    = errors.New("description is required")
	ErrSupplierNameRequired   = errors.New("supplier name is required")
	ErrPurchaseItemsRequired  = errors.New("at least one purchase item is required")
	ErrDuplicateReference     = errors.New("a transaction with this reference already exists")
)
