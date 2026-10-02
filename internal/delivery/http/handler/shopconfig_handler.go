package handler

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"symetra-lab-backend-v2/internal/delivery/http/dto"
	shopUC "symetra-lab-backend-v2/internal/usecase/shopconfig"
)

type ShopConfigHandler struct {
	getShopConfigUC     *shopUC.GetShopConfigUseCase
	updateShopConfigUC  *shopUC.UpdateShopConfigUseCase
	listMarketplacesUC  *shopUC.ListMarketplacesUseCase
	createMarketplaceUC *shopUC.CreateMarketplaceUseCase
	updateMarketplaceUC *shopUC.UpdateMarketplaceUseCase
	deleteMarketplaceUC *shopUC.DeleteMarketplaceUseCase
}

func NewShopConfigHandler(
	getShopConfigUC *shopUC.GetShopConfigUseCase,
	updateShopConfigUC *shopUC.UpdateShopConfigUseCase,
	listMarketplacesUC *shopUC.ListMarketplacesUseCase,
	createMarketplaceUC *shopUC.CreateMarketplaceUseCase,
	updateMarketplaceUC *shopUC.UpdateMarketplaceUseCase,
	deleteMarketplaceUC *shopUC.DeleteMarketplaceUseCase,
) *ShopConfigHandler {
	return &ShopConfigHandler{
		getShopConfigUC:     getShopConfigUC,
		updateShopConfigUC:  updateShopConfigUC,
		listMarketplacesUC:  listMarketplacesUC,
		createMarketplaceUC: createMarketplaceUC,
		updateMarketplaceUC: updateMarketplaceUC,
		deleteMarketplaceUC: deleteMarketplaceUC,
	}
}

func (h *ShopConfigHandler) GetShopConfig(c echo.Context) error {
	userID, err := getUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, dto.Fail("Unauthorized", err.Error()))
	}

	cfg, err := h.getShopConfigUC.Execute(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.Fail("Gagal mengambil konfigurasi bengkel", err.Error()))
	}

	return c.JSON(http.StatusOK, dto.Success(dto.ToShopConfigResponse(cfg)))
}

func (h *ShopConfigHandler) UpdateShopConfig(c echo.Context) error {
	userID, err := getUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, dto.Fail("Unauthorized", err.Error()))
	}

	var req dto.UpdateShopConfigRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail("Format data tidak valid: "+err.Error(), err.Error()))
	}

	cfg, err := h.updateShopConfigUC.Execute(c.Request().Context(), userID, shopUC.UpdateShopConfigInput{
		FilamentPricePerRoll:    req.FilamentPricePerRoll,
		FilamentWeightGrams:     req.FilamentWeightGrams,
		ElectricityTariffPerKwh: req.ElectricityTariffPerKwh,
		PrinterPowerWatts:       req.PrinterPowerWatts,
		PrinterPrice:            req.PrinterPrice,
		PrinterLifespanHours:    req.PrinterLifespanHours,
		FailureBufferPercent:    req.FailureBufferPercent,
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.Fail("Gagal menyimpan konfigurasi: "+err.Error(), err.Error()))
	}

	return c.JSON(http.StatusOK, dto.SuccessWithMsg("Konfigurasi bengkel berhasil diperbarui", dto.ToShopConfigResponse(cfg)))
}

func (h *ShopConfigHandler) ListMarketplaces(c echo.Context) error {
	userID, err := getUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, dto.Fail("Unauthorized", err.Error()))
	}

	platforms, err := h.listMarketplacesUC.Execute(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.Fail("Gagal mengambil platform marketplace", err.Error()))
	}

	resp := make([]dto.MarketplacePlatformResponse, len(platforms))
	for i, p := range platforms {
		resp[i] = dto.ToMarketplaceResponse(p)
	}

	return c.JSON(http.StatusOK, dto.Success(resp))
}

func (h *ShopConfigHandler) CreateMarketplace(c echo.Context) error {
	userID, err := getUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, dto.Fail("Unauthorized", err.Error()))
	}

	var req dto.MarketplacePlatformRequest
	if err := c.Bind(&req); err != nil || req.Name == "" {
		return c.JSON(http.StatusBadRequest, dto.Fail("Format data tidak valid", "Field 'name' wajib diisi"))
	}

	p, err := h.createMarketplaceUC.Execute(c.Request().Context(), userID, shopUC.CreateMarketplaceInput{
		Name:                req.Name,
		CommissionPercent:   req.CommissionPercent,
		PromoFeePercent:     req.PromoFeePercent,
		FreeShippingPercent: req.FreeShippingPercent,
		OrderFeeIDR:         req.OrderFeeIDR,
		IsActive:            req.IsActive,
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.Fail("Gagal membuat platform: "+err.Error(), err.Error()))
	}

	return c.JSON(http.StatusCreated, dto.SuccessWithMsg("Platform marketplace berhasil ditambahkan", dto.ToMarketplaceResponse(p)))
}

func (h *ShopConfigHandler) UpdateMarketplace(c echo.Context) error {
	_, err := getUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, dto.Fail("Unauthorized", err.Error()))
	}

	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail("ID platform tidak valid", err.Error()))
	}

	var req dto.UpdateMarketplacePlatformRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail("Format data tidak valid: "+err.Error(), err.Error()))
	}

	p, err := h.updateMarketplaceUC.Execute(c.Request().Context(), id, shopUC.UpdateMarketplaceInput{
		Name:                req.Name,
		CommissionPercent:   req.CommissionPercent,
		PromoFeePercent:     req.PromoFeePercent,
		FreeShippingPercent: req.FreeShippingPercent,
		OrderFeeIDR:         req.OrderFeeIDR,
		IsActive:            req.IsActive,
	})
	if err != nil {
		return c.JSON(http.StatusNotFound, dto.Fail("Platform tidak ditemukan", err.Error()))
	}

	return c.JSON(http.StatusOK, dto.SuccessWithMsg("Platform marketplace berhasil diperbarui", dto.ToMarketplaceResponse(p)))
}

func (h *ShopConfigHandler) DeleteMarketplace(c echo.Context) error {
	_, err := getUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, dto.Fail("Unauthorized", err.Error()))
	}

	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail("ID platform tidak valid", err.Error()))
	}

	if err := h.deleteMarketplaceUC.Execute(c.Request().Context(), id); err != nil {
		return c.JSON(http.StatusInternalServerError, dto.Fail("Gagal menghapus platform", err.Error()))
	}

	return c.JSON(http.StatusOK, dto.SuccessWithMsg("Platform berhasil dihapus", nil))
}
