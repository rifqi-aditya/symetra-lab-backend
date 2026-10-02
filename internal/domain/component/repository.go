package component

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	FindAll(ctx context.Context, userID uuid.UUID) ([]*Component, error)
	FindByID(ctx context.Context, userID, id uuid.UUID) (*Component, error)
	Create(ctx context.Context, c *Component) error
	Update(ctx context.Context, c *Component) error
	Delete(ctx context.Context, userID, id uuid.UUID, force bool) error
}