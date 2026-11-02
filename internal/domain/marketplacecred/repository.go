package marketplacecred

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	FindByPlatform(ctx context.Context, platform Platform) (*MarketplaceCredential, error)
	FindAll(ctx context.Context) ([]*MarketplaceCredential, error)
	Save(ctx context.Context, cred *MarketplaceCredential) error
	Delete(ctx context.Context, id uuid.UUID) error
}

