package shoptokopedia_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	domainOrder "symetra-lab-backend-v2/internal/domain/order"
	"symetra-lab-backend-v2/internal/domain/product"
	usecaseProduct "symetra-lab-backend-v2/internal/usecase/product"
	usecaseTokopedia "symetra-lab-backend-v2/internal/usecase/shoptokopedia"
)

type mockProductRepo struct {
	bySKU map[string]*product.Product
	byID  map[uuid.UUID]*product.Product
}

func (m *mockProductRepo) FindAll(ctx context.Context, userID uuid.UUID, filter product.Filter) ([]*product.Product, int64, error) {
	return nil, 0, nil
}
func (m *mockProductRepo) FindByID(ctx context.Context, userID, id uuid.UUID) (*product.Product, error) {
	if p, ok := m.byID[id]; ok {
		return p, nil
	}
	return nil, product.ErrNotFound
}
func (m *mockProductRepo) FindBySKU(ctx context.Context, userID uuid.UUID, sku string) (*product.Product, error) {
	if p, ok := m.bySKU[sku]; ok {
		return p, nil
	}
	return nil, product.ErrNotFound
}
func (m *mockProductRepo) Create(ctx context.Context, p *product.Product) error { return nil }
func (m *mockProductRepo) Update(ctx context.Context, p *product.Product) error { return nil }
func (m *mockProductRepo) Delete(ctx context.Context, userID, id uuid.UUID) error { return nil }
func (m *mockProductRepo) GetTotalSoldMap(ctx context.Context, userID uuid.UUID) (map[uuid.UUID]int, error) {
	return nil, nil
}
func (m *mockProductRepo) GetSoldQtyByID(ctx context.Context, userID, productID uuid.UUID) (int, error) {
	return 0, nil
}
func (m *mockProductRepo) FindAllWithoutSKU(ctx context.Context, userID uuid.UUID) ([]*product.Product, error) {
	return nil, nil
}

type mockLedgerSync struct {
	calls int
}

func (m *mockLedgerSync) SyncRunningBalances(ctx context.Context) error {
	m.calls++
	return nil
}

func TestShopTokopediaUseCases_ProcessStandardOrder(t *testing.T) {
	prodID := uuid.New()
	sku := "SKU-CABLE-CLIP"
	p := product.ReconstructProduct(
		prodID,
		uuid.New(),
		"Desk Cable Clip",
		nil,
		&sku,
		nil,
		"Accessories",
		nil,
		nil,
		15,
		"PETG",
		0.5,
		nil,
		1,
		500,
		5000,
		12000,
		50,
		0,
		nil,
		nil,
		time.Now(),
		time.Now(),
	)

	pRepo := &mockProductRepo{
		bySKU: map[string]*product.Product{sku: p},
		byID:  map[uuid.UUID]*product.Product{prodID: p},
	}
	ledger := &mockLedgerSync{}

	costCalc := usecaseProduct.NewCostCalculator(nil, nil, nil, nil)
	matcher := usecaseProduct.NewSKUMatcher(pRepo, costCalc, nil)
	usecase := usecaseTokopedia.NewShopTokopediaUseCases(nil, nil, nil, matcher, ledger)

	stdOrder := &domainOrder.StandardMarketplaceOrder{
		Channel:     domainOrder.ChannelTokopedia,
		ShopID:      54321,
		OrderSN:     "TKPD-2026-001",
		OrderStatus: "TO_SHIP",
		Items: []domainOrder.StandardMarketplaceItem{
			{
				ChannelItemID:   777,
				ItemName:        "Tokopedia Desk Clip",
				ItemSKU:         sku,
				ModelSKU:        sku,
				Quantity:        3,
				OriginalPrice:   12000,
				DiscountedPrice: 10000,
			},
		},
	}

	ctx := context.Background()
	err := usecase.ProcessStandardOrder(ctx, stdOrder, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if stdOrder.Items[0].MappingStatus != "MATCHED" {
		t.Errorf("expected item mapping status MATCHED, got %s", stdOrder.Items[0].MappingStatus)
	}
	if stdOrder.Items[0].ProductID == nil || *stdOrder.Items[0].ProductID != prodID {
		t.Errorf("expected product ID %v, got %v", prodID, stdOrder.Items[0].ProductID)
	}
	if ledger.calls != 1 {
		t.Errorf("expected 1 ledger sync call, got %d", ledger.calls)
	}
}

