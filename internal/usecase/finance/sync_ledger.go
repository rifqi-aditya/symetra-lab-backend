package finance

import (
	"context"

	"symetra-lab-backend-v2/internal/domain/finance"
)

type SyncLedgerUseCase struct {
	repo finance.Repository
}

func NewSyncLedgerUseCase(repo finance.Repository) *SyncLedgerUseCase {
	return &SyncLedgerUseCase{repo: repo}
}

func (uc *SyncLedgerUseCase) Execute(ctx context.Context) error {
	return uc.repo.SyncRunningBalances(ctx)
}

