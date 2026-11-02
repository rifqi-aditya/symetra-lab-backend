package category

import (
	"context"

	"github.com/google/uuid"

	"symetra-lab-backend-v2/internal/domain/category"
)

// =============================================================================
// 1. LIST CATEGORIES USE CASE
// =============================================================================

// ListCategoriesUseCase retrieves all product categories belonging to a specific user.
type ListCategoriesUseCase struct {
	repo category.Repository
}

// NewListCategoriesUseCase creates a new instance of ListCategoriesUseCase.
func NewListCategoriesUseCase(repo category.Repository) *ListCategoriesUseCase {
	return &ListCategoriesUseCase{repo: repo}
}

// Execute retrieves categories from the repository ordered alphabetically by name.
func (uc *ListCategoriesUseCase) Execute(ctx context.Context, userID uuid.UUID) ([]*category.Category, error) {
	return uc.repo.FindAll(ctx, userID)
}

// =============================================================================
// 2. CREATE CATEGORY USE CASE
// =============================================================================

// CreateCategoryUseCase handles the creation and persistence of a new product category.
type CreateCategoryUseCase struct {
	repo category.Repository
}

// NewCreateCategoryUseCase creates a new instance of CreateCategoryUseCase.
func NewCreateCategoryUseCase(repo category.Repository) *CreateCategoryUseCase {
	return &CreateCategoryUseCase{repo: repo}
}

// Execute enforces domain validation rules for the category entity and persists it via the repository.
func (uc *CreateCategoryUseCase) Execute(ctx context.Context, userID uuid.UUID, name string) (*category.Category, error) {
	// 1. Validate core business invariants in the domain entity
	cat, err := category.NewCategory(userID, name)
	if err != nil {
		return nil, err
	}

	// 2. Persist the valid domain entity to database
	if err := uc.repo.Create(ctx, cat); err != nil {
		return nil, err
	}

	return cat, nil
}

// =============================================================================
// 3. UPDATE CATEGORY USE CASE
// =============================================================================

// UpdateCategoryUseCase handles updating the name of an existing product category.
type UpdateCategoryUseCase struct {
	repo category.Repository
}

// NewUpdateCategoryUseCase creates a new instance of UpdateCategoryUseCase.
func NewUpdateCategoryUseCase(repo category.Repository) *UpdateCategoryUseCase {
	return &UpdateCategoryUseCase{repo: repo}
}

// Execute fetches the category by ID, validates the updated name through domain logic,
// and saves the changes to the repository.
func (uc *UpdateCategoryUseCase) Execute(ctx context.Context, userID, id uuid.UUID, name string) (*category.Category, error) {
	// 1. Retrieve the existing category from persistence
	cat, err := uc.repo.FindByID(ctx, userID, id)
	if err != nil {
		return nil, err
	}

	// 2. Mutate entity via domain method (runs domain name validation)
	if err := cat.UpdateName(name); err != nil {
		return nil, err
	}

	// 3. Persist the updated entity
	if err := uc.repo.Update(ctx, cat); err != nil {
		return nil, err
	}

	return cat, nil
}

// =============================================================================
// 4. DELETE CATEGORY USE CASE
// =============================================================================

// DeleteCategoryUseCase handles the deletion of a product category.
type DeleteCategoryUseCase struct {
	repo category.Repository
}

// NewDeleteCategoryUseCase creates a new instance of DeleteCategoryUseCase.
func NewDeleteCategoryUseCase(repo category.Repository) *DeleteCategoryUseCase {
	return &DeleteCategoryUseCase{repo: repo}
}

// Execute deletes the category by ID for the specified user.
func (uc *DeleteCategoryUseCase) Execute(ctx context.Context, userID, id uuid.UUID) error {
	return uc.repo.Delete(ctx, userID, id)
}
