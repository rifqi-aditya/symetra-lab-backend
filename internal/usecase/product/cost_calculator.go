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

type ItemCostBreakdown struct {
	FilamentCost     float64
	EmergencyCost    float64
	ComponentCost    float64
	PackagingCost    float64
	ElectricityCost  float64
	MaintenanceCost  float64
	DepreciationCost float64
	TotalCOGS        float64
	NetProfit        float64
}

type CostCalculator struct {
	shopConfigRepo shopconfig.Repository
	machineRepo    machine.Repository
	componentRepo  component.Repository
	packagingRepo  packaging.Repository
}

func NewCostCalculator(
	sConfigRepo shopconfig.Repository,
	mRepo machine.Repository,
	cRepo component.Repository,
	pkgRepo packaging.Repository,
) *CostCalculator {
	return &CostCalculator{
		shopConfigRepo: sConfigRepo,
		machineRepo:    mRepo,
		componentRepo:  cRepo,
		packagingRepo:  pkgRepo,
	}
}

type CostingContext struct {
	ShopConfig    *costing.ShopConfigInput
	DefaultMach   *costing.MachineCostInput
	Machines      map[uuid.UUID]*costing.MachineCostInput
	CompPrices    map[uuid.UUID]float64
	PackUnitCosts map[uuid.UUID]float64
}

func (cc *CostCalculator) LoadContext(ctx context.Context, userID uuid.UUID) *CostingContext {
	contextCache := &CostingContext{
		Machines:      make(map[uuid.UUID]*costing.MachineCostInput),
		CompPrices:    make(map[uuid.UUID]float64),
		PackUnitCosts: make(map[uuid.UUID]float64),
	}

	if cc.shopConfigRepo != nil {
		if sc, err := cc.shopConfigRepo.GetByUserID(ctx, userID); err == nil && sc != nil {
			contextCache.ShopConfig = &costing.ShopConfigInput{
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

	if cc.machineRepo != nil {
		if ms, err := cc.machineRepo.FindAll(ctx, userID); err == nil {
			for _, m := range ms {
				salvage := 0.0
				if m.SalvageValue() != nil {
					salvage = *m.SalvageValue()
				}
				minput := &costing.MachineCostInput{
					TotalPurchaseCost: m.TotalPurchaseCost(),
					SalvageValue:      salvage,
					LifespanHours:     float64(m.LifespanHours()),
					AvgPowerWatts:     m.AvgPowerWatts(),
				}
				contextCache.Machines[m.ID()] = minput
				if contextCache.DefaultMach == nil {
					contextCache.DefaultMach = minput
				}
			}
		}
	}

	if cc.componentRepo != nil {
		if comps, err := cc.componentRepo.FindAll(ctx, userID); err == nil {
			for _, c := range comps {
				contextCache.CompPrices[c.ID()] = c.PricePerUnit()
			}
		}
	}

	if cc.packagingRepo != nil {
		if pkgs, err := cc.packagingRepo.FindAll(ctx, userID); err == nil {
			for _, p := range pkgs {
				uCost := p.UnitCost()
				if uCost <= 0 && p.PurchaseQuantity() > 0 {
					uCost = p.PurchasePrice() / float64(p.PurchaseQuantity())
				}
				contextCache.PackUnitCosts[p.ID()] = uCost
			}
		}
	}

	return contextCache
}

func (cc *CostCalculator) CalculateProductCosts(
	p *product.Product,
	qty int,
	sellingPrice float64,
	cache *CostingContext,
) ItemCostBreakdown {
	if qty <= 0 {
		qty = 1
	}
	q := float64(qty)

	if p == nil {
		return ItemCostBreakdown{}
	}

	var compInputs []costing.ComponentCostInput
	for _, c := range p.Components() {
		price := 0.0
		if c.ComponentID() != nil && cache != nil {
			price = cache.CompPrices[*c.ComponentID()]
		}
		compInputs = append(compInputs, costing.ComponentCostInput{
			PricePerUnit:  price,
			Quantity:      c.Quantity(),
			MarkupPercent: c.MarkupPercent(),
		})
	}

	var packInputs []costing.PackagingCostInput
	for _, pk := range p.PackagingItems() {
		uCost := 0.0
		if cache != nil {
			uCost = cache.PackUnitCosts[pk.PackagingItemID()]
		}
		packInputs = append(packInputs, costing.PackagingCostInput{
			UnitCost:     uCost,
			QuantityUsed: pk.QuantityUsed(),
		})
	}

	batchSize := p.BatchSize()
	if batchSize <= 0 {
		batchSize = 1
	}

	var machInput *costing.MachineCostInput
	if cache != nil {
		if p.DefaultMachineID() != nil {
			machInput = cache.Machines[*p.DefaultMachineID()]
		}
		if machInput == nil {
			machInput = cache.DefaultMach
		}
	}

	var shopCfgInput *costing.ShopConfigInput
	if cache != nil {
		shopCfgInput = cache.ShopConfig
	}

	cb := costing.Calculate(costing.CalculationInput{
		MaterialType:          p.MaterialType(),
		DefaultWeightGrams:    p.DefaultWeightGrams(),
		DefaultPrintTimeHours: p.DefaultPrintTimeHours(),
		BatchSize:             batchSize,
		PackingFeeIDR:         p.PackingFeeIDR(),
		BaseHPP:               p.BaseHPP(),
		BaseSellingPrice:      p.BaseSellingPrice(),
		Components:            compInputs,
		PackagingItems:        packInputs,
		Machine:               machInput,
		ShopConfig:            shopCfgInput,
	})

	filCost := cb.FilamentCost * q
	emergencyCost := cb.EmergencyFundCost * q
	compCost := cb.HardwareCost * q
	packCost := cb.PackagingCost * q
	elecCost := cb.ElectricityCost * q
	maintCost := cb.MaintenanceCost * q
	depCost := cb.DepreciationCost * q
	totalCOGS := cb.BaseHPP * q

	if totalCOGS == 0 && p.BaseHPP() > 0 {
		totalCOGS = p.BaseHPP() * q
		filCost = totalCOGS
	}

	netProf := (sellingPrice * q) - totalCOGS

	return ItemCostBreakdown{
		FilamentCost:     filCost,
		EmergencyCost:    emergencyCost,
		ComponentCost:    compCost,
		PackagingCost:    packCost,
		ElectricityCost:  elecCost,
		MaintenanceCost:  maintCost,
		DepreciationCost: depCost,
		TotalCOGS:        totalCOGS,
		NetProfit:        netProf,
	}
}

func (cc *CostCalculator) ComputeCost(
	ctx context.Context,
	userID uuid.UUID,
	p *product.Product,
	qty int,
	sellingPrice float64,
) ItemCostBreakdown {
	cache := cc.LoadContext(ctx, userID)
	return cc.CalculateProductCosts(p, qty, sellingPrice, cache)
}

