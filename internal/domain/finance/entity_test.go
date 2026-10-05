package finance_test

import (
	"testing"
	"time"

	"symetra-lab-backend-v2/internal/domain/finance"
)

func TestNewFinanceTransaction(t *testing.T) {
	now := time.Now()

	// Valid transaction
	tx, err := finance.NewFinanceTransaction(
		finance.TypeIncome,
		finance.CategorySalesShopee,
		150000,
		"Order Shopee 123",
		now,
		"SHOPEE_ESCROW",
		"123",
		"Test notes",
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if tx.Amount() != 150000 {
		t.Errorf("expected amount 150000, got %f", tx.Amount())
	}
	if tx.Type() != finance.TypeIncome {
		t.Errorf("expected TypeIncome, got %s", tx.Type())
	}

	// Invalid: zero amount
	_, err = finance.NewFinanceTransaction(
		finance.TypeIncome,
		finance.CategorySalesShopee,
		0,
		"Invalid amount",
		now,
		"", "", "",
	)
	if err != finance.ErrAmountMustBePositive {
		t.Errorf("expected ErrAmountMustBePositive, got %v", err)
	}

	// Invalid: empty description
	_, err = finance.NewFinanceTransaction(
		finance.TypeExpense,
		finance.CategoryFilament,
		50000,
		"",
		now,
		"", "", "",
	)
	if err != finance.ErrDescriptionRequired {
		t.Errorf("expected ErrDescriptionRequired, got %v", err)
	}
}

func TestNewPurchaseOrder(t *testing.T) {
	now := time.Now()

	item1 := finance.NewPurchaseOrderItem(finance.PurchaseItemFilament, "PLA+ Black", 2, "kg", 120000)
	item2 := finance.NewPurchaseOrderItem(finance.PurchaseItemHardware, "Baut M3", 50, "pcs", 500)

	if item1.TotalPrice() != 240000 {
		t.Errorf("expected item1 total 240000, got %f", item1.TotalPrice())
	}
	if item2.TotalPrice() != 25000 {
		t.Errorf("expected item2 total 25000, got %f", item2.TotalPrice())
	}

	po, err := finance.NewPurchaseOrder("Toko 3D", now, "Pembelian filamen", []finance.PurchaseOrderItem{item1, item2})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if po.TotalAmount() != 265000 {
		t.Errorf("expected total amount 265000, got %f", po.TotalAmount())
	}
	if len(po.Items()) != 2 {
		t.Errorf("expected 2 items, got %d", len(po.Items()))
	}

	// Missing supplier
	_, err = finance.NewPurchaseOrder("", now, "", []finance.PurchaseOrderItem{item1})
	if err != finance.ErrSupplierNameRequired {
		t.Errorf("expected ErrSupplierNameRequired, got %v", err)
	}

	// Empty items
	_, err = finance.NewPurchaseOrder("Toko 3D", now, "", nil)
	if err != finance.ErrPurchaseItemsRequired {
		t.Errorf("expected ErrPurchaseItemsRequired, got %v", err)
	}
}

func TestNewCapitalRecord(t *testing.T) {
	now := time.Now()

	cr, err := finance.NewCapitalRecord(finance.CapitalInitial, 5000000, "Modal awal operasional", now, "Dari tabungan")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cr.Amount() != 5000000 {
		t.Errorf("expected 5000000, got %f", cr.Amount())
	}
	if cr.RecordType() != finance.CapitalInitial {
		t.Errorf("expected CapitalInitial, got %s", cr.RecordType())
	}

	// Invalid: negative amount
	_, err = finance.NewCapitalRecord(finance.CapitalAddition, -1000, "Invalid", now, "")
	if err != finance.ErrAmountMustBePositive {
		t.Errorf("expected ErrAmountMustBePositive, got %v", err)
	}
}
