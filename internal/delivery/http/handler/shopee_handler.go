package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"

	"symetra-lab-backend-v2/config"
	"symetra-lab-backend-v2/internal/delivery/http/dto"
	credUC "symetra-lab-backend-v2/internal/usecase/marketplacecred"
	shopeeUC "symetra-lab-backend-v2/internal/usecase/shopee"
	pkgOAuthState "symetra-lab-backend-v2/pkg/oauthstate"
	pkgShopee "symetra-lab-backend-v2/pkg/shopee"
)

type ShopeeHandler struct {
	cfg          *config.Config
	uc           *shopeeUC.ShopeeUseCases
	stateManager *pkgOAuthState.Manager
	credUC       *credUC.UseCase
}

func (h *ShopeeHandler) WithCredUseCase(c *credUC.UseCase) *ShopeeHandler {
	h.credUC = c
	return h
}

func NewShopeeHandler(cfg *config.Config, uc *shopeeUC.ShopeeUseCases) *ShopeeHandler {
	return &ShopeeHandler{
		cfg:          cfg,
		uc:           uc,
		stateManager: pkgOAuthState.NewManager(15 * time.Minute),
	}
}

func (h *ShopeeHandler) GetAuthURL(c echo.Context) error {
	baseURL, err := h.uc.GetAuthURL(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.Fail("Failed to build Shopee auth URL", err.Error()))
	}

	// Buat anti-CSRF state token
	state := ""
	if h.stateManager != nil {
		if st, err := h.stateManager.GenerateState(); err == nil {
			state = st
		}
	}

	finalURL := baseURL
	if state != "" {
		if strings.Contains(finalURL, "?") {
			finalURL = fmt.Sprintf("%s&state=%s", finalURL, url.QueryEscape(state))
		} else {
			finalURL = fmt.Sprintf("%s?state=%s", finalURL, url.QueryEscape(state))
		}
	}

	if c.QueryParam("redirect") == "true" || strings.Contains(c.Request().Header.Get("Accept"), "text/html") {
		return c.Redirect(http.StatusTemporaryRedirect, finalURL)
	}

	return c.JSON(http.StatusOK, dto.Success(map[string]string{
		"auth_url": finalURL,
		"state":    state,
		"message":  "Open this URL to authorize Shopee store",
	}))
}

func (h *ShopeeHandler) HandleCallback(c echo.Context) error {
	code := c.QueryParam("code")
	shopIDStr := c.QueryParam("shop_id")
	state := c.QueryParam("state")
	errParam := c.QueryParam("error")
	errMsg := c.QueryParam("msg")
	format := c.QueryParam("format") // jika "json", return JSON response bukan redirect (untuk automated testing)

	// Tentukan target redirect Frontend
	feURL := "http://localhost:5173"
	if h.cfg != nil && h.cfg.FrontendURL != "" {
		feURL = strings.TrimRight(h.cfg.FrontendURL, "/")
	}
	targetBase := fmt.Sprintf("%s/marketplace", feURL)

	// 1. Cek jika Shopee mengembalikan error
	if errParam != "" || errMsg != "" {
		msg := errMsg
		if msg == "" {
			msg = errParam
		}
		if format == "json" {
			return c.JSON(http.StatusBadRequest, dto.Fail("Shopee authorization error", msg))
		}
		return c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s?status=error&message=%s", targetBase, url.QueryEscape(msg)))
	}

	// 2. Validasi Anti-CSRF OAuth State jika disediakan
	if state != "" && h.stateManager != nil {
		if !h.stateManager.ValidateAndConsume(state) {
			if format == "json" {
				return c.JSON(http.StatusBadRequest, dto.Fail("Invalid or expired OAuth state", "CSRF verification failed"))
			}
			return c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s?status=error&message=%s", targetBase, url.QueryEscape("Sesi otorisasi kedaluwarsa atau tidak valid (CSRF check)")))
		}
	}

	// 3. Validasi parameter code dan shop_id
	if code == "" || shopIDStr == "" {
		if format == "json" {
			return c.JSON(http.StatusBadRequest, dto.Fail("Invalid callback parameters", "code and shop_id required"))
		}
		return c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s?status=error&message=%s", targetBase, url.QueryEscape("Kode otorisasi atau ID toko tidak ditemukan")))
	}

	shopID, err := strconv.ParseUint(shopIDStr, 10, 64)
	if err != nil {
		if format == "json" {
			return c.JSON(http.StatusBadRequest, dto.Fail("Invalid shop_id", err.Error()))
		}
		return c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s?status=error&message=%s", targetBase, url.QueryEscape("ID toko tidak valid")))
	}

	// 4. Lakukan pertukaran token dengan Shopee Open Platform
	shop, err := h.uc.HandleCallback(c.Request().Context(), code, shopID)
	if err != nil {
		if format == "json" {
			return c.JSON(http.StatusBadRequest, dto.Fail("Failed to exchange token", err.Error()))
		}
		return c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s?status=error&message=%s", targetBase, url.QueryEscape(err.Error())))
	}

	// 5. Berhasil terhubung! Redirect ke halaman Marketplace Frontend
	if format == "json" {
		return c.JSON(http.StatusOK, dto.Success(dto.ToShopResponse(shop)))
	}
	return c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s?status=success&shop_id=%d", targetBase, shop.ShopID()))
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

// HandlePushWebhook menerima push notification HTTP POST / GET dari Shopee Open Platform
// Menangani notifikasi code == 3 (order_status_push) dan verifikasi URL callback
func (h *ShopeeHandler) HandlePushWebhook(c echo.Context) error {
	req := c.Request()

	// 1. Dukung HTTP GET (sering digunakan sistem platform untuk verifikasi ping / liveness)
	if req.Method == http.MethodGet {
		return c.JSON(http.StatusOK, map[string]string{
			"status":  "success",
			"message": "Shopee webhook endpoint is active",
		})
	}

	rawBody, err := io.ReadAll(req.Body)
	if err != nil {
		log.Printf("[Shopee Webhook] Error reading request body: %v", err)
		return c.JSON(http.StatusOK, map[string]string{"status": "error_reading_body"})
	}

	// Jika body kosong (misal connection handshake test push dari Shopee)
	if len(rawBody) == 0 {
		return c.JSON(http.StatusOK, map[string]string{
			"status":  "success",
			"message": "Endpoint handshake verified",
		})
	}

	// 2. Verifikasi Signature Shopee (Header: Authorization atau X-Shopee-Signature)
	authHeader := req.Header.Get("Authorization")
	if authHeader == "" {
		authHeader = req.Header.Get("X-Shopee-Signature")
	}

	scheme := "http"
	if req.TLS != nil || req.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	var fullURL string
	if strings.HasPrefix(req.RequestURI, "http://") || strings.HasPrefix(req.RequestURI, "https://") {
		fullURL = req.RequestURI
	} else {
		fullURL = fmt.Sprintf("%s://%s%s", scheme, req.Host, req.RequestURI)
	}

	var candidateKeys []string
	if h.credUC != nil {
		if dbKey := h.credUC.GetShopeePushKey(req.Context()); dbKey != "" {
			candidateKeys = append(candidateKeys, dbKey)
		}
		if dbPartnerKey := h.credUC.GetShopeePartnerKey(req.Context()); dbPartnerKey != "" {
			candidateKeys = append(candidateKeys, dbPartnerKey)
		}
	}
	if h.cfg != nil {
		if h.cfg.ShopeePushPartnerKey != "" {
			candidateKeys = append(candidateKeys, h.cfg.ShopeePushPartnerKey)
		}
		if h.cfg.ShopeePartnerKey != "" {
			candidateKeys = append(candidateKeys, h.cfg.ShopeePartnerKey)
		}
	}

	// Fallback to verified static push key and partner key to guarantee webhook processing
	// against DB credential typos or browser autofill overwrites
	candidateKeys = append(candidateKeys,
		"4d68734e79635058547376725a6b505765555279544867476d57456f54774455",
		"shpk676b516461596a455257534a557973446969466f44455268686f6d727756",
	)

	// De-duplicate non-empty candidate keys
	uniqueKeys := make([]string, 0, len(candidateKeys))
	seen := make(map[string]bool)
	for _, k := range candidateKeys {
		kTrim := strings.TrimSpace(k)
		if kTrim != "" && !seen[kTrim] {
			seen[kTrim] = true
			uniqueKeys = append(uniqueKeys, kTrim)
		}
	}

	if len(uniqueKeys) > 0 {
		if authHeader == "" {
			log.Printf("[Shopee Webhook] REJECTED: Missing signature header (URL: %s)", fullURL)
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"status":  "unauthorized",
				"message": "Missing webhook signature header",
			})
		}

		isValid := false
		for _, key := range uniqueKeys {
			if pkgShopee.VerifyWebhookSignature(fullURL, rawBody, authHeader, key) {
				isValid = true
				break
			}
		}

		if !isValid {
			maskedKeys := make([]string, len(uniqueKeys))
			for i, k := range uniqueKeys {
				if len(k) >= 8 {
					maskedKeys[i] = fmt.Sprintf("len=%d(%s...%s)", len(k), k[:4], k[len(k)-4:])
				} else {
					maskedKeys[i] = fmt.Sprintf("len=%d", len(k))
				}
			}
			log.Printf("[Shopee Webhook] REJECTED: Invalid signature! URL: %s, Auth: %s (tested %d keys: %v), RawBody: %s",
				fullURL, authHeader, len(uniqueKeys), maskedKeys, string(rawBody))
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"status":  "unauthorized",
				"message": "Invalid webhook signature",
			})
		}
		log.Printf("[Shopee Webhook] Signature verified successfully")
	} else {
		log.Printf("[Shopee Webhook] WARNING: No partner key found in DB or ENV. Cannot verify signature for URL: %s", fullURL)
	}

	// 3. Parse Webhook Push Payload
	var push pkgShopee.WebhookPushPayload
	if err := json.Unmarshal(rawBody, &push); err != nil {
		log.Printf("[Shopee Webhook] Non-standard payload or test ping: %v (raw: %s)", err, string(rawBody))
		// Shopee verifier mengharapkan status 2xx
		return c.JSON(http.StatusOK, map[string]interface{}{
			"status": "received",
			"raw":    string(rawBody),
		})
	}

	orderSN := push.Data.GetOrderSN()
	log.Printf("[Shopee Webhook] Received push event code=%d, shop_id=%d, order_sn=%s, status=%s",
		push.Code, push.ShopID, orderSN, push.Data.Status)

	// Code 3 = order_status_push atau payload mengandung orderSN
	if push.Code == 3 || orderSN != "" {
		ctx := req.Context()
		status := push.Data.Status
		shopID := push.ShopID

		// Jalankan proses sinkronisasi order & escrow realtime
		updatedOrder, err := h.uc.HandleOrderStatusPush(ctx, uint64(shopID), orderSN, status)
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

