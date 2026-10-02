package packaging

import (
	"context"

	"github.com/google/uuid"

	"symetra-lab-backend-v2/internal/domain/packaging"
)

type ListPackagingUseCase struct {
	repo packaging.Repository
}

func NewListPackagingUseCase(repo packaging.Repository) *ListPackagingUseCase {
	return &ListPackagingUseCase{repo: repo}
}

func (uc *ListPackagingUseCase) Execute(ctx context.Context, userID uuid.UUID) ([]*packaging.PackagingItem, error) {
	if userID == uuid.Nil {
		return nil, packaging.ErrUserIDRequired
	}
	return uc.repo.FindAll(ctx, userID)
}

type GetPackagingUseCase struct {
	repo packaging.Repository
}

func NewGetPackagingUseCase(repo packaging.Repository) *GetPackagingUseCase {
	return &GetPackagingUseCase{repo: repo}
}

func (uc *GetPackagingUseCase) Execute(ctx context.Context, userID, id uuid.UUID) (*packaging.PackagingItem, error) {
	if userID == uuid.Nil {
		return nil, packaging.ErrUserIDRequired
	}
	return uc.repo.FindByID(ctx, userID, id)
}

type ListPackagingPresetsUseCase struct {
	repo packaging.Repository
}

func NewListPackagingPresetsUseCase(repo packaging.Repository) *ListPackagingPresetsUseCase {
	return &ListPackagingPresetsUseCase{repo: repo}
}

func (uc *ListPackagingPresetsUseCase) Execute(ctx context.Context, userID uuid.UUID) ([]*packaging.PackagingPreset, error) {
	if userID == uuid.Nil {
		return nil, packaging.ErrUserIDRequired
	}
	return uc.repo.FindAllPresets(ctx, userID)
}