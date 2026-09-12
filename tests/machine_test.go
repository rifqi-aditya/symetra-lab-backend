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

func TestMachineCRUDAndMaintenance(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := SetupTestDB(t)
	handler := handlers.NewMachineHandler(db)

	router := gin.New()
	v1 := router.Group("/api/v1")
	{
		machines := v1.Group("/machines")
		{
			machines.GET("", handler.GetAllMachines)
			machines.GET("/:id", handler.GetMachineByID)
			machines.POST("", handler.CreateMachine)
			machines.PUT("/:id", handler.UpdateMachine)
			machines.PATCH("/:id/state", handler.UpdateMachineState)
			machines.DELETE("/:id", handler.DeleteMachine)

			// Sub-rute suku cadang
			machines.GET("/:id/parts", handler.GetPartsByMachine)
			machines.POST("/:id/parts", handler.CreatePart)
			machines.PUT("/parts/:part_id", handler.UpdatePart)
			machines.POST("/parts/:part_id/replace", handler.ReplacePart)
			machines.DELETE("/parts/:part_id", handler.DeletePart)
		}
	}

	// 1. Test Create Machine (POST /api/v1/machines)
	brand := "Bambu Lab"
	createMachineReq := handlers.CreateMachineRequest{
		Name:              "Bambu Lab P1S Combo",
		Brand:             &brand,
		TotalPurchaseCost: 12500000,
		AvgPowerWatts:     350,
	}
	bodyBytes, _ := json.Marshal(createMachineReq)

	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("POST", "/api/v1/machines", bytes.NewBuffer(bodyBytes))
	req1.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w1, req1)

	assert.Equal(t, http.StatusCreated, w1.Code)

	var mResp map[string]interface{}
	_ = json.Unmarshal(w1.Body.Bytes(), &mResp)
	mData := mResp["data"].(map[string]interface{})
	machineID := mData["id"].(string)

	assert.Equal(t, "Bambu Lab P1S Combo", mData["name"])

	// 2. Test Get All Machines (GET /api/v1/machines)
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/api/v1/machines", nil)
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)

	// 3. Test Update State (PATCH /api/v1/machines/:id/state)
	stateReq := handlers.UpdateStateRequest{State: "PRINTING"}
	stateBytes, _ := json.Marshal(stateReq)

	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest("PATCH", "/api/v1/machines/"+machineID+"/state", bytes.NewBuffer(stateBytes))
	req3.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w3, req3)

	assert.Equal(t, http.StatusOK, w3.Code)

	var stateResp map[string]interface{}
	_ = json.Unmarshal(w3.Body.Bytes(), &stateResp)
	sData := stateResp["data"].(map[string]interface{})
	assert.Equal(t, "PRINTING", sData["state"])

	// 4. Test Add Maintenance Part (POST /api/v1/machines/:id/parts)
	lifespan := 1000.0
	stock := 3
	hoursUsed := 250.0
	partReq := handlers.CreatePartRequest{
		PartName:         "Hardened Steel Nozzle 0.4mm",
		CostIDR:          185000,
		LifespanHours:    &lifespan,
		StockQuantity:    &stock,
		HoursUsedCurrent: &hoursUsed,
	}
	partBytes, _ := json.Marshal(partReq)

	w4 := httptest.NewRecorder()
	req4, _ := http.NewRequest("POST", "/api/v1/machines/"+machineID+"/parts", bytes.NewBuffer(partBytes))
	req4.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w4, req4)

	assert.Equal(t, http.StatusCreated, w4.Code)

	var partCreatedResp map[string]interface{}
	_ = json.Unmarshal(w4.Body.Bytes(), &partCreatedResp)
	pData := partCreatedResp["data"].(map[string]interface{})
	partID := pData["id"].(string)

	// 5. Test Replace Maintenance Part (POST /api/v1/machines/parts/:part_id/replace)
	w5 := httptest.NewRecorder()
	req5, _ := http.NewRequest("POST", "/api/v1/machines/parts/"+partID+"/replace", nil)
	router.ServeHTTP(w5, req5)

	assert.Equal(t, http.StatusOK, w5.Code)

	var replaceResp map[string]interface{}
	_ = json.Unmarshal(w5.Body.Bytes(), &replaceResp)
	rData := replaceResp["data"].(map[string]interface{})
	assert.Equal(t, 0.0, rData["hours_used_current"])
	assert.Equal(t, 2, int(rData["remaining_stock"].(float64)))

	// 6. Test Delete Part (DELETE /api/v1/machines/parts/:part_id)
	w6 := httptest.NewRecorder()
	req6, _ := http.NewRequest("DELETE", "/api/v1/machines/parts/"+partID, nil)
	router.ServeHTTP(w6, req6)
	assert.Equal(t, http.StatusOK, w6.Code)

	// 7. Test Delete Machine (DELETE /api/v1/machines/:id)
	w7 := httptest.NewRecorder()
	req7, _ := http.NewRequest("DELETE", "/api/v1/machines/"+machineID, nil)
	router.ServeHTTP(w7, req7)
	assert.Equal(t, http.StatusOK, w7.Code)
}
