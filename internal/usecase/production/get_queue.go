package production

import (
	"symetra-lab-backend-v2/internal/domain/production"
)

type GetProductionQueueUseCase struct {
	repo production.Repository
}

func NewGetProductionQueueUseCase(repo production.Repository) *GetProductionQueueUseCase {
	return &GetProductionQueueUseCase{repo: repo}
}

func (uc *GetProductionQueueUseCase) Execute() ([]*production.ProductionQueueItem, error) {
	return uc.repo.GetUnifiedQueue()
}
