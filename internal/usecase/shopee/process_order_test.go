package shopee_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"symetra-lab-backend-v2/internal/domain/product"
	"symetra-lab-backend-v2/internal/domain/shopee"
	usecaseProduct "symetra-lab-backend-v2/internal/usecase/product"
	usecaseShopee "symetra-lab-backend-v2/internal/usecase/shopee"
)

type mockShopeeRepo struct {
	orders   map[string]*shopee.ShopeeOrder
	saveErr  error
	linkErr  error
	histMaps map[uint64]*uuid.UUID
}

func (m *mockShopeeRepo) FindShop(ctx context.Context, shopID uint64) (*shopee.ShopeeShop, error) {
	return nil, nil
}
func (m *mockShopeeRepo) FindDefaultShop(ctx context.Context) (*shopee.ShopeeShop, error) {
	return nil, nil
}
func (m *mockShopeeRepo) FindAllShops(ctx context.Context) ([]*shopee.ShopeeShop, error) {
	return nil, nil
}
func (m *mockShopeeRepo) SaveShop(ctx context.Context, shop *shopee.ShopeeShop) error {
	return nil
}
func (m *mockShopeeRepo) FindOrder(ctx context.Context, orderSN string) (*shopee.ShopeeOrder, error) {
	if ord, ok := m.orders[orderSN]; ok {
		return ord, nil
	}
	return nil, nil
}
func (m *mockShopeeRepo) FindOrders(ctx context.Context, status, carrier, search string) ([]*shopee.ShopeeOrder, error) {
	return nil, nil
}
func (m *mockShopeeRepo) SaveOrder(ctx context.Context, order *shopee.ShopeeOrder) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	if m.orders == nil {
		m.orders = make(map[string]*shopee.ShopeeOrder)
	}
	m.orders[order.OrderSN()] = order
	return nil
}
func (m *mockShopeeRepo) LinkSKU(ctx context.Context, itemID, modelID uint64, productID uuid.UUID) error {
	return m.linkErr
}
func (m *mockShopeeRepo) FindHistoricalProductID(ctx context.Context, channelItemID, channelModelID uint64) (*uuid.UUID, error) {
	if id, ok := m.histMaps[channelItemID]; ok {
		return id, nil
	}
	return nil, nil
}
func (m *mockShopeeRepo) GetCashflowSummary(ctx context.Context) (*shopee.CashflowSummary, error) {
	return nil, nil
}
func (m *mockShopeeRepo) RecalculateFinances(ctx context.Context) (int, error) {
	return 0, nil
}
func (m *mockShopeeRepo) RecalculateSingleOrder(ctx context.Context, orderID uuid.UUID) error {
	return nil
}

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
	syncCalls int
}

func (m *mockLedgerSync) SyncRunningBalances(ctx context.Context) error {
	m.syncCalls++
	return nil
}

func TestProcessShopeeOrderUseCase_Execute(t *testing.T) {
	prodID := uuid.New()
	sku := "SKU-VIBRO-PAD"
	p := product.ReconstructProduct(
		prodID,
		uuid.New(),
		"Anti-Vibration Pad",
		nil,
		&sku,
		nil,
		"Accessories",
		nil,
		nil,
		50,
		"TPU",
		1.0,
		nil,
		1,
		500,
		15000,
		35000,
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
	sRepo := &mockShopeeRepo{
		orders:   make(map[string]*shopee.ShopeeOrder),
		histMaps: make(map[uint64]*uuid.UUID),
	}
	ledger := &mockLedgerSync{}

	costCalc := usecaseProduct.NewCostCalculator(nil, nil, nil, nil)
	matcher := usecaseProduct.NewSKUMatcher(pRepo, costCalc, sRepo)
	usecase := usecaseShopee.NewProcessShopeeOrderUseCase(sRepo, matcher, ledger)

	item := shopee.ReconstructOrderItem(
		1,
		"241029ABC123",
		1001,
		"Shopee Pad",
		sku,
		0,
		"",
		sku,
		2,
		35000,
		30000,
		nil,
		"",
		"UNMAPPED",
		0, 0, 0, 0, 0, 0, 0, 0, 0,
		time.Now(),
		time.Now(),
	)

	escrow := shopee.ReconstructOrderEscrow(
		"241029ABC123",
		55000, // Escrow amount
		60000, // Selling price
		3000,  // Commission fee
		"Std", 5.0,
		2000, // Service fee
		"Std", 3.0,
		0, 0, 0,
		5000, // Total marketplace fee
		0, 0, 0, 0, 0, 0, 0, 0,
		"RELEASED",
		time.Now(),
		time.Now(),
	)

	order := shopee.ReconstructOrder(
		"241029ABC123",
		123,
		"COMPLETED",
		456,
		"buyer1",
		"",
		0,
		nil,
		"J&T",
		"JT123456",
		60000,
		"",
		0, 0, 0,
		[]shopee.ShopeeOrderItem{item},
		escrow,
		time.Now(),
		time.Now(),
	)

	ctx := context.Background()
	err := usecase.Execute(ctx, order, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify items mapped
	savedOrder := sRepo.orders["241029ABC123"]
	if savedOrder == nil {
		t.Fatal("expected order to be saved in repository")
	}

	savedItems := savedOrder.Items()
	if len(savedItems) != 1 {
		t.Fatalf("expected 1 item, got %d", len(savedItems))
	}
	if savedItems[0].MappingStatus() != "MATCHED" {
		t.Errorf("expected item mapping status MATCHED, got %s", savedItems[0].MappingStatus())
	}
	if savedItems[0].ProductID() == nil || *savedItems[0].ProductID() != prodID {
		t.Errorf("expected product ID %v, got %v", prodID, savedItems[0].ProductID())
	}

	// Verify ledger synchronizer was called
	if ledger.syncCalls != 1 {
		t.Errorf("expected 1 ledger sync call, got %d", ledger.syncCalls)
	}

	// Verify escrow funds computed
	savedEscrow := savedOrder.Escrow()
	if savedEscrow == nil {
		t.Fatal("expected escrow to be populated")
	}
	if savedEscrow.KasFilamen() <= 0 {
		t.Errorf("expected KasFilamen > 0, got %f", savedEscrow.KasFilamen())
	}
	if savedEscrow.KasCadanganRisiko() <= 0 {
		t.Errorf("expected KasCadanganRisiko > 0, got %f", savedEscrow.KasCadanganRisiko())
	}
}

