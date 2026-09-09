package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"symetra-lab-backend/models"
	"symetra-lab-backend/pkg/shopee"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// OrderHandler mengelola endpoint sinkronisasi dan query pesanan Shopee
type OrderHandler struct {
	DB           *gorm.DB
	ShopeeClient *shopee.Client
}

// NewOrderHandler membuat instance baru OrderHandler
func NewOrderHandler(db *gorm.DB, client *shopee.Client) *OrderHandler {
	return &OrderHandler{
		DB:           db,
		ShopeeClient: client,
	}
}

// getValidShopToken mengambil data shop dari DB dan otomatis refresh access token jika sudah kadaluarsa
func (h *OrderHandler) getValidShopToken(shopID uint64) (*models.Shop, error) {
	var shop models.Shop
	if err := h.DB.Where("shop_id = ?", shopID).First(&shop).Error; err != nil {
		return nil, fmt.Errorf("toko dengan shop_id %d belum terdaftar di sistem", shopID)
	}

	// Cek apakah access token sudah kadaluarsa (atau mendekati 5 menit sebelum expired)
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
			return nil, fmt.Errorf("gagal menyimpan token baru ke database: %w", err)
		}
	}

	return &shop, nil
}

// SyncOrders menarik order dari Shopee API, memperkaya dengan detail item & escrow, lalu menyimpannya ke Supabase
// POST /api/v1/shopee/shops/:shop_id/sync-orders
func (h *OrderHandler) SyncOrders(c *gin.Context) {
	shopIDParam := c.Param("shop_id")
	shopID, err := strconv.ParseUint(shopIDParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format shop_id tidak valid", "status": "error"})
		return
	}

	shop, err := h.getValidShopToken(shopID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error(), "status": "error"})
		return
	}

	// Filter rentang waktu: default 15 hari terakhir
	days := 15
	if daysQuery := c.Query("days"); daysQuery != "" {
		if d, err := strconv.Atoi(daysQuery); err == nil && d > 0 {
			days = d
		}
	}

	now := time.Now()
	timeTo := now.Unix()
	timeFrom := now.AddDate(0, 0, -days).Unix()
	orderStatus := c.Query("status") // opsional: READY_TO_SHIP, COMPLETED, dll

	// 1. Panggil GetOrderList
	orderListResp, err := h.ShopeeClient.GetOrderList(shop.AccessToken, shop.ShopID, timeFrom, timeTo, orderStatus, "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   fmt.Sprintf("Gagal menarik order list: %v", err),
			"status":  "error",
		})
		return
	}

	totalFound := len(orderListResp.Response.OrderList)
	if totalFound == 0 {
		c.JSON(http.StatusOK, gin.H{
			"message":      "Tidak ada pesanan ditemukan pada rentang waktu yang dipilih",
			"synced_count": 0,
			"status":       "success",
		})
		return
	}

	// Kumpulkan order_sn untuk ditarik detailnya (maks 50 per batch)
	var orderSNs []string
	for _, item := range orderListResp.Response.OrderList {
		orderSNs = append(orderSNs, item.OrderSN)
	}

	// 2. Panggil GetOrderDetail secara batch
	detailResp, err := h.ShopeeClient.GetOrderDetail(shop.AccessToken, shop.ShopID, orderSNs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   fmt.Sprintf("Gagal menarik order detail: %v", err),
			"status":  "error",
		})
		return
	}

	syncedOrders := make([]models.ShopeeOrder, 0, len(detailResp.Response.OrderList))

	// 3. Iterasi setiap order, ambil escrow (finansial), lalu simpan ke DB
	for _, od := range detailResp.Response.OrderList {
		var shipByDateTime *time.Time
		if od.ShipByDate > 0 {
			t := time.Unix(od.ShipByDate, 0)
			shipByDateTime = &t
		}

		orderModel := models.ShopeeOrder{
			OrderSN:           od.OrderSN,
			ShopID:            shop.ShopID,
			OrderStatus:       od.OrderStatus,
			BuyerUserID:       od.BuyerUserID,
			BuyerUsername:     od.BuyerUsername,
			MessageToSeller:   od.MessageToSeller,
			ShipByDate:        od.ShipByDate,
			ShipByDateTime:    shipByDateTime,
			ShippingCarrier:   od.ShippingCarrier,
			TotalAmount:       od.TotalAmount,
			BuyerCancelReason: od.BuyerCancelReason,
			CreateTimeShopee:  od.CreateTime,
			UpdateTimeShopee:  od.UpdateTime,
			UpdatedAt:         time.Now(),
		}

		// Items
		var items []models.ShopeeOrderItem
		for _, it := range od.ItemList {
			items = append(items, models.ShopeeOrderItem{
				OrderSN:         od.OrderSN,
				ItemID:          it.ItemID,
				ItemName:        it.ItemName,
				ItemSKU:         it.ItemSKU,
				ModelID:         it.ModelID,
				ModelName:       it.ModelName,
				ModelSKU:        it.ModelSKU,
				Quantity:        it.ModelQuantityPurchased,
				OriginalPrice:   it.ModelOriginalPrice,
				DiscountedPrice: it.ModelDiscountedPrice,
			})
		}
		orderModel.Items = items

		// 4. Tarik Escrow Detail (Finansial & Potongan Shopee)
		escrowResp, err := h.ShopeeClient.GetEscrowDetail(shop.AccessToken, shop.ShopID, od.OrderSN)
		if err == nil && escrowResp != nil {
			income := escrowResp.Response.OrderIncome
			sellingPrice := income.OrderSellingPrice
			if sellingPrice == 0 {
				sellingPrice = income.SellingPrice
			}

			// Hitung persentase komisi kategori & service fee
			var commPct, servPct float64
			if sellingPrice > 0 {
				commPct = (income.CommissionFee / sellingPrice) * 100
				servPct = (income.ServiceFee / sellingPrice) * 100
			}

			var commRuleName string
			if len(income.NetCommissionFeeInfo) > 0 {
				commRuleName = income.NetCommissionFeeInfo[0].RuleDisplayName
			}

			var servRuleName string
			if len(income.NetServiceFeeInfo) > 0 {
				servRuleName = income.NetServiceFeeInfo[0].RuleDisplayName
			}

			orderModel.Escrow = &models.ShopeeOrderEscrow{
				OrderSN:                  od.OrderSN,
				EscrowAmount:             income.EscrowAmount,
				SellingPrice:             sellingPrice,
				CommissionFee:            income.CommissionFee,
				CommissionRuleName:       commRuleName,
				CommissionPercentage:     commPct,
				ServiceFee:               income.ServiceFee,
				ServiceRuleName:          servRuleName,
				ServicePercentage:        servPct,
				SellerTransactionFee:     income.SellerTransactionFee,
				SellerOrderProcessingFee: income.SellerOrderProcessingFee,
				SellerVoucherDiscount:    income.VoucherFromSeller,
				UpdatedAt:                time.Now(),
			}
		}

		// 5. Upsert ke Database PostgreSQL (GORM Clauses)
		err = h.DB.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "order_sn"}},
			UpdateAll: true,
		}).Create(&orderModel).Error

		if err == nil {
			// Refresh items (hapus lama, pasang baru agar tidak duplikat)
			_ = h.DB.Where("order_sn = ?", od.OrderSN).Delete(&models.ShopeeOrderItem{})
			if len(items) > 0 {
				_ = h.DB.Create(&items)
			}
			if orderModel.Escrow != nil {
				_ = h.DB.Clauses(clause.OnConflict{
					Columns:   []clause.Column{{Name: "order_sn"}},
					UpdateAll: true,
				}).Create(orderModel.Escrow)
			}
			syncedOrders = append(syncedOrders, orderModel)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      fmt.Sprintf("Berhasil sinkronisasi %d pesanan dari Shopee", len(syncedOrders)),
		"synced_count": len(syncedOrders),
		"status":       "success",
		"data":         syncedOrders,
	})
}

// GetOrders mengambil daftar pesanan dari database dengan filter status, pencarian catatan pembeli, dan sorting deadline
// GET /api/v1/shopee/shops/:shop_id/orders
func (h *OrderHandler) GetOrders(c *gin.Context) {
	shopIDParam := c.Param("shop_id")
	shopID, err := strconv.ParseUint(shopIDParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format shop_id tidak valid", "status": "error"})
		return
	}

	query := h.DB.Model(&models.ShopeeOrder{}).Where("shop_id = ?", shopID)

	// Filter order_status
	if status := c.Query("status"); status != "" {
		query = query.Where("order_status = ?", status)
	}

	// Filter pencarian teks (Order SN, username pembeli, atau catatan pembeli)
	if search := c.Query("search"); search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("order_sn ILIKE ? OR buyer_username ILIKE ? OR message_to_seller ILIKE ?", searchPattern, searchPattern, searchPattern)
	}

	// Sorting
	sortBy := c.Query("sort_by")
	switch sortBy {
	case "deadline":
		// Prioritaskan yang deadline kirimnya paling dekat! (SLA anti-penalty)
		query = query.Where("ship_by_date > 0").Order("ship_by_date ASC")
	case "amount":
		query = query.Order("total_amount DESC")
	default:
		// Default: yang paling baru dibuat
		query = query.Order("create_time_shopee DESC, created_at DESC")
	}

	// Pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	var total int64
	query.Count(&total)

	var orders []models.ShopeeOrder
	if err := query.Preload("Items").Preload("Escrow").Offset(offset).Limit(pageSize).Find(&orders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data pesanan", "status": "error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"orders":      orders,
			"total":       total,
			"page":        page,
			"page_size":   pageSize,
			"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

// GetOrderDetail mengambil detail lengkap 1 pesanan berdasarkan order_sn
// GET /api/v1/shopee/orders/:order_sn
func (h *OrderHandler) GetOrderDetail(c *gin.Context) {
	orderSN := c.Param("order_sn")
	if orderSN == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "order_sn tidak boleh kosong", "status": "error"})
		return
	}

	var order models.ShopeeOrder
	if err := h.DB.Preload("Items").Preload("Escrow").Where("order_sn = ?", orderSN).First(&order).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Pesanan tidak ditemukan", "status": "error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   order,
	})
}
