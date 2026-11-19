package product

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"symetra-lab-backend-v2/internal/domain/category"
	"symetra-lab-backend-v2/internal/domain/component"
	"symetra-lab-backend-v2/internal/domain/costing"
	"symetra-lab-backend-v2/internal/domain/machine"
	"symetra-lab-backend-v2/internal/domain/packaging"
	"symetra-lab-backend-v2/internal/domain/product"
	"symetra-lab-backend-v2/internal/domain/shopconfig"
)

type ComponentInput struct {
	ComponentID   *uuid.UUID
	Quantity      float64
	MarkupPercent float64
}

type PackagingItemInput struct {
	PackagingItemID uuid.UUID
	QuantityUsed    float64
}

type CreateProductInput struct {
	ParentID              *uuid.UUID
	VariantName           string
	Name                  string
	ParentSKU             *string
	SKU                   *string
	Description           *string
	Category              string
	CategoryID            *uuid.UUID
	ThumbnailURL          *string
	DesignLink            *string
	DefaultWeightGrams    float64
	MaterialType          string
	DefaultPrintTimeHours float64
	BatchSize             int
	TargetMarginPercent   int
	Components            []ComponentInput
	PackagingItems        []PackagingItemInput
}

type CreateProductUseCase struct {
	productRepo    product.Repository
	shopConfigRepo shopconfig.Repository
	machineRepo    machine.Repository
	componentRepo  component.Repository
	packagingRepo  packaging.Repository
	categoryRepo   category.Repository
}

func NewCreateProductUseCase(
	pRepo product.Repository,
	sConfigRepo shopconfig.Repository,
	mRepo machine.Repository,
	cRepo component.Repository,
	pkgRepo packaging.Repository,
	catRepo category.Repository,
) *CreateProductUseCase {
	return &CreateProductUseCase{
		productRepo:    pRepo,
		shopConfigRepo: sConfigRepo,
		machineRepo:    mRepo,
		componentRepo:  cRepo,
		packagingRepo:  pkgRepo,
		categoryRepo:   catRepo,
	}
}

func (uc *CreateProductUseCase) Execute(ctx context.Context, userID uuid.UUID, input CreateProductInput) (*product.Product, costing.ProductCostBreakdown, error) {
	if userID == uuid.Nil {
		return nil, costing.ProductCostBreakdown{}, product.ErrUserIDRequired
	}
	cleanName := strings.TrimSpace(input.Name)
	if cleanName == "" && (input.ParentID == nil || *input.ParentID == uuid.Nil) {
		return nil, costing.ProductCostBreakdown{}, product.ErrProductNameRequired
	}

	// SKU check
	var cleanSKU *string
	if input.SKU != nil && strings.TrimSpace(*input.SKU) != "" {
		val := strings.TrimSpace(strings.ToUpper(*input.SKU))
		cleanSKU = &val

		existing, _ := uc.productRepo.FindBySKU(ctx, userID, *cleanSKU)
		if existing != nil {
			return nil, costing.ProductCostBreakdown{}, fmt.Errorf("SKU '%s' sudah digunakan oleh produk '%s'", *cleanSKU, existing.Name())
		}
	}

	// 1. Resolve Parent Product
	var parent *product.ParentProduct
	if input.ParentID != nil && *input.ParentID != uuid.Nil {
		p, err := uc.productRepo.FindParentByID(ctx, userID, *input.ParentID)
		if err != nil {
			return nil, costing.ProductCostBreakdown{}, fmt.Errorf("parent product not found: %w", err)
		}
		parent = p
	} else {
		// Resolve category ID
		var catID *uuid.UUID = input.CategoryID
		if (catID == nil || *catID == uuid.Nil) && strings.TrimSpace(input.Category) != "" && uc.categoryRepo != nil {
			cats, _ := uc.categoryRepo.FindAll(ctx, userID)
			for _, c := range cats {
				if strings.EqualFold(strings.TrimSpace(c.Name()), strings.TrimSpace(input.Category)) {
					id := c.ID()
					catID = &id
					break
				}
			}
		}

		pSKU := ""
		if input.ParentSKU != nil && strings.TrimSpace(*input.ParentSKU) != "" {
			pSKU = strings.TrimSpace(strings.ToUpper(*input.ParentSKU))
		} else if cleanSKU != nil {
			pSKU = *cleanSKU
		}

		newParent, err := product.NewParentProduct(
			userID,
			cleanName,
			pSKU,
			catID,
			input.Description,
			input.ThumbnailURL,
			input.DesignLink,
		)
		if err != nil {
			return nil, costing.ProductCostBreakdown{}, err
		}

		if err := uc.productRepo.CreateParent(ctx, newParent); err != nil {
			return nil, costing.ProductCostBreakdown{}, fmt.Errorf("failed to create parent product: %w", err)
		}
		parent = newParent
	}

	prodID := uuid.New()
	now := time.Now()

	comps := make([]*product.ProductComponent, len(input.Components))
	var compCostInputs []costing.ComponentCostInput
	for i, c := range input.Components {
		comps[i] = product.ReconstructProductComponent(
			uuid.New(),
			prodID,
			c.ComponentID,
			c.Quantity,
			now,
		)
		price := 0.0
		markup := 0.0
		if c.ComponentID != nil && uc.componentRepo != nil {
			if compEntity, err := uc.componentRepo.FindByID(ctx, userID, *c.ComponentID); err == nil && compEntity != nil {
				price = compEntity.PricePerUnit()
				markup = compEntity.DefaultMarkupPercent()
			}
		}
		compCostInputs = append(compCostInputs, costing.ComponentCostInput{
			PricePerUnit:  price,
			Quantity:      c.Quantity,
			MarkupPercent: markup,
		})
	}

	packs := make([]*product.ProductPackagingItem, len(input.PackagingItems))
	var pkgCostInputs []costing.PackagingCostInput
	for i, p := range input.PackagingItems {
		packs[i] = product.ReconstructProductPackagingItem(
			uuid.New(),
			prodID,
			p.PackagingItemID,
			p.QuantityUsed,
			now,
		)
		unitCost := 0.0
		if uc.packagingRepo != nil {
			if pkgEntity, err := uc.packagingRepo.FindByID(ctx, userID, p.PackagingItemID); err == nil && pkgEntity != nil {
				unitCost = pkgEntity.UnitCost()
				if unitCost <= 0 {
					unitCost = pkgEntity.CalculateUnitCost()
				}
			}
		}
		pkgCostInputs = append(pkgCostInputs, costing.PackagingCostInput{
			UnitCost:     unitCost,
			QuantityUsed: p.QuantityUsed,
		})
	}

	// ShopConfig
	var sCfgInput *costing.ShopConfigInput
	if uc.shopConfigRepo != nil {
		if sc, err := uc.shopConfigRepo.GetByUserID(ctx, userID); err == nil && sc != nil {
			sCfgInput = &costing.ShopConfigInput{
				FilamentPricePerRoll:    sc.FilamentPricePerRoll(),
				FilamentWeightGrams:     sc.FilamentWeightGrams(),
				ElectricityTariffPerKwh: sc.ElectricityTariffPerKwh(),
				PrinterPowerWatts:       sc.PrinterPowerWatts(),
				PrinterPrice:            sc.PrinterPrice(),
				PrinterLifespanHours:    sc.PrinterLifespanHours(),
				FailureBufferPercent:    sc.FailureBufferPercent(),
			}
		}
	}

	// Marketplace
	var mpInput *costing.MarketplacePlatformInput
	if uc.shopConfigRepo != nil {
		if mp, err := uc.shopConfigRepo.GetActiveMarketplace(ctx, userID, "shopee"); err == nil && mp != nil {
			mpInput = &costing.MarketplacePlatformInput{
				CommissionPercent:   mp.CommissionPercent(),
				PromoFeePercent:     mp.PromoFeePercent(),
				FreeShippingPercent: mp.FreeShippingPercent(),
				OrderFeeIDR:         mp.OrderFeeIDR(),
				IsActive:            mp.IsActive(),
			}
		}
	}

	margin := input.TargetMarginPercent
	if margin <= 0 {
		margin = 30
	}

	batchSize := input.BatchSize
	if batchSize <= 0 {
		batchSize = 1
	}

	cb := costing.Calculate(costing.CalculationInput{
		MaterialType:          input.MaterialType,
		DefaultWeightGrams:    input.DefaultWeightGrams,
		DefaultPrintTimeHours: input.DefaultPrintTimeHours,
		BatchSize:             batchSize,
		TargetMarginPercent:   margin,
		Components:            compCostInputs,
		PackagingItems:        pkgCostInputs,
		ShopConfig:            sCfgInput,
		Marketplace:           mpInput,
	})

	varName := strings.TrimSpace(input.VariantName)
	if varName == "" {
		varName = "Standard"
	}

	mat := strings.TrimSpace(input.MaterialType)
	if mat == "" {
		mat = "PLA"
	}

	prod := product.ReconstructProduct(
		prodID,
		parent.ID(),
		varName,
		cleanSKU,
		mat,
		input.DefaultWeightGrams,
		input.DefaultPrintTimeHours,
		batchSize,
		margin,
		cb.BaseHPP,
		cb.BaseSellingPrice,
		parent,
		comps,
		packs,
		now,
		now,
	)

	if err := uc.productRepo.Create(ctx, prod); err != nil {
		return nil, costing.ProductCostBreakdown{}, err
	}

	return prod, cb, nil
}
