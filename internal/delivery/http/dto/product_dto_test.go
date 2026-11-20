package dto_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"symetra-lab-backend-v2/internal/delivery/http/dto"
	"symetra-lab-backend-v2/internal/domain/costing"
	"symetra-lab-backend-v2/internal/domain/product"
)

func TestToProductResponse_VirtualPlateFields(t *testing.T) {
	now := time.Now()
	pID := uuid.New()
	parentID := uuid.New()
	sku := "SKU-TEST-001"

	// Skenario:
	// Produk per-unit memiliki berat 20.0 gram dan durasi cetak 1.0 jam.
	// Diproduksi dalam 1 plate slicer berisi 5 pcs (batch_size = 5).
	p := product.ReconstructProduct(
		pID,
		parentID,
		"Standard",
		&sku,
		"PLA",
		20.0, // default_weight_grams (per unit)
		1.0,  // default_print_time_hours (per unit)
		5,    // batch_size
		30,
		5000,
		7500,
		nil,
		nil,
		nil,
		now,
		now,
	)

	resp := dto.ToProductResponse(p, 10, costing.ProductCostBreakdown{})

	// Verifikasi nilai per unit dipertahankan
	if resp.DefaultWeightGrams != 20.0 {
		t.Errorf("expected DefaultWeightGrams 20.0, got %f", resp.DefaultWeightGrams)
	}
	if resp.DefaultPrintTimeHours != 1.0 {
		t.Errorf("expected DefaultPrintTimeHours 1.0, got %f", resp.DefaultPrintTimeHours)
	}

	// Verifikasi nilai virtual 1 plate slicer dihitung dengan benar (20 * 5 = 100g, 1 * 5 = 5 jam)
	expectedPlateWeight := 100.0
	if resp.PlateWeightGrams != expectedPlateWeight {
		t.Errorf("expected PlateWeightGrams %f, got %f", expectedPlateWeight, resp.PlateWeightGrams)
	}

	expectedPlateHours := 5.0
	if resp.PlatePrintTimeHours != expectedPlateHours {
		t.Errorf("expected PlatePrintTimeHours %f, got %f", expectedPlateHours, resp.PlatePrintTimeHours)
	}
}
