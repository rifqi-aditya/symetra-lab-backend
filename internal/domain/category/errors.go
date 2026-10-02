package category

import "errors"

var (
	ErrUserIDRequired   = errors.New("user ID is required")
	ErrCategoryNameEmpty = errors.New("category name cannot be empty")
	ErrNotFound          = errors.New("category not found")
)