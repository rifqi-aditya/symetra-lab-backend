package handler

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"symetra-lab-backend-v2/internal/delivery/http/dto"
	packagingUC "symetra-lab-backend-v2/internal/usecase/packaging"
)

type PackagingHandler struct {
	listPackagingUC   *packagingUC.ListPackagingUseCase
	getPackagingUC    *packagingUC.GetPackagingUseCase
	createPackagingUC *packagingUC.CreatePackagingItemUseCase
	updatePackagingUC *packagingUC.UpdatePackagingItemUseCase
	deletePackagingUC *packagingUC.DeletePackagingItemUseCase

	listPresetsUC   *packagingUC.ListPackagingPresetsUseCase
	createPresetUC  *packagingUC.CreatePresetUseCase
	updatePresetUC  *packagingUC.UpdatePresetUseCase
	deletePresetUC  *packagingUC.DeletePresetUseCase
}

func NewPackagingHandler(
	listPackagingUC *packagingUC.ListPackagingUseCase,
	getPackagingUC *packagingUC.GetPackagingUseCase,
	createPackagingUC *packagingUC.CreatePackagingItemUseCase,
	updatePackagingUC *packagingUC.UpdatePackagingItemUseCase,
	deletePackagingUC *packagingUC.DeletePackagingItemUseCase,
	listPresetsUC *packagingUC.ListPackagingPresetsUseCase,
	createPresetUC *packagingUC.CreatePresetUseCase,
	updatePresetUC *packagingUC.UpdatePresetUseCase,
	deletePresetUC *packagingUC.DeletePresetUseCase,
) *PackagingHandler {
	return &PackagingHandler{
		listPackagingUC:   listPackagingUC,
		getPackagingUC:    getPackagingUC,
		createPackagingUC: createPackagingUC,
		updatePackagingUC: updatePackagingUC,
		deletePackagingUC: deletePackagingUC,
		listPresetsUC:     listPresetsUC,
		createPresetUC:    createPresetUC,
		updatePresetUC:    updatePresetUC,
		deletePresetUC:    deletePresetUC,
	}
}

// Packaging Items
func (h *PackagingHandler) ListPackagingItems(c echo.Context) error {
	userID, err := getUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, dto.Fail("Unauthorized", err.Error()))
	}

	items, err := h.listPackagingUC.Execute(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.Fail("Failed to fetch packaging items", err.Error()))
	}

	resp := make([]dto.PackagingItemResponse, len(items))
	for i, it := range items {
		resp[i] = dto.ToPackagingItemResponse(it)
	}
	return c.JSON(http.StatusOK, dto.Success(resp))
}

func (h *PackagingHandler) GetPackagingItemByID(c echo.Context) error {
	userID, err := getUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, dto.Fail("Unauthorized", err.Error()))
	}

	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail("Invalid packaging item ID", err.Error()))
	}

	item, err := h.getPackagingUC.Execute(c.Request().Context(), userID, id)
	if err != nil {
		return c.JSON(http.StatusNotFound, dto.Fail("Packaging item not found", err.Error()))
	}

	return c.JSON(http.StatusOK, dto.Success(dto.ToPackagingItemResponse(item)))
}

func (h *PackagingHandler) CreatePackagingItem(c echo.Context) error {
	userID, err := getUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, dto.Fail("Unauthorized", err.Error()))
	}

	var req dto.CreatePackagingItemRequest
	if err := c.Bind(&req); err != nil || req.Name == "" {
		return c.JSON(http.StatusBadRequest, dto.Fail("Invalid payload", "Name is required"))
	}

	item, err := h.createPackagingUC.Execute(c.Request().Context(), userID, packagingUC.CreatePackagingItemInput{
		Name:             req.Name,
		Category:         req.Category,
		UnitType:         req.UnitType,
		PurchasePrice:    req.PurchasePrice,
		PurchaseQuantity: req.PurchaseQuantity,
		StockQuantity:    req.StockQuantity,
	})
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail("Failed to create packaging item", err.Error()))
	}

	return c.JSON(http.StatusCreated, dto.Success(dto.ToPackagingItemResponse(item)))
}

func (h *PackagingHandler) UpdatePackagingItem(c echo.Context) error {
	userID, err := getUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, dto.Fail("Unauthorized", err.Error()))
	}

	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail("Invalid packaging item ID", err.Error()))
	}

	var req dto.UpdatePackagingItemRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail("Invalid payload", err.Error()))
	}

	existing, err := h.getPackagingUC.Execute(c.Request().Context(), userID, id)
	if err != nil {
		return c.JSON(http.StatusNotFound, dto.Fail("Packaging item not found", err.Error()))
	}

	name := existing.Name()
	if req.Name != nil && *req.Name != "" {
		name = *req.Name
	}
	cat := existing.Category()
	if req.Category != nil {
		cat = *req.Category
	}
	uType := existing.UnitType()
	if req.UnitType != nil {
		uType = *req.UnitType
	}
	price := existing.PurchasePrice()
	if req.PurchasePrice != nil && *req.PurchasePrice >= 0 {
		price = *req.PurchasePrice
	}
	qty := existing.PurchaseQuantity()
	if req.PurchaseQuantity != nil && *req.PurchaseQuantity > 0 {
		qty = *req.PurchaseQuantity
	}
	stock := existing.StockQuantity()
	if req.StockQuantity != nil && *req.StockQuantity >= 0 {
		stock = *req.StockQuantity
	}

	item, err := h.updatePackagingUC.Execute(c.Request().Context(), userID, packagingUC.UpdatePackagingItemInput{
		ID:               id,
		Name:             name,
		Category:         cat,
		UnitType:         uType,
		PurchasePrice:    price,
		PurchaseQuantity: qty,
		StockQuantity:    stock,
	})
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail("Failed to update packaging item", err.Error()))
	}

	return c.JSON(http.StatusOK, dto.Success(dto.ToPackagingItemResponse(item)))
}

func (h *PackagingHandler) DeletePackagingItem(c echo.Context) error {
	userID, err := getUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, dto.Fail("Unauthorized", err.Error()))
	}

	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail("Invalid packaging item ID", err.Error()))
	}

	force := c.QueryParam("force") == "true"
	if err := h.deletePackagingUC.Execute(c.Request().Context(), userID, id, force); err != nil {
		return c.JSON(http.StatusConflict, dto.Fail("Failed to delete packaging item", err.Error()))
	}

	return c.JSON(http.StatusOK, dto.Success(map[string]string{"message": "Packaging item deleted successfully"}))
}

// Presets
func (h *PackagingHandler) ListPackagingPresets(c echo.Context) error {
	userID, err := getUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, dto.Fail("Unauthorized", err.Error()))
	}

	presets, err := h.listPresetsUC.Execute(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.Fail("Failed to fetch packaging presets", err.Error()))
	}

	resp := make([]dto.PackagingPresetResponse, len(presets))
	for i, p := range presets {
		resp[i] = dto.ToPackagingPresetResponse(p)
	}
	return c.JSON(http.StatusOK, dto.Success(resp))
}

func (h *PackagingHandler) CreatePackagingPreset(c echo.Context) error {
	userID, err := getUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, dto.Fail("Unauthorized", err.Error()))
	}

	var req dto.CreatePresetRequest
	if err := c.Bind(&req); err != nil || req.Name == "" {
		return c.JSON(http.StatusBadRequest, dto.Fail("Invalid payload", "Preset name is required"))
	}

	items := make([]packagingUC.PresetItemInput, len(req.Items))
	for i, it := range req.Items {
		pkgID, _ := uuid.Parse(it.PackagingItemID)
		items[i] = packagingUC.PresetItemInput{
			PackagingItemID: pkgID,
			QuantityUsed:    it.QuantityUsed,
		}
	}

	preset, err := h.createPresetUC.Execute(c.Request().Context(), userID, packagingUC.CreatePresetInput{
		Name:        req.Name,
		Description: req.Description,
		Items:       items,
	})
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail("Failed to create packaging preset", err.Error()))
	}

	return c.JSON(http.StatusCreated, dto.Success(dto.ToPackagingPresetResponse(preset)))
}

func (h *PackagingHandler) UpdatePackagingPreset(c echo.Context) error {
	userID, err := getUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, dto.Fail("Unauthorized", err.Error()))
	}

	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail("Invalid packaging preset ID", err.Error()))
	}

	var req dto.UpdatePresetRequest
	if err := c.Bind(&req); err != nil || req.Name == "" {
		return c.JSON(http.StatusBadRequest, dto.Fail("Invalid payload", "Preset name is required"))
	}

	items := make([]packagingUC.PresetItemInput, len(req.Items))
	for i, it := range req.Items {
		pkgID, _ := uuid.Parse(it.PackagingItemID)
		items[i] = packagingUC.PresetItemInput{
			PackagingItemID: pkgID,
			QuantityUsed:    it.QuantityUsed,
		}
	}

	preset, err := h.updatePresetUC.Execute(c.Request().Context(), userID, packagingUC.UpdatePresetInput{
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
		Items:       items,
	})
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail("Failed to update packaging preset", err.Error()))
	}

	return c.JSON(http.StatusOK, dto.Success(dto.ToPackagingPresetResponse(preset)))
}

func (h *PackagingHandler) DeletePackagingPreset(c echo.Context) error {
	userID, err := getUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, dto.Fail("Unauthorized", err.Error()))
	}

	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail("Invalid packaging preset ID", err.Error()))
	}

	force := c.QueryParam("force") == "true"
	if err := h.deletePresetUC.Execute(c.Request().Context(), userID, id, force); err != nil {
		return c.JSON(http.StatusConflict, dto.Fail("Failed to delete packaging preset", err.Error()))
	}

	return c.JSON(http.StatusOK, dto.Success(map[string]string{"message": "Packaging preset deleted successfully"}))
}