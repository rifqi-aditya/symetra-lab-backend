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

func TestSKUMatcher_MultiStackVariants(t *testing.T) {
	parentSKU := "RDM"
	sku2S := "RDM-2S"
	sku3S := "RDM-3S"

	id2S := uuid.New()
	p2S := product.ReconstructProduct(
		id2S,
		uuid.New(),
		"Rak Display Modular - 2 Stack",
		&parentSKU,
		&sku2S,
		nil,
		"Display",
		nil,
		nil,
		55,
		"PLA",
		2.25,
		nil,
		1,
		2000,
		16500,
		45000,
		30,
		0,
		nil,
		nil,
		time.Now(),
		time.Now(),
	)

	id3S := uuid.New()
	p3S := product.ReconstructProduct(
		id3S,
		uuid.New(),
		"Rak Display Modular - 3 Stack",
		&parentSKU,
		&sku3S,
		nil,
		"Display",
		nil,
		nil,
		80,
		"PLA",
		3.5,
		nil,
		1,
		2000,
		24200,
		65000,
		30,
		0,
		nil,
		nil,
		time.Now(),
		time.Now(),
	)

	repo := &mockProductRepo{
		products: map[string]*product.Product{
			sku2S: p2S,
			sku3S: p3S,
		},
		byID: map[uuid.UUID]*product.Product{
			id2S: p2S,
			id3S: p3S,
		},
	}

	calc := usecaseProduct.NewCostCalculator(nil, nil, nil, nil)
	matcher := usecaseProduct.NewSKUMatcher(repo, calc, nil)
	ctx := context.Background()

	// 1. Customer buys "Side View - 2 Stack", Shopee sends ModelSKU = RDM-2S
	resSide2 := matcher.MatchItem(ctx, usecaseProduct.MatchItemInput{
		UserID:        uuid.New(),
		ModelSKU:      "RDM-2S",
		ItemSKU:       "RDM",
		ChannelItemID: 55501,
		ItemName:      "Rak Display Hot Wheels (Side View - 2 Stack)",
		Quantity:      1,
		SellingPrice:  45000,
	}, nil)

	if resSide2.MappingStatus != "MATCHED" {
		t.Fatalf("expected MATCHED for Side View 2 Stack, got %s", resSide2.MappingStatus)
	}
	if *resSide2.ProductID != id2S {
		t.Errorf("expected product ID %v, got %v", id2S, *resSide2.ProductID)
	}
	if resSide2.WeightGrams != 55.0 {
		t.Errorf("expected 55g for 2 Stack, got %f", resSide2.WeightGrams)
	}

	// 2. Customer buys "Front View - 2 Stack", Shopee sends ModelSKU = RDM-2S
	resFront2 := matcher.MatchItem(ctx, usecaseProduct.MatchItemInput{
		UserID:        uuid.New(),
		ModelSKU:      "RDM-2S",
		ItemSKU:       "RDM",
		ChannelItemID: 55501,
		ItemName:      "Rak Display Hot Wheels (Front View - 2 Stack)",
		Quantity:      1,
		SellingPrice:  45000,
	}, nil)

	if resFront2.MappingStatus != "MATCHED" || *resFront2.ProductID != id2S {
		t.Errorf("expected Front View 2 Stack to match id2S, got %v", resFront2.ProductID)
	}

	// 3. Customer buys "Front View - 3 Stack", Shopee sends ModelSKU = RDM-3S
	resFront3 := matcher.MatchItem(ctx, usecaseProduct.MatchItemInput{
		UserID:        uuid.New(),
		ModelSKU:      "RDM-3S",
		ItemSKU:       "RDM",
		ChannelItemID: 55501,
		ItemName:      "Rak Display Hot Wheels (Front View - 3 Stack)",
		Quantity:      2,
		SellingPrice:  65000,
	}, nil)

	if resFront3.MappingStatus != "MATCHED" {
		t.Fatalf("expected MATCHED for Front View 3 Stack, got %s", resFront3.MappingStatus)
	}
	if *resFront3.ProductID != id3S {
		t.Errorf("expected product ID %v, got %v", id3S, *resFront3.ProductID)
	}
	if resFront3.WeightGrams != 160.0 { // 80g * 2 qty
		t.Errorf("expected 160g for 2 units of 3 Stack, got %f", resFront3.WeightGrams)
	}
}
