package category

import (
	"context"

	"github.com/google/uuid"

	"symetra-lab-backend-v2/internal/domain/category"
)

type ListCategoriesUseCase struct {
	repo category.Repository
}

func NewListCategoriesUseCase(repo category.Repository) *ListCategoriesUseCase {
	return &ListCategoriesUseCase{repo: repo}
}

func (uc *ListCategoriesUseCase) Execute(ctx context.Context, userID uuid.UUID) ([]*category.Category, error) {
	return uc.repo.FindAll(ctx, userID)
}

type CreateCategoryUseCase struct {
	repo category.Repository
}

func NewCreateCategoryUseCase(repo category.Repository) *CreateCategoryUseCase {
	return &CreateCategoryUseCase{repo: repo}
}

func (uc *CreateCategoryUseCase) Execute(ctx context.Context, userID uuid.UUID, name string) (*category.Category, error) {
	cat, err := category.NewCategory(userID, name)
	if err != nil {
		return nil, err
	}
	if err := uc.repo.Create(ctx, cat); err != nil {
		return nil, err
	}
	return cat, nil
}

type UpdateCategoryUseCase struct {
	repo category.Repository
}

func NewUpdateCategoryUseCase(repo category.Repository) *UpdateCategoryUseCase {
	return &UpdateCategoryUseCase{repo: repo}
}

func (uc *UpdateCategoryUseCase) Execute(ctx context.Context, userID, id uuid.UUID, name string) (*category.Category, error) {
	cat, err := uc.repo.FindByID(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	if err := cat.UpdateName(name); err != nil {
		return nil, err
	}
	if err := uc.repo.Update(ctx, cat); err != nil {
		return nil, err
	}
	return cat, nil
}

type DeleteCategoryUseCase struct {
	repo category.Repository
}

func NewDeleteCategoryUseCase(repo category.Repository) *DeleteCategoryUseCase {
	return &DeleteCategoryUseCase{repo: repo}
}

func (uc *DeleteCategoryUseCase) Execute(ctx context.Context, userID, id uuid.UUID) error {
	return uc.repo.Delete(ctx, userID, id)
}