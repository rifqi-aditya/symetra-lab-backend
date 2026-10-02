package product

import (
	"context"

	"github.com/google/uuid"
)

// Repository defines the contract for Product persistence.
type Repository interface {
	FindAll(ctx context.Context, userID uuid.UUID, filter Filter) ([]*Product, int64, error)
	FindByID(ctx context.Context, userID, id uuid.UUID) (*Product, error)
	FindBySKU(ctx context.Context, userID uuid.UUID, sku string) (*Product, error)
	Create(ctx context.Context, p *Product) error
	Update(ctx context.Context, p *Product) error
	Delete(ctx context.Context, userID, id uuid.UUID) error
	GetTotalSoldMap(ctx context.Context, userID uuid.UUID) (map[uuid.UUID]int, error)
	GetSoldQtyByID(ctx context.Context, userID, productID uuid.UUID) (int, error)
	FindAllWithoutSKU(ctx context.Context, userID uuid.UUID) ([]*Product, error)
}

// Filter defines query parameters for searching products.
type Filter struct {
	Search    string
	Category  string
	ParentSKU string
	Limit     int
	Offset    int
}