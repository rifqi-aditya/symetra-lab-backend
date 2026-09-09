package shopee

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetShippingParameter(t *testing.T) {
	partnerID := int64(123456)
	partnerKey := "secret_key"
	shopID := uint64(999888)
	accessToken := "test_token"
	orderSN := "260909LOGIS001"

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/logistics/get_shipping_parameter" {
			t.Errorf("Path tidak sesuai: %s", r.URL.Path)
		}
		if r.URL.Query().Get("order_sn") != orderSN {
			t.Errorf("order_sn query salah: %s", r.URL.Query().Get("order_sn"))
		}

		jsonResp := `{
			"error": "",
			"message": "",
			"response": {
				"info_needed": {
					"dropoff": ["branch_id"],
					"pickup": []
				},
				"dropoff": {
					"branch_list": [
						{
							"branch_id": 5001,
							"region": "ID",
							"state": "DKI Jakarta",
							"city": "Jakarta Selatan",
							"address": "SPX Hub Tebet"
						}
					]
				}
			}
		}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(jsonResp))
	}))
	defer mockServer.Close()

	client := NewClient(partnerID, partnerKey, false, "http://localhost/callback")
	client.BaseURL = mockServer.URL

	resp, err := client.GetShippingParameter(accessToken, shopID, orderSN)
	if err != nil {
		t.Fatalf("GetShippingParameter mengembalikan error: %v", err)
	}

	if len(resp.Response.Dropoff.BranchList) != 1 {
		t.Fatalf("Diharapkan 1 dropoff branch, didapat %d", len(resp.Response.Dropoff.BranchList))
	}

	branch := resp.Response.Dropoff.BranchList[0]
	if branch.BranchID != 5001 {
		t.Errorf("BranchID salah: %d", branch.BranchID)
	}
	if branch.Address != "SPX Hub Tebet" {
		t.Errorf("Alamat cabang salah: %s", branch.Address)
	}
}

func TestShipOrder(t *testing.T) {
	partnerID := int64(123456)
	partnerKey := "secret_key"
	shopID := uint64(999888)
	accessToken := "test_token"
	orderSN := "260909LOGIS001"

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/logistics/ship_order" {
			t.Errorf("Path tidak sesuai: %s", r.URL.Path)
		}

		jsonResp := `{
			"error": "",
			"message": "success",
			"request_id": "req_ship_123"
		}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(jsonResp))
	}))
	defer mockServer.Close()

	client := NewClient(partnerID, partnerKey, false, "http://localhost/callback")
	client.BaseURL = mockServer.URL

	req := ShopeeShipOrderRequest{
		OrderSN: orderSN,
		Dropoff: &ShopeeShipDropoff{
			BranchID: 5001,
		},
	}

	resp, err := client.ShipOrder(accessToken, shopID, req)
	if err != nil {
		t.Fatalf("ShipOrder error: %v", err)
	}

	if resp.Error != "" {
		t.Errorf("ShipOrder mengembalikan error: %s", resp.Error)
	}
}

func TestGetTrackingNumber(t *testing.T) {
	partnerID := int64(123456)
	partnerKey := "secret_key"
	shopID := uint64(999888)
	accessToken := "test_token"
	orderSN := "260909LOGIS001"

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/logistics/get_tracking_number" {
			t.Errorf("Path tidak sesuai: %s", r.URL.Path)
		}

		jsonResp := `{
			"error": "",
			"message": "",
			"response": {
				"tracking_number": "SPXID04829103948"
			}
		}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(jsonResp))
	}))
	defer mockServer.Close()

	client := NewClient(partnerID, partnerKey, false, "http://localhost/callback")
	client.BaseURL = mockServer.URL

	resp, err := client.GetTrackingNumber(accessToken, shopID, orderSN)
	if err != nil {
		t.Fatalf("GetTrackingNumber error: %v", err)
	}

	if resp.Response.TrackingNumber != "SPXID04829103948" {
		t.Errorf("Tracking number salah: %s", resp.Response.TrackingNumber)
	}
}

func TestDownloadShippingDocument(t *testing.T) {
	partnerID := int64(123456)
	partnerKey := "secret_key"
	shopID := uint64(999888)
	accessToken := "test_token"
	orderSN := "260909LOGIS001"

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/logistics/download_shipping_document" {
			t.Errorf("Path tidak sesuai: %s", r.URL.Path)
		}

		// Kirim mock byte PDF
		w.Header().Set("Content-Type", "application/pdf")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("%PDF-1.4 Mock Thermal Label"))
	}))
	defer mockServer.Close()

	client := NewClient(partnerID, partnerKey, false, "http://localhost/callback")
	client.BaseURL = mockServer.URL

	pdfBytes, err := client.DownloadShippingDocument(accessToken, shopID, orderSN)
	if err != nil {
		t.Fatalf("DownloadShippingDocument error: %v", err)
	}

	if string(pdfBytes) != "%PDF-1.4 Mock Thermal Label" {
		t.Errorf("Isi PDF tidak sesuai: %s", string(pdfBytes))
	}
}
