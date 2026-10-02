package packaging

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	FindAll(ctx context.Context, userID uuid.UUID) ([]*PackagingItem, error)
	FindByID(ctx context.Context, userID, id uuid.UUID) (*PackagingItem, error)
	Create(ctx context.Context, p *PackagingItem) error
	Update(ctx context.Context, p *PackagingItem) error
	Delete(ctx context.Context, userID, id uuid.UUID, force bool) error

	FindAllPresets(ctx context.Context, userID uuid.UUID) ([]*PackagingPreset, error)
	FindPresetByID(ctx context.Context, userID, id uuid.UUID) (*PackagingPreset, error)
	CreatePreset(ctx context.Context, p *PackagingPreset) error
	UpdatePreset(ctx context.Context, p *PackagingPreset) error
	DeletePreset(ctx context.Context, userID, id uuid.UUID, force bool) error
}