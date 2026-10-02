package handler

import (
	"math"
	"net/http"

	"github.com/labstack/echo/v4"

	"symetra-lab-backend-v2/internal/delivery/http/dto"
	productionUC "symetra-lab-backend-v2/internal/usecase/production"
)

type ProductionHandler struct {
	getQueueUC    *productionUC.GetProductionQueueUseCase
	completeJobUC *productionUC.CompleteJobUseCase
}

func NewProductionHandler(
	getQueueUC *productionUC.GetProductionQueueUseCase,
	completeJobUC *productionUC.CompleteJobUseCase,
) *ProductionHandler {
	return &ProductionHandler{
		getQueueUC:    getQueueUC,
		completeJobUC: completeJobUC,
	}
}

func (h *ProductionHandler) GetQueue(c echo.Context) error {
	queue, err := h.getQueueUC.Execute()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.Fail("Gagal mengambil antrean produksi: "+err.Error(), err.Error()))
	}

	var totalPrintHours float64
	urgentCount := 0
	items := make([]dto.ProductionQueueItemResponse, len(queue))
	for i, q := range queue {
		items[i] = dto.ToProductionQueueItemResponse(q)
		totalPrintHours += q.TotalPrintHours
		if q.IsUrgent {
			urgentCount++
		}
	}

	resp := dto.ProductionQueueSummaryResponse{
		Queue:             items,
		TotalJobs:         len(items),
		TotalHoursWaiting: math.Round(totalPrintHours*100) / 100,
		UrgentJobsCount:   urgentCount,
	}

	return c.JSON(http.StatusOK, dto.Success(resp))
}

func (h *ProductionHandler) CompleteJob(c echo.Context) error {
	var req dto.CompletePrintJobRequest
	if err := c.Bind(&req); err != nil || req.Source == "" || req.JobID == "" {
		return c.JSON(http.StatusBadRequest, dto.Fail("Validasi input gagal: source dan job_id wajib diisi", "Invalid input"))
	}

	res, err := h.completeJobUC.Execute(req.Source, req.JobID, req.MachineID, req.FilamentID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail(err.Error(), err.Error()))
	}

	return c.JSON(http.StatusOK, dto.Success(dto.ToCompletePrintJobResponse(res)))
}
