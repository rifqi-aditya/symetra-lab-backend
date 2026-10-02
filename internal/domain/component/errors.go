package component

import "errors"

var (
	ErrUserIDRequired        = errors.New("user ID is required")
	ErrComponentNameRequired = errors.New("component name is required")
	ErrComponentNotFound     = errors.New("component not found")
	ErrInvalidPrice          = errors.New("price per unit cannot be negative")
)