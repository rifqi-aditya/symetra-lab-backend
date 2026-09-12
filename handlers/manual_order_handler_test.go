package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"symetra-lab-backend/models"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func setupTestManualOrderDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	err = db.AutoMigrate(
		&models.Order{},
		&models.OrderItem{},
		&models.OrderFilament{},
		&models.OrderComponent{},
		&models.Product{},
		&models.ProductFilament{},
		&models.ProductComponent{},
		&models.ProductPackagingItem{},
		&models.FilamentProfile{},
		&models.Filament{},
		&models.Machine{},
		&models.Component{},
		&models.ShopConfig{},
		&models.MarketplacePlatform{},
	)
	assert.NoError(t, err)

	// Seed shop config
	db.Create(&models.ShopConfig{
		ID:                      "cfg-manual",
		UserID:                  DefaultAdminUserID,
		FilamentPricePerRoll:    188000,
		FilamentWeightGrams:     1000,
		ElectricityTariffPerKwh: 1700,
		PrinterPowerWatts:       150,
		PrinterPrice:            7500000,
		PrinterLifespanHours:    10000,
		FailureBufferPercent:    10,
	})

	// Seed machine
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

	// Seed filament
	fil := models.Filament{
		ID:           "fil-01",
		UserID:       DefaultAdminUserID,
		ColorName:    "Hitam PLA+",
		PricePerRoll: 188000,
	}
	db.Create(&fil)

	// Seed product
	sku := "KEY-SMILNIGH-STD"
	parentSKU := "KEY-SMILNIGH"
	prod := models.Product{
		ID:                    "prod-manual-01",
		UserID:                DefaultAdminUserID,
		Name:                  "Smiley Keychain",
		ParentSKU:             &parentSKU,
		SKU:                   &sku,
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
		FilamentID:      &fil.ID,
		WeightUsedGrams: 10.0,
	})

	return db
}

func TestManualOrderCRUDAndStatusUpdates(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTestManualOrderDB(t)
	handler := NewManualOrderHandler(db)

	router := gin.New()
	router.GET("/api/v1/orders", handler.GetOrders)
	router.GET("/api/v1/orders/:id", handler.GetOrderByID)
	router.POST("/api/v1/orders", handler.CreateOrder)
	router.PATCH("/api/v1/orders/:id/status", handler.UpdateOrderStatus)
	router.PATCH("/api/v1/orders/:id/payment", handler.UpdatePaymentStatus)
	router.DELETE("/api/v1/orders/:id", handler.DeleteOrder)

	// 1. Test CreateOrder
	prodID := "prod-manual-01"
	createPayload := models.CreateManualOrderRequest{
		CustomerName:    "Budi Santoso",
		CustomerContact: "081234567890",
		Notes:           "Warna hitam doff ya",
		PaymentStatus:   "UNPAID",
		Items: []models.CreateManualOrderItemInput{
			{
				ProductID:    &prodID,
				ProductName:  "Smiley Keychain",
				Quantity:     2,
				SellingPrice: 15000,
			},
		},
	}
	bytesIn, _ := json.Marshal(createPayload)
	wPost := httptest.NewRecorder()
	reqPost, _ := http.NewRequest("POST", "/api/v1/orders", bytes.NewBuffer(bytesIn))
	reqPost.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(wPost, reqPost)

	assert.Equal(t, http.StatusCreated, wPost.Code)

	var postResp struct {
		Status string       `json:"status"`
		Data   models.Order `json:"data"`
	}
	_ = json.Unmarshal(wPost.Body.Bytes(), &postResp)
	assert.Equal(t, "success", postResp.Status)
	assert.Equal(t, "Budi Santoso", postResp.Data.CustomerName)
	assert.Equal(t, 30000.0, postResp.Data.TotalRevenue)
	assert.Greater(t, postResp.Data.TotalHPP, 0.0)
	assert.Greater(t, postResp.Data.TotalProfit, 0.0)
	orderID := postResp.Data.ID

	// 2. Test GetOrders
	wList := httptest.NewRecorder()
	reqList, _ := http.NewRequest("GET", "/api/v1/orders?search=Budi", nil)
	router.ServeHTTP(wList, reqList)
	assert.Equal(t, http.StatusOK, wList.Code)

	// 3. Test GetOrderByID
	wGet := httptest.NewRecorder()
	reqGet, _ := http.NewRequest("GET", "/api/v1/orders/"+orderID, nil)
	router.ServeHTTP(wGet, reqGet)
	assert.Equal(t, http.StatusOK, wGet.Code)

	// 4. Test UpdateOrderStatus
	statusPayload := models.UpdateOrderStatusRequest{
		Status: "IN_PRODUCTION",
	}
	stBytes, _ := json.Marshal(statusPayload)
	wPatchStatus := httptest.NewRecorder()
	reqPatchStatus, _ := http.NewRequest("PATCH", "/api/v1/orders/"+orderID+"/status", bytes.NewBuffer(stBytes))
	reqPatchStatus.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(wPatchStatus, reqPatchStatus)
	assert.Equal(t, http.StatusOK, wPatchStatus.Code)

	// 5. Test UpdatePaymentStatus
	payPayload := models.UpdatePaymentStatusRequest{
		PaymentStatus: "PAID",
	}
	payBytes, _ := json.Marshal(payPayload)
	wPatchPay := httptest.NewRecorder()
	reqPatchPay, _ := http.NewRequest("PATCH", "/api/v1/orders/"+orderID+"/payment", bytes.NewBuffer(payBytes))
	reqPatchPay.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(wPatchPay, reqPatchPay)
	assert.Equal(t, http.StatusOK, wPatchPay.Code)

	// 6. Test DeleteOrder
	wDel := httptest.NewRecorder()
	reqDel, _ := http.NewRequest("DELETE", "/api/v1/orders/"+orderID, nil)
	router.ServeHTTP(wDel, reqDel)
	assert.Equal(t, http.StatusOK, wDel.Code)
}
