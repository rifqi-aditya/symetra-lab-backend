package handler

import (
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"symetra-lab-backend-v2/internal/delivery/http/dto"
	machineUC "symetra-lab-backend-v2/internal/usecase/machine"
)

type MachineHandler struct {
	listUC        *machineUC.ListMachinesUseCase
	getUC         *machineUC.GetMachineUseCase
	createUC      *machineUC.CreateMachineUseCase
	updateUC      *machineUC.UpdateMachineUseCase
	updateStateUC *machineUC.UpdateMachineStateUseCase
	deleteUC      *machineUC.DeleteMachineUseCase
	listPartsUC   *machineUC.ListPartsUseCase
	createPartUC  *machineUC.CreatePartUseCase
	updatePartUC  *machineUC.UpdatePartUseCase
	replacePartUC *machineUC.ReplacePartUseCase
	deletePartUC  *machineUC.DeletePartUseCase
}

func NewMachineHandler(
	listUC *machineUC.ListMachinesUseCase,
	getUC *machineUC.GetMachineUseCase,
	createUC *machineUC.CreateMachineUseCase,
	updateUC *machineUC.UpdateMachineUseCase,
	updateStateUC *machineUC.UpdateMachineStateUseCase,
	deleteUC *machineUC.DeleteMachineUseCase,
	listPartsUC *machineUC.ListPartsUseCase,
	createPartUC *machineUC.CreatePartUseCase,
	updatePartUC *machineUC.UpdatePartUseCase,
	replacePartUC *machineUC.ReplacePartUseCase,
	deletePartUC *machineUC.DeletePartUseCase,
) *MachineHandler {
	return &MachineHandler{
		listUC:        listUC,
		getUC:         getUC,
		createUC:      createUC,
		updateUC:      updateUC,
		updateStateUC: updateStateUC,
		deleteUC:      deleteUC,
		listPartsUC:   listPartsUC,
		createPartUC:  createPartUC,
		updatePartUC:  updatePartUC,
		replacePartUC: replacePartUC,
		deletePartUC:  deletePartUC,
	}
}

func (h *MachineHandler) ListMachines(c echo.Context) error {
	userID, err := getUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, dto.Fail("Unauthorized", err.Error()))
	}

	machines, err := h.listUC.Execute(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.Fail("Failed to fetch machines", err.Error()))
	}

	filterState := strings.ToUpper(strings.TrimSpace(c.QueryParam("state")))
	resp := make([]dto.MachineResponse, 0, len(machines))
	for _, m := range machines {
		if filterState != "" && strings.ToUpper(m.CurrentState()) != filterState {
			continue
		}
		resp = append(resp, dto.ToMachineResponse(m))
	}

	return c.JSON(http.StatusOK, dto.Success(resp))
}

func (h *MachineHandler) GetMachineByID(c echo.Context) error {
	userID, err := getUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, dto.Fail("Unauthorized", err.Error()))
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail("Invalid machine ID", err.Error()))
	}

	m, err := h.getUC.Execute(c.Request().Context(), userID, id)
	if err != nil {
		return c.JSON(http.StatusNotFound, dto.Fail("Machine not found", err.Error()))
	}

	return c.JSON(http.StatusOK, dto.Success(dto.ToMachineResponse(m)))
}

func (h *MachineHandler) Create(c echo.Context) error {
	userID, err := getUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, dto.Fail("Unauthorized", err.Error()))
	}

	var req dto.CreateMachineRequest
	if err := c.Bind(&req); err != nil || req.Name == "" {
		return c.JSON(http.StatusBadRequest, dto.Fail("Invalid payload", "Machine name is required"))
	}

	lifespan := 5000
	if req.LifespanHours > 0 {
		lifespan = req.LifespanHours
	}

	m, err := h.createUC.Execute(c.Request().Context(), userID, machineUC.CreateMachineInput{
		Name:              req.Name,
		Brand:             req.Brand,
		TotalPurchaseCost: req.TotalPurchaseCost,
		LifespanHours:     lifespan,
		AvgPowerWatts:     req.AvgPowerWatts,
	})
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail("Failed to create machine", err.Error()))
	}

	return c.JSON(http.StatusCreated, dto.Success(dto.ToMachineResponse(m)))
}

func (h *MachineHandler) Update(c echo.Context) error {
	userID, err := getUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, dto.Fail("Unauthorized", err.Error()))
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail("Invalid machine ID", err.Error()))
	}

	existing, err := h.getUC.Execute(c.Request().Context(), userID, id)
	if err != nil {
		return c.JSON(http.StatusNotFound, dto.Fail("Machine not found", err.Error()))
	}

	var req dto.UpdateMachineRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail("Invalid payload", err.Error()))
	}

	name := existing.Name()
	if req.Name != nil && *req.Name != "" {
		name = *req.Name
	}
	brand := existing.Brand()
	if req.Brand != nil {
		brand = req.Brand
	}
	cost := existing.TotalPurchaseCost()
	if req.TotalPurchaseCost != nil && *req.TotalPurchaseCost >= 0 {
		cost = *req.TotalPurchaseCost
	}
	salvage := existing.SalvageValue()
	if req.SalvageValue != nil {
		salvage = req.SalvageValue
	}
	lifespan := existing.LifespanHours()
	if req.LifespanHours != nil && *req.LifespanHours > 0 {
		lifespan = *req.LifespanHours
	}
	purchaseDate := existing.PurchaseDate()
	if req.PurchaseDate != nil {
		purchaseDate = req.PurchaseDate
	}
	hoursUsed := existing.TotalHoursUsed()
	if req.TotalHoursUsed != nil && *req.TotalHoursUsed >= 0 {
		hoursUsed = *req.TotalHoursUsed
	}
	power := existing.AvgPowerWatts()
	if req.AvgPowerWatts != nil && *req.AvgPowerWatts >= 0 {
		power = *req.AvgPowerWatts
	}
	buffer := existing.MaintenanceBufferPerHour()
	if req.MaintenanceBufferPerHour != nil {
		buffer = req.MaintenanceBufferPerHour
	}
	failure := existing.FailureRatePercent()
	if req.FailureRatePercent != nil {
		failure = req.FailureRatePercent
	}
	bvx := existing.BuildVolumeX()
	if req.BuildVolumeX != nil {
		bvx = req.BuildVolumeX
	}
	bvy := existing.BuildVolumeY()
	if req.BuildVolumeY != nil {
		bvy = req.BuildVolumeY
	}
	bvz := existing.BuildVolumeZ()
	if req.BuildVolumeZ != nil {
		bvz = req.BuildVolumeZ
	}
	speed := existing.SpeedMultiplier()
	if req.SpeedMultiplier != nil {
		speed = req.SpeedMultiplier
	}
	state := existing.CurrentState()
	if req.CurrentState != nil && *req.CurrentState != "" {
		state = *req.CurrentState
	}
	elecCost := existing.ElectricityCostPerHour()
	if req.ElectricityCostPerHour != nil {
		elecCost = req.ElectricityCostPerHour
	}

	m, err := h.updateUC.Execute(c.Request().Context(), userID, machineUC.UpdateMachineInput{
		ID:                       id,
		Name:                     name,
		Brand:                    brand,
		TotalPurchaseCost:        cost,
		SalvageValue:             salvage,
		LifespanHours:            lifespan,
		PurchaseDate:             purchaseDate,
		TotalHoursUsed:           hoursUsed,
		AvgPowerWatts:            power,
		MaintenanceBufferPerHour: buffer,
		FailureRatePercent:       failure,
		BuildVolumeX:             bvx,
		BuildVolumeY:             bvy,
		BuildVolumeZ:             bvz,
		SpeedMultiplier:          speed,
		CurrentState:             state,
		ElectricityCostPerHour:   elecCost,
	})
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail("Failed to update machine", err.Error()))
	}

	return c.JSON(http.StatusOK, dto.Success(dto.ToMachineResponse(m)))
}

func (h *MachineHandler) UpdateState(c echo.Context) error {
	userID, err := getUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, dto.Fail("Unauthorized", err.Error()))
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail("Invalid machine ID", err.Error()))
	}

	var req dto.UpdateMachineStateRequest
	if err := c.Bind(&req); err != nil || req.State == "" {
		return c.JSON(http.StatusBadRequest, dto.Fail("Invalid payload", "State is required"))
	}

	if err := h.updateStateUC.Execute(c.Request().Context(), userID, id, req.State); err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail("Failed to update machine state", err.Error()))
	}

	return c.JSON(http.StatusOK, dto.Success(map[string]string{
		"id":    id.String(),
		"state": req.State,
	}))
}

func (h *MachineHandler) Delete(c echo.Context) error {
	userID, err := getUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, dto.Fail("Unauthorized", err.Error()))
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail("Invalid machine ID", err.Error()))
	}

	if err := h.deleteUC.Execute(c.Request().Context(), userID, id); err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail("Failed to delete machine", err.Error()))
	}

	return c.JSON(http.StatusOK, dto.Success(map[string]string{
		"message": "Machine deleted successfully",
	}))
}

func (h *MachineHandler) GetPartsByMachine(c echo.Context) error {
	machineID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail("Invalid machine ID", err.Error()))
	}

	parts, err := h.listPartsUC.Execute(c.Request().Context(), machineID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.Fail("Failed to fetch parts", err.Error()))
	}

	resp := make([]dto.MachineMaintenancePartResponse, len(parts))
	for i, p := range parts {
		resp[i] = dto.ToPartResponse(p)
	}

	return c.JSON(http.StatusOK, dto.Success(resp))
}

func (h *MachineHandler) CreatePart(c echo.Context) error {
	machineID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail("Invalid machine ID", err.Error()))
	}

	var req dto.CreatePartRequest
	if err := c.Bind(&req); err != nil || req.PartName == "" {
		return c.JSON(http.StatusBadRequest, dto.Fail("Invalid payload", "Part name is required"))
	}

	lifespan := 500.0
	if req.LifespanHours > 0 {
		lifespan = req.LifespanHours
	}

	part, err := h.createPartUC.Execute(c.Request().Context(), machineUC.CreatePartInput{
		MachineID:     machineID,
		PartName:      req.PartName,
		CostIDR:       req.CostIDR,
		LifespanHours: lifespan,
		StockQuantity: req.StockQuantity,
	})
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail("Failed to create part", err.Error()))
	}

	return c.JSON(http.StatusCreated, dto.Success(dto.ToPartResponse(part)))
}

func (h *MachineHandler) UpdatePart(c echo.Context) error {
	partID, err := uuid.Parse(c.Param("part_id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail("Invalid part ID", err.Error()))
	}

	var req dto.UpdatePartRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail("Invalid payload", err.Error()))
	}

	name := ""
	if req.PartName != nil {
		name = *req.PartName
	}
	cost := 0.0
	if req.CostIDR != nil {
		cost = *req.CostIDR
	}
	lifespan := 0.0
	if req.LifespanHours != nil {
		lifespan = *req.LifespanHours
	}
	stock := 0
	if req.StockQuantity != nil {
		stock = *req.StockQuantity
	}

	part, err := h.updatePartUC.Execute(c.Request().Context(), machineUC.UpdatePartInput{
		ID:            partID,
		PartName:      name,
		CostIDR:       cost,
		LifespanHours: lifespan,
		StockQuantity: stock,
	})
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail("Failed to update part", err.Error()))
	}

	return c.JSON(http.StatusOK, dto.Success(dto.ToPartResponse(part)))
}

func (h *MachineHandler) ReplacePart(c echo.Context) error {
	partID, err := uuid.Parse(c.Param("part_id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail("Invalid part ID", err.Error()))
	}

	var req dto.ReplacePartRequest
	_ = c.Bind(&req)

	cost := 0.0
	if req.CostIDR != nil {
		cost = *req.CostIDR
	}

	part, err := h.replacePartUC.Execute(c.Request().Context(), partID, cost)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail("Failed to replace part", err.Error()))
	}

	return c.JSON(http.StatusOK, dto.Success(dto.ToPartResponse(part)))
}

func (h *MachineHandler) DeletePart(c echo.Context) error {
	partID, err := uuid.Parse(c.Param("part_id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail("Invalid part ID", err.Error()))
	}

	if err := h.deletePartUC.Execute(c.Request().Context(), partID); err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail("Failed to delete part", err.Error()))
	}

	return c.JSON(http.StatusOK, dto.Success(map[string]string{
		"message": "Part deleted successfully",
	}))
}
