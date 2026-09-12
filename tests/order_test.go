package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"symetra-lab-backend/handlers"
	"symetra-lab-backend/models"
	"symetra-lab-backend/pkg/costing"
	"symetra-lab-backend/pkg/shopee"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestOrderHandlerInvalidShopID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	client := shopee.NewClient(123456, "secret", false, "http://localhost/callback")
	handler := handlers.NewOrderHandler(nil, client)

	router := gin.New()
	router.POST("/api/v1/shopee/shops/:shop_id/sync-orders", handler.SyncOrders)
	router.GET("/api/v1/shopee/shops/:shop_id/orders", handler.GetOrders)
	router.GET("/api/v1/shopee/shops/:shop_id/raw-orders", handler.GetRawOrders)

	// Test 1: sync-orders with non-numeric shop_id
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("POST", "/api/v1/shopee/shops/invalid_id/sync-orders", nil)
	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusBadRequest, w1.Code)

	// Test 2: get-orders with non-numeric shop_id
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/api/v1/shopee/shops/invalid_id/orders", nil)
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusBadRequest, w2.Code)

	// Test 3: raw-orders with non-numeric shop_id
	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest("GET", "/api/v1/shopee/shops/invalid_id/raw-orders", nil)
	router.ServeHTTP(w3, req3)
	assert.Equal(t, http.StatusBadRequest, w3.Code)
}

func TestAllocateShopeeOrderFinancesAndBuckets(t *testing.T) {
	db := SetupTestDB(t)

	// Seed Machine
	brand := "Bambu Lab"
	elecCost := 255.0
	machine := models.Machine{
		ID:                     "mach-01",
		UserID:                 DefaultAdminUserID,
		Name:                   "Bambu Lab A1",
		Brand:                  &brand,
		TotalPurchaseCost:      7500000,
		LifespanHours:          10000,
		AvgPowerWatts:          150,
		ElectricityCostPerHour: &elecCost,
	}
	db.Create(&machine)

	// Seed Filament
	filament := models.Filament{
		ID:           "fil-01",
		UserID:       DefaultAdminUserID,
		ColorName:    "Hitam PLA+",
		PricePerRoll: 188000,
	}
	db.Create(&filament)

	// Seed Component
	comp := models.Component{
		ID:           "comp-01",
		UserID:       DefaultAdminUserID,
		Name:         "Gantungan Kunci",
		PricePerUnit: 350,
	}
	db.Create(&comp)

	// Seed Product with SKU
	testSKU := "KEY-SMILNIGH-STD"
	testParentSKU := "KEY-SMILNIGH"
	prod := models.Product{
		ID:                    "prod-test-01",
		UserID:                DefaultAdminUserID,
		Name:                  "Smiley Keychain",
		ParentSKU:             &testParentSKU,
		SKU:                   &testSKU,
		DefaultWeightGrams:    10.0,
		DefaultPrintTimeHours: 1.0,
		DefaultMachineID:      &machine.ID,
		BaseHPP:               3500,
		BaseSellingPrice:      12000,
	}
	db.Create(&prod)

	// Link Product Filaments & Components
	db.Create(&models.ProductFilament{
		ID:              "pf-01",
		ProductID:       prod.ID,
		FilamentID:      &filament.ID,
		WeightUsedGrams: 10.0,
	})
	db.Create(&models.ProductComponent{
		ID:          "pc-01",
		ProductID:   prod.ID,
		ComponentID: &comp.ID,
		Quantity:    1,
	})

	order := models.ShopeeOrder{
		OrderSN:     "SN-TEST-001",
		ShopID:      711996297,
		OrderStatus: "COMPLETED",
		TotalAmount: 30000,
		Items: []models.ShopeeOrderItem{
			{
				OrderSN:         "SN-TEST-001",
				ItemID:          1001,
				ItemName:        "Smiley Night Fury Keychain",
				ModelSKU:        "KEY-SMILNIGH-STD", // Harus match dengan produk di DB!
				Quantity:        2,
				OriginalPrice:   15000,
				DiscountedPrice: 15000,
			},
		},
		Escrow: &models.ShopeeOrderEscrow{
			OrderSN:                  "SN-TEST-001",
			EscrowAmount:             25500, // Dana bersih masuk seller setelah potongan admin
			SellingPrice:             30000,
			CommissionFee:            2400,
			ServiceFee:               1200,
			SellerTransactionFee:     900,
			SellerOrderProcessingFee: 0,
		},
	}

	err := costing.AllocateShopeeOrderFinances(&order, db)
	assert.NoError(t, err)

	// Verifikasi Item Matched
	item := order.Items[0]
	assert.Equal(t, "MATCHED", item.MappingStatus)
	assert.Equal(t, "KEY-SMILNIGH-STD", item.MatchedSKU)
	assert.NotNil(t, item.ProductID)
	assert.Equal(t, "prod-test-01", *item.ProductID)
	assert.Greater(t, item.FilamentCost, 0.0)
	assert.Greater(t, item.HardwareCost, 0.0)
	assert.Greater(t, item.MachineCost, 0.0)
	assert.Greater(t, item.BaseHPP, 0.0)

	// Verifikasi Alokasi 5 Ember Kas pada Escrow
	escrow := order.Escrow
	assert.Greater(t, escrow.TotalHPP, 0.0)
	assert.Greater(t, escrow.TotalFilamentCost, 0.0)
	assert.Greater(t, escrow.TotalMachineCost, 0.0)
	// Laba bersih = EscrowAmount (25.500) - Total HPP
	expectedNetProfit := 25500.0 - escrow.TotalHPP
	assert.InDelta(t, expectedNetProfit, escrow.NetProfit, 0.05)
	assert.Greater(t, escrow.ProfitMarginPercent, 0.0)
}

func TestOrderHandlerLinkItemSKUAndCashflowSummary(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := SetupTestDB(t)

	// Seed Machine, Filament, Component, Product
	brand := "Bambu Lab"
	elecCost := 255.0
	machine := models.Machine{
		ID:                     "mach-01",
		UserID:                 DefaultAdminUserID,
		Name:                   "Bambu Lab A1",
		Brand:                  &brand,
		TotalPurchaseCost:      7500000,
		LifespanHours:          10000,
		AvgPowerWatts:          150,
		ElectricityCostPerHour: &elecCost,
	}
	db.Create(&machine)

	filament := models.Filament{
		ID:           "fil-01",
		UserID:       DefaultAdminUserID,
		ColorName:    "Hitam PLA+",
		PricePerRoll: 188000,
	}
	db.Create(&filament)

	comp := models.Component{
		ID:           "comp-01",
		UserID:       DefaultAdminUserID,
		Name:         "Gantungan Kunci",
		PricePerUnit: 350,
	}
	db.Create(&comp)

	testSKU := "KEY-SMILNIGH-STD"
	testParentSKU := "KEY-SMILNIGH"
	prod := models.Product{
		ID:                    "prod-test-01",
		UserID:                DefaultAdminUserID,
		Name:                  "Smiley Keychain",
		ParentSKU:             &testParentSKU,
		SKU:                   &testSKU,
		DefaultWeightGrams:    10.0,
		DefaultPrintTimeHours: 1.0,
		DefaultMachineID:      &machine.ID,
		BaseHPP:               3500,
		BaseSellingPrice:      12000,
	}
	db.Create(&prod)

	db.Create(&models.ProductFilament{
		ID:              "pf-01",
		ProductID:       prod.ID,
		FilamentID:      &filament.ID,
		WeightUsedGrams: 10.0,
	})
	db.Create(&models.ProductComponent{
		ID:          "pc-01",
		ProductID:   prod.ID,
		ComponentID: &comp.ID,
		Quantity:    1,
	})

	client := shopee.NewClient(123456, "secret", false, "http://localhost/callback")
	handler := handlers.NewOrderHandler(db, client)

	router := gin.New()
	router.POST("/api/v1/shopee/orders/:order_sn/items/:item_id/link-sku", handler.LinkItemSKU)
	router.GET("/api/v1/shopee/financial/cashflow-summary", handler.GetCashflowSummary)
	router.POST("/api/v1/shopee/financial/recalculate", handler.RecalculateFinances)

	// Seed pesanan dengan item UNMAPPED
	orderSN := "SN-UNMAPPED-999"
	order := models.ShopeeOrder{
		OrderSN:     orderSN,
		ShopID:      711996297,
		OrderStatus: "COMPLETED",
		TotalAmount: 18000,
		CreatedAt:   time.Now(),
	}
	db.Create(&order)

	item := models.ShopeeOrderItem{
		ID:              1,
		OrderSN:         orderSN,
		ItemID:          8888,
		ItemName:        "Gantungan Kunci Kustom Tanpa SKU",
		ModelSKU:        "", // Kosong
		Quantity:        1,
		OriginalPrice:   18000,
		DiscountedPrice: 18000,
		MappingStatus:   "UNMAPPED",
	}
	db.Create(&item)

	escrow := models.ShopeeOrderEscrow{
		OrderSN:      orderSN,
		EscrowAmount: 15500,
		SellingPrice: 18000,
		CreatedAt:    time.Now(),
	}
	db.Create(&escrow)

	// 1. Uji LinkItemSKU
	linkPayload := models.LinkSKURequest{
		ProductID: "prod-test-01",
	}
	payloadBytes, _ := json.Marshal(linkPayload)

	wLink := httptest.NewRecorder()
	reqLink, _ := http.NewRequest("POST", "/api/v1/shopee/orders/"+orderSN+"/items/8888/link-sku", bytes.NewBuffer(payloadBytes))
	reqLink.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(wLink, reqLink)

	assert.Equal(t, http.StatusOK, wLink.Code)

	var linkResp map[string]interface{}
	err := json.Unmarshal(wLink.Body.Bytes(), &linkResp)
	assert.NoError(t, err)
	assert.Equal(t, "success", linkResp["status"])

	// Cek bahwa di DB item sekarang berstatus MANUAL_LINKED dan memiliki base_hpp > 0
	var updatedItem models.ShopeeOrderItem
	db.Where("order_sn = ?", orderSN).First(&updatedItem)
	assert.Equal(t, "MANUAL_LINKED", updatedItem.MappingStatus)
	assert.Equal(t, "KEY-SMILNIGH-STD", updatedItem.MatchedSKU)
	assert.Greater(t, updatedItem.BaseHPP, 0.0)

	// 2. Uji Cashflow Summary Endpoint
	wSum := httptest.NewRecorder()
	reqSum, _ := http.NewRequest("GET", "/api/v1/shopee/financial/cashflow-summary?shop_id=711996297", nil)
	router.ServeHTTP(wSum, reqSum)

	assert.Equal(t, http.StatusOK, wSum.Code)

	var sumResp struct {
		Status string                         `json:"status"`
		Data   models.CashflowSummaryResponse `json:"data"`
	}
	err = json.Unmarshal(wSum.Body.Bytes(), &sumResp)
	assert.NoError(t, err)
	assert.Equal(t, "success", sumResp.Status)
	assert.Equal(t, int64(1), sumResp.Data.TotalOrders)
	assert.Equal(t, 18000.0, sumResp.Data.TotalGrossSales)
	assert.Equal(t, 15500.0, sumResp.Data.TotalEscrowNetIn)
	assert.Greater(t, sumResp.Data.BucketFilament, 0.0)
	assert.Greater(t, sumResp.Data.BucketNetProfit, 0.0)

	// 3. Uji Recalculate Finances Endpoint
	wRecalc := httptest.NewRecorder()
	reqRecalc, _ := http.NewRequest("POST", "/api/v1/shopee/financial/recalculate", nil)
	router.ServeHTTP(wRecalc, reqRecalc)

	assert.Equal(t, http.StatusOK, wRecalc.Code)
}
