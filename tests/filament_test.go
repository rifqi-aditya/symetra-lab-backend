package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"symetra-lab-backend/handlers"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestFilamentCRUDAndSyncStock(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := SetupTestDB(t)
	handler := handlers.NewFilamentHandler(db)

	router := gin.New()
	v1 := router.Group("/api/v1")
	{
		filaments := v1.Group("/filaments")
		{
			filaments.GET("", handler.GetAllFilaments)
			filaments.GET("/:id", handler.GetFilamentByID)
			filaments.POST("", handler.CreateFilament)
			filaments.PUT("/:id", handler.UpdateFilament)
			filaments.DELETE("/:id", handler.DeleteFilament)
			filaments.POST("/:id/sync-stock", handler.SyncStock)
		}
		v1.GET("/filament-profiles", handler.GetProfiles)
	}

	// 1. Test Create Filament (POST /api/v1/filaments)
	emptySpool := 200.0
	nozzleTemp := 215
	bedTemp := 60
	spoolWeight := 1000.0

	createReq := handlers.CreateFilamentRequest{
		Brand:                 "Sunlu",
		MaterialType:          "PLA+",
		NozzleTemp:            &nozzleTemp,
		BedTemp:               &bedTemp,
		EmptySpoolWeightGrams: &emptySpool,
		SpoolWeightGrams:      &spoolWeight,
		ColorName:             "Matte Navy Blue",
		ColorHex:              "#1B263B",
		PricePerRoll:          145000,
	}
	bodyBytes, _ := json.Marshal(createReq)

	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("POST", "/api/v1/filaments", bytes.NewBuffer(bodyBytes))
	req1.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w1, req1)

	assert.Equal(t, http.StatusCreated, w1.Code)

	var createdResp map[string]interface{}
	_ = json.Unmarshal(w1.Body.Bytes(), &createdResp)
	data := createdResp["data"].(map[string]interface{})
	filamentID := data["id"].(string)

	assert.Equal(t, "Matte Navy Blue", data["color_name"])
	assert.Equal(t, "Sunlu", data["brand"])

	// 2. Test Get All Filaments (GET /api/v1/filaments)
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/api/v1/filaments", nil)
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)

	var listResp map[string]interface{}
	_ = json.Unmarshal(w2.Body.Bytes(), &listResp)
	total := int(listResp["total"].(float64))
	assert.Equal(t, 1, total)

	// 3. Test Sync Stock Timbangan Fisik (POST /api/v1/filaments/:id/sync-stock)
	syncReq := handlers.SyncStockRequest{
		GrossWeightGrams: 750.0,
	}
	syncBytes, _ := json.Marshal(syncReq)

	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest("POST", "/api/v1/filaments/"+filamentID+"/sync-stock", bytes.NewBuffer(syncBytes))
	req3.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w3, req3)
	assert.Equal(t, http.StatusOK, w3.Code)

	var syncResp map[string]interface{}
	_ = json.Unmarshal(w3.Body.Bytes(), &syncResp)
	syncData := syncResp["data"].(map[string]interface{})
	calculatedNet := syncData["calculated_net_grams"].(float64)
	assert.Equal(t, 550.0, calculatedNet)

	// 4. Test Update Filament (PUT /api/v1/filaments/:id)
	newPrice := 150000.0
	newColor := "Deep Navy Blue"
	updateReq := handlers.UpdateFilamentRequest{
		ColorName:    &newColor,
		PricePerRoll: &newPrice,
	}
	updateBytes, _ := json.Marshal(updateReq)

	w4 := httptest.NewRecorder()
	req4, _ := http.NewRequest("PUT", "/api/v1/filaments/"+filamentID, bytes.NewBuffer(updateBytes))
	req4.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w4, req4)
	assert.Equal(t, http.StatusOK, w4.Code)

	// 5. Test Get Master Profiles (GET /api/v1/filament-profiles)
	w5 := httptest.NewRecorder()
	req5, _ := http.NewRequest("GET", "/api/v1/filament-profiles", nil)
	router.ServeHTTP(w5, req5)
	assert.Equal(t, http.StatusOK, w5.Code)

	// 6. Test Delete Filament (DELETE /api/v1/filaments/:id)
	w6 := httptest.NewRecorder()
	req6, _ := http.NewRequest("DELETE", "/api/v1/filaments/"+filamentID, nil)
	router.ServeHTTP(w6, req6)
	assert.Equal(t, http.StatusOK, w6.Code)

	// Verifikasi filamen sudah terhapus
	w7 := httptest.NewRecorder()
	req7, _ := http.NewRequest("GET", "/api/v1/filaments/"+filamentID, nil)
	router.ServeHTTP(w7, req7)
	assert.Equal(t, http.StatusNotFound, w7.Code)
}
