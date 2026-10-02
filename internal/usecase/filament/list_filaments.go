package filament

import (
	"context"

	"github.com/google/uuid"

	"symetra-lab-backend-v2/internal/domain/filament"
)

type ListFilamentsUseCase struct {
	repo filament.Repository
}

func NewListFilamentsUseCase(repo filament.Repository) *ListFilamentsUseCase {
	return &ListFilamentsUseCase{repo: repo}
}

func (uc *ListFilamentsUseCase) Execute(ctx context.Context, userID uuid.UUID) ([]*filament.Filament, error) {
	if userID == uuid.Nil {
		return nil, filament.ErrUserIDRequired
	}
	return uc.repo.FindAll(ctx, userID)
}

type GetFilamentUseCase struct {
	repo filament.Repository
}

func NewGetFilamentUseCase(repo filament.Repository) *GetFilamentUseCase {
	return &GetFilamentUseCase{repo: repo}
}

func (uc *GetFilamentUseCase) Execute(ctx context.Context, userID, id uuid.UUID) (*filament.Filament, error) {
	if userID == uuid.Nil {
		return nil, filament.ErrUserIDRequired
	}
	return uc.repo.FindByID(ctx, userID, id)
}

type ListFilamentProfilesUseCase struct {
	repo filament.Repository
}

func NewListFilamentProfilesUseCase(repo filament.Repository) *ListFilamentProfilesUseCase {
	return &ListFilamentProfilesUseCase{repo: repo}
}

func (uc *ListFilamentProfilesUseCase) Execute(ctx context.Context, userID uuid.UUID) ([]*filament.FilamentProfile, error) {
	if userID == uuid.Nil {
		return nil, filament.ErrUserIDRequired
	}
	return uc.repo.FindAllProfiles(ctx, userID)
}