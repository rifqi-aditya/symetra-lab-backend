package machine

import (
	"context"

	"github.com/google/uuid"

	"symetra-lab-backend-v2/internal/domain/machine"
)

type CreateMachineInput struct {
	Name              string
	Brand             *string
	TotalPurchaseCost float64
	LifespanHours     int
	AvgPowerWatts     int
}

type CreateMachineUseCase struct {
	repo machine.Repository
}

func NewCreateMachineUseCase(repo machine.Repository) *CreateMachineUseCase {
	return &CreateMachineUseCase{repo: repo}
}

func (uc *CreateMachineUseCase) Execute(ctx context.Context, userID uuid.UUID, input CreateMachineInput) (*machine.Machine, error) {
	m, err := machine.NewMachine(
		userID,
		input.Name,
		input.Brand,
		input.TotalPurchaseCost,
		input.LifespanHours,
		input.AvgPowerWatts,
	)
	if err != nil {
		return nil, err
	}
	if err := uc.repo.Create(ctx, m); err != nil {
		return nil, err
	}
	return uc.repo.FindByID(ctx, userID, m.ID())
}

type UpdateMachineInput struct {
	ID                       uuid.UUID
	Name                     string
	Brand                    *string
	TotalPurchaseCost        float64
	SalvageValue             *float64
	LifespanHours            int
	PurchaseDate             *string
	TotalHoursUsed           float64
	AvgPowerWatts            int
	MaintenanceBufferPerHour *float64
	FailureRatePercent       *float64
	BuildVolumeX             *float64
	BuildVolumeY             *float64
	BuildVolumeZ             *float64
	SpeedMultiplier          *float64
	CurrentState             string
	ElectricityCostPerHour   *float64
}

type UpdateMachineUseCase struct {
	repo machine.Repository
}

func NewUpdateMachineUseCase(repo machine.Repository) *UpdateMachineUseCase {
	return &UpdateMachineUseCase{repo: repo}
}

func (uc *UpdateMachineUseCase) Execute(ctx context.Context, userID uuid.UUID, input UpdateMachineInput) (*machine.Machine, error) {
	existing, err := uc.repo.FindByID(ctx, userID, input.ID)
	if err != nil {
		return nil, err
	}

	m := machine.ReconstructMachine(
		existing.ID(),
		existing.UserID(),
		input.Name,
		input.Brand,
		input.TotalPurchaseCost,
		input.SalvageValue,
		input.LifespanHours,
		input.PurchaseDate,
		input.TotalHoursUsed,
		input.AvgPowerWatts,
		input.MaintenanceBufferPerHour,
		input.FailureRatePercent,
		input.BuildVolumeX,
		input.BuildVolumeY,
		input.BuildVolumeZ,
		input.SpeedMultiplier,
		input.CurrentState,
		input.ElectricityCostPerHour,
		existing.MaintenanceParts(),
		existing.CreatedAt(),
		existing.UpdatedAt(),
	)

	if err := uc.repo.Update(ctx, m); err != nil {
		return nil, err
	}
	return uc.repo.FindByID(ctx, userID, m.ID())
}

type UpdateMachineStateUseCase struct {
	repo machine.Repository
}

func NewUpdateMachineStateUseCase(repo machine.Repository) *UpdateMachineStateUseCase {
	return &UpdateMachineStateUseCase{repo: repo}
}

func (uc *UpdateMachineStateUseCase) Execute(ctx context.Context, userID, id uuid.UUID, state string) error {
	return uc.repo.UpdateState(ctx, userID, id, state)
}

type DeleteMachineUseCase struct {
	repo machine.Repository
}

func NewDeleteMachineUseCase(repo machine.Repository) *DeleteMachineUseCase {
	return &DeleteMachineUseCase{repo: repo}
}

func (uc *DeleteMachineUseCase) Execute(ctx context.Context, userID, id uuid.UUID) error {
	return uc.repo.Delete(ctx, userID, id)
}

// Parts
type ListPartsUseCase struct {
	repo machine.Repository
}

func NewListPartsUseCase(repo machine.Repository) *ListPartsUseCase {
	return &ListPartsUseCase{repo: repo}
}

func (uc *ListPartsUseCase) Execute(ctx context.Context, machineID uuid.UUID) ([]*machine.MachineMaintenancePart, error) {
	return uc.repo.FindPartsByMachineID(ctx, machineID)
}

type CreatePartInput struct {
	MachineID     uuid.UUID
	PartName      string
	CostIDR       float64
	LifespanHours float64
	StockQuantity int
}

type CreatePartUseCase struct {
	repo machine.Repository
}

func NewCreatePartUseCase(repo machine.Repository) *CreatePartUseCase {
	return &CreatePartUseCase{repo: repo}
}

func (uc *CreatePartUseCase) Execute(ctx context.Context, input CreatePartInput) (*machine.MachineMaintenancePart, error) {
	part, err := machine.NewMaintenancePart(
		input.MachineID,
		input.PartName,
		input.CostIDR,
		input.LifespanHours,
		input.StockQuantity,
	)
	if err != nil {
		return nil, err
	}
	if err := uc.repo.CreatePart(ctx, part); err != nil {
		return nil, err
	}
	return part, nil
}

type UpdatePartInput struct {
	ID            uuid.UUID
	PartName      string
	CostIDR       float64
	LifespanHours float64
	StockQuantity int
}

type UpdatePartUseCase struct {
	repo machine.Repository
}

func NewUpdatePartUseCase(repo machine.Repository) *UpdatePartUseCase {
	return &UpdatePartUseCase{repo: repo}
}

func (uc *UpdatePartUseCase) Execute(ctx context.Context, input UpdatePartInput) (*machine.MachineMaintenancePart, error) {
	part, err := uc.repo.FindPartByID(ctx, input.ID)
	if err != nil {
		return nil, err
	}
	part.Update(input.PartName, input.CostIDR, input.LifespanHours, input.StockQuantity)
	if err := uc.repo.UpdatePart(ctx, part); err != nil {
		return nil, err
	}
	return part, nil
}

type ReplacePartUseCase struct {
	repo machine.Repository
}

func NewReplacePartUseCase(repo machine.Repository) *ReplacePartUseCase {
	return &ReplacePartUseCase{repo: repo}
}

func (uc *ReplacePartUseCase) Execute(ctx context.Context, partID uuid.UUID, cost float64) (*machine.MachineMaintenancePart, error) {
	return uc.repo.ReplacePart(ctx, partID, cost)
}

type DeletePartUseCase struct {
	repo machine.Repository
}

func NewDeletePartUseCase(repo machine.Repository) *DeletePartUseCase {
	return &DeletePartUseCase{repo: repo}
}

func (uc *DeletePartUseCase) Execute(ctx context.Context, partID uuid.UUID) error {
	return uc.repo.DeletePart(ctx, partID)
}