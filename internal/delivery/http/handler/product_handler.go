package handler

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"symetra-lab-backend-v2/internal/delivery/http/dto"
	"symetra-lab-backend-v2/internal/domain/product"
	productUC "symetra-lab-backend-v2/internal/usecase/product"
)

type ProductHandler struct {
	listUC       *productUC.ListProductsUseCase
	getUC        *productUC.GetProductUseCase
	createUC     *productUC.CreateProductUseCase
	updateUC     *productUC.UpdateProductUseCase
	deleteUC     *productUC.DeleteProductUseCase
	autoGenSKUUC *productUC.AutoGenerateSKUsUseCase
}

func NewProductHandler(
	listUC *productUC.ListProductsUseCase,
	getUC *productUC.GetProductUseCase,
	createUC *productUC.CreateProductUseCase,
	updateUC *productUC.UpdateProductUseCase,
	deleteUC *productUC.DeleteProductUseCase,
	autoGenSKUUC *productUC.AutoGenerateSKUsUseCase,
) *ProductHandler {
	return &ProductHandler{
		listUC:       listUC,
		getUC:        getUC,
		createUC:     createUC,
		updateUC:     updateUC,
		deleteUC:     deleteUC,
		autoGenSKUUC: autoGenSKUUC,
	}
}

func getUserID(c echo.Context) (uuid.UUID, error) {
	if ctxVal := c.Get("user_id"); ctxVal != nil {
		if id, ok := ctxVal.(uuid.UUID); ok && id != uuid.Nil {
			return id, nil
		}
		if s, ok := ctxVal.(string); ok && s != "" {
			return uuid.Parse(s)
		}
	}

	val := c.Request().Header.Get("X-User-ID")
	if val != "" {
		return uuid.Parse(val)
	}

	return uuid.Nil, errors.New("user_id not found")
}

func (h *ProductHandler) List(c echo.Context) error {
	userID, err := getUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, dto.Fail("Unauthorized", err.Error()))
	}

	page, _ := strconv.Atoi(c.QueryParam("page"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	isPaginated := page > 0 || limit > 0
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 25
	}
	if limit > 100 {
		limit = 100
	}

	filter := product.Filter{
		Search:   c.QueryParam("search"),
		Category: c.QueryParam("category"),
	}
	if isPaginated {
		filter.Limit = limit
		filter.Offset = (page - 1) * limit
	}

	output, err := h.listUC.Execute(c.Request().Context(), userID, filter)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.Fail("Failed to retrieve products", err.Error()))
	}

	items := make([]dto.ProductResponse, len(output.Items))
	for i, item := range output.Items {
		items[i] = dto.ToProductResponse(item.Product, item.TotalSold, item.CostBreakdown)
	}

	if isPaginated {
		totalPages := 0
		if output.TotalCount > 0 {
			totalPages = int(math.Ceil(float64(output.TotalCount) / float64(limit)))
		}
		return c.JSON(http.StatusOK, dto.SuccessPaginated(items, output.TotalCount, page, limit, totalPages))
	}

	return c.JSON(http.StatusOK, dto.Success(items))
}

func (h *ProductHandler) GetByID(c echo.Context) error {
	userID, err := getUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, dto.Fail("Unauthorized", err.Error()))
	}

	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail("Invalid product ID", err.Error()))
	}

	detail, err := h.getUC.Execute(c.Request().Context(), userID, id)
	if err != nil {
		return c.JSON(http.StatusNotFound, dto.Fail("Product not found", err.Error()))
	}

	return c.JSON(http.StatusOK, dto.Success(dto.ToProductResponse(detail.Product, detail.TotalSold, detail.CostBreakdown)))
}

func (h *ProductHandler) GetBySKU(c echo.Context) error {
	userID, err := getUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, dto.Fail("Unauthorized", err.Error()))
	}

	sku := c.Param("sku")
	detail, err := h.getUC.ExecuteBySKU(c.Request().Context(), userID, sku)
	if err != nil {
		return c.JSON(http.StatusNotFound, dto.Fail("Product not found", err.Error()))
	}

	return c.JSON(http.StatusOK, dto.Success(dto.ToProductResponse(detail.Product, detail.TotalSold, detail.CostBreakdown)))
}

func parseUUIDPtr(val *string) *uuid.UUID {
	if val == nil {
		return nil
	}
	s := strings.TrimSpace(*val)
	if s == "" || strings.EqualFold(s, "null") || strings.EqualFold(s, "none") {
		return nil
	}
	if parsed, err := uuid.Parse(s); err == nil {
		return &parsed
	}
	return nil
}

func (h *ProductHandler) Create(c echo.Context) error {
	userID, err := getUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, dto.Fail("Unauthorized", err.Error()))
	}

	var req dto.CreateProductRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail("Invalid payload", err.Error()))
	}

	comps := make([]productUC.ComponentInput, len(req.Components))
	for i, comp := range req.Components {
		comps[i] = productUC.ComponentInput{
			ComponentID:   parseUUIDPtr(comp.ComponentID),
			Quantity:      comp.Quantity,
			MarkupPercent: comp.MarkupPercent,
		}
	}

	packs := make([]productUC.PackagingItemInput, len(req.PackagingItems))
	for i, pack := range req.PackagingItems {
		pkgID, _ := uuid.Parse(pack.PackagingItemID)
		packs[i] = productUC.PackagingItemInput{
			PackagingItemID: pkgID,
			QuantityUsed:    pack.QuantityUsed,
		}
	}

	input := productUC.CreateProductInput{
		Name:                  req.Name,
		ParentSKU:             req.ParentSKU,
		SKU:                   req.SKU,
		Description:           req.Description,
		Category:              req.Category,
		ThumbnailURL:          req.ThumbnailURL,
		DesignLink:            req.DesignLink,
		DefaultWeightGrams:    req.DefaultWeightGrams,
		MaterialType:          req.MaterialType,
		DefaultPrintTimeHours: req.DefaultPrintTimeHours,
		DefaultMachineID:      parseUUIDPtr(req.DefaultMachineID),
		PackagingPresetID:     parseUUIDPtr(req.PackagingPresetID),
		BatchSize:             req.BatchSize,
		PackingFeeIDR:         req.PackingFeeIDR,
		TargetMarginPercent:   req.TargetMarginPercent,
		Components:            comps,
		PackagingItems:        packs,
	}

	prod, cb, err := h.createUC.Execute(c.Request().Context(), userID, input)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail("Failed to create product", err.Error()))
	}

	return c.JSON(http.StatusCreated, dto.Success(dto.ToProductResponse(prod, 0, cb)))
}

func (h *ProductHandler) Update(c echo.Context) error {
	userID, err := getUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, dto.Fail("Unauthorized", err.Error()))
	}

	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail("Invalid product ID", err.Error()))
	}

	var req dto.UpdateProductRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail("Invalid payload", err.Error()))
	}

	var comps *[]productUC.ComponentInput
	if req.Components != nil {
		cList := make([]productUC.ComponentInput, len(*req.Components))
		for i, comp := range *req.Components {
			cList[i] = productUC.ComponentInput{
				ComponentID:   parseUUIDPtr(comp.ComponentID),
				Quantity:      comp.Quantity,
				MarkupPercent: comp.MarkupPercent,
			}
		}
		comps = &cList
	}

	var packs *[]productUC.PackagingItemInput
	if req.PackagingItems != nil {
		pList := make([]productUC.PackagingItemInput, len(*req.PackagingItems))
		for i, pack := range *req.PackagingItems {
			pkgID, _ := uuid.Parse(pack.PackagingItemID)
			pList[i] = productUC.PackagingItemInput{
				PackagingItemID: pkgID,
				QuantityUsed:    pack.QuantityUsed,
			}
		}
		packs = &pList
	}

	input := productUC.UpdateProductInput{
		ID:                    id,
		Name:                  req.Name,
		ParentSKU:             req.ParentSKU,
		SKU:                   req.SKU,
		Description:           req.Description,
		Category:              req.Category,
		ThumbnailURL:          req.ThumbnailURL,
		DesignLink:            req.DesignLink,
		DefaultWeightGrams:    req.DefaultWeightGrams,
		MaterialType:          req.MaterialType,
		DefaultPrintTimeHours: req.DefaultPrintTimeHours,
		DefaultMachineID:      parseUUIDPtr(req.DefaultMachineID),
		PackagingPresetID:     parseUUIDPtr(req.PackagingPresetID),
		BatchSize:             req.BatchSize,
		PackingFeeIDR:         req.PackingFeeIDR,
		TargetMarginPercent:   req.TargetMarginPercent,
		Components:            comps,
		PackagingItems:        packs,
	}

	prod, cb, err := h.updateUC.Execute(c.Request().Context(), userID, input)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail("Failed to update product", err.Error()))
	}

	return c.JSON(http.StatusOK, dto.Success(dto.ToProductResponse(prod, 0, cb)))
}

func (h *ProductHandler) Delete(c echo.Context) error {
	userID, err := getUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, dto.Fail("Unauthorized", err.Error()))
	}

	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail("Invalid product ID", err.Error()))
	}

	if err := h.deleteUC.Execute(c.Request().Context(), userID, id); err != nil {
		return c.JSON(http.StatusInternalServerError, dto.Fail("Failed to delete product", err.Error()))
	}

	return c.JSON(http.StatusOK, dto.Success(map[string]string{
		"message": "Product and BOM successfully deleted",
	}))
}

func (h *ProductHandler) AutoGenerateDraftSKUs(c echo.Context) error {
	userID, err := getUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, dto.Fail("Unauthorized", err.Error()))
	}

	count, err := h.autoGenSKUUC.Execute(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.Fail("Failed to generate SKUs", err.Error()))
	}

	return c.JSON(http.StatusOK, dto.Success(map[string]interface{}{
		"message":       fmt.Sprintf("Draft SKUs generated for %d products", count),
		"total_updated": count,
	}))
}