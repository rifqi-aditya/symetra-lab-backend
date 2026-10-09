package order

import (
	"context"
	"time"

	"github.com/google/uuid"

	"symetra-lab-backend-v2/internal/domain/finance"
)

type Repository interface {
	FindAll(ctx context.Context, userID uuid.UUID) ([]*Order, error)
	FindByID(ctx context.Context, userID, id uuid.UUID) (*Order, error)
	FindByOrderNumber(ctx context.Context, userID uuid.UUID, orderNumber string) (*Order, error)
	Create(ctx context.Context, o *Order) error
	UpdateStatus(ctx context.Context, userID, id uuid.UUID, status string) error
	UpdatePaymentStatus(ctx context.Context, userID, id uuid.UUID, paymentStatus string) error
	Delete(ctx context.Context, userID, id uuid.UUID) error
	GetOrderAllocationSummary(ctx context.Context, userID uuid.UUID, channel string, dateFrom, dateTo *time.Time) (*finance.OrderAllocationSummary, error)
	GetNextManualOrderNumber(ctx context.Context, userID uuid.UUID, t time.Time) (string, error)
}

