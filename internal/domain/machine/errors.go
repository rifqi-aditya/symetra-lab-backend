package machine

import "errors"

var (
	ErrUserIDRequired      = errors.New("user ID is required")
	ErrMachineNameRequired = errors.New("machine name is required")
	ErrMachineNotFound     = errors.New("machine not found")
	ErrPartNotFound        = errors.New("maintenance part not found")
	ErrInvalidLifespan     = errors.New("lifespan hours must be greater than zero")
)