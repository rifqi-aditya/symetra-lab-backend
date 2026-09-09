package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestMachineCRUDAndMaintenance(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTestDB(t)
	handler := NewMachineHandler(db)

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
	createMachineReq := CreateMachineRequest{
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

	if w1.Code != http.StatusCreated {
		t.Fatalf("Create machine gagal! Status: %d, Body: %s", w1.Code, w1.Body.String())
	}

	var mResp map[string]interface{}
	_ = json.Unmarshal(w1.Body.Bytes(), &mResp)
	mData := mResp["data"].(map[string]interface{})
	machineID := mData["id"].(string)

	if mData["name"] != "Bambu Lab P1S Combo" {
		t.Errorf("Ekspektasi nama mesin 'Bambu Lab P1S Combo', didapat: %v", mData["name"])
	}

	// 2. Test Get All Machines (GET /api/v1/machines)
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/api/v1/machines", nil)
	router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Fatalf("Get all machines gagal! Status: %d", w2.Code)
	}

	// 3. Test Update State (PATCH /api/v1/machines/:id/state)
	stateReq := UpdateStateRequest{State: "PRINTING"}
	stateBytes, _ := json.Marshal(stateReq)

	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest("PATCH", "/api/v1/machines/"+machineID+"/state", bytes.NewBuffer(stateBytes))
	req3.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w3, req3)

	if w3.Code != http.StatusOK {
		t.Fatalf("Update machine state gagal! Status: %d, Body: %s", w3.Code, w3.Body.String())
	}

	var stateResp map[string]interface{}
	_ = json.Unmarshal(w3.Body.Bytes(), &stateResp)
	sData := stateResp["data"].(map[string]interface{})
	if sData["state"] != "PRINTING" {
		t.Errorf("Ekspektasi state 'PRINTING', didapat: %v", sData["state"])
	}

	// 4. Test Add Maintenance Part (POST /api/v1/machines/:id/parts)
	lifespan := 1000.0
	stock := 3
	hoursUsed := 250.0
	partReq := CreatePartRequest{
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

	if w4.Code != http.StatusCreated {
		t.Fatalf("Create part gagal! Status: %d, Body: %s", w4.Code, w4.Body.String())
	}

	var partCreatedResp map[string]interface{}
	_ = json.Unmarshal(w4.Body.Bytes(), &partCreatedResp)
	pData := partCreatedResp["data"].(map[string]interface{})
	partID := pData["id"].(string)

	// 5. Test Replace Maintenance Part (POST /api/v1/machines/parts/:part_id/replace)
	// Saat suku cadang diganti: jam pemakaian reset ke 0, stok berkurang dari 3 menjadi 2
	w5 := httptest.NewRecorder()
	req5, _ := http.NewRequest("POST", "/api/v1/machines/parts/"+partID+"/replace", nil)
	router.ServeHTTP(w5, req5)

	if w5.Code != http.StatusOK {
		t.Fatalf("Replace part gagal! Status: %d, Body: %s", w5.Code, w5.Body.String())
	}

	var replaceResp map[string]interface{}
	_ = json.Unmarshal(w5.Body.Bytes(), &replaceResp)
	rData := replaceResp["data"].(map[string]interface{})
	if rData["hours_used_current"].(float64) != 0 {
		t.Errorf("Ekspektasi hours_used_current reset ke 0, didapat: %v", rData["hours_used_current"])
	}
	if int(rData["remaining_stock"].(float64)) != 2 {
		t.Errorf("Ekspektasi remaining_stock berkurang jadi 2, didapat: %v", rData["remaining_stock"])
	}

	// 6. Test Delete Part (DELETE /api/v1/machines/parts/:part_id)
	w6 := httptest.NewRecorder()
	req6, _ := http.NewRequest("DELETE", "/api/v1/machines/parts/"+partID, nil)
	router.ServeHTTP(w6, req6)

	if w6.Code != http.StatusOK {
		t.Fatalf("Delete part gagal! Status: %d", w6.Code)
	}

	// 7. Test Delete Machine (DELETE /api/v1/machines/:id)
	w7 := httptest.NewRecorder()
	req7, _ := http.NewRequest("DELETE", "/api/v1/machines/"+machineID, nil)
	router.ServeHTTP(w7, req7)

	if w7.Code != http.StatusOK {
		t.Fatalf("Delete machine gagal! Status: %d", w7.Code)
	}
}
