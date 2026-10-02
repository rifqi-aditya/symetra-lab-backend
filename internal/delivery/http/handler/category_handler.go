package handler

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"symetra-lab-backend-v2/internal/delivery/http/dto"
	categoryUC "symetra-lab-backend-v2/internal/usecase/category"
)

type CategoryHandler struct {
	listUC   *categoryUC.ListCategoriesUseCase
	createUC *categoryUC.CreateCategoryUseCase
	updateUC *categoryUC.UpdateCategoryUseCase
	deleteUC *categoryUC.DeleteCategoryUseCase
}

func NewCategoryHandler(
	listUC *categoryUC.ListCategoriesUseCase,
	createUC *categoryUC.CreateCategoryUseCase,
	updateUC *categoryUC.UpdateCategoryUseCase,
	deleteUC *categoryUC.DeleteCategoryUseCase,
) *CategoryHandler {
	return &CategoryHandler{
		listUC:   listUC,
		createUC: createUC,
		updateUC: updateUC,
		deleteUC: deleteUC,
	}
}

func (h *CategoryHandler) List(c echo.Context) error {
	userID, _ := getUserID(c)
	items, err := h.listUC.Execute(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.Fail("Failed to fetch categories", err.Error()))
	}
	resp := make([]dto.CategoryResponse, len(items))
	for i, cat := range items {
		resp[i] = dto.ToCategoryResponse(cat)
	}
	return c.JSON(http.StatusOK, dto.Success(resp))
}

func (h *CategoryHandler) Create(c echo.Context) error {
	userID, err := getUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, dto.Fail("Unauthorized", err.Error()))
	}

	var req dto.CategoryRequest
	if err := c.Bind(&req); err != nil || req.Name == "" {
		return c.JSON(http.StatusBadRequest, dto.Fail("Invalid payload", "Category name is required"))
	}

	cat, err := h.createUC.Execute(c.Request().Context(), userID, req.Name)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail("Failed to create category", err.Error()))
	}

	return c.JSON(http.StatusCreated, dto.Success(dto.ToCategoryResponse(cat)))
}

func (h *CategoryHandler) Update(c echo.Context) error {
	userID, err := getUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, dto.Fail("Unauthorized", err.Error()))
	}

	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail("Invalid category ID", err.Error()))
	}

	var req dto.CategoryRequest
	if err := c.Bind(&req); err != nil || req.Name == "" {
		return c.JSON(http.StatusBadRequest, dto.Fail("Invalid payload", "Category name is required"))
	}

	cat, err := h.updateUC.Execute(c.Request().Context(), userID, id, req.Name)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail("Failed to update category", err.Error()))
	}

	return c.JSON(http.StatusOK, dto.Success(dto.ToCategoryResponse(cat)))
}

func (h *CategoryHandler) Delete(c echo.Context) error {
	userID, err := getUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, dto.Fail("Unauthorized", err.Error()))
	}

	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail("Invalid category ID", err.Error()))
	}

	if err := h.deleteUC.Execute(c.Request().Context(), userID, id); err != nil {
		return c.JSON(http.StatusInternalServerError, dto.Fail("Failed to delete category", err.Error()))
	}

	return c.JSON(http.StatusOK, dto.Success(map[string]string{"message": "Category deleted successfully"}))
}