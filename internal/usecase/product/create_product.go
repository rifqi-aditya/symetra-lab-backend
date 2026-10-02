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
	Name                  string
	ParentSKU             *string
	SKU                   *string
	Description           *string
	Category              string
	ThumbnailURL          *string
	DesignLink            *string
	DefaultWeightGrams    float64
	MaterialType          string
	DefaultPrintTimeHours float64
	DefaultMachineID      *uuid.UUID
	PackagingPresetID     *uuid.UUID
	BatchSize             int
	PackingFeeIDR         int
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
}

func NewCreateProductUseCase(
	pRepo product.Repository,
	sConfigRepo shopconfig.Repository,
	mRepo machine.Repository,
	cRepo component.Repository,
	pkgRepo packaging.Repository,
) *CreateProductUseCase {
	return &CreateProductUseCase{
		productRepo:    pRepo,
		shopConfigRepo: sConfigRepo,
		machineRepo:    mRepo,
		componentRepo:  cRepo,
		packagingRepo:  pkgRepo,
	}
}

func (uc *CreateProductUseCase) Execute(ctx context.Context, userID uuid.UUID, input CreateProductInput) (*product.Product, costing.ProductCostBreakdown, error) {
	if userID == uuid.Nil {
		return nil, costing.ProductCostBreakdown{}, product.ErrUserIDRequired
	}
	cleanName := strings.TrimSpace(input.Name)
	if cleanName == "" {
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

	// ShopConfig
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

	// Marketplace
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

	// Machine
	var machInput *costing.MachineCostInput
	if input.DefaultMachineID != nil {
		if m, err := uc.machineRepo.FindByID(ctx, userID, *input.DefaultMachineID); err == nil && m != nil {
			machInput = &costing.MachineCostInput{
				TotalPurchaseCost: m.TotalPurchaseCost(),
				LifespanHours:     float64(m.LifespanHours()),
				AvgPowerWatts:     m.AvgPowerWatts(),
			}
		}
	}

	hasPreset := false
	presetTotal := 0.0
	if input.PackagingPresetID != nil {
		if preset, err := uc.packagingRepo.FindPresetByID(ctx, userID, *input.PackagingPresetID); err == nil && preset != nil {
			hasPreset = true
			presetTotal = preset.TotalCost()
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
		PackingFeeIDR:         input.PackingFeeIDR,
		TargetMarginPercent:   margin,
		Components:            compCostInputs,
		PackagingItems:        pkgCostInputs,
		HasPackagingPreset:    hasPreset,
		PackagingPresetTotal:  presetTotal,
		Machine:               machInput,
		ShopConfig:            sCfgInput,
		Marketplace:           mpInput,
	})

	prod := product.ReconstructProduct(
		prodID,
		userID,
		cleanName,
		input.ParentSKU,
		cleanSKU,
		input.Description,
		input.Category,
		input.ThumbnailURL,
		input.DesignLink,
		input.DefaultWeightGrams,
		input.MaterialType,
		input.DefaultPrintTimeHours,
		input.DefaultMachineID,
		input.PackagingPresetID,
		batchSize,
		input.PackingFeeIDR,
		cb.BaseHPP,
		cb.BaseSellingPrice,
		margin,
		0,
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