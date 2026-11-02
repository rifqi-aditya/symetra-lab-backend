package product_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"symetra-lab-backend-v2/internal/domain/product"
	usecaseProduct "symetra-lab-backend-v2/internal/usecase/product"
)

type mockProductRepo struct {
	products map[string]*product.Product
	byID     map[uuid.UUID]*product.Product
}

func (m *mockProductRepo) FindAll(ctx context.Context, userID uuid.UUID, filter product.Filter) ([]*product.Product, int64, error) {
	var list []*product.Product
	for _, p := range m.byID {
		list = append(list, p)
	}
	return list, int64(len(list)), nil
}

func (m *mockProductRepo) FindByID(ctx context.Context, userID, id uuid.UUID) (*product.Product, error) {
	if p, ok := m.byID[id]; ok {
		return p, nil
	}
	return nil, product.ErrNotFound
}

func (m *mockProductRepo) FindBySKU(ctx context.Context, userID uuid.UUID, sku string) (*product.Product, error) {
	if p, ok := m.products[sku]; ok {
		return p, nil
	}
	return nil, product.ErrNotFound
}

func (m *mockProductRepo) Create(ctx context.Context, p *product.Product) error { return nil }
func (m *mockProductRepo) Update(ctx context.Context, p *product.Product) error { return nil }
func (m *mockProductRepo) Delete(ctx context.Context, userID, id uuid.UUID) error { return nil }
func (m *mockProductRepo) GetTotalSoldMap(ctx context.Context, userID uuid.UUID) (map[uuid.UUID]int, error) {
	return map[uuid.UUID]int{}, nil
}
func (m *mockProductRepo) GetSoldQtyByID(ctx context.Context, userID, productID uuid.UUID) (int, error) {
	return 0, nil
}
func (m *mockProductRepo) FindAllWithoutSKU(ctx context.Context, userID uuid.UUID) ([]*product.Product, error) {
	return nil, nil
}

type mockMappingFinder struct {
	hist map[uint64]*uuid.UUID
}

func (m *mockMappingFinder) FindHistoricalProductID(ctx context.Context, channelItemID, channelModelID uint64) (*uuid.UUID, error) {
	if id, ok := m.hist[channelItemID]; ok {
		return id, nil
	}
	return nil, nil
}

func TestSKUMatcher_MatchBySKU(t *testing.T) {
	prodID := uuid.New()
	sku := "SKU-3D-DRAGON"
	p := product.ReconstructProduct(
		prodID,
		uuid.New(),
		"Articulated Dragon",
		nil,
		&sku,
		nil,
		"Toys",
		nil,
		nil,
		150,
		"PLA",
		4.5,
		nil,
		nil,
		1,
		2000,
		50000,
		100000,
		50,
		0,
		nil,
		nil,
		time.Now(),
		time.Now(),
	)

	repo := &mockProductRepo{
		products: map[string]*product.Product{sku: p},
		byID:     map[uuid.UUID]*product.Product{prodID: p},
	}

	calc := usecaseProduct.NewCostCalculator(nil, nil, nil, nil)
	matcher := usecaseProduct.NewSKUMatcher(repo, calc, nil)

	ctx := context.Background()
	res := matcher.MatchItem(ctx, usecaseProduct.MatchItemInput{
		UserID:        uuid.New(),
		ModelSKU:      sku,
		ChannelItemID: 12345,
		ItemName:      "Shopee Dragon Toy",
		Quantity:      2,
		SellingPrice:  90000,
	}, nil)

	if res.MappingStatus != "MATCHED" {
		t.Errorf("expected MATCHED, got %s", res.MappingStatus)
	}
	if res.ProductID == nil || *res.ProductID != prodID {
		t.Errorf("expected product ID %v, got %v", prodID, res.ProductID)
	}
	if res.MatchedSKU != sku {
		t.Errorf("expected SKU %s, got %s", sku, res.MatchedSKU)
	}
	if res.ProductName != "Articulated Dragon" {
		t.Errorf("expected Articulated Dragon, got %s", res.ProductName)
	}
	if res.WeightGrams != 300.0 {
		t.Errorf("expected 300g, got %f", res.WeightGrams)
	}
}

func TestSKUMatcher_HistoricalFallback(t *testing.T) {
	prodID := uuid.New()
	sku := "SKU-BENCHY"
	p := product.ReconstructProduct(
		prodID,
		uuid.New(),
		"3DBenchy",
		nil,
		&sku,
		nil,
		"Boats",
		nil,
		nil,
		20,
		"PLA",
		1.0,
		nil,
		nil,
		1,
		1000,
		10000,
		25000,
		40,
		0,
		nil,
		nil,
		time.Now(),
		time.Now(),
	)

	repo := &mockProductRepo{
		products: map[string]*product.Product{sku: p},
		byID:     map[uuid.UUID]*product.Product{prodID: p},
	}

	finder := &mockMappingFinder{
		hist: map[uint64]*uuid.UUID{99999: &prodID},
	}

	calc := usecaseProduct.NewCostCalculator(nil, nil, nil, nil)
	matcher := usecaseProduct.NewSKUMatcher(repo, calc, finder)

	ctx := context.Background()
	// Candidate SKU is unknown/empty, but channel_item_id = 99999 is in historical mapping
	res := matcher.MatchItem(ctx, usecaseProduct.MatchItemInput{
		UserID:        uuid.New(),
		ModelSKU:      "",
		ItemSKU:       "",
		ChannelItemID: 99999,
		ItemName:      "Custom Miniature Boat",
		Quantity:      1,
		SellingPrice:  25000,
	}, nil)

	if res.MappingStatus != "MATCHED" {
		t.Errorf("expected MATCHED, got %s", res.MappingStatus)
	}
	if res.ProductID == nil || *res.ProductID != prodID {
		t.Errorf("expected product ID %v, got %v", prodID, res.ProductID)
	}
	if res.MatchedSKU != sku {
		t.Errorf("expected SKU %s, got %s", sku, res.MatchedSKU)
	}
	if res.ProductName != "3DBenchy" {
		t.Errorf("expected 3DBenchy, got %s", res.ProductName)
	}
}
