package handlers

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"symetra-lab-backend/models"
	"symetra-lab-backend/pkg/costing"
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

		// 5. Kalkulasi Alokasi Kas Bengkel & Auto-Match SKU ke Master Produk
		_ = costing.AllocateShopeeOrderFinances(&orderModel, h.DB)

		// 6. Upsert ke Database PostgreSQL (GORM Clauses)
		err = h.DB.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "order_sn"}},
			UpdateAll: true,
		}).Create(&orderModel).Error

		if err == nil {
			// Refresh items (hapus lama, pasang baru lengkap dengan snapshot HPP & kas terhitung)
			_ = h.DB.Where("order_sn = ?", od.OrderSN).Delete(&models.ShopeeOrderItem{})
			if len(orderModel.Items) > 0 {
				_ = h.DB.Create(&orderModel.Items)
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
	if err := query.Preload("Items.Product").Preload("Escrow").Offset(offset).Limit(pageSize).Find(&orders).Error; err != nil {
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
	if err := h.DB.Preload("Items.Product").Preload("Escrow").Where("order_sn = ?", orderSN).First(&order).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Pesanan tidak ditemukan", "status": "error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   order,
	})
}

// GetRawOrders mengambil respon mentah (raw JSON) get_order_list & get_order_detail dari Shopee tanpa menyimpan ke database
// GET /api/v1/shopee/shops/:shop_id/raw-orders
func (h *OrderHandler) GetRawOrders(c *gin.Context) {
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

	// Filter opsional dari query params:
	// ?status=READY_TO_SHIP | PROCESSED | SHIPPED | COMPLETED | CANCELLED
	// ?order_sn=260912T07VRYRJ (untuk cek 1 pesanan tertentu langsung)
	statusFilter := strings.ToUpper(strings.TrimSpace(c.Query("status")))
	specificOrderSN := strings.TrimSpace(c.Query("order_sn"))

	// Jika ada parameter order_sn spesifik, langsung tarik detail pesanan tersebut
	if specificOrderSN != "" {
		detailRaw, _ := h.ShopeeClient.CallShopeeAPIRaw("/api/v2/order/get_order_detail", shop.AccessToken, shop.ShopID, map[string]string{
			"order_sn_list":            specificOrderSN,
			"response_optional_fields": "buyer_user_id,buyer_username,item_list,recipient_address,shipping_carrier,total_amount,payment_method,estimated_shipping_fee,message_to_seller",
		})
		var detailJSON map[string]interface{}
		_ = json.Unmarshal(detailRaw, &detailJSON)

		escrowRaw, _ := h.ShopeeClient.CallShopeeAPIRaw("/api/v2/payment/get_escrow_detail", shop.AccessToken, shop.ShopID, map[string]string{
			"order_sn": specificOrderSN,
		})
		var escrowJSON map[string]interface{}
		_ = json.Unmarshal(escrowRaw, &escrowJSON)

		c.JSON(http.StatusOK, gin.H{
			"status":                   "success",
			"mode":                     "SPECIFIC_ORDER_INSPECTION",
			"shop_id":                  shop.ShopID,
			"order_sn":                 specificOrderSN,
			"order_detail":             detailJSON,
			"escrow_detail":            escrowJSON,
		})
		return
	}

	// 14 hari terakhir (Shopee v2 batas time range maksimal 15 hari)
	timeTo := time.Now().Unix()
	timeFrom := time.Now().AddDate(0, 0, -14).Unix()

	extraParams := map[string]string{
		"time_range_field":         "create_time",
		"time_from":                fmt.Sprintf("%d", timeFrom),
		"time_to":                  fmt.Sprintf("%d", timeTo),
		"page_size":                "20",
		"response_optional_fields": "order_status",
	}
	if statusFilter != "" {
		extraParams["order_status"] = statusFilter
	}

	// 1. Ambil raw order list
	orderListRaw, err := h.ShopeeClient.CallShopeeAPIRaw("/api/v2/order/get_order_list", shop.AccessToken, shop.ShopID, extraParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghubungi get_order_list Shopee: " + err.Error(), "status": "error"})
		return
	}

	var listJSON map[string]interface{}
	_ = json.Unmarshal(orderListRaw, &listJSON)

	// Kumpulkan order_sn untuk mengambil detail sampel
	var orderSNs []string
	if respMap, ok := listJSON["response"].(map[string]interface{}); ok {
		if orderList, ok := respMap["order_list"].([]interface{}); ok {
			for _, o := range orderList {
				if oMap, ok := o.(map[string]interface{}); ok {
					if sn, ok := oMap["order_sn"].(string); ok && sn != "" {
						orderSNs = append(orderSNs, sn)
					}
				}
			}
		}
	}

	// 2. Ambil raw order detail sampel (jika ada pesanan)
	var detailJSON map[string]interface{}
	var escrowJSON map[string]interface{}
	if len(orderSNs) > 0 {
		limit := len(orderSNs)
		if limit > 5 {
			limit = 5
		}
		detailRaw, _ := h.ShopeeClient.CallShopeeAPIRaw("/api/v2/order/get_order_detail", shop.AccessToken, shop.ShopID, map[string]string{
			"order_sn_list":            strings.Join(orderSNs[:limit], ","),
			"response_optional_fields": "buyer_user_id,buyer_username,item_list,recipient_address,shipping_carrier,total_amount,payment_method,estimated_shipping_fee,message_to_seller",
		})
		_ = json.Unmarshal(detailRaw, &detailJSON)

		// Ambil 1 sampel rincian escrow fee pesanan pertama
		escrowRaw, _ := h.ShopeeClient.CallShopeeAPIRaw("/api/v2/payment/get_escrow_detail", shop.AccessToken, shop.ShopID, map[string]string{
			"order_sn": orderSNs[0],
		})
		_ = json.Unmarshal(escrowRaw, &escrowJSON)
	}

	c.JSON(http.StatusOK, gin.H{
		"status":                "success",
		"mode":                  "RAW_INSPECTION_ONLY (NO DATABASE WRITE)",
		"shop_id":               shop.ShopID,
		"time_range": gin.H{
			"time_from": time.Unix(timeFrom, 0).Format("2006-01-02 15:04:05"),
			"time_to":   time.Unix(timeTo, 0).Format("2006-01-02 15:04:05"),
		},
		"total_orders_found":    len(orderSNs),
		"shopee_raw_order_list": listJSON,
		"shopee_raw_order_detail_sample": detailJSON,
		"shopee_raw_escrow_sample":       escrowJSON,
	})
}

// LinkItemSKU menghubungkan item pesanan Shopee yang belum terpetakan ke master produk fisik
// POST /api/v1/shopee/orders/:order_sn/items/:item_id/link-sku
func (h *OrderHandler) LinkItemSKU(c *gin.Context) {
	orderSN := c.Param("order_sn")
	itemIDStr := c.Param("item_id")
	itemID, err := strconv.ParseUint(itemIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format item_id tidak valid", "status": "error"})
		return
	}

	var req models.LinkSKURequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "product_id wajib diisi", "status": "error"})
		return
	}

	// 1. Cek apakah produk fisik valid
	var product models.Product
	if err := h.DB.Where("id = ?", req.ProductID).First(&product).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Produk fisik tidak ditemukan", "status": "error"})
		return
	}

	// 2. Cek apakah item pesanan ada
	var item models.ShopeeOrderItem
	if err := h.DB.Where("order_sn = ? AND (item_id = ? OR id = ?)", orderSN, itemID, itemID).First(&item).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Item pesanan tidak ditemukan", "status": "error"})
		return
	}

	// 3. Update mapping status & product_id
	item.ProductID = &product.ID
	skuVal := ""
	if product.SKU != nil {
		skuVal = *product.SKU
	}
	item.MatchedSKU = skuVal
	item.MappingStatus = "MANUAL_LINKED"
	if err := h.DB.Save(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan link produk", "status": "error"})
		return
	}

	// 4. Muat ulang seluruh order dan hitung ulang alokasi kas
	var order models.ShopeeOrder
	if err := h.DB.Preload("Items").Preload("Escrow").Where("order_sn = ?", orderSN).First(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memuat pesanan", "status": "error"})
		return
	}

	_ = costing.AllocateShopeeOrderFinances(&order, h.DB)

	// Simpan hasil kalkulasi baru
	for _, it := range order.Items {
		_ = h.DB.Save(&it)
	}
	if order.Escrow != nil {
		_ = h.DB.Save(order.Escrow)
	}

	// Reload dengan relasi product
	_ = h.DB.Preload("Items.Product").Preload("Escrow").Where("order_sn = ?", orderSN).First(&order)

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Berhasil memetakan item %s ke produk %s (%s)", item.ItemName, product.Name, skuVal),
		"status":  "success",
		"data":    order,
	})
}

// GetCashflowSummary mengembalikan rekapitulasi 5 ember kas bengkel berdasarkan pesanan Shopee
// GET /api/v1/shopee/financial/cashflow-summary
func (h *OrderHandler) GetCashflowSummary(c *gin.Context) {
	shopIDParam := c.Query("shop_id")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	query := h.DB.Table("shopee_orders").
		Joins("JOIN shopee_order_escrows ON shopee_orders.order_sn = shopee_order_escrows.order_sn")

	if shopIDParam != "" {
		if shopID, err := strconv.ParseUint(shopIDParam, 10, 64); err == nil {
			query = query.Where("shopee_orders.shop_id = ?", shopID)
		}
	}

	if startDate != "" {
		query = query.Where("shopee_orders.created_at >= ?", startDate)
	}
	if endDate != "" {
		query = query.Where("shopee_orders.created_at <= ?", endDate)
	}

	type SummaryResult struct {
		TotalOrders             int64   `gorm:"column:total_orders"`
		TotalGrossSales         float64 `gorm:"column:total_gross_sales"`
		TotalCommissionFee      float64 `gorm:"column:total_commission_fee"`
		TotalServiceFee         float64 `gorm:"column:total_service_fee"`
		TotalTxnFee             float64 `gorm:"column:total_txn_fee"`
		TotalProcessingFee      float64 `gorm:"column:total_processing_fee"`
		TotalVoucherDiscount    float64 `gorm:"column:total_voucher_discount"`
		TotalEscrowNetIn        float64 `gorm:"column:total_escrow_net_in"`
		TotalHPP                float64 `gorm:"column:total_hpp"`
		BucketFilament          float64 `gorm:"column:bucket_filament"`
		BucketHardwarePackaging float64 `gorm:"column:bucket_hardware_packaging"`
		BucketMachineCost       float64 `gorm:"column:bucket_machine_cost"`
		BucketNetProfit         float64 `gorm:"column:bucket_net_profit"`
	}

	var res SummaryResult
	err := query.Select(`
		COUNT(shopee_orders.order_sn) AS total_orders,
		COALESCE(SUM(shopee_order_escrows.selling_price), 0) AS total_gross_sales,
		COALESCE(SUM(shopee_order_escrows.commission_fee), 0) AS total_commission_fee,
		COALESCE(SUM(shopee_order_escrows.service_fee), 0) AS total_service_fee,
		COALESCE(SUM(shopee_order_escrows.seller_transaction_fee), 0) AS total_txn_fee,
		COALESCE(SUM(shopee_order_escrows.seller_order_processing_fee), 0) AS total_processing_fee,
		COALESCE(SUM(shopee_order_escrows.seller_voucher_discount), 0) AS total_voucher_discount,
		COALESCE(SUM(shopee_order_escrows.escrow_amount), 0) AS total_escrow_net_in,
		COALESCE(SUM(shopee_order_escrows.total_hpp), 0) AS total_hpp,
		COALESCE(SUM(shopee_order_escrows.total_filament_cost), 0) AS bucket_filament,
		COALESCE(SUM(shopee_order_escrows.total_hardware_packaging_cost), 0) AS bucket_hardware_packaging,
		COALESCE(SUM(shopee_order_escrows.total_machine_cost), 0) AS bucket_machine_cost,
		COALESCE(SUM(shopee_order_escrows.net_profit), 0) AS bucket_net_profit
	`).Scan(&res).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghitung ringkasan kas: " + err.Error(), "status": "error"})
		return
	}

	// Hitung jumlah item yang belum terpetakan (unmapped)
	var unmappedCount int64
	h.DB.Model(&models.ShopeeOrderItem{}).Where("mapping_status = 'UNMAPPED'").Count(&unmappedCount)

	totalMarketplaceFees := res.TotalCommissionFee + res.TotalServiceFee + res.TotalTxnFee + res.TotalProcessingFee + res.TotalVoucherDiscount

	avgMargin := 0.0
	if res.TotalGrossSales > 0 {
		avgMargin = (res.BucketNetProfit / res.TotalGrossSales) * 100
	}

	resp := models.CashflowSummaryResponse{
		TotalOrders:              res.TotalOrders,
		TotalGrossSales:          res.TotalGrossSales,
		TotalMarketplaceFees:     math.Round(totalMarketplaceFees*100) / 100,
		TotalEscrowNetIn:         res.TotalEscrowNetIn,
		TotalHPP:                 res.TotalHPP,
		BucketFilament:           res.BucketFilament,
		BucketHardwarePackaging:  res.BucketHardwarePackaging,
		BucketMachineElectricity: res.BucketMachineCost,
		BucketNetProfit:          res.BucketNetProfit,
		AverageProfitMargin:      math.Round(avgMargin*100) / 100,
		UnmappedItemsCount:       unmappedCount,
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   resp,
	})
}

// RecalculateFinances menghitung ulang seluruh pesanan Shopee di database menggunakan rumus alokasi kas & SKU matching terbaru
// POST /api/v1/shopee/financial/recalculate
func (h *OrderHandler) RecalculateFinances(c *gin.Context) {
	orderSN := c.Query("order_sn")

	var orders []models.ShopeeOrder
	query := h.DB.Preload("Items").Preload("Escrow")
	if orderSN != "" {
		query = query.Where("order_sn = ?", orderSN)
	}

	if err := query.Find(&orders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memuat pesanan", "status": "error"})
		return
	}

	recalculatedCount := 0
	for i := range orders {
		order := &orders[i]
		_ = costing.AllocateShopeeOrderFinances(order, h.DB)

		for _, it := range order.Items {
			_ = h.DB.Save(&it)
		}
		if order.Escrow != nil {
			_ = h.DB.Save(order.Escrow)
		}
		recalculatedCount++
	}

	c.JSON(http.StatusOK, gin.H{
		"message":            fmt.Sprintf("Berhasil menghitung ulang alokasi kas untuk %d pesanan", recalculatedCount),
		"recalculated_count": recalculatedCount,
		"status":             "success",
	})
}

