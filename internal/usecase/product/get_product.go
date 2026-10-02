package product

import (
	"context"

	"github.com/google/uuid"

	"symetra-lab-backend-v2/internal/domain/component"
	"symetra-lab-backend-v2/internal/domain/costing"
	"symetra-lab-backend-v2/internal/domain/machine"
	"symetra-lab-backend-v2/internal/domain/packaging"
	"symetra-lab-backend-v2/internal/domain/product"
	"symetra-lab-backend-v2/internal/domain/shopconfig"
)

type ProductDetailOutput struct {
	Product       *product.Product
	TotalSold     int
	CostBreakdown costing.ProductCostBreakdown
}

type GetProductUseCase struct {
	productRepo    product.Repository
	shopConfigRepo shopconfig.Repository
	machineRepo    machine.Repository
	componentRepo  component.Repository
	packagingRepo  packaging.Repository
}

func NewGetProductUseCase(
	pRepo product.Repository,
	sConfigRepo shopconfig.Repository,
	mRepo machine.Repository,
	cRepo component.Repository,
	pkgRepo packaging.Repository,
) *GetProductUseCase {
	return &GetProductUseCase{
		productRepo:    pRepo,
		shopConfigRepo: sConfigRepo,
		machineRepo:    mRepo,
		componentRepo:  cRepo,
		packagingRepo:  pkgRepo,
	}
}

func (uc *GetProductUseCase) Execute(ctx context.Context, userID, id uuid.UUID) (*ProductDetailOutput, error) {
	p, err := uc.productRepo.FindByID(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	return uc.buildDetailOutput(ctx, userID, p)
}

func (uc *GetProductUseCase) ExecuteBySKU(ctx context.Context, userID uuid.UUID, sku string) (*ProductDetailOutput, error) {
	p, err := uc.productRepo.FindBySKU(ctx, userID, sku)
	if err != nil {
		return nil, err
	}
	return uc.buildDetailOutput(ctx, userID, p)
}

func (uc *GetProductUseCase) buildDetailOutput(ctx context.Context, userID uuid.UUID, p *product.Product) (*ProductDetailOutput, error) {
	sold, _ := uc.productRepo.GetSoldQtyByID(ctx, userID, p.ID())

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
	if p.DefaultMachineID() != nil {
		if m, err := uc.machineRepo.FindByID(ctx, userID, *p.DefaultMachineID()); err == nil && m != nil {
			machInput = &costing.MachineCostInput{
				TotalPurchaseCost: m.TotalPurchaseCost(),
				LifespanHours:     float64(m.LifespanHours()),
				AvgPowerWatts:     m.AvgPowerWatts(),
			}
		}
	}

	// Components
	var compInputs []costing.ComponentCostInput
	for _, pc := range p.Components() {
		price := 0.0
		if pc.ComponentID() != nil {
			if c, err := uc.componentRepo.FindByID(ctx, userID, *pc.ComponentID()); err == nil && c != nil {
				price = c.PricePerUnit()
			}
		}
		compInputs = append(compInputs, costing.ComponentCostInput{
			PricePerUnit:  price,
			Quantity:      pc.Quantity(),
			MarkupPercent: pc.MarkupPercent(),
		})
	}

	// Packaging
	var pkgInputs []costing.PackagingCostInput
	for _, pi := range p.PackagingItems() {
		unitCost := 0.0
		if pkgItem, err := uc.packagingRepo.FindByID(ctx, userID, pi.PackagingItemID()); err == nil && pkgItem != nil {
			unitCost = pkgItem.UnitCost()
			if unitCost <= 0 {
				unitCost = pkgItem.CalculateUnitCost()
			}
		}
		pkgInputs = append(pkgInputs, costing.PackagingCostInput{
			UnitCost:     unitCost,
			QuantityUsed: pi.QuantityUsed(),
		})
	}

	// Preset
	hasPreset := false
	presetTotal := 0.0
	if p.PackagingPresetID() != nil {
		if preset, err := uc.packagingRepo.FindPresetByID(ctx, userID, *p.PackagingPresetID()); err == nil && preset != nil {
			hasPreset = true
			presetTotal = preset.TotalCost()
		}
	}

	cb := costing.Calculate(costing.CalculationInput{
		MaterialType:          p.MaterialType(),
		DefaultWeightGrams:    p.DefaultWeightGrams(),
		DefaultPrintTimeHours: p.DefaultPrintTimeHours(),
		BatchSize:             p.BatchSize(),
		PackingFeeIDR:         p.PackingFeeIDR(),
		TargetMarginPercent:   p.TargetMarginPercent(),
		BaseHPP:               p.BaseHPP(),
		BaseSellingPrice:      p.BaseSellingPrice(),
		Components:            compInputs,
		PackagingItems:        pkgInputs,
		HasPackagingPreset:    hasPreset,
		PackagingPresetTotal:  presetTotal,
		Machine:               machInput,
		ShopConfig:            sCfgInput,
		Marketplace:           mpInput,
	})

	return &ProductDetailOutput{
		Product:       p,
		TotalSold:     sold,
		CostBreakdown: cb,
	}, nil
}