package filament

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	FindAll(ctx context.Context, userID uuid.UUID) ([]*Filament, error)
	FindByID(ctx context.Context, userID, id uuid.UUID) (*Filament, error)
	Create(ctx context.Context, f *Filament) error
	Update(ctx context.Context, f *Filament) error
	Delete(ctx context.Context, userID, id uuid.UUID) error
	SyncStock(ctx context.Context, userID, id uuid.UUID, stockGrams float64, isWeighed bool) (*Filament, error)

	FindAllProfiles(ctx context.Context, userID uuid.UUID) ([]*FilamentProfile, error)
	FindProfileByID(ctx context.Context, userID, id uuid.UUID) (*FilamentProfile, error)
	FindOrCreateProfile(ctx context.Context, userID uuid.UUID, brand, materialType string, spoolWeight float64) (*FilamentProfile, error)

	FindAllMaterialRates(ctx context.Context, userID uuid.UUID) ([]*FilamentMaterialRate, error)
	FindMaterialRateByID(ctx context.Context, userID, id uuid.UUID) (*FilamentMaterialRate, error)
	CreateMaterialRate(ctx context.Context, r *FilamentMaterialRate) error
	UpdateMaterialRate(ctx context.Context, r *FilamentMaterialRate) error
	DeleteMaterialRate(ctx context.Context, userID, id uuid.UUID) error
}