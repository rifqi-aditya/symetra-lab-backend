package shopconfig

import (
	"context"

	"github.com/google/uuid"

	"symetra-lab-backend-v2/internal/domain/shopconfig"
)

type ListMarketplacesUseCase struct {
	repo shopconfig.Repository
}

func NewListMarketplacesUseCase(repo shopconfig.Repository) *ListMarketplacesUseCase {
	return &ListMarketplacesUseCase{repo: repo}
}

func (uc *ListMarketplacesUseCase) Execute(ctx context.Context, userID uuid.UUID) ([]*shopconfig.MarketplacePlatform, error) {
	return uc.repo.ListMarketplaces(ctx, userID)
}
