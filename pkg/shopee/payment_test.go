package shopee

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetEscrowDetail(t *testing.T) {
	partnerID := int64(123456)
	partnerKey := "secret_key"
	shopID := uint64(999888)
	accessToken := "test_access_token"
	orderSN := "260909TEST001"

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/payment/get_escrow_detail" {
			t.Errorf("Path tidak sesuai: %s", r.URL.Path)
		}

		q := r.URL.Query()
		if q.Get("order_sn") != orderSN {
			t.Errorf("order_sn tidak sesuai: %s", q.Get("order_sn"))
		}

		jsonResp := `{
			"error": "",
			"message": "",
			"response": {
				"order_sn": "260909TEST001",
				"order_income": {
					"escrow_amount": 131800.00,
					"buyer_total_amount": 150000.00,
					"order_selling_price": 150000.00,
					"selling_price": 150000.00,
					"seller_order_processing_fee": 1250.00,
					"commission_fee": 9750.00,
					"service_fee": 6000.00,
					"seller_transaction_fee": 6000.00,
					"voucher_from_seller": 0.00,
					"seller_discount": 0.00,
					"net_commission_fee_info": [
						{
							"rule_id": 101,
							"fee_amount": 9750.00,
							"rule_display_name": "Biaya Komisi Kategori Hobi & Koleksi"
						}
					],
					"net_service_fee_info": [
						{
							"rule_id": 202,
							"fee_amount": 6000.00,
							"rule_display_name": "Program Gratis Ongkir XTRA",
							"category": "Free Shipping Special"
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

	resp, err := client.GetEscrowDetail(accessToken, shopID, orderSN)
	if err != nil {
		t.Fatalf("GetEscrowDetail mengembalikan error tak terduga: %v", err)
	}

	income := resp.Response.OrderIncome
	if income.EscrowAmount != 131800.00 {
		t.Errorf("EscrowAmount salah: %.2f", income.EscrowAmount)
	}
	if income.SellerOrderProcessingFee != 1250.00 {
		t.Errorf("Biaya per pesanan (Rp 1.250) salah: %.2f", income.SellerOrderProcessingFee)
	}
	if income.CommissionFee != 9750.00 {
		t.Errorf("CommissionFee salah: %.2f", income.CommissionFee)
	}
	if len(income.NetCommissionFeeInfo) != 1 || income.NetCommissionFeeInfo[0].RuleDisplayName != "Biaya Komisi Kategori Hobi & Koleksi" {
		t.Errorf("Nama aturan komisi kategori salah")
	}

	// Validasi kalkulasi persentase
	sellingPrice := income.OrderSellingPrice
	commPct := (income.CommissionFee / sellingPrice) * 100
	if commPct != 6.5 {
		t.Errorf("Kalkulasi persentase komisi kategori salah: diharapkan 6.5%%, didapat %.2f%%", commPct)
	}

	servPct := (income.ServiceFee / sellingPrice) * 100
	if servPct != 4.0 {
		t.Errorf("Kalkulasi persentase layanan XTRA salah: diharapkan 4.0%%, didapat %.2f%%", servPct)
	}
}
