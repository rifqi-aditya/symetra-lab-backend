package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestPackagingItemCRUD(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTestDB(t)
	handler := NewPackagingHandler(db)

	router := gin.New()
	v1 := router.Group("/api/v1")
	{
		items := v1.Group("/packaging-items")
		{
			items.GET("", handler.GetAllPackagingItems)
			items.GET("/:id", handler.GetPackagingItemByID)
			items.POST("", handler.CreatePackagingItem)
			items.PUT("/:id", handler.UpdatePackagingItem)
			items.DELETE("/:id", handler.DeletePackagingItem)
		}
	}

	// 1. Test Create Packaging Item (POST /api/v1/packaging-items)
	// Beli 1 pack kardus diecut isi 50 seharga 25.750 -> unit_cost harus 515.0
	cat := "BOX"
	unitType := "PCS"
	qty := 50.0
	stock := 50.0
	createReq := CreatePackagingItemRequest{
		Name:             "Diecut Kecil",
		Category:         &cat,
		UnitType:         &unitType,
		PurchasePrice:    25750,
		PurchaseQuantity: &qty,
		StockQuantity:    &stock,
	}
	bodyBytes, _ := json.Marshal(createReq)

	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("POST", "/api/v1/packaging-items", bytes.NewBuffer(bodyBytes))
	req1.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w1, req1)

	if w1.Code != http.StatusCreated {
		t.Fatalf("Create packaging item gagal! Status: %d, Body: %s", w1.Code, w1.Body.String())
	}

	var createdResp map[string]interface{}
	_ = json.Unmarshal(w1.Body.Bytes(), &createdResp)
	data := createdResp["data"].(map[string]interface{})
	itemID := data["id"].(string)

	if data["unit_cost"].(float64) != 515.0 {
		t.Errorf("Ekspektasi unit_cost = 515.0, didapat: %v", data["unit_cost"])
	}

	// 2. Test Get All Packaging Items (GET /api/v1/packaging-items)
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/api/v1/packaging-items", nil)
	router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Fatalf("Get all packaging items gagal! Status: %d", w2.Code)
	}

	var listResp map[string]interface{}
	_ = json.Unmarshal(w2.Body.Bytes(), &listResp)
	if int(listResp["total"].(float64)) != 1 {
		t.Errorf("Ekspektasi total = 1, didapat: %v", listResp["total"])
	}

	// 3. Test Filter Kategori
	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest("GET", "/api/v1/packaging-items?category=BOX", nil)
	router.ServeHTTP(w3, req3)

	var filterResp map[string]interface{}
	_ = json.Unmarshal(w3.Body.Bytes(), &filterResp)
	if int(filterResp["total"].(float64)) != 1 {
		t.Errorf("Ekspektasi kategori BOX = 1, didapat: %v", filterResp["total"])
	}

	// 4. Test Update (PUT /api/v1/packaging-items/:id)
	newPrice := 30000.0 // 30.000 / 50 = 600
	updateReq := UpdatePackagingItemRequest{
		PurchasePrice: &newPrice,
	}
	updateBytes, _ := json.Marshal(updateReq)

	w4 := httptest.NewRecorder()
	req4, _ := http.NewRequest("PUT", "/api/v1/packaging-items/"+itemID, bytes.NewBuffer(updateBytes))
	req4.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w4, req4)

	if w4.Code != http.StatusOK {
		t.Fatalf("Update packaging item gagal! Status: %d", w4.Code)
	}

	var updateResp map[string]interface{}
	_ = json.Unmarshal(w4.Body.Bytes(), &updateResp)
	uData := updateResp["data"].(map[string]interface{})
	if uData["unit_cost"].(float64) != 600.0 {
		t.Errorf("Ekspektasi unit_cost baru = 600.0, didapat: %v", uData["unit_cost"])
	}

	// 5. Test Delete (DELETE /api/v1/packaging-items/:id)
	w5 := httptest.NewRecorder()
	req5, _ := http.NewRequest("DELETE", "/api/v1/packaging-items/"+itemID, nil)
	router.ServeHTTP(w5, req5)

	if w5.Code != http.StatusOK {
		t.Fatalf("Delete packaging item gagal! Status: %d", w5.Code)
	}
}

func TestPackagingPresetCRUDAndTotalCost(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTestDB(t)
	handler := NewPackagingHandler(db)

	router := gin.New()
	v1 := router.Group("/api/v1")
	{
		items := v1.Group("/packaging-items")
		{
			items.POST("", handler.CreatePackagingItem)
		}
		presets := v1.Group("/packaging-presets")
		{
			presets.GET("", handler.GetAllPresets)
			presets.GET("/:id", handler.GetPresetByID)
			presets.POST("", handler.CreatePreset)
			presets.PUT("/:id", handler.UpdatePreset)
			presets.DELETE("/:id", handler.DeletePreset)
		}
	}

	// 1. Buat 2 item bahan kemasan dulu
	item1Req := CreatePackagingItemRequest{
		Name:          "Plastik Clip Hologram",
		PurchasePrice: 464,
	}
	b1, _ := json.Marshal(item1Req)
	wItem1 := httptest.NewRecorder()
	rItem1, _ := http.NewRequest("POST", "/api/v1/packaging-items", bytes.NewBuffer(b1))
	rItem1.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(wItem1, rItem1)
	var resp1 map[string]interface{}
	_ = json.Unmarshal(wItem1.Body.Bytes(), &resp1)
	item1ID := resp1["data"].(map[string]interface{})["id"].(string)

	item2Req := CreatePackagingItemRequest{
		Name:          "Mailer Putih",
		PurchasePrice: 895,
	}
	b2, _ := json.Marshal(item2Req)
	wItem2 := httptest.NewRecorder()
	rItem2, _ := http.NewRequest("POST", "/api/v1/packaging-items", bytes.NewBuffer(b2))
	rItem2.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(wItem2, rItem2)
	var resp2 map[string]interface{}
	_ = json.Unmarshal(wItem2.Body.Bytes(), &resp2)
	item2ID := resp2["data"].(map[string]interface{})["id"].(string)

	// 2. Buat Preset "Keychain" yang berisi:
	// - 1x Plastik Clip Hologram (Rp 464)
	// - 1x Mailer Putih (Rp 895)
	// Total biaya packing harus = 1359.0
	desc := "Paket packing standar gantungan kunci"
	createPresetReq := CreatePresetRequest{
		Name:        "Keychain",
		Description: &desc,
		Items: []CreatePresetItemInput{
			{PackagingItemID: item1ID, QuantityUsed: 1},
			{PackagingItemID: item2ID, QuantityUsed: 1},
		},
	}
	presetBytes, _ := json.Marshal(createPresetReq)

	wPreset := httptest.NewRecorder()
	rPreset, _ := http.NewRequest("POST", "/api/v1/packaging-presets", bytes.NewBuffer(presetBytes))
	rPreset.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(wPreset, rPreset)

	if wPreset.Code != http.StatusCreated {
		t.Fatalf("Create preset gagal! Status: %d, Body: %s", wPreset.Code, wPreset.Body.String())
	}

	var pResp map[string]interface{}
	_ = json.Unmarshal(wPreset.Body.Bytes(), &pResp)
	pData := pResp["data"].(map[string]interface{})
	presetID := pData["id"].(string)

	if pData["total_cost"].(float64) != 1359.0 {
		t.Errorf("Ekspektasi total_cost = 1359.0 (464 + 895), didapat: %v", pData["total_cost"])
	}

	// 3. Test Get All Presets (GET /api/v1/packaging-presets)
	wAll := httptest.NewRecorder()
	rAll, _ := http.NewRequest("GET", "/api/v1/packaging-presets", nil)
	router.ServeHTTP(wAll, rAll)

	if wAll.Code != http.StatusOK {
		t.Fatalf("Get all presets gagal! Status: %d", wAll.Code)
	}

	// 4. Test Get Preset By ID (GET /api/v1/packaging-presets/:id)
	wSingle := httptest.NewRecorder()
	rSingle, _ := http.NewRequest("GET", "/api/v1/packaging-presets/"+presetID, nil)
	router.ServeHTTP(wSingle, rSingle)

	if wSingle.Code != http.StatusOK {
		t.Fatalf("Get preset by id gagal! Status: %d", wSingle.Code)
	}

	// 5. Test Delete Preset (DELETE /api/v1/packaging-presets/:id)
	wDel := httptest.NewRecorder()
	rDel, _ := http.NewRequest("DELETE", "/api/v1/packaging-presets/"+presetID, nil)
	router.ServeHTTP(wDel, rDel)

	if wDel.Code != http.StatusOK {
		t.Fatalf("Delete preset gagal! Status: %d", wDel.Code)
	}
}
