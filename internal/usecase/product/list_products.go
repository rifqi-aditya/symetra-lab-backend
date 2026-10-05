package product

import (
	"context"
	"sort"
	"sync"

	"github.com/google/uuid"

	"symetra-lab-backend-v2/internal/domain/costing"
	"symetra-lab-backend-v2/internal/domain/product"
)

type ProductItemOutput struct {
	Product       *product.Product
	TotalSold     int
	CostBreakdown costing.ProductCostBreakdown
}

type ListProductsOutput struct {
	Items      []ProductItemOutput
	TotalCount int64
}

type ListProductsUseCase struct {
	productRepo product.Repository
}

func NewListProductsUseCase(pRepo product.Repository) *ListProductsUseCase {
	return &ListProductsUseCase{
		productRepo: pRepo,
	}
}

func (uc *ListProductsUseCase) Execute(ctx context.Context, userID uuid.UUID, filter product.Filter) (*ListProductsOutput, error) {
	var (
		products []*product.Product
		total    int64
		errProd  error
		soldMap  map[uuid.UUID]int
	)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		products, total, errProd = uc.productRepo.FindAll(ctx, userID, filter)
	}()

	go func() {
		defer wg.Done()
		var errSold error
		soldMap, errSold = uc.productRepo.GetTotalSoldMap(ctx, userID)
		if errSold != nil {
			soldMap = make(map[uuid.UUID]int)
		}
	}()

	wg.Wait()

	if errProd != nil {
		return nil, errProd
	}

	items := make([]ProductItemOutput, len(products))
	for i, p := range products {
		cb := costing.ProductCostBreakdown{
			BaseHPP:             p.BaseHPP(),
			BaseSellingPrice:    p.BaseSellingPrice(),
			TargetMarginPercent: p.TargetMarginPercent(),
		}
		items[i] = ProductItemOutput{
			Product:       p,
			TotalSold:     soldMap[p.ID()],
			CostBreakdown: cb,
		}
	}

	// Sort: TotalSold DESC -> CreatedAt DESC (Best Sellers first)
	sort.Slice(items, func(i, j int) bool {
		if items[i].TotalSold != items[j].TotalSold {
			return items[i].TotalSold > items[j].TotalSold
		}
		return items[i].Product.CreatedAt().After(items[j].Product.CreatedAt())
	})

	return &ListProductsOutput{
		Items:      items,
		TotalCount: total,
	}, nil
}
