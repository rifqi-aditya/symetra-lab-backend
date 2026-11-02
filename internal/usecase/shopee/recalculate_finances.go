package shopee

import (
	"context"

	"symetra-lab-backend-v2/internal/domain/shopee"
)

type RecalculateFinancesUseCase struct {
	repo       shopee.Repository
	ledgerSync LedgerSynchronizer
}

func NewRecalculateFinancesUseCase(
	repo shopee.Repository,
	ledgerSync LedgerSynchronizer,
) *RecalculateFinancesUseCase {
	return &RecalculateFinancesUseCase{
		repo:       repo,
		ledgerSync: ledgerSync,
	}
}

func (uc *RecalculateFinancesUseCase) Execute(ctx context.Context) (int, error) {
	count, err := uc.repo.RecalculateFinances(ctx)
	if err != nil {
		return count, err
	}
	if uc.ledgerSync != nil {
		_ = uc.ledgerSync.SyncRunningBalances(ctx)
	}
	return count, nil
}

