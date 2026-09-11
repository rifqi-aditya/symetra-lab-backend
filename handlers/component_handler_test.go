package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestComponentCRUDAndCalculations(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTestDB(t)
	handler := NewComponentHandler(db)

	router := gin.New()
	v1 := router.Group("/api/v1")
	{
		comps := v1.Group("/components")
		{
			comps.GET("", handler.GetAllComponents)
			comps.GET("/:id", handler.GetComponentByID)
			comps.POST("", handler.CreateComponent)
			comps.PUT("/:id", handler.UpdateComponent)
			comps.DELETE("/:id", handler.DeleteComponent)
		}
	}

	// 1. Test Create Component (POST /api/v1/components)
	markup := 25.0
	desc := "Blue Switch Keyboard Tactile Clicky"
	createReq := CreateComponentRequest{
		Name:                 "Blue Switch Keyboard",
		PricePerUnit:         1600,
		DefaultMarkupPercent: &markup,
		Description:          &desc,
	}
	bodyBytes, _ := json.Marshal(createReq)

	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("POST", "/api/v1/components", bytes.NewBuffer(bodyBytes))
	req1.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w1, req1)

	if w1.Code != http.StatusCreated {
		t.Fatalf("Create component gagal! Status: %d, Body: %s", w1.Code, w1.Body.String())
	}

	var createdResp map[string]interface{}
	_ = json.Unmarshal(w1.Body.Bytes(), &createdResp)
	cData := createdResp["data"].(map[string]interface{})
	compID := cData["id"].(string)

	if cData["name"] != "Blue Switch Keyboard" {
		t.Errorf("Ekspektasi nama 'Blue Switch Keyboard', didapat: %v", cData["name"])
	}

	// Cek kalkulasi selling price: 1600 * (1 + 0.25) = 2000
	sellingPrice := cData["calculated_selling_price"].(float64)
	if sellingPrice != 2000.0 {
		t.Errorf("Ekspektasi calculated_selling_price = 2000.0, didapat: %v", sellingPrice)
	}

	// 2. Test Get All Components (GET /api/v1/components)
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/api/v1/components", nil)
	router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Fatalf("Get all components gagal! Status: %d", w2.Code)
	}

	var listResp map[string]interface{}
	_ = json.Unmarshal(w2.Body.Bytes(), &listResp)
	if int(listResp["total"].(float64)) != 1 {
		t.Errorf("Ekspektasi total = 1, didapat: %v", listResp["total"])
	}

	// 3. Test Search Query Filter
	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest("GET", "/api/v1/components?q=Switch", nil)
	router.ServeHTTP(w3, req3)

	var searchResp map[string]interface{}
	_ = json.Unmarshal(w3.Body.Bytes(), &searchResp)
	if int(searchResp["total"].(float64)) != 1 {
		t.Errorf("Ekspektasi search 'Switch' menghasilkan 1 item, didapat: %v", searchResp["total"])
	}

	w3NotFound := httptest.NewRecorder()
	req3NotFound, _ := http.NewRequest("GET", "/api/v1/components?q=Dinamo", nil)
	router.ServeHTTP(w3NotFound, req3NotFound)

	var searchEmptyResp map[string]interface{}
	_ = json.Unmarshal(w3NotFound.Body.Bytes(), &searchEmptyResp)
	if int(searchEmptyResp["total"].(float64)) != 0 {
		t.Errorf("Ekspektasi search 'Dinamo' menghasilkan 0 item, didapat: %v", searchEmptyResp["total"])
	}

	// 4. Test Get Component By ID (GET /api/v1/components/:id)
	w4 := httptest.NewRecorder()
	req4, _ := http.NewRequest("GET", "/api/v1/components/"+compID, nil)
	router.ServeHTTP(w4, req4)

	if w4.Code != http.StatusOK {
		t.Fatalf("Get component by id gagal! Status: %d", w4.Code)
	}

	// 5. Test Update Component (PUT /api/v1/components/:id)
	// Ubah harga per unit menjadi 2000 dan markup menjadi 50% -> selling price = 3000
	newPrice := 2000.0
	newMarkup := 50.0
	updateReq := UpdateComponentRequest{
		PricePerUnit:         &newPrice,
		DefaultMarkupPercent: &newMarkup,
	}
	updateBytes, _ := json.Marshal(updateReq)

	w5 := httptest.NewRecorder()
	req5, _ := http.NewRequest("PUT", "/api/v1/components/"+compID, bytes.NewBuffer(updateBytes))
	req5.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w5, req5)

	if w5.Code != http.StatusOK {
		t.Fatalf("Update component gagal! Status: %d, Body: %s", w5.Code, w5.Body.String())
	}

	var updateResp map[string]interface{}
	_ = json.Unmarshal(w5.Body.Bytes(), &updateResp)
	uData := updateResp["data"].(map[string]interface{})
	if uData["calculated_selling_price"].(float64) != 3000.0 {
		t.Errorf("Ekspektasi calculated_selling_price pasca-update = 3000.0, didapat: %v", uData["calculated_selling_price"])
	}

	// 6. Test Delete Component (DELETE /api/v1/components/:id)
	w6 := httptest.NewRecorder()
	req6, _ := http.NewRequest("DELETE", "/api/v1/components/"+compID, nil)
	router.ServeHTTP(w6, req6)

	if w6.Code != http.StatusOK {
		t.Fatalf("Delete component gagal! Status: %d", w6.Code)
	}

	// Verifikasi sudah terhapus (404)
	w7 := httptest.NewRecorder()
	req7, _ := http.NewRequest("GET", "/api/v1/components/"+compID, nil)
	router.ServeHTTP(w7, req7)

	if w7.Code != http.StatusNotFound {
		t.Errorf("Ekspektasi 404 setelah delete, didapat: %d", w7.Code)
	}
}
