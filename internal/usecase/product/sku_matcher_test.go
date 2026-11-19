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
func (m *mockProductRepo) FindAllParents(ctx context.Context, userID uuid.UUID, filter product.Filter) ([]*product.ParentProduct, int64, error) {
	return nil, 0, nil
}
func (m *mockProductRepo) FindParentByID(ctx context.Context, userID, id uuid.UUID) (*product.ParentProduct, error) {
	return nil, nil
}
func (m *mockProductRepo) FindParentBySKU(ctx context.Context, userID uuid.UUID, parentSKU string) (*product.ParentProduct, error) {
	return nil, nil
}
func (m *mockProductRepo) CreateParent(ctx context.Context, pp *product.ParentProduct) error { return nil }
func (m *mockProductRepo) UpdateParent(ctx context.Context, pp *product.ParentProduct) error { return nil }
func (m *mockProductRepo) DeleteParent(ctx context.Context, userID, id uuid.UUID) error { return nil }
func (m *mockProductRepo) FindVariantsByParentID(ctx context.Context, parentID uuid.UUID) ([]*product.Product, error) {
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

func makeTestProduct(id, parentID, userID uuid.UUID, parentName, variantName, sku, parentSKU, category string, weight, printTime, hpp, price float64) *product.Product {
	var skuPtr *string
	if sku != "" {
		skuPtr = &sku
	}
	parent := product.ReconstructParentProduct(
		parentID,
		userID,
		parentSKU,
		parentName,
		nil,
		category,
		nil,
		nil,
		nil,
		nil,
		time.Now(),
		time.Now(),
	)
	p := product.ReconstructProduct(
		id,
		parentID,
		variantName,
		skuPtr,
		"PLA",
		weight,
		printTime,
		1,
		30,
		hpp,
		price,
		parent,
		nil,
		nil,
		time.Now(),
		time.Now(),
	)
	return p
}

func TestSKUMatcher_MatchBySKU(t *testing.T) {
	prodID := uuid.New()
	parentID := uuid.New()
	userID := uuid.New()
	sku := "SKU-3D-DRAGON"
	p := makeTestProduct(
		prodID, parentID, userID,
		"Articulated Dragon", "Standard", sku, "PAR-DRAGON", "Toys",
		150, 4.5, 50000, 100000,
	)

	repo := &mockProductRepo{
		products: map[string]*product.Product{sku: p},
		byID:     map[uuid.UUID]*product.Product{prodID: p},
	}

	calc := usecaseProduct.NewCostCalculator(nil, nil, nil, nil)
	matcher := usecaseProduct.NewSKUMatcher(repo, calc, nil)

	ctx := context.Background()
	res := matcher.MatchItem(ctx, usecaseProduct.MatchItemInput{
		UserID:        userID,
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
	parentID := uuid.New()
	userID := uuid.New()
	sku := "SKU-BENCHY"
	p := makeTestProduct(
		prodID, parentID, userID,
		"3DBenchy", "Standard", sku, "PAR-BENCHY", "Boats",
		20, 1.0, 10000, 25000,
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
		UserID:        userID,
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
	id3S := uuid.New()
	parentID := uuid.New()
	userID := uuid.New()

	p2S := makeTestProduct(
		id2S, parentID, userID,
		"Rak Display Modular", "2 Stack", sku2S, parentSKU, "Display",
		55, 2.25, 16500, 45000,
	)

	p3S := makeTestProduct(
		id3S, parentID, userID,
		"Rak Display Modular", "3 Stack", sku3S, parentSKU, "Display",
		80, 3.5, 24200, 65000,
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
		UserID:        userID,
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
		UserID:        userID,
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
		UserID:        userID,
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
