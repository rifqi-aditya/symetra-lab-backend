package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"symetra-lab-backend-v2/config"
	"symetra-lab-backend-v2/internal/delivery/http/dto"
	shopeeUC "symetra-lab-backend-v2/internal/usecase/shopee"
	pkgShopee "symetra-lab-backend-v2/pkg/shopee"
)

type ShopeeHandler struct {
	cfg *config.Config
	uc  *shopeeUC.ShopeeUseCases
}

func NewShopeeHandler(cfg *config.Config, uc *shopeeUC.ShopeeUseCases) *ShopeeHandler {
	return &ShopeeHandler{cfg: cfg, uc: uc}
}

func (h *ShopeeHandler) GetAuthURL(c echo.Context) error {
	url, err := h.uc.GetAuthURL()
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail("Failed to build Shopee auth URL", err.Error()))
	}

	if c.QueryParam("redirect") == "true" {
		return c.Redirect(http.StatusTemporaryRedirect, url)
	}

	return c.JSON(http.StatusOK, dto.Success(map[string]string{
		"auth_url": url,
		"message":  "Open this URL to authorize Shopee store",
	}))
}

func (h *ShopeeHandler) HandleCallback(c echo.Context) error {
	code := c.QueryParam("code")
	shopIDStr := c.QueryParam("shop_id")

	if code == "" || shopIDStr == "" {
		return c.JSON(http.StatusBadRequest, dto.Fail("Invalid callback parameters", "code and shop_id required"))
	}

	shopID, err := strconv.ParseUint(shopIDStr, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail("Invalid shop_id", err.Error()))
	}

	shop, err := h.uc.HandleCallback(c.Request().Context(), code, shopID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail("Failed to exchange token", err.Error()))
	}

	return c.JSON(http.StatusOK, dto.Success(dto.ToShopResponse(shop)))
}

func (h *ShopeeHandler) ListShops(c echo.Context) error {
	shops, err := h.uc.ListShops(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.Fail("Failed to list shops", err.Error()))
	}

	resp := make([]dto.ShopeeShopResponse, len(shops))
	for i, s := range shops {
		resp[i] = dto.ToShopResponse(s)
	}

	return c.JSON(http.StatusOK, dto.Success(resp))
}

func (h *ShopeeHandler) RefreshToken(c echo.Context) error {
	shopID, err := strconv.ParseUint(c.Param("shop_id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail("Invalid shop_id", err.Error()))
	}

	shop, err := h.uc.RefreshToken(c.Request().Context(), shopID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail("Failed to refresh token", err.Error()))
	}

	return c.JSON(http.StatusOK, dto.Success(dto.ToShopResponse(shop)))
}

func (h *ShopeeHandler) ListOrders(c echo.Context) error {
	status := c.QueryParam("status")
	carrier := c.QueryParam("carrier")
	search := c.QueryParam("search")

	orders, err := h.uc.ListOrders(c.Request().Context(), status, carrier, search)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.Fail("Failed to fetch Shopee orders", err.Error()))
	}

	resp := make([]dto.ShopeeOrderResponse, len(orders))
	for i, o := range orders {
		resp[i] = dto.ToShopeeOrderResponse(o)
	}

	return c.JSON(http.StatusOK, dto.Success(resp))
}

func (h *ShopeeHandler) GetOrderDetail(c echo.Context) error {
	orderSN := c.Param("order_sn")
	if orderSN == "" {
		return c.JSON(http.StatusBadRequest, dto.Fail("order_sn required", nil))
	}

	ord, err := h.uc.GetOrderDetail(c.Request().Context(), orderSN)
	if err != nil {
		return c.JSON(http.StatusNotFound, dto.Fail("Shopee order not found", err.Error()))
	}

	return c.JSON(http.StatusOK, dto.Success(dto.ToShopeeOrderResponse(ord)))
}

func (h *ShopeeHandler) SyncShopee(c echo.Context) error {
	count, err := h.uc.SyncShopeeOrders(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.Fail("Failed to sync Shopee orders", err.Error()))
	}

	return c.JSON(http.StatusOK, dto.Success(map[string]interface{}{
		"message":     fmt.Sprintf("Synchronized %d Shopee orders", count),
		"order_count": count,
	}))
}

func (h *ShopeeHandler) LinkSKU(c echo.Context) error {
	var req dto.LinkSKURequest
	if err := c.Bind(&req); err != nil || req.ShopeeItemID == 0 {
		return c.JSON(http.StatusBadRequest, dto.Fail("Invalid payload", "shopee_item_id and product_id required"))
	}

	if err := h.uc.LinkSKU(c.Request().Context(), req.ShopeeItemID, req.ShopeeModelID, req.ProductID); err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail("Failed to link SKU", err.Error()))
	}

	return c.JSON(http.StatusOK, dto.Success(map[string]string{
		"message": "SKU successfully linked",
	}))
}

func (h *ShopeeHandler) ShipOrder(c echo.Context) error {
	orderSN := c.Param("order_sn")
	if orderSN == "" {
		return c.JSON(http.StatusBadRequest, dto.Fail("order_sn required", nil))
	}

	resp, err := h.uc.ShipOrder(c.Request().Context(), orderSN)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail("Failed to ship order", err.Error()))
	}

	return c.JSON(http.StatusOK, dto.Success(resp))
}

func (h *ShopeeHandler) DownloadShippingLabel(c echo.Context) error {
	orderSN := c.Param("order_sn")
	if orderSN == "" {
		return c.JSON(http.StatusBadRequest, dto.Fail("order_sn required", nil))
	}

	data, err := h.uc.DownloadShippingLabel(c.Request().Context(), orderSN)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail("Failed to download label", err.Error()))
	}

	c.Response().Header().Set("Content-Type", "application/pdf")
	return c.Blob(http.StatusOK, "application/pdf", data)
}

func (h *ShopeeHandler) GetCashflowSummary(c echo.Context) error {
	summary, err := h.uc.GetCashflowSummary(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.Fail("Failed to retrieve cashflow summary", err.Error()))
	}

	return c.JSON(http.StatusOK, dto.Success(dto.ToCashflowSummaryResponse(summary)))
}

func (h *ShopeeHandler) RecalculateFinances(c echo.Context) error {
	count, err := h.uc.RecalculateFinances(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.Fail("Failed to recalculate finances", err.Error()))
	}

	return c.JSON(http.StatusOK, dto.Success(map[string]interface{}{
		"message":             fmt.Sprintf("Recalculated %d orders", count),
		"recalculated_orders": count,
	}))
}

// HandlePushWebhook menerima push notification HTTP POST dari Shopee Open Platform
// Menangani notifikasi code == 3 (order_status_push)
func (h *ShopeeHandler) HandlePushWebhook(c echo.Context) error {
	req := c.Request()
	rawBody, err := io.ReadAll(req.Body)
	if err != nil {
		log.Printf("[Shopee Webhook] Error reading request body: %v", err)
		return c.JSON(http.StatusBadRequest, dto.Fail("Invalid body", err.Error()))
	}

	// 1. Verifikasi Signature Shopee (Header: Authorization atau X-Shopee-Signature)
	authHeader := req.Header.Get("Authorization")
	if authHeader == "" {
		authHeader = req.Header.Get("X-Shopee-Signature")
	}

	// Buat full URL request untuk pencocokan signature
	scheme := "http"
	if req.TLS != nil || req.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	fullURL := fmt.Sprintf("%s://%s%s", scheme, req.Host, req.RequestURI)

	if h.cfg != nil && h.cfg.ShopeePartnerKey != "" && authHeader != "" {
		isValid := pkgShopee.VerifyWebhookSignature(fullURL, rawBody, authHeader, h.cfg.ShopeePartnerKey)
		if !isValid {
			log.Printf("[Shopee Webhook] Signature mismatch! URL: %s, Auth: %s", fullURL, authHeader)
			// Return 401 jika signature tidak cocok
			return c.JSON(http.StatusUnauthorized, dto.Fail("Unauthorized webhook signature", nil))
		}
	}

	// 2. Parse Webhook Push Payload
	var push pkgShopee.WebhookPushPayload
	if err := json.Unmarshal(rawBody, &push); err != nil {
		log.Printf("[Shopee Webhook] Failed to decode JSON payload: %v (raw: %s)", err, string(rawBody))
		return c.JSON(http.StatusBadRequest, dto.Fail("Invalid JSON payload", err.Error()))
	}

	log.Printf("[Shopee Webhook] Received push event code=%d, shop_id=%d, order_sn=%s, status=%s",
		push.Code, push.ShopID, push.Data.OrderSN, push.Data.Status)

	// Code 3 = order_status_push
	if push.Code == 3 || push.Data.OrderSN != "" {
		ctx := req.Context()
		orderSN := push.Data.OrderSN
		status := push.Data.Status
		shopID := push.ShopID

		// Jalankan proses sinkronisasi order & escrow realtime
		updatedOrder, err := h.uc.HandleOrderStatusPush(ctx, shopID, orderSN, status)
		if err != nil {
			log.Printf("[Shopee Webhook] Error syncing order %s: %v", orderSN, err)
			// Shopee mengharapkan status 200 OK agar tidak me-retry terus menerus
			return c.JSON(http.StatusOK, map[string]interface{}{
				"status":   "error_processing",
				"message":  err.Error(),
				"order_sn": orderSN,
			})
		}

		log.Printf("[Shopee Webhook] Order %s successfully updated to status %s", orderSN, updatedOrder.OrderStatus())
		return c.JSON(http.StatusOK, map[string]interface{}{
			"status":   "success",
			"code":     push.Code,
			"order_sn": orderSN,
		})
	}

	// Event lain selain order_status_push cukup di-ACK dengan 200 OK
	return c.JSON(http.StatusOK, map[string]interface{}{
		"status": "ignored",
		"code":   push.Code,
	})
}

