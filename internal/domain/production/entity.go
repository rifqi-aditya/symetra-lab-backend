package production

import (
	"time"
)

type ProductionQueueItem struct {
	JobID             string
	Source            string // "SHOPEE" or "MANUAL"
	OrderIdentifier   string
	CustomerName      string
	ItemName          string
	VariantSKU        string
	Quantity          int
	WeightGrams       float64
	PrintTimeHours    float64
	TotalPrintHours   float64
	AssignedMachineID *string
	AssignedMachine   string
	Status            string
	Deadline          *time.Time
	IsUrgent          bool
	CreatedAt         time.Time
}

type CompleteJobResult struct {
	Status             string
	Message            string
	DeductedFilaments  []string
	AddedMachineHours  float64
	LowStockWarning    bool
	MaintenanceWarning bool
}

type Repository interface {
	GetUnifiedQueue() ([]*ProductionQueueItem, error)
	CompleteJob(source, jobID, machineID, filamentID string) (*CompleteJobResult, error)
}
