package product

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"symetra-lab-backend-v2/internal/domain/component"
	"symetra-lab-backend-v2/internal/domain/costing"
	"symetra-lab-backend-v2/internal/domain/machine"
	"symetra-lab-backend-v2/internal/domain/packaging"
	"symetra-lab-backend-v2/internal/domain/product"
	"symetra-lab-backend-v2/internal/domain/shopconfig"
)

type UpdateProductInput struct {
	ID                    uuid.UUID
	Name                  *string
	ParentSKU             *string
	SKU                   *string
	Description           *string
	Category              *string
	ThumbnailURL          *string
	DesignLink            *string
	DefaultWeightGrams    *float64
	MaterialType          *string
	DefaultPrintTimeHours *float64
	DefaultMachineID      *uuid.UUID
	PackagingPresetID     *uuid.UUID
	BatchSize             *int
	PackingFeeIDR         *int
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
}

func NewUpdateProductUseCase(
	pRepo product.Repository,
	sConfigRepo shopconfig.Repository,
	mRepo machine.Repository,
	cRepo component.Repository,
	pkgRepo packaging.Repository,
) *UpdateProductUseCase {
	return &UpdateProductUseCase{
		productRepo:    pRepo,
		shopConfigRepo: sConfigRepo,
		machineRepo:    mRepo,
		componentRepo:  cRepo,
		packagingRepo:  pkgRepo,
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

	name := existing.Name()
	if input.Name != nil {
		name = strings.TrimSpace(*input.Name)
	}

	desc := existing.Description()
	if input.Description != nil {
		desc = input.Description
	}

	cat := existing.Category()
	if input.Category != nil {
		cat = strings.TrimSpace(*input.Category)
	}

	thumb := existing.ThumbnailURL()
	if input.ThumbnailURL != nil {
		thumb = input.ThumbnailURL
	}

	design := existing.DesignLink()
	if input.DesignLink != nil {
		design = input.DesignLink
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

	machID := existing.DefaultMachineID()
	if input.DefaultMachineID != nil {
		machID = input.DefaultMachineID
	}

	presetID := existing.PackagingPresetID()
	if input.PackagingPresetID != nil {
		presetID = input.PackagingPresetID
	}

	batch := existing.BatchSize()
	if input.BatchSize != nil && *input.BatchSize > 0 {
		batch = *input.BatchSize
	}

	packingFee := existing.PackingFeeIDR()
	if input.PackingFeeIDR != nil {
		packingFee = *input.PackingFeeIDR
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
			if c.ComponentID != nil {
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
			if pc.ComponentID() != nil {
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
			if pkgEntity, err := uc.packagingRepo.FindByID(ctx, userID, p.PackagingItemID); err == nil && pkgEntity != nil {
				unitCost = pkgEntity.UnitCost()
				if unitCost <= 0 {
					unitCost = pkgEntity.CalculateUnitCost()
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
			if pkgEntity, err := uc.packagingRepo.FindByID(ctx, userID, pi.PackagingItemID()); err == nil && pkgEntity != nil {
				unitCost = pkgEntity.UnitCost()
				if unitCost <= 0 {
					unitCost = pkgEntity.CalculateUnitCost()
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

	var mpInput *costing.MarketplacePlatformInput
	if mp, err := uc.shopConfigRepo.GetActiveMarketplace(ctx, userID, "shopee"); err == nil && mp != nil {
		mpInput = &costing.MarketplacePlatformInput{
			CommissionPercent:   mp.CommissionPercent(),
			PromoFeePercent:     mp.PromoFeePercent(),
			FreeShippingPercent: mp.FreeShippingPercent(),
			OrderFeeIDR:         mp.OrderFeeIDR(),
			IsActive:            mp.IsActive(),
		}
	}

	var machInput *costing.MachineCostInput
	if machID != nil {
		if m, err := uc.machineRepo.FindByID(ctx, userID, *machID); err == nil && m != nil {
			machInput = &costing.MachineCostInput{
				TotalPurchaseCost: m.TotalPurchaseCost(),
				LifespanHours:     float64(m.LifespanHours()),
				AvgPowerWatts:     m.AvgPowerWatts(),
			}
		}
	}

	hasPreset := false
	presetTotal := 0.0
	if presetID != nil {
		if preset, err := uc.packagingRepo.FindPresetByID(ctx, userID, *presetID); err == nil && preset != nil {
			hasPreset = true
			presetTotal = preset.TotalCost()
		}
	}

	cb := costing.Calculate(costing.CalculationInput{
		MaterialType:          mat,
		DefaultWeightGrams:    weight,
		DefaultPrintTimeHours: printTime,
		BatchSize:             batch,
		PackingFeeIDR:         packingFee,
		TargetMarginPercent:   margin,
		BaseHPP:               existing.BaseHPP(),
		BaseSellingPrice:      existing.BaseSellingPrice(),
		Components:            compCostInputs,
		PackagingItems:        pkgCostInputs,
		HasPackagingPreset:    hasPreset,
		PackagingPresetTotal:  presetTotal,
		Machine:               machInput,
		ShopConfig:            sCfgInput,
		Marketplace:           mpInput,
	})

	updatedProd := product.ReconstructProduct(
		existing.ID(),
		existing.UserID(),
		name,
		existing.ParentSKU(),
		cleanSKU,
		desc,
		cat,
		thumb,
		design,
		weight,
		mat,
		printTime,
		machID,
		presetID,
		batch,
		packingFee,
		cb.BaseHPP,
		cb.BaseSellingPrice,
		margin,
		existing.TimesOrdered(),
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