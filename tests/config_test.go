package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"symetra-lab-backend/handlers"
	"symetra-lab-backend/models"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestConfigHandlerShopAndMarketplaces(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := SetupTestDB(t)
	handler := handlers.NewConfigHandler(db)

	router := gin.New()
	router.GET("/api/v1/config/shop", handler.GetShopConfig)
	router.PUT("/api/v1/config/shop", handler.UpdateShopConfig)
	router.GET("/api/v1/config/marketplaces", handler.GetMarketplaces)
	router.POST("/api/v1/config/marketplaces", handler.CreateMarketplace)
	router.PUT("/api/v1/config/marketplaces/:id", handler.UpdateMarketplace)
	router.DELETE("/api/v1/config/marketplaces/:id", handler.DeleteMarketplace)

	// 1. Test GetShopConfig
	wGet := httptest.NewRecorder()
	reqGet, _ := http.NewRequest("GET", "/api/v1/config/shop", nil)
	router.ServeHTTP(wGet, reqGet)
	assert.Equal(t, http.StatusOK, wGet.Code)

	var getResp struct {
		Data models.ShopConfig `json:"data"`
	}
	_ = json.Unmarshal(wGet.Body.Bytes(), &getResp)
	assert.Equal(t, 1700.0, getResp.Data.ElectricityTariffPerKwh)

	// 2. Test UpdateShopConfig
	updatePayload := models.ShopConfig{
		ElectricityTariffPerKwh: 1850,
		FailureBufferPercent:    15,
		PrinterPowerWatts:       250,
	}
	upBytes, _ := json.Marshal(updatePayload)
	wPut := httptest.NewRecorder()
	reqPut, _ := http.NewRequest("PUT", "/api/v1/config/shop", bytes.NewBuffer(upBytes))
	reqPut.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(wPut, reqPut)
	assert.Equal(t, http.StatusOK, wPut.Code)

	var putResp struct {
		Data models.ShopConfig `json:"data"`
	}
	_ = json.Unmarshal(wPut.Body.Bytes(), &putResp)
	assert.Equal(t, 1850.0, putResp.Data.ElectricityTariffPerKwh)
	assert.Equal(t, 15.0, putResp.Data.FailureBufferPercent)

	// 3. Test CreateMarketplace
	mpPayload := models.MarketplacePlatform{
		Name:              "Tokopedia",
		CommissionPercent: 8.5,
		PromoFeePercent:   3.0,
		OrderFeeIDR:       1000,
		IsActive:          true,
	}
	mpBytes, _ := json.Marshal(mpPayload)
	wMpPost := httptest.NewRecorder()
	reqMpPost, _ := http.NewRequest("POST", "/api/v1/config/marketplaces", bytes.NewBuffer(mpBytes))
	reqMpPost.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(wMpPost, reqMpPost)
	assert.Equal(t, http.StatusCreated, wMpPost.Code)

	var mpPostResp struct {
		Data models.MarketplacePlatform `json:"data"`
	}
	_ = json.Unmarshal(wMpPost.Body.Bytes(), &mpPostResp)
	assert.Equal(t, "Tokopedia", mpPostResp.Data.Name)
	createdID := mpPostResp.Data.ID

	// 4. Test GetMarketplaces
	wMpList := httptest.NewRecorder()
	reqMpList, _ := http.NewRequest("GET", "/api/v1/config/marketplaces", nil)
	router.ServeHTTP(wMpList, reqMpList)
	assert.Equal(t, http.StatusOK, wMpList.Code)

	// 5. Test UpdateMarketplace
	upMpPayload := models.MarketplacePlatform{
		CommissionPercent: 9.0,
	}
	upMpBytes, _ := json.Marshal(upMpPayload)
	wMpPut := httptest.NewRecorder()
	reqMpPut, _ := http.NewRequest("PUT", "/api/v1/config/marketplaces/"+createdID, bytes.NewBuffer(upMpBytes))
	reqMpPut.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(wMpPut, reqMpPut)
	assert.Equal(t, http.StatusOK, wMpPut.Code)

	// 6. Test DeleteMarketplace
	wMpDel := httptest.NewRecorder()
	reqMpDel, _ := http.NewRequest("DELETE", "/api/v1/config/marketplaces/"+createdID, nil)
	router.ServeHTTP(wMpDel, reqMpDel)
	assert.Equal(t, http.StatusOK, wMpDel.Code)
}
