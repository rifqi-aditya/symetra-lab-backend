package shopconfig

import (
	"context"

	"github.com/google/uuid"

	"symetra-lab-backend-v2/internal/domain/shopconfig"
)

type UpdateMarketplaceInput struct {
	Name                *string
	CommissionPercent   *float64
	PromoFeePercent     *float64
	FreeShippingPercent *float64
	OrderFeeIDR         *float64
	IsActive            *bool
}

type UpdateMarketplaceUseCase struct {
	repo shopconfig.Repository
}

func NewUpdateMarketplaceUseCase(repo shopconfig.Repository) *UpdateMarketplaceUseCase {
	return &UpdateMarketplaceUseCase{repo: repo}
}

func (uc *UpdateMarketplaceUseCase) Execute(ctx context.Context, id uuid.UUID, input UpdateMarketplaceInput) (*shopconfig.MarketplacePlatform, error) {
	p, err := uc.repo.GetMarketplaceByID(ctx, id)
	if err != nil {
		return nil, err
	}

	p.Update(
		input.Name,
		input.CommissionPercent,
		input.PromoFeePercent,
		input.FreeShippingPercent,
		input.OrderFeeIDR,
		input.IsActive,
	)

	if err := uc.repo.UpdateMarketplace(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}
