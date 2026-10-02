package component

import (
	"context"

	"github.com/google/uuid"

	"symetra-lab-backend-v2/internal/domain/component"
)

type CreateComponentInput struct {
	Name                 string
	PricePerUnit         float64
	DefaultMarkupPercent float64
	Description          *string
}

type CreateComponentUseCase struct {
	repo component.Repository
}

func NewCreateComponentUseCase(repo component.Repository) *CreateComponentUseCase {
	return &CreateComponentUseCase{repo: repo}
}

func (uc *CreateComponentUseCase) Execute(ctx context.Context, userID uuid.UUID, input CreateComponentInput) (*component.Component, error) {
	comp, err := component.NewComponent(userID, input.Name, input.PricePerUnit, input.DefaultMarkupPercent, input.Description)
	if err != nil {
		return nil, err
	}
	if err := uc.repo.Create(ctx, comp); err != nil {
		return nil, err
	}
	return comp, nil
}

type UpdateComponentInput struct {
	ID                   uuid.UUID
	Name                 string
	PricePerUnit         float64
	DefaultMarkupPercent float64
	Description          *string
}

type UpdateComponentUseCase struct {
	repo component.Repository
}

func NewUpdateComponentUseCase(repo component.Repository) *UpdateComponentUseCase {
	return &UpdateComponentUseCase{repo: repo}
}

func (uc *UpdateComponentUseCase) Execute(ctx context.Context, userID uuid.UUID, input UpdateComponentInput) (*component.Component, error) {
	comp, err := uc.repo.FindByID(ctx, userID, input.ID)
	if err != nil {
		return nil, err
	}
	if err := comp.Update(input.Name, input.PricePerUnit, input.DefaultMarkupPercent, input.Description); err != nil {
		return nil, err
	}
	if err := uc.repo.Update(ctx, comp); err != nil {
		return nil, err
	}
	return comp, nil
}

type DeleteComponentUseCase struct {
	repo component.Repository
}

func NewDeleteComponentUseCase(repo component.Repository) *DeleteComponentUseCase {
	return &DeleteComponentUseCase{repo: repo}
}

func (uc *DeleteComponentUseCase) Execute(ctx context.Context, userID, id uuid.UUID, force bool) error {
	return uc.repo.Delete(ctx, userID, id, force)
}