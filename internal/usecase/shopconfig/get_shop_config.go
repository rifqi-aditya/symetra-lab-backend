package shopconfig

import (
	"context"

	"github.com/google/uuid"

	"symetra-lab-backend-v2/internal/domain/shopconfig"
)

type GetShopConfigUseCase struct {
	repo shopconfig.Repository
}

func NewGetShopConfigUseCase(repo shopconfig.Repository) *GetShopConfigUseCase {
	return &GetShopConfigUseCase{repo: repo}
}

func (uc *GetShopConfigUseCase) Execute(ctx context.Context, userID uuid.UUID) (*shopconfig.ShopConfig, error) {
	return uc.repo.GetByUserID(ctx, userID)
}
