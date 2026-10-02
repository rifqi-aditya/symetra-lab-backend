package filament

import (
	"context"

	"github.com/google/uuid"

	"symetra-lab-backend-v2/internal/domain/filament"
)

type CreateFilamentInput struct {
	Brand                  string
	MaterialType           string
	SpoolWeightGrams       float64
	ProfileID              *uuid.UUID
	ColorName              string
	ColorHex               string
	SKU                    *string
	PricePerRoll           float64
	CurrentStockGrams      float64
	LowStockThresholdGrams float64
}

type CreateFilamentUseCase struct {
	repo filament.Repository
}

func NewCreateFilamentUseCase(repo filament.Repository) *CreateFilamentUseCase {
	return &CreateFilamentUseCase{repo: repo}
}

func (uc *CreateFilamentUseCase) Execute(ctx context.Context, userID uuid.UUID, input CreateFilamentInput) (*filament.Filament, error) {
	profileID := input.ProfileID
	if profileID == nil || *profileID == uuid.Nil {
		p, err := uc.repo.FindOrCreateProfile(ctx, userID, input.Brand, input.MaterialType, input.SpoolWeightGrams)
		if err != nil {
			return nil, err
		}
		pid := p.ID()
		profileID = &pid
	}

	threshold := input.LowStockThresholdGrams
	if threshold <= 0 {
		threshold = 200
	}
	f, err := filament.NewFilament(
		userID,
		profileID,
		input.ColorName,
		input.ColorHex,
		input.SKU,
		input.PricePerRoll,
		input.CurrentStockGrams,
		threshold,
	)
	if err != nil {
		return nil, err
	}
	if err := uc.repo.Create(ctx, f); err != nil {
		return nil, err
	}
	return uc.repo.FindByID(ctx, userID, f.ID())
}

type UpdateFilamentInput struct {
	ID                     uuid.UUID
	ProfileID              *uuid.UUID
	ColorName              string
	ColorHex               string
	SKU                    *string
	PricePerRoll           float64
	CurrentStockGrams      float64
	LowStockThresholdGrams float64
}

type UpdateFilamentUseCase struct {
	repo filament.Repository
}

func NewUpdateFilamentUseCase(repo filament.Repository) *UpdateFilamentUseCase {
	return &UpdateFilamentUseCase{repo: repo}
}

func (uc *UpdateFilamentUseCase) Execute(ctx context.Context, userID uuid.UUID, input UpdateFilamentInput) (*filament.Filament, error) {
	f, err := uc.repo.FindByID(ctx, userID, input.ID)
	if err != nil {
		return nil, err
	}
	f.Update(
		input.ProfileID,
		input.ColorName,
		input.ColorHex,
		input.SKU,
		input.PricePerRoll,
		input.CurrentStockGrams,
		input.LowStockThresholdGrams,
	)
	if err := uc.repo.Update(ctx, f); err != nil {
		return nil, err
	}
	return uc.repo.FindByID(ctx, userID, f.ID())
}

type DeleteFilamentUseCase struct {
	repo filament.Repository
}

func NewDeleteFilamentUseCase(repo filament.Repository) *DeleteFilamentUseCase {
	return &DeleteFilamentUseCase{repo: repo}
}

func (uc *DeleteFilamentUseCase) Execute(ctx context.Context, userID, id uuid.UUID) error {
	return uc.repo.Delete(ctx, userID, id)
}

type SyncStockUseCase struct {
	repo filament.Repository
}

func NewSyncStockUseCase(repo filament.Repository) *SyncStockUseCase {
	return &SyncStockUseCase{repo: repo}
}

func (uc *SyncStockUseCase) Execute(ctx context.Context, userID, id uuid.UUID, grams float64, isWeighed bool) (*filament.Filament, error) {
	return uc.repo.SyncStock(ctx, userID, id, grams, isWeighed)
}

// Material Rates
type ListMaterialRatesUseCase struct {
	repo filament.Repository
}

func NewListMaterialRatesUseCase(repo filament.Repository) *ListMaterialRatesUseCase {
	return &ListMaterialRatesUseCase{repo: repo}
}

func (uc *ListMaterialRatesUseCase) Execute(ctx context.Context, userID uuid.UUID) ([]*filament.FilamentMaterialRate, error) {
	return uc.repo.FindAllMaterialRates(ctx, userID)
}

type CreateMaterialRateInput struct {
	MaterialType string
	PricePerGram float64
	IsDefault    bool
	Description  *string
}

type CreateMaterialRateUseCase struct {
	repo filament.Repository
}

func NewCreateMaterialRateUseCase(repo filament.Repository) *CreateMaterialRateUseCase {
	return &CreateMaterialRateUseCase{repo: repo}
}

func (uc *CreateMaterialRateUseCase) Execute(ctx context.Context, userID uuid.UUID, input CreateMaterialRateInput) (*filament.FilamentMaterialRate, error) {
	rate, err := filament.NewMaterialRate(userID, input.MaterialType, input.PricePerGram, input.IsDefault, input.Description)
	if err != nil {
		return nil, err
	}
	if err := uc.repo.CreateMaterialRate(ctx, rate); err != nil {
		return nil, err
	}
	return rate, nil
}

type UpdateMaterialRateInput struct {
	ID           uuid.UUID
	PricePerGram float64
	IsDefault    bool
	Description  *string
}

type UpdateMaterialRateUseCase struct {
	repo filament.Repository
}

func NewUpdateMaterialRateUseCase(repo filament.Repository) *UpdateMaterialRateUseCase {
	return &UpdateMaterialRateUseCase{repo: repo}
}

func (uc *UpdateMaterialRateUseCase) Execute(ctx context.Context, userID uuid.UUID, input UpdateMaterialRateInput) (*filament.FilamentMaterialRate, error) {
	rate, err := uc.repo.FindMaterialRateByID(ctx, userID, input.ID)
	if err != nil {
		return nil, err
	}
	rate.Update(input.PricePerGram, input.IsDefault, input.Description)
	if err := uc.repo.UpdateMaterialRate(ctx, rate); err != nil {
		return nil, err
	}
	return rate, nil
}

type DeleteMaterialRateUseCase struct {
	repo filament.Repository
}

func NewDeleteMaterialRateUseCase(repo filament.Repository) *DeleteMaterialRateUseCase {
	return &DeleteMaterialRateUseCase{repo: repo}
}

func (uc *DeleteMaterialRateUseCase) Execute(ctx context.Context, userID, id uuid.UUID) error {
	return uc.repo.DeleteMaterialRate(ctx, userID, id)
}