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

type UpdateProductInput struct {
	ID                    uuid.UUID
	VariantName           *string
	Name                  *string
	ParentSKU             *string
	SKU                   *string
	Description           *string
	Category              *string
	CategoryID            *uuid.UUID
	ThumbnailURL          *string
	DesignLink            *string
	DefaultWeightGrams    *float64
	MaterialType          *string
	DefaultPrintTimeHours *float64
	BatchSize             *int
	TargetMarginPercent   *int
	Components            *[]ComponentInput
	PackagingItems        *[]PackagingItemInput
}

type UpdateProductUseCase struct {
	productRepo    product.Repository
	shopConfigRepo shopconfig.Repository
	machineRepo    machine.Repository
	componentRepo  component.Repository
	packagingRepo  packaging.Repository
	categoryRepo   category.Repository
}

func NewUpdateProductUseCase(
	pRepo product.Repository,
	sConfigRepo shopconfig.Repository,
	mRepo machine.Repository,
	cRepo component.Repository,
	pkgRepo packaging.Repository,
	catRepo category.Repository,
) *UpdateProductUseCase {
	return &UpdateProductUseCase{
		productRepo:    pRepo,
		shopConfigRepo: sConfigRepo,
		machineRepo:    mRepo,
		componentRepo:  cRepo,
		packagingRepo:  pkgRepo,
		categoryRepo:   catRepo,
	}
}

func (uc *UpdateProductUseCase) Execute(ctx context.Context, userID uuid.UUID, input UpdateProductInput) (*product.Product, costing.ProductCostBreakdown, error) {
	existing, err := uc.productRepo.FindByID(ctx, userID, input.ID)
	if err != nil {
		return nil, costing.ProductCostBreakdown{}, err
	}

	// Check SKU if provided
	cleanSKU := existing.SKU()
	if input.SKU != nil && strings.TrimSpace(*input.SKU) != "" {
		val := strings.TrimSpace(strings.ToUpper(*input.SKU))
		foundWithSKU, _ := uc.productRepo.FindBySKU(ctx, userID, val)
		if foundWithSKU != nil && foundWithSKU.ID() != existing.ID() {
			return nil, costing.ProductCostBreakdown{}, fmt.Errorf("SKU '%s' sudah digunakan oleh produk lain", val)
		}
		cleanSKU = &val
	}

	// 1. Update Parent Product if attributes are present
	parent := existing.Parent()
	if parent == nil && existing.ParentID() != uuid.Nil {
		parent, _ = uc.productRepo.FindParentByID(ctx, userID, existing.ParentID())
	}

	if parent != nil {
		parentModified := false
		pName := parent.Name()
		if input.Name != nil && strings.TrimSpace(*input.Name) != "" && strings.TrimSpace(*input.Name) != pName {
			pName = strings.TrimSpace(*input.Name)
			parentModified = true
		}

		pSKU := parent.ParentSKU()
		if input.ParentSKU != nil && strings.TrimSpace(*input.ParentSKU) != "" && strings.TrimSpace(strings.ToUpper(*input.ParentSKU)) != pSKU {
			pSKU = strings.TrimSpace(strings.ToUpper(*input.ParentSKU))
			parentModified = true
		}

		pCatID := parent.CategoryID()
		if input.CategoryID != nil && *input.CategoryID != uuid.Nil {
			pCatID = input.CategoryID
			parentModified = true
		} else if input.Category != nil && strings.TrimSpace(*input.Category) != "" && uc.categoryRepo != nil {
			cats, _ := uc.categoryRepo.FindAll(ctx, userID)
			for _, c := range cats {
				if strings.EqualFold(strings.TrimSpace(c.Name()), strings.TrimSpace(*input.Category)) {
					id := c.ID()
					pCatID = &id
					parentModified = true
					break
				}
			}
		}

		pDesc := parent.Description()
		if input.Description != nil {
			pDesc = input.Description
			parentModified = true
		}

		pThumb := parent.ThumbnailURL()
		if input.ThumbnailURL != nil {
			pThumb = input.ThumbnailURL
			parentModified = true
		}

		pDesign := parent.DesignLink()
		if input.DesignLink != nil {
			pDesign = input.DesignLink
			parentModified = true
		}

		if parentModified {
			if err := parent.Update(pName, pSKU, pCatID, pDesc, pThumb, pDesign); err == nil {
				_ = uc.productRepo.UpdateParent(ctx, parent)
			}
		}
	}

	// 2. Variant attributes
	varName := existing.VariantName()
	if input.VariantName != nil && strings.TrimSpace(*input.VariantName) != "" {
		varName = strings.TrimSpace(*input.VariantName)
	}

	weight := existing.DefaultWeightGrams()
	if input.DefaultWeightGrams != nil {
		weight = *input.DefaultWeightGrams
	}

	mat := existing.MaterialType()
	if input.MaterialType != nil {
		mat = strings.TrimSpace(*input.MaterialType)
	}

	printTime := existing.DefaultPrintTimeHours()
	if input.DefaultPrintTimeHours != nil {
		printTime = *input.DefaultPrintTimeHours
	}

	batch := existing.BatchSize()
	if input.BatchSize != nil && *input.BatchSize > 0 {
		batch = *input.BatchSize
	}

	margin := existing.TargetMarginPercent()
	if input.TargetMarginPercent != nil && *input.TargetMarginPercent > 0 {
		margin = *input.TargetMarginPercent
	}

	// Components
	now := time.Now()
	var comps []*product.ProductComponent
	var compCostInputs []costing.ComponentCostInput
	if input.Components != nil {
		comps = make([]*product.ProductComponent, len(*input.Components))
		for i, c := range *input.Components {
			comps[i] = product.ReconstructProductComponent(
				uuid.New(),
				existing.ID(),
				c.ComponentID,
				c.Quantity,
				c.MarkupPercent,
				now,
			)
			price := 0.0
			if c.ComponentID != nil && uc.componentRepo != nil {
				if compEntity, err := uc.componentRepo.FindByID(ctx, userID, *c.ComponentID); err == nil && compEntity != nil {
					price = compEntity.PricePerUnit()
				}
			}
			compCostInputs = append(compCostInputs, costing.ComponentCostInput{
				PricePerUnit:  price,
				Quantity:      c.Quantity,
				MarkupPercent: c.MarkupPercent,
			})
		}
	} else {
		comps = existing.Components()
		for _, pc := range comps {
			price := 0.0
			if pc.ComponentID() != nil && uc.componentRepo != nil {
				if compEntity, err := uc.componentRepo.FindByID(ctx, userID, *pc.ComponentID()); err == nil && compEntity != nil {
					price = compEntity.PricePerUnit()
				}
			}
			compCostInputs = append(compCostInputs, costing.ComponentCostInput{
				PricePerUnit:  price,
				Quantity:      pc.Quantity(),
				MarkupPercent: pc.MarkupPercent(),
			})
		}
	}

	// Packaging
	var packs []*product.ProductPackagingItem
	var pkgCostInputs []costing.PackagingCostInput
	if input.PackagingItems != nil {
		packs = make([]*product.ProductPackagingItem, len(*input.PackagingItems))
		for i, p := range *input.PackagingItems {
			packs[i] = product.ReconstructProductPackagingItem(
				uuid.New(),
				existing.ID(),
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
	} else {
		packs = existing.PackagingItems()
		for _, pi := range packs {
			unitCost := 0.0
			if uc.packagingRepo != nil {
				if pkgEntity, err := uc.packagingRepo.FindByID(ctx, userID, pi.PackagingItemID()); err == nil && pkgEntity != nil {
					unitCost = pkgEntity.UnitCost()
					if unitCost <= 0 {
						unitCost = pkgEntity.CalculateUnitCost()
					}
				}
			}
			pkgCostInputs = append(pkgCostInputs, costing.PackagingCostInput{
				UnitCost:     unitCost,
				QuantityUsed: pi.QuantityUsed(),
			})
		}
	}

	// Config & Machine
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

	cb := costing.Calculate(costing.CalculationInput{
		MaterialType:          mat,
		DefaultWeightGrams:    weight,
		DefaultPrintTimeHours: printTime,
		BatchSize:             batch,
		TargetMarginPercent:   margin,
		BaseHPP:               existing.BaseHPP(),
		BaseSellingPrice:      existing.BaseSellingPrice(),
		Components:            compCostInputs,
		PackagingItems:        pkgCostInputs,
		ShopConfig:            sCfgInput,
		Marketplace:           mpInput,
	})

	updatedProd := product.ReconstructProduct(
		existing.ID(),
		existing.ParentID(),
		varName,
		cleanSKU,
		mat,
		weight,
		printTime,
		batch,
		margin,
		cb.BaseHPP,
		cb.BaseSellingPrice,
		parent,
		comps,
		packs,
		existing.CreatedAt(),
		now,
	)

	if err := uc.productRepo.Update(ctx, updatedProd); err != nil {
		return nil, costing.ProductCostBreakdown{}, err
	}

	return updatedProd, cb, nil
}
