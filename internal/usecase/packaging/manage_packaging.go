package packaging

import (
	"context"
	"time"

	"github.com/google/uuid"

	"symetra-lab-backend-v2/internal/domain/packaging"
)

type CreatePackagingItemInput struct {
	Name             string
	Category         string
	UnitType         string
	PurchasePrice    float64
	PurchaseQuantity float64
	StockQuantity    float64
}

type CreatePackagingItemUseCase struct {
	repo packaging.Repository
}

func NewCreatePackagingItemUseCase(repo packaging.Repository) *CreatePackagingItemUseCase {
	return &CreatePackagingItemUseCase{repo: repo}
}

func (uc *CreatePackagingItemUseCase) Execute(ctx context.Context, userID uuid.UUID, input CreatePackagingItemInput) (*packaging.PackagingItem, error) {
	item, err := packaging.NewPackagingItem(
		userID,
		input.Name,
		input.Category,
		input.UnitType,
		input.PurchasePrice,
		input.PurchaseQuantity,
		input.StockQuantity,
	)
	if err != nil {
		return nil, err
	}
	if err := uc.repo.Create(ctx, item); err != nil {
		return nil, err
	}
	return item, nil
}

type UpdatePackagingItemInput struct {
	ID               uuid.UUID
	Name             string
	Category         string
	UnitType         string
	PurchasePrice    float64
	PurchaseQuantity float64
	StockQuantity    float64
}

type UpdatePackagingItemUseCase struct {
	repo packaging.Repository
}

func NewUpdatePackagingItemUseCase(repo packaging.Repository) *UpdatePackagingItemUseCase {
	return &UpdatePackagingItemUseCase{repo: repo}
}

func (uc *UpdatePackagingItemUseCase) Execute(ctx context.Context, userID uuid.UUID, input UpdatePackagingItemInput) (*packaging.PackagingItem, error) {
	item, err := uc.repo.FindByID(ctx, userID, input.ID)
	if err != nil {
		return nil, err
	}
	if err := item.Update(input.Name, input.Category, input.UnitType, input.PurchasePrice, input.PurchaseQuantity, input.StockQuantity); err != nil {
		return nil, err
	}
	if err := uc.repo.Update(ctx, item); err != nil {
		return nil, err
	}
	return item, nil
}

type DeletePackagingItemUseCase struct {
	repo packaging.Repository
}

func NewDeletePackagingItemUseCase(repo packaging.Repository) *DeletePackagingItemUseCase {
	return &DeletePackagingItemUseCase{repo: repo}
}

func (uc *DeletePackagingItemUseCase) Execute(ctx context.Context, userID, id uuid.UUID, force bool) error {
	return uc.repo.Delete(ctx, userID, id, force)
}

type PresetItemInput struct {
	PackagingItemID uuid.UUID
	QuantityUsed    float64
}

type CreatePresetInput struct {
	Name        string
	Description *string
	Items       []PresetItemInput
}

type CreatePresetUseCase struct {
	repo packaging.Repository
}

func NewCreatePresetUseCase(repo packaging.Repository) *CreatePresetUseCase {
	return &CreatePresetUseCase{repo: repo}
}

func (uc *CreatePresetUseCase) Execute(ctx context.Context, userID uuid.UUID, input CreatePresetInput) (*packaging.PackagingPreset, error) {
	now := time.Now()
	presetID := uuid.New()
	items := make([]*packaging.PackagingPresetItem, len(input.Items))
	for i, it := range input.Items {
		pkgItem, _ := uc.repo.FindByID(ctx, userID, it.PackagingItemID)
		items[i] = packaging.ReconstructPresetItem(
			uuid.New(),
			presetID,
			it.PackagingItemID,
			it.QuantityUsed,
			pkgItem,
			now,
		)
	}

	preset := packaging.ReconstructPackagingPreset(
		presetID,
		userID,
		input.Name,
		input.Description,
		items,
		now,
		now,
	)

	if err := uc.repo.CreatePreset(ctx, preset); err != nil {
		return nil, err
	}
	return preset, nil
}

type UpdatePresetInput struct {
	ID          uuid.UUID
	Name        string
	Description *string
	Items       []PresetItemInput
}

type UpdatePresetUseCase struct {
	repo packaging.Repository
}

func NewUpdatePresetUseCase(repo packaging.Repository) *UpdatePresetUseCase {
	return &UpdatePresetUseCase{repo: repo}
}

func (uc *UpdatePresetUseCase) Execute(ctx context.Context, userID uuid.UUID, input UpdatePresetInput) (*packaging.PackagingPreset, error) {
	existing, err := uc.repo.FindPresetByID(ctx, userID, input.ID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	items := make([]*packaging.PackagingPresetItem, len(input.Items))
	for i, it := range input.Items {
		pkgItem, _ := uc.repo.FindByID(ctx, userID, it.PackagingItemID)
		items[i] = packaging.ReconstructPresetItem(
			uuid.New(),
			existing.ID(),
			it.PackagingItemID,
			it.QuantityUsed,
			pkgItem,
			now,
		)
	}

	existing.Update(input.Name, input.Description, items)
	if err := uc.repo.UpdatePreset(ctx, existing); err != nil {
		return nil, err
	}
	return existing, nil
}

type DeletePresetUseCase struct {
	repo packaging.Repository
}

func NewDeletePresetUseCase(repo packaging.Repository) *DeletePresetUseCase {
	return &DeletePresetUseCase{repo: repo}
}

func (uc *DeletePresetUseCase) Execute(ctx context.Context, userID, id uuid.UUID, force bool) error {
	return uc.repo.DeletePreset(ctx, userID, id, force)
}