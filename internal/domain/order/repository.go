package order

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	FindAll(ctx context.Context, userID uuid.UUID) ([]*Order, error)
	FindByID(ctx context.Context, userID, id uuid.UUID) (*Order, error)
	Create(ctx context.Context, o *Order) error
	UpdateStatus(ctx context.Context, userID, id uuid.UUID, status string) error
	UpdatePaymentStatus(ctx context.Context, userID, id uuid.UUID, paymentStatus string) error
	Delete(ctx context.Context, userID, id uuid.UUID) error
}
