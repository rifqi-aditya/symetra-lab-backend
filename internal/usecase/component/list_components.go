package component

import (
	"context"

	"github.com/google/uuid"

	"symetra-lab-backend-v2/internal/domain/component"
)

type ListComponentsUseCase struct {
	repo component.Repository
}

func NewListComponentsUseCase(repo component.Repository) *ListComponentsUseCase {
	return &ListComponentsUseCase{repo: repo}
}

func (uc *ListComponentsUseCase) Execute(ctx context.Context, userID uuid.UUID) ([]*component.Component, error) {
	if userID == uuid.Nil {
		return nil, component.ErrUserIDRequired
	}
	return uc.repo.FindAll(ctx, userID)
}

type GetComponentUseCase struct {
	repo component.Repository
}

func NewGetComponentUseCase(repo component.Repository) *GetComponentUseCase {
	return &GetComponentUseCase{repo: repo}
}

func (uc *GetComponentUseCase) Execute(ctx context.Context, userID, id uuid.UUID) (*component.Component, error) {
	if userID == uuid.Nil {
		return nil, component.ErrUserIDRequired
	}
	return uc.repo.FindByID(ctx, userID, id)
}