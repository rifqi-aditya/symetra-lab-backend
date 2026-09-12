package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"symetra-lab-backend/models"
	"symetra-lab-backend/pkg/costing"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func setupTestProductDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	err = db.AutoMigrate(
		&models.ProductCategory{},
		&models.Product{},
		&models.ProductFilament{},
		&models.ProductComponent{},
		&models.ProductPackagingItem{},
		&models.FilamentProfile{},
		&models.Filament{},
		&models.Component{},
		&models.PackagingItem{},
		&models.PackagingPreset{},
		&models.PackagingPresetItem{},
		&models.Machine{},
		&models.ShopConfig{},
		&models.MarketplacePlatform{},
	)
	assert.NoError(t, err)

	// Seed shop config & marketplace
	db.Create(&models.ShopConfig{
		ID:                      "cfg-1",
		UserID:                  DefaultAdminUserID,
		FilamentPricePerRoll:    150000,
		FilamentWeightGrams:     1000,
		ElectricityTariffPerKwh: 1700,
		PrinterPowerWatts:       200,
		PrinterPrice:            5000000,
		PrinterLifespanHours:    3000,
		FailureBufferPercent:    10,
	})

	db.Create(&models.MarketplacePlatform{
		ID:                  "shopee-1",
		UserID:              DefaultAdminUserID,
		Name:                "Shopee",
		CommissionPercent:   9.5,
		PromoFeePercent:     4.5,
		FreeShippingPercent: 0,
		OrderFeeIDR:         1250,
		IsActive:            true,
	})

	return db
}

func TestFormatScalableSKU(t *testing.T) {
	// Skenario 1: Produk single belum punya varian -> otomatis dapat -STD
	parent, full := costing.FormatScalableSKU("BOX-B18650", "")
	assert.Equal(t, "BOX-B18650", parent)
	assert.Equal(t, "BOX-B18650-STD", full)

	// Skenario 2: Produk berkembang punya varian (4 Slot Hitam)
	parent2, full2 := costing.FormatScalableSKU("BOX-B18650", "4S-BLK")
	assert.Equal(t, "BOX-B18650", parent2)
	assert.Equal(t, "BOX-B18650-4S-BLK", full2)
}

func TestCalculateCostBreakdown(t *testing.T) {
	cfg := &models.ShopConfig{
		FilamentPricePerRoll:    150000, // Rp 150/gr
		FilamentWeightGrams:     1000,
		ElectricityTariffPerKwh: 1700,
		PrinterPowerWatts:       200,
		PrinterPrice:            5000000, // Rp 1.666,67/jam
		PrinterLifespanHours:    3000,
		FailureBufferPercent:    10, // 10%
	}

	shopee := &models.MarketplacePlatform{
		Name:              "Shopee",
		CommissionPercent: 9.5,
		PromoFeePercent:   4.5, // total 14%
		OrderFeeIDR:       1250,
		IsActive:          true,
	}

	prod := &models.Product{
		DefaultWeightGrams:    40,  // 40 gr + buffer 10% = 44 gr * 150 = 6600
		DefaultPrintTimeHours: 2.0, // Listrik: 0.2 kW * 2h * 1700 = 680. Depresiasi: (5jt/3000)*2 = 3333.33
		TargetMarginPercent:   30,
	}

	breakdown := costing.CalculateCostBreakdown(prod, cfg, shopee)

	assert.InDelta(t, 6600.0, breakdown.FilamentCost, 1.0)
	assert.InDelta(t, 680.0, breakdown.ElectricityCost, 1.0)
	assert.InDelta(t, 3333.33, breakdown.DepreciationCost, 1.0)
	assert.Greater(t, breakdown.BaseHPP, 10000.0)
	assert.Greater(t, breakdown.ShopeeRecommended, breakdown.BaseSellingPrice)
}

func TestProductCRUDAndBySKU(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTestProductDB(t)
	handler := NewProductHandler(db)

	r := gin.Default()
	r.GET("/products", handler.ListProducts)
	r.GET("/products/by-sku/:sku", handler.GetProductBySKU)
	r.POST("/products", handler.CreateProduct)
	r.DELETE("/products/:id", handler.DeleteProduct)

	// 1. Create Product
	parentSKU := "KEY-TOOTH"
	sku := "KEY-TOOTH-STD"
	payload := CreateProductPayload{
		Name:                  "Toothless Flexi Keychain",
		ParentSKU:             &parentSKU,
		SKU:                   &sku,
		Category:              "Keychain",
		DefaultWeightGrams:    15,
		DefaultPrintTimeHours: 1.0,
		TargetMarginPercent:   35,
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest(http.MethodPost, "/products", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var createResp struct {
		Data models.ProductResponse `json:"data"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &createResp)
	assert.NoError(t, err)
	assert.Equal(t, "KEY-TOOTH-STD", *createResp.Data.SKU)
	assert.Equal(t, "KEY-TOOTH", *createResp.Data.ParentSKU)

	// 2. Fetch by SKU
	req2, _ := http.NewRequest(http.MethodGet, "/products/by-sku/KEY-TOOTH-STD", nil)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusOK, w2.Code)
	var fetchResp struct {
		Data models.ProductResponse `json:"data"`
	}
	err = json.Unmarshal(w2.Body.Bytes(), &fetchResp)
	assert.NoError(t, err)
	assert.Equal(t, "Toothless Flexi Keychain", fetchResp.Data.Name)

	// 3. Test Duplicate SKU rejected
	reqDup, _ := http.NewRequest(http.MethodPost, "/products", bytes.NewBuffer(body))
	reqDup.Header.Set("Content-Type", "application/json")
	wDup := httptest.NewRecorder()
	r.ServeHTTP(wDup, reqDup)

	assert.Equal(t, http.StatusConflict, wDup.Code)

	// 4. Delete Product
	reqDel, _ := http.NewRequest(http.MethodDelete, "/products/"+createResp.Data.ID, nil)
	wDel := httptest.NewRecorder()
	r.ServeHTTP(wDel, reqDel)

	assert.Equal(t, http.StatusOK, wDel.Code)
}
