package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"symetra-lab-backend/pkg/shopee"

	"github.com/gin-gonic/gin"
)

func TestOrderHandlerInvalidShopID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	client := shopee.NewClient(123456, "secret", false, "http://localhost/callback")
	handler := NewOrderHandler(nil, client)

	router := gin.New()
	router.POST("/api/v1/shopee/shops/:shop_id/sync-orders", handler.SyncOrders)
	router.GET("/api/v1/shopee/shops/:shop_id/orders", handler.GetOrders)

	// Test 1: sync-orders with non-numeric shop_id
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("POST", "/api/v1/shopee/shops/invalid_id/sync-orders", nil)
	router.ServeHTTP(w1, req1)

	if w1.Code != http.StatusBadRequest {
		t.Errorf("Diharapkan status 400 Bad Request, didapat %d", w1.Code)
	}

	var resp1 map[string]interface{}
	_ = json.Unmarshal(w1.Body.Bytes(), &resp1)
	if resp1["status"] != "error" {
		t.Errorf("Diharapkan status 'error', didapat %v", resp1["status"])
	}

	// Test 2: get-orders with non-numeric shop_id
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/api/v1/shopee/shops/invalid_id/orders", nil)
	router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusBadRequest {
		t.Errorf("Diharapkan status 400 Bad Request, didapat %d", w2.Code)
	}
}
