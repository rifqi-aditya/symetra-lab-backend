package dto

import (
	"time"

	"symetra-lab-backend-v2/internal/domain/production"
)

type ProductionQueueItemResponse struct {
	JobID             string     `json:"job_id"`
	Source            string     `json:"source"`
	OrderIdentifier   string     `json:"order_identifier"`
	CustomerName      string     `json:"customer_name"`
	ItemName          string     `json:"item_name"`
	VariantSKU        string     `json:"variant_sku,omitempty"`
	Quantity          int        `json:"quantity"`
	WeightGrams       float64    `json:"weight_grams"`
	PrintTimeHours    float64    `json:"print_time_hours"`
	TotalPrintHours   float64    `json:"total_print_hours"`
	AssignedMachineID *string    `json:"assigned_machine_id,omitempty"`
	AssignedMachine   string     `json:"assigned_machine_name,omitempty"`
	Status            string     `json:"status"`
	Deadline          *time.Time `json:"deadline,omitempty"`
	IsUrgent          bool       `json:"is_urgent"`
	CreatedAt         time.Time  `json:"created_at"`
}

func ToProductionQueueItemResponse(item *production.ProductionQueueItem) ProductionQueueItemResponse {
	return ProductionQueueItemResponse{
		JobID:             item.JobID,
		Source:            item.Source,
		OrderIdentifier:   item.OrderIdentifier,
		CustomerName:      item.CustomerName,
		ItemName:          item.ItemName,
		VariantSKU:        item.VariantSKU,
		Quantity:          item.Quantity,
		WeightGrams:       item.WeightGrams,
		PrintTimeHours:    item.PrintTimeHours,
		TotalPrintHours:   item.TotalPrintHours,
		AssignedMachineID: item.AssignedMachineID,
		AssignedMachine:   item.AssignedMachine,
		Status:            item.Status,
		Deadline:          item.Deadline,
		IsUrgent:          item.IsUrgent,
		CreatedAt:         item.CreatedAt,
	}
}

type ProductionQueueSummaryResponse struct {
	Queue             []ProductionQueueItemResponse `json:"queue"`
	TotalJobs         int                           `json:"total_jobs"`
	TotalHoursWaiting float64                       `json:"total_hours_waiting"`
	UrgentJobsCount   int                           `json:"urgent_jobs_count"`
}

type CompletePrintJobRequest struct {
	Source     string `json:"source"`
	JobID      string `json:"job_id"`
	MachineID  string `json:"machine_id"`
	FilamentID string `json:"filament_id,omitempty"`
}

type CompletePrintJobResponse struct {
	Status             string   `json:"status"`
	Message            string   `json:"message"`
	DeductedFilaments  []string `json:"deducted_filaments"`
	AddedMachineHours  float64  `json:"added_machine_hours"`
	LowStockWarning    bool     `json:"low_stock_warning"`
	MaintenanceWarning bool     `json:"maintenance_warning"`
}

func ToCompletePrintJobResponse(r *production.CompleteJobResult) CompletePrintJobResponse {
	return CompletePrintJobResponse{
		Status:             r.Status,
		Message:            r.Message,
		DeductedFilaments:  r.DeductedFilaments,
		AddedMachineHours:  r.AddedMachineHours,
		LowStockWarning:    r.LowStockWarning,
		MaintenanceWarning: r.MaintenanceWarning,
	}
}
