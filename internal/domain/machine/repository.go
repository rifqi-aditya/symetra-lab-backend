package machine

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	FindAll(ctx context.Context, userID uuid.UUID) ([]*Machine, error)
	FindByID(ctx context.Context, userID, id uuid.UUID) (*Machine, error)
	Create(ctx context.Context, m *Machine) error
	Update(ctx context.Context, m *Machine) error
	UpdateState(ctx context.Context, userID, id uuid.UUID, state string) error
	Delete(ctx context.Context, userID, id uuid.UUID) error

	FindPartsByMachineID(ctx context.Context, machineID uuid.UUID) ([]*MachineMaintenancePart, error)
	FindPartByID(ctx context.Context, partID uuid.UUID) (*MachineMaintenancePart, error)
	CreatePart(ctx context.Context, part *MachineMaintenancePart) error
	UpdatePart(ctx context.Context, part *MachineMaintenancePart) error
	ReplacePart(ctx context.Context, partID uuid.UUID, cost float64) (*MachineMaintenancePart, error)
	DeletePart(ctx context.Context, partID uuid.UUID) error
}