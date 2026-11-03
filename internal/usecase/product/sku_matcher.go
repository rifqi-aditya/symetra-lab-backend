package product

import (
	"context"
	"strings"

	"github.com/google/uuid"

	"symetra-lab-backend-v2/internal/domain/product"
)

type ChannelMappingFinder interface {
	FindHistoricalProductID(ctx context.Context, channelItemID, channelModelID uint64) (*uuid.UUID, error)
}

type MatchedItemResult struct {
	ProductID      *uuid.UUID
	MatchedSKU     string
	MappingStatus  string // "MATCHED" or "UNMAPPED"
	ProductName    string
	ItemSKU        string
	WeightGrams    float64
	PrintTimeHours float64
	Costs          ItemCostBreakdown
}

type SKUMatcher struct {
	productRepo    product.Repository
	costCalculator *CostCalculator
	mappingFinder  ChannelMappingFinder
}

func NewSKUMatcher(
	pRepo product.Repository,
	costCalc *CostCalculator,
	mappingFinder ChannelMappingFinder,
) *SKUMatcher {
	return &SKUMatcher{
		productRepo:    pRepo,
		costCalculator: costCalc,
		mappingFinder:  mappingFinder,
	}
}

type MatchItemInput struct {
	UserID         uuid.UUID
	ModelSKU       string
	ItemSKU        string
	ChannelItemID  uint64
	ChannelModelID uint64
	ItemName       string
	Quantity       int
	SellingPrice   float64
	ExistingProdID *uuid.UUID
}

func (m *SKUMatcher) MatchItem(
	ctx context.Context,
	input MatchItemInput,
	cache *CostingContext,
) MatchedItemResult {
	qty := input.Quantity
	if qty <= 0 {
		qty = 1
	}

	res := MatchedItemResult{
		MappingStatus: "UNMAPPED",
		ProductName:   input.ItemName,
		ItemSKU:       input.ItemSKU,
	}

	var p *product.Product
	var err error

	// 1. If existing product ID is already bound, use it
	if input.ExistingProdID != nil && *input.ExistingProdID != uuid.Nil {
		p, _ = m.productRepo.FindByID(ctx, input.UserID, *input.ExistingProdID)
	}

	// 2. Try matching by Candidate SKU (ModelSKU first, then ItemSKU)
	if p == nil {
		targetSKU := strings.TrimSpace(input.ModelSKU)
		if targetSKU == "" {
			targetSKU = strings.TrimSpace(input.ItemSKU)
		}
		if targetSKU != "" {
			p, err = m.productRepo.FindBySKU(ctx, input.UserID, targetSKU)
			if err != nil {
				p = nil
			}
		}
	}

	// 3. Fallback to historical channel mapping if not found by SKU
	if p == nil && m.mappingFinder != nil && input.ChannelItemID > 0 {
		if histID, hErr := m.mappingFinder.FindHistoricalProductID(ctx, input.ChannelItemID, input.ChannelModelID); hErr == nil && histID != nil {
			p, _ = m.productRepo.FindByID(ctx, input.UserID, *histID)
		}
	}

	// 4. If product resolved, calculate costs and populated fields
	if p != nil {
		prodID := p.ID()
		res.ProductID = &prodID
		if p.SKU() != nil {
			res.MatchedSKU = *p.SKU()
			if res.ItemSKU == "" {
				res.ItemSKU = *p.SKU()
			}
		}
		res.MappingStatus = "MATCHED"
		if p.Name() != "" {
			res.ProductName = p.Name()
		}
		res.WeightGrams = p.DefaultWeightGrams() * float64(qty)
		res.PrintTimeHours = p.DefaultPrintTimeHours() * float64(qty)

		if m.costCalculator != nil {
			if cache == nil {
				userUUID := input.UserID
				if userUUID == uuid.Nil && p != nil {
					userUUID = p.UserID()
				}
				cache = m.costCalculator.LoadContext(ctx, userUUID)
			}
			res.Costs = m.costCalculator.CalculateProductCosts(p, qty, input.SellingPrice, cache)
		}
	}

	return res
}

func (m *SKUMatcher) CostCalculator() *CostCalculator {
	return m.costCalculator
}

