package shopconfig

import (
	"context"

	"github.com/google/uuid"

	"symetra-lab-backend-v2/internal/domain/shopconfig"
)

type UpdateShopConfigInput struct {
	FilamentPricePerRoll    float64
	FilamentWeightGrams     float64
	ElectricityTariffPerKwh float64
	PrinterPowerWatts       float64
	PrinterPrice            float64
	PrinterLifespanHours    float64
	FailureBufferPercent    float64
}

type UpdateShopConfigUseCase struct {
	repo shopconfig.Repository
}

func NewUpdateShopConfigUseCase(repo shopconfig.Repository) *UpdateShopConfigUseCase {
	return &UpdateShopConfigUseCase{repo: repo}
}

func (uc *UpdateShopConfigUseCase) Execute(ctx context.Context, userID uuid.UUID, input UpdateShopConfigInput) (*shopconfig.ShopConfig, error) {
	cfg, err := uc.repo.GetByUserID(ctx, userID)
	if err != nil {
		cfg = shopconfig.NewShopConfig(
			userID,
			input.FilamentPricePerRoll,
			input.FilamentWeightGrams,
			input.ElectricityTariffPerKwh,
			input.PrinterPowerWatts,
			input.PrinterPrice,
			input.PrinterLifespanHours,
			input.FailureBufferPercent,
		)
	} else {
		cfg.Update(
			input.FilamentPricePerRoll,
			input.FilamentWeightGrams,
			input.ElectricityTariffPerKwh,
			input.PrinterPowerWatts,
			input.PrinterPrice,
			input.PrinterLifespanHours,
			input.FailureBufferPercent,
		)
	}

	if err := uc.repo.Save(ctx, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
