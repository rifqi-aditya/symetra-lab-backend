package dashboard

import (
	"context"

	"symetra-lab-backend-v2/internal/domain/dashboard"
)

type GetDashboardOverviewUseCase struct {
	repo dashboard.Repository
}

func NewGetDashboardOverviewUseCase(repo dashboard.Repository) *GetDashboardOverviewUseCase {
	return &GetDashboardOverviewUseCase{repo: repo}
}

func (uc *GetDashboardOverviewUseCase) Execute(ctx context.Context) (*dashboard.DashboardOverview, error) {
	return uc.repo.GetOverview(ctx)
}
