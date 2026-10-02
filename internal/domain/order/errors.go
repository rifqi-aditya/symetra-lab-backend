package order

import "errors"

var (
	ErrOrderNotFound        = errors.New("order not found")
	ErrUserIDRequired       = errors.New("user ID is required")
	ErrCustomerNameRequired = errors.New("customer name is required")
	ErrOrderItemsRequired   = errors.New("at least one order item is required")
	ErrInvalidStatus        = errors.New("invalid order status")
)
