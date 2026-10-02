package product

import "errors"

var (
	ErrUserIDRequired      = errors.New("user ID is required")
	ErrProductNameRequired = errors.New("product name is required")
	ErrCategoryRequired    = errors.New("category is required")
	ErrNotFound            = errors.New("product not found")
	ErrSKUAlreadyExists    = errors.New("SKU already exists")
	ErrInvalidWeight       = errors.New("default weight cannot be negative")
	ErrInvalidPrintTime    = errors.New("default print time cannot be negative")
	ErrProductInUseInOrder = errors.New("product cannot be deleted because it is recorded in order history")
)