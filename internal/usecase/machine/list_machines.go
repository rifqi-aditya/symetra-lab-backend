package machine

import (
	"context"

	"github.com/google/uuid"

	"symetra-lab-backend-v2/internal/domain/machine"
)

type ListMachinesUseCase struct {
	repo machine.Repository
}

func NewListMachinesUseCase(repo machine.Repository) *ListMachinesUseCase {
	return &ListMachinesUseCase{repo: repo}
}

func (uc *ListMachinesUseCase) Execute(ctx context.Context, userID uuid.UUID) ([]*machine.Machine, error) {
	if userID == uuid.Nil {
		return nil, machine.ErrUserIDRequired
	}
	return uc.repo.FindAll(ctx, userID)
}

type GetMachineUseCase struct {
	repo machine.Repository
}

func NewGetMachineUseCase(repo machine.Repository) *GetMachineUseCase {
	return &GetMachineUseCase{repo: repo}
}

func (uc *GetMachineUseCase) Execute(ctx context.Context, userID, id uuid.UUID) (*machine.Machine, error) {
	if userID == uuid.Nil {
		return nil, machine.ErrUserIDRequired
	}
	return uc.repo.FindByID(ctx, userID, id)
}