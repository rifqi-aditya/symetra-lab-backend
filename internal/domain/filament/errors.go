package filament

import "errors"

var (
	ErrUserIDRequired    = errors.New("user ID is required")
	ErrColorNameRequired = errors.New("color name is required")
	ErrColorHexRequired  = errors.New("color hex is required")
	ErrFilamentNotFound  = errors.New("filament not found")
	ErrProfileNotFound   = errors.New("filament profile not found")
	ErrInvalidStock      = errors.New("stock cannot be negative")
)