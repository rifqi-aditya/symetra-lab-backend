package shopee

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetOrderList(t *testing.T) {
	partnerID := int64(123456)
	partnerKey := "secret_key"
	shopID := uint64(999888)
	accessToken := "test_access_token"

	// Mock server HTTP untuk Shopee API
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Validasi path
		if r.URL.Path != "/api/v2/order/get_order_list" {
			t.Errorf("Path tidak sesuai: didapat %s", r.URL.Path)
		}

		// Validasi query parameters
		q := r.URL.Query()
		if q.Get("partner_id") != "123456" {
			t.Errorf("partner_id tidak sesuai: %s", q.Get("partner_id"))
		}
		if q.Get("access_token") != accessToken {
			t.Errorf("access_token tidak sesuai: %s", q.Get("access_token"))
		}
		if q.Get("shop_id") != fmt.Sprintf("%d", shopID) {
			t.Errorf("shop_id tidak sesuai: %s", q.Get("shop_id"))
		}
		if q.Get("sign") == "" {
			t.Errorf("sign tidak boleh kosong")
		}

		// Response JSON mock
		jsonResp := `{
			"error": "",
			"message": "",
			"response": {
				"more": false,
				"next_cursor": "",
				"order_list": [
					{
						"order_sn": "260909TEST001",
						"order_status": "READY_TO_SHIP"
					}
				]
			}
		}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(jsonResp))
	}))
	defer mockServer.Close()

	client := NewClient(partnerID, partnerKey, false, "http://localhost/callback")
	client.BaseURL = mockServer.URL

	resp, err := client.GetOrderList(accessToken, shopID, 1788800000, 1788900000, "READY_TO_SHIP", "")
	if err != nil {
		t.Fatalf("GetOrderList mengembalikan error tak terduga: %v", err)
	}

	if len(resp.Response.OrderList) != 1 {
		t.Fatalf("Diharapkan 1 order, didapat %d", len(resp.Response.OrderList))
	}

	order := resp.Response.OrderList[0]
	if order.OrderSN != "260909TEST001" {
		t.Errorf("order_sn tidak sesuai: %s", order.OrderSN)
	}
	if order.OrderStatus != "READY_TO_SHIP" {
		t.Errorf("order_status tidak sesuai: %s", order.OrderStatus)
	}
}

func TestGetOrderDetail(t *testing.T) {
	partnerID := int64(123456)
	partnerKey := "secret_key"
	shopID := uint64(999888)
	accessToken := "test_access_token"

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/order/get_order_detail" {
			t.Errorf("Path tidak sesuai: %s", r.URL.Path)
		}

		q := r.URL.Query()
		if q.Get("order_sn_list") != "260909TEST001,260909TEST002" {
			t.Errorf("order_sn_list tidak sesuai: %s", q.Get("order_sn_list"))
		}

		jsonResp := `{
			"error": "",
			"message": "",
			"response": {
				"order_list": [
					{
						"order_sn": "260909TEST001",
						"order_status": "READY_TO_SHIP",
						"buyer_user_id": 45678,
						"buyer_username": "buyer_symetra",
						"message_to_seller": "Tolong cetak warna matte black infill 30%",
						"ship_by_date": 1789000000,
						"shipping_carrier": "SPX Standard",
						"total_amount": 150000.00,
						"item_list": [
							{
								"item_id": 1001,
								"item_name": "Stand Handphone 3D Articulated",
								"item_sku": "SKU-STAND-01",
								"model_id": 2001,
								"model_name": "Matte Black - PLA+",
								"model_sku": "SKU-STAND-BLK",
								"model_quantity_purchased": 2,
								"model_original_price": 75000.00,
								"model_discounted_price": 75000.00
							}
						]
					}
				]
			}
		}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(jsonResp))
	}))
	defer mockServer.Close()

	client := NewClient(partnerID, partnerKey, false, "http://localhost/callback")
	client.BaseURL = mockServer.URL

	resp, err := client.GetOrderDetail(accessToken, shopID, []string{"260909TEST001", "260909TEST002"})
	if err != nil {
		t.Fatalf("GetOrderDetail mengembalikan error tak terduga: %v", err)
	}

	if len(resp.Response.OrderList) != 1 {
		t.Fatalf("Diharapkan 1 order detail, didapat %d", len(resp.Response.OrderList))
	}

	detail := resp.Response.OrderList[0]
	if detail.OrderSN != "260909TEST001" {
		t.Errorf("OrderSN salah: %s", detail.OrderSN)
	}
	if detail.MessageToSeller != "Tolong cetak warna matte black infill 30%" {
		t.Errorf("Catatan kustom salah: %s", detail.MessageToSeller)
	}
	if len(detail.ItemList) != 1 {
		t.Fatalf("Diharapkan 1 item, didapat %d", len(detail.ItemList))
	}
	item := detail.ItemList[0]
	if item.ModelName != "Matte Black - PLA+" {
		t.Errorf("Nama varian salah: %s", item.ModelName)
	}
	if item.ModelQuantityPurchased != 2 {
		t.Errorf("Quantity salah: %d", item.ModelQuantityPurchased)
	}
}
