package handler

import (
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"symetra-lab-backend-v2/internal/delivery/http/dto"
	filamentUC "symetra-lab-backend-v2/internal/usecase/filament"
)

type FilamentHandler struct {
	listFilamentsUC *filamentUC.ListFilamentsUseCase
	getFilamentUC   *filamentUC.GetFilamentUseCase
	createUC        *filamentUC.CreateFilamentUseCase
	updateUC        *filamentUC.UpdateFilamentUseCase
	deleteUC        *filamentUC.DeleteFilamentUseCase
	syncStockUC     *filamentUC.SyncStockUseCase

	listProfilesUC *filamentUC.ListFilamentProfilesUseCase

	listRatesUC   *filamentUC.ListMaterialRatesUseCase
	createRateUC  *filamentUC.CreateMaterialRateUseCase
	updateRateUC  *filamentUC.UpdateMaterialRateUseCase
	deleteRateUC  *filamentUC.DeleteMaterialRateUseCase
}

func NewFilamentHandler(
	listFilamentsUC *filamentUC.ListFilamentsUseCase,
	getFilamentUC *filamentUC.GetFilamentUseCase,
	createUC *filamentUC.CreateFilamentUseCase,
	updateUC *filamentUC.UpdateFilamentUseCase,
	deleteUC *filamentUC.DeleteFilamentUseCase,
	syncStockUC *filamentUC.SyncStockUseCase,
	listProfilesUC *filamentUC.ListFilamentProfilesUseCase,
	listRatesUC *filamentUC.ListMaterialRatesUseCase,
	createRateUC *filamentUC.CreateMaterialRateUseCase,
	updateRateUC *filamentUC.UpdateMaterialRateUseCase,
	deleteRateUC *filamentUC.DeleteMaterialRateUseCase,
) *FilamentHandler {
	return &FilamentHandler{
		listFilamentsUC: listFilamentsUC,
		getFilamentUC:   getFilamentUC,
		createUC:        createUC,
		updateUC:        updateUC,
		deleteUC:        deleteUC,
		syncStockUC:     syncStockUC,
		listProfilesUC:  listProfilesUC,
		listRatesUC:     listRatesUC,
		createRateUC:    createRateUC,
		updateRateUC:    updateRateUC,
		deleteRateUC:    deleteRateUC,
	}
}

func (h *FilamentHandler) ListFilaments(c echo.Context) error {
	userID, err := getUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, dto.Fail("Unauthorized", err.Error()))
	}

	items, err := h.listFilamentsUC.Execute(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.Fail("Failed to fetch filaments", err.Error()))
	}

	resp := make([]dto.FilamentResponse, len(items))
	for i, f := range items {
		resp[i] = dto.ToFilamentResponse(f)
	}
	return c.JSON(http.StatusOK, dto.Success(resp))
}

func (h *FilamentHandler) GetFilamentByID(c echo.Context) error {
	userID, err := getUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, dto.Fail("Unauthorized", err.Error()))
	}

	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail("Invalid filament ID", err.Error()))
	}

	item, err := h.getFilamentUC.Execute(c.Request().Context(), userID, id)
	if err != nil {
		return c.JSON(http.StatusNotFound, dto.Fail("Filament not found", err.Error()))
	}

	return c.JSON(http.StatusOK, dto.Success(dto.ToFilamentResponse(item)))
}

func (h *FilamentHandler) Create(c echo.Context) error {
	userID, err := getUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, dto.Fail("Unauthorized", err.Error()))
	}

	var req dto.CreateFilamentRequest
	if err := c.Bind(&req); err != nil || req.ColorName == "" {
		return c.JSON(http.StatusBadRequest, dto.Fail("Invalid payload", "Color name is required"))
	}

	var profID *uuid.UUID
	if req.ProfileID != nil && strings.TrimSpace(*req.ProfileID) != "" {
		if pID, err := uuid.Parse(strings.TrimSpace(*req.ProfileID)); err == nil {
			profID = &pID
		}
	}

	brand := "Generic"
	if req.Brand != nil && *req.Brand != "" {
		brand = *req.Brand
	}
	mat := "PLA"
	if req.MaterialType != nil && *req.MaterialType != "" {
		mat = *req.MaterialType
	}
	stock := 1000.0
	if req.CurrentStockGrams != nil {
		stock = *req.CurrentStockGrams
	} else if req.RemainingWeightGrams != nil {
		stock = *req.RemainingWeightGrams
	}
	spoolWeight := 1000.0
	if req.SpoolWeightGrams != nil && *req.SpoolWeightGrams > 0 {
		spoolWeight = *req.SpoolWeightGrams
	}
	threshold := 200.0
	if req.LowStockThresholdGrams != nil && *req.LowStockThresholdGrams > 0 {
		threshold = *req.LowStockThresholdGrams
	}

	f, err := h.createUC.Execute(c.Request().Context(), userID, filamentUC.CreateFilamentInput{
		Brand:                  brand,
		MaterialType:           mat,
		SpoolWeightGrams:       spoolWeight,
		ProfileID:              profID,
		ColorName:              req.ColorName,
		ColorHex:               req.ColorHex,
		SKU:                    req.SKU,
		PricePerRoll:           req.PricePerRoll,
		CurrentStockGrams:      stock,
		LowStockThresholdGrams: threshold,
	})
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail("Failed to create filament", err.Error()))
	}

	return c.JSON(http.StatusCreated, dto.Success(dto.ToFilamentResponse(f)))
}

func (h *FilamentHandler) Update(c echo.Context) error {
	userID, err := getUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, dto.Fail("Unauthorized", err.Error()))
	}

	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail("Invalid filament ID", err.Error()))
	}

	var req dto.UpdateFilamentRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail("Invalid payload", err.Error()))
	}

	existing, err := h.getFilamentUC.Execute(c.Request().Context(), userID, id)
	if err != nil {
		return c.JSON(http.StatusNotFound, dto.Fail("Filament not found", err.Error()))
	}

	profID := existing.ProfileID()
	if req.ProfileID != nil {
		if strings.TrimSpace(*req.ProfileID) == "" {
			profID = nil
		} else if pID, err := uuid.Parse(strings.TrimSpace(*req.ProfileID)); err == nil {
			profID = &pID
		}
	}

	colorName := existing.ColorName()
	if req.ColorName != nil && *req.ColorName != "" {
		colorName = *req.ColorName
	}
	colorHex := existing.ColorHex()
	if req.ColorHex != nil && *req.ColorHex != "" {
		colorHex = *req.ColorHex
	}
	sku := existing.SKU()
	if req.SKU != nil {
		sku = req.SKU
	}
	price := existing.PricePerRoll()
	if req.PricePerRoll != nil && *req.PricePerRoll >= 0 {
		price = *req.PricePerRoll
	}
	stock := existing.CurrentStockGrams()
	if req.CurrentStockGrams != nil && *req.CurrentStockGrams >= 0 {
		stock = *req.CurrentStockGrams
	}
	threshold := existing.LowStockThresholdGrams()
	if req.LowStockThresholdGrams != nil && *req.LowStockThresholdGrams >= 0 {
		threshold = *req.LowStockThresholdGrams
	}

	f, err := h.updateUC.Execute(c.Request().Context(), userID, filamentUC.UpdateFilamentInput{
		ID:                     id,
		ProfileID:              profID,
		ColorName:              colorName,
		ColorHex:               colorHex,
		SKU:                    sku,
		PricePerRoll:           price,
		CurrentStockGrams:      stock,
		LowStockThresholdGrams: threshold,
	})
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail("Failed to update filament", err.Error()))
	}

	return c.JSON(http.StatusOK, dto.Success(dto.ToFilamentResponse(f)))
}

func (h *FilamentHandler) Delete(c echo.Context) error {
	userID, err := getUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, dto.Fail("Unauthorized", err.Error()))
	}

	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail("Invalid filament ID", err.Error()))
	}

	if err := h.deleteUC.Execute(c.Request().Context(), userID, id); err != nil {
		return c.JSON(http.StatusInternalServerError, dto.Fail("Failed to delete filament", err.Error()))
	}

	return c.JSON(http.StatusOK, dto.Success(map[string]string{"message": "Filament deleted successfully"}))
}

func (h *FilamentHandler) SyncStock(c echo.Context) error {
	userID, err := getUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, dto.Fail("Unauthorized", err.Error()))
	}

	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail("Invalid filament ID", err.Error()))
	}

	var req dto.SyncStockRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail("Invalid payload", err.Error()))
	}

	stockGrams := 0.0
	isWeighed := req.IsWeighed
	if req.CurrentStockGrams != nil {
		stockGrams = *req.CurrentStockGrams
	} else if req.RemainingWeightGrams != nil {
		stockGrams = *req.RemainingWeightGrams
	} else if req.GrossWeightGrams != nil {
		isWeighed = true
		stockGrams = *req.GrossWeightGrams
	}

	f, err := h.syncStockUC.Execute(c.Request().Context(), userID, id, stockGrams, isWeighed)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail("Failed to sync stock", err.Error()))
	}

	return c.JSON(http.StatusOK, dto.Success(dto.ToFilamentResponse(f)))
}

func (h *FilamentHandler) ListFilamentProfiles(c echo.Context) error {
	userID, err := getUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, dto.Fail("Unauthorized", err.Error()))
	}

	profiles, err := h.listProfilesUC.Execute(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.Fail("Failed to fetch filament profiles", err.Error()))
	}

	resp := make([]dto.FilamentProfileResponse, len(profiles))
	for i, p := range profiles {
		resp[i] = dto.ToFilamentProfileResponse(p)
	}
	return c.JSON(http.StatusOK, dto.Success(resp))
}

// Material Rates
func (h *FilamentHandler) ListMaterialRates(c echo.Context) error {
	userID, err := getUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, dto.Fail("Unauthorized", err.Error()))
	}

	rates, err := h.listRatesUC.Execute(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.Fail("Failed to fetch material rates", err.Error()))
	}

	resp := make([]dto.FilamentMaterialRateResponse, len(rates))
	for i, r := range rates {
		resp[i] = dto.ToMaterialRateResponse(r)
	}
	return c.JSON(http.StatusOK, dto.Success(resp))
}

func (h *FilamentHandler) CreateMaterialRate(c echo.Context) error {
	userID, err := getUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, dto.Fail("Unauthorized", err.Error()))
	}

	var req dto.CreateMaterialRateRequest
	if err := c.Bind(&req); err != nil || req.MaterialType == "" {
		return c.JSON(http.StatusBadRequest, dto.Fail("Invalid payload", "Material type is required"))
	}

	rate, err := h.createRateUC.Execute(c.Request().Context(), userID, filamentUC.CreateMaterialRateInput{
		MaterialType: req.MaterialType,
		PricePerGram: req.PricePerGram,
		IsDefault:    req.IsDefault,
		Description:  req.Description,
	})
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail("Failed to create material rate", err.Error()))
	}

	return c.JSON(http.StatusCreated, dto.Success(dto.ToMaterialRateResponse(rate)))
}

func (h *FilamentHandler) UpdateMaterialRate(c echo.Context) error {
	userID, err := getUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, dto.Fail("Unauthorized", err.Error()))
	}

	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail("Invalid material rate ID", err.Error()))
	}

	var req dto.UpdateMaterialRateRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail("Invalid payload", err.Error()))
	}

	rate, err := h.updateRateUC.Execute(c.Request().Context(), userID, filamentUC.UpdateMaterialRateInput{
		ID:           id,
		PricePerGram: req.PricePerGram,
		IsDefault:    req.IsDefault,
		Description:  req.Description,
	})
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail("Failed to update material rate", err.Error()))
	}

	return c.JSON(http.StatusOK, dto.Success(dto.ToMaterialRateResponse(rate)))
}

func (h *FilamentHandler) DeleteMaterialRate(c echo.Context) error {
	userID, err := getUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, dto.Fail("Unauthorized", err.Error()))
	}

	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail("Invalid material rate ID", err.Error()))
	}

	if err := h.deleteRateUC.Execute(c.Request().Context(), userID, id); err != nil {
		return c.JSON(http.StatusInternalServerError, dto.Fail("Failed to delete material rate", err.Error()))
	}

	return c.JSON(http.StatusOK, dto.Success(map[string]string{"message": "Material rate deleted successfully"}))
}