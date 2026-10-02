package handler

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"symetra-lab-backend-v2/internal/delivery/http/dto"
	componentUC "symetra-lab-backend-v2/internal/usecase/component"
)

type ComponentHandler struct {
	listUC   *componentUC.ListComponentsUseCase
	getUC    *componentUC.GetComponentUseCase
	createUC *componentUC.CreateComponentUseCase
	updateUC *componentUC.UpdateComponentUseCase
	deleteUC *componentUC.DeleteComponentUseCase
}

func NewComponentHandler(
	listUC *componentUC.ListComponentsUseCase,
	getUC *componentUC.GetComponentUseCase,
	createUC *componentUC.CreateComponentUseCase,
	updateUC *componentUC.UpdateComponentUseCase,
	deleteUC *componentUC.DeleteComponentUseCase,
) *ComponentHandler {
	return &ComponentHandler{
		listUC:   listUC,
		getUC:    getUC,
		createUC: createUC,
		updateUC: updateUC,
		deleteUC: deleteUC,
	}
}

func (h *ComponentHandler) ListComponents(c echo.Context) error {
	userID, err := getUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, dto.Fail("Unauthorized", err.Error()))
	}

	items, err := h.listUC.Execute(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.Fail("Failed to fetch components", err.Error()))
	}

	resp := make([]dto.ComponentResponse, len(items))
	for i, comp := range items {
		resp[i] = dto.ToComponentResponse(comp)
	}
	return c.JSON(http.StatusOK, dto.Success(resp))
}

func (h *ComponentHandler) GetComponentByID(c echo.Context) error {
	userID, err := getUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, dto.Fail("Unauthorized", err.Error()))
	}

	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail("Invalid component ID", err.Error()))
	}

	item, err := h.getUC.Execute(c.Request().Context(), userID, id)
	if err != nil {
		return c.JSON(http.StatusNotFound, dto.Fail("Component not found", err.Error()))
	}

	return c.JSON(http.StatusOK, dto.Success(dto.ToComponentResponse(item)))
}

func (h *ComponentHandler) Create(c echo.Context) error {
	userID, err := getUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, dto.Fail("Unauthorized", err.Error()))
	}

	var req dto.CreateComponentRequest
	if err := c.Bind(&req); err != nil || req.Name == "" {
		return c.JSON(http.StatusBadRequest, dto.Fail("Invalid payload", "Component name is required"))
	}

	markup := 0.0
	if req.DefaultMarkupPercent != nil {
		markup = *req.DefaultMarkupPercent
	}

	comp, err := h.createUC.Execute(c.Request().Context(), userID, componentUC.CreateComponentInput{
		Name:                 req.Name,
		PricePerUnit:         req.PricePerUnit,
		DefaultMarkupPercent: markup,
		Description:          req.Description,
	})
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail("Failed to create component", err.Error()))
	}

	return c.JSON(http.StatusCreated, dto.Success(dto.ToComponentResponse(comp)))
}

func (h *ComponentHandler) Update(c echo.Context) error {
	userID, err := getUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, dto.Fail("Unauthorized", err.Error()))
	}

	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail("Invalid component ID", err.Error()))
	}

	var req dto.UpdateComponentRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail("Invalid payload", err.Error()))
	}

	existing, err := h.getUC.Execute(c.Request().Context(), userID, id)
	if err != nil {
		return c.JSON(http.StatusNotFound, dto.Fail("Component not found", err.Error()))
	}

	name := existing.Name()
	if req.Name != nil && *req.Name != "" {
		name = *req.Name
	}
	price := existing.PricePerUnit()
	if req.PricePerUnit != nil && *req.PricePerUnit >= 0 {
		price = *req.PricePerUnit
	}
	markup := existing.DefaultMarkupPercent()
	if req.DefaultMarkupPercent != nil {
		markup = *req.DefaultMarkupPercent
	}
	desc := existing.Description()
	if req.Description != nil {
		desc = req.Description
	}

	comp, err := h.updateUC.Execute(c.Request().Context(), userID, componentUC.UpdateComponentInput{
		ID:                   id,
		Name:                 name,
		PricePerUnit:         price,
		DefaultMarkupPercent: markup,
		Description:          desc,
	})
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail("Failed to update component", err.Error()))
	}

	return c.JSON(http.StatusOK, dto.Success(dto.ToComponentResponse(comp)))
}

func (h *ComponentHandler) Delete(c echo.Context) error {
	userID, err := getUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, dto.Fail("Unauthorized", err.Error()))
	}

	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail("Invalid component ID", err.Error()))
	}

	force := c.QueryParam("force") == "true"
	if err := h.deleteUC.Execute(c.Request().Context(), userID, id, force); err != nil {
		return c.JSON(http.StatusConflict, dto.Fail("Failed to delete component", err.Error()))
	}

	return c.JSON(http.StatusOK, dto.Success(map[string]string{
		"message": "Component successfully deleted",
	}))
}