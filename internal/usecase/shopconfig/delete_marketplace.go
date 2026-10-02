package shopconfig

import (
	"context"

	"github.com/google/uuid"

	"symetra-lab-backend-v2/internal/domain/shopconfig"
)

type DeleteMarketplaceUseCase struct {
	repo shopconfig.Repository
}

func NewDeleteMarketplaceUseCase(repo shopconfig.Repository) *DeleteMarketplaceUseCase {
	return &DeleteMarketplaceUseCase{repo: repo}
}

func (uc *DeleteMarketplaceUseCase) Execute(ctx context.Context, id uuid.UUID) error {
	return uc.repo.DeleteMarketplace(ctx, id)
}
