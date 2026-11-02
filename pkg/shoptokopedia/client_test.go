package shoptokopedia_test

import (
	"testing"

	"symetra-lab-backend-v2/pkg/shoptokopedia"
)

func TestClient_GenerateSignature(t *testing.T) {
	client := shoptokopedia.NewClient("test_app_key", "test_app_secret", false, "http://localhost:8080")

	path := "/api/v2/token/get"
	params := map[string]string{
		"app_key":    "test_app_key",
		"timestamp":  "1690000000",
		"auth_code":  "code_123",
		"grant_type": "authorized_code",
	}

	sig := client.GenerateSignature(path, params, nil)
	if sig == "" {
		t.Fatal("expected non-empty signature")
	}

	// Verify idempotency
	sig2 := client.GenerateSignature(path, params, nil)
	if sig != sig2 {
		t.Errorf("expected deterministic signature: %s != %s", sig, sig2)
	}
}

func TestConvertToStandardOrder(t *testing.T) {
	raw := shoptokopedia.RawOrderDetail{
		ID:     "1234567890",
		Status: "111", // AWAITING_SHIPMENT
		ShippingProvider: "J&T Express",
		TrackingNumber:   "JT99999",
		LineItems: []shoptokopedia.RawOrderLineItem{
			{
				ProductID:     "1001",
				SKUID:         "2001",
				ProductName:   "Articulated Slug",
				SellerSKU:     "SLUG-PURPLE",
				Quantity:      2,
				OriginalPrice: "25000",
				SalePrice:     "22000",
			},
		},
	}
	raw.Payment.SubTotal = "44000"
	raw.Payment.TotalAmount = "50000"
	raw.RecipientAddress.Name = "Budi Santoso"

	ord := shoptokopedia.ConvertToStandardOrder(raw, "TOKOPEDIA", 998877)

	if ord.Channel != "TOKOPEDIA" {
		t.Errorf("expected TOKOPEDIA channel, got %s", ord.Channel)
	}
	if ord.OrderStatus != "TO_SHIP" {
		t.Errorf("expected TO_SHIP status, got %s", ord.OrderStatus)
	}
	if ord.BuyerUsername != "Budi Santoso" {
		t.Errorf("expected Budi Santoso, got %s", ord.BuyerUsername)
	}
	if len(ord.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(ord.Items))
	}
	if ord.Items[0].CandidateSKU() != "SLUG-PURPLE" {
		t.Errorf("expected candidate SKU SLUG-PURPLE, got %s", ord.Items[0].CandidateSKU())
	}
	if ord.Financials.GrossAmount != 44000 {
		t.Errorf("expected GrossAmount 44000, got %f", ord.Financials.GrossAmount)
	}
}

