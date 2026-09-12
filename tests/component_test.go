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

func TestComponentCRUDAndCalculations(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := SetupTestDB(t)
	handler := handlers.NewComponentHandler(db)

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
	createReq := handlers.CreateComponentRequest{
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

	assert.Equal(t, http.StatusCreated, w1.Code)

	var createdResp map[string]interface{}
	_ = json.Unmarshal(w1.Body.Bytes(), &createdResp)
	cData := createdResp["data"].(map[string]interface{})
	compID := cData["id"].(string)

	assert.Equal(t, "Blue Switch Keyboard", cData["name"])
	assert.Equal(t, 2000.0, cData["calculated_selling_price"])

	// 2. Test Get All Components (GET /api/v1/components)
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/api/v1/components", nil)
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)

	var listResp map[string]interface{}
	_ = json.Unmarshal(w2.Body.Bytes(), &listResp)
	assert.Equal(t, 1, int(listResp["total"].(float64)))

	// 3. Test Search Query Filter
	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest("GET", "/api/v1/components?q=Switch", nil)
	router.ServeHTTP(w3, req3)

	var searchResp map[string]interface{}
	_ = json.Unmarshal(w3.Body.Bytes(), &searchResp)
	assert.Equal(t, 1, int(searchResp["total"].(float64)))

	w3NotFound := httptest.NewRecorder()
	req3NotFound, _ := http.NewRequest("GET", "/api/v1/components?q=Dinamo", nil)
	router.ServeHTTP(w3NotFound, req3NotFound)

	var searchEmptyResp map[string]interface{}
	_ = json.Unmarshal(w3NotFound.Body.Bytes(), &searchEmptyResp)
	assert.Equal(t, 0, int(searchEmptyResp["total"].(float64)))

	// 4. Test Get Component By ID (GET /api/v1/components/:id)
	w4 := httptest.NewRecorder()
	req4, _ := http.NewRequest("GET", "/api/v1/components/"+compID, nil)
	router.ServeHTTP(w4, req4)
	assert.Equal(t, http.StatusOK, w4.Code)

	// 5. Test Update Component (PUT /api/v1/components/:id)
	newPrice := 2000.0
	newMarkup := 50.0
	updateReq := handlers.UpdateComponentRequest{
		PricePerUnit:         &newPrice,
		DefaultMarkupPercent: &newMarkup,
	}
	updateBytes, _ := json.Marshal(updateReq)

	w5 := httptest.NewRecorder()
	req5, _ := http.NewRequest("PUT", "/api/v1/components/"+compID, bytes.NewBuffer(updateBytes))
	req5.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w5, req5)
	assert.Equal(t, http.StatusOK, w5.Code)

	var updateResp map[string]interface{}
	_ = json.Unmarshal(w5.Body.Bytes(), &updateResp)
	uData := updateResp["data"].(map[string]interface{})
	assert.Equal(t, 3000.0, uData["calculated_selling_price"])

	// 6. Test Delete Component (DELETE /api/v1/components/:id)
	w6 := httptest.NewRecorder()
	req6, _ := http.NewRequest("DELETE", "/api/v1/components/"+compID, nil)
	router.ServeHTTP(w6, req6)
	assert.Equal(t, http.StatusOK, w6.Code)

	// Verifikasi sudah terhapus (404)
	w7 := httptest.NewRecorder()
	req7, _ := http.NewRequest("GET", "/api/v1/components/"+compID, nil)
	router.ServeHTTP(w7, req7)
	assert.Equal(t, http.StatusNotFound, w7.Code)
}
