package packaging

import "errors"

var (
	ErrUserIDRequired        = errors.New("user ID is required")
	ErrPackagingItemNotFound = errors.New("packaging item not found")
	ErrPresetNotFound        = errors.New("packaging preset not found")
	ErrInvalidQuantity       = errors.New("quantity must be greater than zero")
)