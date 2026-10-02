package shopconfig

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	GetByUserID(ctx context.Context, userID uuid.UUID) (*ShopConfig, error)
	Save(ctx context.Context, cfg *ShopConfig) error
	GetActiveMarketplace(ctx context.Context, userID uuid.UUID, platformName string) (*MarketplacePlatform, error)
	ListMarketplaces(ctx context.Context, userID uuid.UUID) ([]*MarketplacePlatform, error)
	GetMarketplaceByID(ctx context.Context, id uuid.UUID) (*MarketplacePlatform, error)
	CreateMarketplace(ctx context.Context, p *MarketplacePlatform) error
	UpdateMarketplace(ctx context.Context, p *MarketplacePlatform) error
	DeleteMarketplace(ctx context.Context, id uuid.UUID) error
}
