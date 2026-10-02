package shopconfig

import (
	"context"

	"github.com/google/uuid"

	"symetra-lab-backend-v2/internal/domain/shopconfig"
)

type CreateMarketplaceInput struct {
	Name                string
	CommissionPercent   float64
	PromoFeePercent     float64
	FreeShippingPercent float64
	OrderFeeIDR         float64
	IsActive            bool
}

type CreateMarketplaceUseCase struct {
	repo shopconfig.Repository
}

func NewCreateMarketplaceUseCase(repo shopconfig.Repository) *CreateMarketplaceUseCase {
	return &CreateMarketplaceUseCase{repo: repo}
}

func (uc *CreateMarketplaceUseCase) Execute(ctx context.Context, userID uuid.UUID, input CreateMarketplaceInput) (*shopconfig.MarketplacePlatform, error) {
	p := shopconfig.NewMarketplacePlatform(
		userID,
		input.Name,
		input.CommissionPercent,
		input.PromoFeePercent,
		input.FreeShippingPercent,
		input.OrderFeeIDR,
		input.IsActive,
	)

	if err := uc.repo.CreateMarketplace(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}
