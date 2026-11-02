package shopee

import (
	"context"

	"github.com/google/uuid"

	"symetra-lab-backend-v2/internal/domain/shopee"
)

type LinkShopeeSKUUseCase struct {
	repo       shopee.Repository
	ledgerSync LedgerSynchronizer
}

func NewLinkShopeeSKUUseCase(
	repo shopee.Repository,
	ledgerSync LedgerSynchronizer,
) *LinkShopeeSKUUseCase {
	return &LinkShopeeSKUUseCase{
		repo:       repo,
		ledgerSync: ledgerSync,
	}
}

func (uc *LinkShopeeSKUUseCase) Execute(
	ctx context.Context,
	itemID, modelID uint64,
	productID uuid.UUID,
) error {
	if err := uc.repo.LinkSKU(ctx, itemID, modelID, productID); err != nil {
		return err
	}
	if uc.ledgerSync != nil {
		_ = uc.ledgerSync.SyncRunningBalances(ctx)
	}
	return nil
}

