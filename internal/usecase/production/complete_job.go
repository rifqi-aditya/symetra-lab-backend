package production

import (
	"symetra-lab-backend-v2/internal/domain/production"
)

type CompleteJobUseCase struct {
	repo production.Repository
}

func NewCompleteJobUseCase(repo production.Repository) *CompleteJobUseCase {
	return &CompleteJobUseCase{repo: repo}
}

func (uc *CompleteJobUseCase) Execute(source, jobID, machineID, filamentID string) (*production.CompleteJobResult, error) {
	return uc.repo.CompleteJob(source, jobID, machineID, filamentID)
}
