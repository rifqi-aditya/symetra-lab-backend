package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"symetra-lab-backend/models"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func setupTestProductionDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	err = db.AutoMigrate(
		&models.ShopeeOrder{},
		&models.ShopeeOrderItem{},
		&models.Order{},
		&models.OrderItem{},
		&models.OrderFilament{},
		&models.Product{},
		&models.ProductFilament{},
		&models.FilamentProfile{},
		&models.Filament{},
		&models.Machine{},
		&models.MachineMaintenancePart{},
		&models.ShopConfig{},
	)
	assert.NoError(t, err)

	// Machine with 0 hours used
	brand := "Bambu Lab"
	elecCost := 255.0
	machine := models.Machine{
		ID:                     "mach-prod-01",
		UserID:                 DefaultAdminUserID,
		Name:                   "Bambu Lab A1",
		Brand:                  &brand,
		TotalPurchaseCost:      7500000,
		LifespanHours:          10000,
		AvgPowerWatts:          150,
		ElectricityCostPerHour: &elecCost,
		TotalHoursUsed:         0,
		CurrentState:           "IDLE",
	}
	db.Create(&machine)

	// Filament with 1000g stock
	fil := models.Filament{
		ID:                "fil-prod-01",
		UserID:            DefaultAdminUserID,
		ColorName:         "Hitam PLA+",
		PricePerRoll:      188000,
		CurrentStockGrams: 1000.0,
	}
	db.Create(&fil)

	// Product (10g weight, 1.5h print time)
	sku := "KEY-SMILNIGH-STD"
	parentSKU := "KEY-SMILNIGH"
	prod := models.Product{
		ID:                    "prod-prod-01",
		UserID:                DefaultAdminUserID,
		Name:                  "Smiley Keychain",
		ParentSKU:             &parentSKU,
		SKU:                   &sku,
		DefaultWeightGrams:    10.0,
		DefaultPrintTimeHours: 1.5,
		DefaultMachineID:      &machine.ID,
	}
	db.Create(&prod)

	db.Create(&models.ProductFilament{
		ID:              "pf-prod-01",
		ProductID:       prod.ID,
		FilamentID:      &fil.ID,
		WeightUsedGrams: 10.0,
	})

	return db
}

func TestProductionQueueAndJobCompletion(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTestProductionDB(t)
	handler := NewProductionHandler(db)

	router := gin.New()
	router.GET("/api/v1/production/queue", handler.GetQueue)
	router.POST("/api/v1/production/complete-job", handler.CompleteJob)

	// 1. Seed 1 Shopee Order (READY_TO_SHIP)
	deadline := time.Now().Add(12 * time.Hour) // Urgent SLA!
	shopeeOrder := models.ShopeeOrder{
		OrderSN:        "SN-SHOPEE-QUEUE-1",
		ShopID:         711996297,
		OrderStatus:    "READY_TO_SHIP",
		BuyerUsername:  "shopee_customer",
		ShipByDateTime: &deadline,
		Items: []models.ShopeeOrderItem{
			{
				ID:        555,
				OrderSN:   "SN-SHOPEE-QUEUE-1",
				ItemName:  "Smiley Keychain Shopee",
				ModelSKU:  "KEY-SMILNIGH-STD",
				Quantity:  1,
				ProductID: stringPtr("prod-prod-01"),
			},
		},
	}
	db.Create(&shopeeOrder)

	// 2. Seed 1 Manual Order (PENDING)
	manualOrder := models.Order{
		ID:           "manual-ord-queue-1",
		UserID:       DefaultAdminUserID,
		OrderNumber:  "ORD-20260912-1111",
		CustomerName: "Customer Offline",
		Status:       "PENDING",
		CreatedAt:    time.Now(),
		Items: []models.OrderItem{
			{
				ID:             "manual-item-1",
				OrderID:        "manual-ord-queue-1",
				ProductID:      stringPtr("prod-prod-01"),
				ProductName:    "Smiley Keychain Offline",
				Quantity:       2,
				WeightGrams:    10.0,
				PrintTimeHours: 1.5,
				MachineID:      stringPtr("mach-prod-01"),
			},
		},
	}
	db.Create(&manualOrder)

	// 3. Test GetQueue (Antrean Cetak Terpadu)
	wQueue := httptest.NewRecorder()
	reqQueue, _ := http.NewRequest("GET", "/api/v1/production/queue", nil)
	router.ServeHTTP(wQueue, reqQueue)

	assert.Equal(t, http.StatusOK, wQueue.Code)

	var queueResp struct {
		Status string `json:"status"`
		Data   struct {
			Queue             []models.ProductionQueueItem `json:"queue"`
			TotalJobs         int                          `json:"total_jobs"`
			TotalHoursWaiting float64                      `json:"total_hours_waiting"`
			UrgentJobsCount   int                          `json:"urgent_jobs_count"`
		} `json:"data"`
	}
	_ = json.Unmarshal(wQueue.Body.Bytes(), &queueResp)
	assert.Equal(t, "success", queueResp.Status)
	assert.Equal(t, 2, queueResp.Data.TotalJobs)
	assert.GreaterOrEqual(t, queueResp.Data.UrgentJobsCount, 1)

	// 4. Test CompleteJob untuk manual job (Quantity: 2 x 10g = 20g filamen, 2 x 1.5h = 3.0h mesin)
	completePayload := models.CompletePrintJobRequest{
		Source: "MANUAL",
		JobID:  "manual-item-1",
	}
	compBytes, _ := json.Marshal(completePayload)
	wComp := httptest.NewRecorder()
	reqComp, _ := http.NewRequest("POST", "/api/v1/production/complete-job", bytes.NewBuffer(compBytes))
	reqComp.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(wComp, reqComp)

	assert.Equal(t, http.StatusOK, wComp.Code)

	var compResp struct {
		Status string                          `json:"status"`
		Data   models.CompletePrintJobResponse `json:"data"`
	}
	_ = json.Unmarshal(wComp.Body.Bytes(), &compResp)
	assert.Equal(t, "success", compResp.Status)
	assert.Equal(t, 3.0, compResp.Data.AddedMachineHours)

	// 5. Cek di DB bahwa stok filamen berkurang (1000g - 20g = 980g)
	var updatedFil models.Filament
	db.Where("id = ?", "fil-prod-01").First(&updatedFil)
	assert.Equal(t, 980.0, updatedFil.CurrentStockGrams)

	// 6. Cek di DB bahwa jam pakai mesin bertambah (+3 jam)
	var updatedMach models.Machine
	db.Where("id = ?", "mach-prod-01").First(&updatedMach)
	assert.Equal(t, 3.0, updatedMach.TotalHoursUsed)
}

func stringPtr(s string) *string {
	return &s
}
