package product

import (
	"context"

	"github.com/google/uuid"

	"symetra-lab-backend-v2/internal/domain/product"
)

type DeleteProductUseCase struct {
	productRepo product.Repository
}

func NewDeleteProductUseCase(pRepo product.Repository) *DeleteProductUseCase {
	return &DeleteProductUseCase{productRepo: pRepo}
}

func (uc *DeleteProductUseCase) Execute(ctx context.Context, userID, id uuid.UUID) error {
	return uc.productRepo.Delete(ctx, userID, id)
}