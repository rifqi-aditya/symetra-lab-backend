package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"symetra-lab-backend-v2/internal/delivery/http/dto"
	dashboardUC "symetra-lab-backend-v2/internal/usecase/dashboard"
)

type DashboardHandler struct {
	getOverviewUC *dashboardUC.GetDashboardOverviewUseCase
}

func NewDashboardHandler(getOverviewUC *dashboardUC.GetDashboardOverviewUseCase) *DashboardHandler {
	return &DashboardHandler{
		getOverviewUC: getOverviewUC,
	}
}

func (h *DashboardHandler) GetOverview(c echo.Context) error {
	overview, err := h.getOverviewUC.Execute(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.Fail("Failed to retrieve dashboard overview", err.Error()))
	}

	return c.JSON(http.StatusOK, dto.Success(dto.ToDashboardOverviewResponse(overview)))
}
