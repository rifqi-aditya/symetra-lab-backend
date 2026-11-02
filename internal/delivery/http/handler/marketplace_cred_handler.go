package handler

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"symetra-lab-backend-v2/internal/delivery/http/dto"
	credUC "symetra-lab-backend-v2/internal/usecase/marketplacecred"
)

type MarketplaceCredHandler struct {
	uc *credUC.UseCase
}

func NewMarketplaceCredHandler(uc *credUC.UseCase) *MarketplaceCredHandler {
	return &MarketplaceCredHandler{uc: uc}
}

func (h *MarketplaceCredHandler) ListCredentials(c echo.Context) error {
	creds, err := h.uc.ListCredentials(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.Fail("Failed to list marketplace credentials", err.Error()))
	}

	res := make([]dto.MarketplaceCredentialResponse, len(creds))
	for i, cr := range creds {
		res[i] = dto.ToMarketplaceCredentialResponse(cr)
	}

	return c.JSON(http.StatusOK, dto.Success(res))
}

func (h *MarketplaceCredHandler) SaveCredential(c echo.Context) error {
	var req dto.SaveMarketplaceCredentialRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail("Invalid payload", err.Error()))
	}

	saved, err := h.uc.SaveCredential(
		c.Request().Context(),
		req.Platform,
		req.PartnerID,
		req.PartnerKey,
		req.RedirectURL,
		req.IsProduction,
	)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail("Failed to save marketplace credential", err.Error()))
	}

	return c.JSON(http.StatusOK, dto.Success(dto.ToMarketplaceCredentialResponse(saved)))
}

func (h *MarketplaceCredHandler) DeleteCredential(c echo.Context) error {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail("Invalid UUID", err.Error()))
	}

	if err := h.uc.DeleteCredential(c.Request().Context(), id); err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail("Failed to delete marketplace credential", err.Error()))
	}

	return c.JSON(http.StatusOK, dto.Success(map[string]string{"message": "Marketplace credential deleted"}))
}

