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
	"gorm.io/gorm"
)

// setupTestDB membuat in-memory SQLite database terisolasi untuk unit testing
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Gagal inisialisasi in-memory SQLite: %v", err)
	}

	err = db.AutoMigrate(
		&models.FilamentProfile{},
		&models.Filament{},
		&models.Machine{},
		&models.MachineMaintenancePart{},
	)
	if err != nil {
		t.Fatalf("Gagal auto-migrate tabel test: %v", err)
	}

	return db
}

func TestFilamentCRUDAndSyncStock(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTestDB(t)
	handler := NewFilamentHandler(db)

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

	createReq := CreateFilamentRequest{
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

	if w1.Code != http.StatusCreated {
		t.Fatalf("Create filamen gagal! Status: %d, Body: %s", w1.Code, w1.Body.String())
	}

	var createdResp map[string]interface{}
	_ = json.Unmarshal(w1.Body.Bytes(), &createdResp)
	data := createdResp["data"].(map[string]interface{})
	filamentID := data["id"].(string)

	if data["color_name"] != "Matte Navy Blue" {
		t.Errorf("Ekspektasi color_name 'Matte Navy Blue', didapat: %v", data["color_name"])
	}
	if data["brand"] != "Sunlu" {
		t.Errorf("Ekspektasi brand 'Sunlu' ter-flatten, didapat: %v", data["brand"])
	}

	// 2. Test Get All Filaments (GET /api/v1/filaments)
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/api/v1/filaments", nil)
	router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Fatalf("Get all filaments gagal! Status: %d", w2.Code)
	}

	var listResp map[string]interface{}
	_ = json.Unmarshal(w2.Body.Bytes(), &listResp)
	total := int(listResp["total"].(float64))
	if total != 1 {
		t.Errorf("Ekspektasi total filamen = 1, didapat: %d", total)
	}

	// 3. Test Sync Stock Timbangan Fisik (POST /api/v1/filaments/:id/sync-stock)
	// Berat kotor di timbangan = 750 gram. Berat spool kosong = 200 gram -> Sisa filamen = 550 gram.
	syncReq := SyncStockRequest{
		GrossWeightGrams: 750.0,
	}
	syncBytes, _ := json.Marshal(syncReq)

	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest("POST", "/api/v1/filaments/"+filamentID+"/sync-stock", bytes.NewBuffer(syncBytes))
	req3.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w3, req3)

	if w3.Code != http.StatusOK {
		t.Fatalf("Sync stock gagal! Status: %d, Body: %s", w3.Code, w3.Body.String())
	}

	var syncResp map[string]interface{}
	_ = json.Unmarshal(w3.Body.Bytes(), &syncResp)
	syncData := syncResp["data"].(map[string]interface{})
	calculatedNet := syncData["calculated_net_grams"].(float64)
	if calculatedNet != 550.0 {
		t.Errorf("Ekspektasi calculated_net_grams = 550.0 (750 - 200), didapat: %v", calculatedNet)
	}

	// 4. Test Update Filament (PUT /api/v1/filaments/:id)
	newPrice := 150000.0
	newColor := "Deep Navy Blue"
	updateReq := UpdateFilamentRequest{
		ColorName:    &newColor,
		PricePerRoll: &newPrice,
	}
	updateBytes, _ := json.Marshal(updateReq)

	w4 := httptest.NewRecorder()
	req4, _ := http.NewRequest("PUT", "/api/v1/filaments/"+filamentID, bytes.NewBuffer(updateBytes))
	req4.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w4, req4)

	if w4.Code != http.StatusOK {
		t.Fatalf("Update filamen gagal! Status: %d", w4.Code)
	}

	// 5. Test Get Master Profiles (GET /api/v1/filament-profiles)
	w5 := httptest.NewRecorder()
	req5, _ := http.NewRequest("GET", "/api/v1/filament-profiles", nil)
	router.ServeHTTP(w5, req5)

	if w5.Code != http.StatusOK {
		t.Fatalf("Get profiles gagal! Status: %d", w5.Code)
	}

	// 6. Test Delete Filament (DELETE /api/v1/filaments/:id)
	w6 := httptest.NewRecorder()
	req6, _ := http.NewRequest("DELETE", "/api/v1/filaments/"+filamentID, nil)
	router.ServeHTTP(w6, req6)

	if w6.Code != http.StatusOK {
		t.Fatalf("Delete filamen gagal! Status: %d", w6.Code)
	}

	// Verifikasi filamen sudah terhapus
	w7 := httptest.NewRecorder()
	req7, _ := http.NewRequest("GET", "/api/v1/filaments/"+filamentID, nil)
	router.ServeHTTP(w7, req7)

	if w7.Code != http.StatusNotFound {
		t.Errorf("Ekspektasi 404 Not Found setelah delete, didapat: %d", w7.Code)
	}
}
