package handlers

import (
	"fmt"
	"net/http"
	"time"

	"symetra-lab-backend/models"
	"symetra-lab-backend/pkg/shopee"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// LogisticsHandler mengelola endpoint pengiriman dan pencetakan label resi thermal
type LogisticsHandler struct {
	DB           *gorm.DB
	ShopeeClient *shopee.Client
}

// NewLogisticsHandler membuat instance baru LogisticsHandler
func NewLogisticsHandler(db *gorm.DB, client *shopee.Client) *LogisticsHandler {
	return &LogisticsHandler{
		DB:           db,
		ShopeeClient: client,
	}
}

// getValidShopToken mengambil data shop dari DB dan otomatis refresh token jika kadaluarsa
func (h *LogisticsHandler) getValidShopToken(shopID uint64) (*models.Shop, error) {
	var shop models.Shop
	if err := h.DB.Where("shop_id = ?", shopID).First(&shop).Error; err != nil {
		return nil, fmt.Errorf("toko dengan shop_id %d belum terdaftar", shopID)
	}

	if shop.IsTokenExpired() {
		refreshResp, err := h.ShopeeClient.RefreshAccessToken(shop.RefreshToken, shop.ShopID)
		if err != nil {
			return nil, fmt.Errorf("gagal refresh token otomatis: %w", err)
		}

		shop.AccessToken = refreshResp.AccessToken
		shop.RefreshToken = refreshResp.RefreshToken
		shop.AccessTokenExpiresAt = time.Now().Add(time.Duration(refreshResp.ExpireIn) * time.Second)
		shop.UpdatedAt = time.Now()

		if err := h.DB.Save(&shop).Error; err != nil {
			return nil, fmt.Errorf("gagal menyimpan token baru: %w", err)
		}
	}

	return &shop, nil
}

// ShipOrder memproses tombol "Atur Pengiriman", mengubah status pesanan menjadi PROCESSED dan menerbitkan nomor resi
// POST /api/v1/shopee/orders/:order_sn/ship
func (h *LogisticsHandler) ShipOrder(c *gin.Context) {
	orderSN := c.Param("order_sn")
	if orderSN == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "order_sn tidak boleh kosong", "status": "error"})
		return
	}

	var order models.ShopeeOrder
	if err := h.DB.Where("order_sn = ?", orderSN).First(&order).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Pesanan tidak ditemukan di database", "status": "error"})
		return
	}

	shop, err := h.getValidShopToken(order.ShopID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error(), "status": "error"})
		return
	}

	// 1. Cek opsi pengiriman via shipping parameter
	paramResp, err := h.ShopeeClient.GetShippingParameter(shop.AccessToken, shop.ShopID, orderSN)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Gagal cek shipping parameter: %v", err), "status": "error"})
		return
	}

	// 2. Siapkan payload request ship order
	shipReq := shopee.ShopeeShipOrderRequest{
		OrderSN: orderSN,
	}

	// Prioritaskan dropoff jika tersedia
	if len(paramResp.Response.Dropoff.BranchList) > 0 {
		shipReq.Dropoff = &shopee.ShopeeShipDropoff{
			BranchID: paramResp.Response.Dropoff.BranchList[0].BranchID,
		}
	} else if len(paramResp.Response.Pickup.AddressList) > 0 {
		pickupAddr := paramResp.Response.Pickup.AddressList[0]
		shipReq.Pickup = &shopee.ShopeeShipPickup{
			AddressID: pickupAddr.AddressID,
		}
		if len(pickupAddr.TimeSlotList) > 0 {
			shipReq.Pickup.PickupTimeID = pickupAddr.TimeSlotList[0].PickupTimeID
		}
	} else {
		// Default dropoff
		shipReq.Dropoff = &shopee.ShopeeShipDropoff{}
	}

	// 3. Konfirmasi Atur Pengiriman ke Shopee
	shipResp, err := h.ShopeeClient.ShipOrder(shop.AccessToken, shop.ShopID, shipReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Gagal atur pengiriman di Shopee: %v", err), "status": "error"})
		return
	}
	if shipResp.Error != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": shipResp.Message, "status": "error"})
		return
	}

	// 4. Ambil nomor resi resmi
	var trackingNumber string
	trackResp, err := h.ShopeeClient.GetTrackingNumber(shop.AccessToken, shop.ShopID, orderSN)
	if err == nil && trackResp != nil {
		trackingNumber = trackResp.Response.TrackingNumber
	}

	// 5. Update status di database Supabase
	order.OrderStatus = "PROCESSED"
	if trackingNumber != "" {
		order.TrackingNumber = trackingNumber
	}
	order.UpdatedAt = time.Now()
	_ = h.DB.Save(&order)

	c.JSON(http.StatusOK, gin.H{
		"status":          "success",
		"message":         "Pengiriman berhasil diatur",
		"order_sn":        orderSN,
		"order_status":    "PROCESSED",
		"tracking_number": trackingNumber,
	})
}

// DownloadShippingLabel mengambil file PDF label resi thermal (100x150 mm) dan langsung menyajikannya ke browser
// GET /api/v1/shopee/orders/:order_sn/shipping-label
func (h *LogisticsHandler) DownloadShippingLabel(c *gin.Context) {
	orderSN := c.Param("order_sn")
	if orderSN == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "order_sn tidak boleh kosong", "status": "error"})
		return
	}

	var order models.ShopeeOrder
	if err := h.DB.Where("order_sn = ?", orderSN).First(&order).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Pesanan tidak ditemukan", "status": "error"})
		return
	}

	shop, err := h.getValidShopToken(order.ShopID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error(), "status": "error"})
		return
	}

	// Buat task dokumen terlebih dahulu (idempotent)
	_, _ = h.ShopeeClient.CreateShippingDocument(shop.AccessToken, shop.ShopID, orderSN)

	// Download PDF dari Shopee
	pdfBytes, err := h.ShopeeClient.DownloadShippingDocument(shop.AccessToken, shop.ShopID, orderSN)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Gagal download label resi: %v", err), "status": "error"})
		return
	}

	// Sajikan file PDF secara langsung
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf("inline; filename=\"resi_%s.pdf\"", orderSN))
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}
