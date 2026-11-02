package shopee

import (
	"context"
	"time"

	"github.com/google/uuid"

	"symetra-lab-backend-v2/internal/domain/finance"
	"symetra-lab-backend-v2/internal/domain/shopee"
	usecaseProduct "symetra-lab-backend-v2/internal/usecase/product"
)

type LedgerSynchronizer interface {
	SyncRunningBalances(ctx context.Context) error
}

type ProcessShopeeOrderUseCase struct {
	repo       shopee.Repository
	skuMatcher *usecaseProduct.SKUMatcher
	ledgerSync LedgerSynchronizer
}

func NewProcessShopeeOrderUseCase(
	repo shopee.Repository,
	skuMatcher *usecaseProduct.SKUMatcher,
	ledgerSync LedgerSynchronizer,
) *ProcessShopeeOrderUseCase {
	return &ProcessShopeeOrderUseCase{
		repo:       repo,
		skuMatcher: skuMatcher,
		ledgerSync: ledgerSync,
	}
}

func (uc *ProcessShopeeOrderUseCase) Execute(
	ctx context.Context,
	order *shopee.ShopeeOrder,
	cache *usecaseProduct.CostingContext,
) error {
	if order == nil {
		return nil
	}

	// 1. Calculate items total price from the store perspective
	var itemsTotal float64
	for _, it := range order.Items() {
		p := it.DiscountedPrice()
		if p <= 0 {
			p = it.OriginalPrice()
		}
		itemsTotal += p * float64(it.Quantity())
	}

	// 2. Determine order revenue and payment status
	var grossAmount, netAmount, channelFee float64
	paymentStatus := "UNPAID"
	isRefunded := false

	existingOrder, _ := uc.repo.FindOrder(ctx, order.OrderSN())

	if order.Escrow() != nil && order.Escrow().FinancialStatus() == "REFUNDED" {
		isRefunded = true
	} else if order.Escrow() != nil && order.Escrow().EscrowAmount() == 0 && (order.OrderStatus() == "COMPLETED" || order.OrderStatus() == "SHIPPED") && (itemsTotal > 0 || (existingOrder != nil && existingOrder.Escrow() != nil && existingOrder.Escrow().EscrowAmount() > 0)) {
		isRefunded = true
	}

	if isRefunded {
		grossAmount = 0
		netAmount = 0
		channelFee = 0
		paymentStatus = "REFUNDED"
	} else if order.Escrow() != nil && order.Escrow().EscrowAmount() > 0 {
		e := order.Escrow()
		grossAmount = e.SellingPrice()
		if grossAmount <= 0 {
			grossAmount = itemsTotal
		}
		if grossAmount <= 0 {
			grossAmount = e.EscrowAmount() + e.TotalMarketplaceFee()
		}
		netAmount = e.EscrowAmount()
		channelFee = e.TotalMarketplaceFee()
		if e.FinancialStatus() == "RELEASED" {
			paymentStatus = "PAID"
		}
	} else {
		if itemsTotal > 0 {
			grossAmount = itemsTotal
		} else {
			grossAmount = order.TotalAmount()
		}

		if existingOrder != nil && existingOrder.Escrow() != nil && existingOrder.Escrow().EscrowAmount() > 0 {
			netAmount = existingOrder.Escrow().EscrowAmount()
			channelFee = existingOrder.Escrow().TotalMarketplaceFee()
			paymentStatus = "PAID"
			if grossAmount <= 0 && existingOrder.Escrow().SellingPrice() > 0 {
				grossAmount = existingOrder.Escrow().SellingPrice()
			}
		} else {
			if channelFee > 0 {
				netAmount = grossAmount - channelFee
			} else {
				netAmount = grossAmount
			}
		}
	}

	if grossAmount <= 0 && itemsTotal > 0 {
		grossAmount = itemsTotal
	}
	if grossAmount <= 0 && netAmount > 0 {
		grossAmount = netAmount + channelFee
	}

	if paymentStatus == "PAID" && order.PaidAt() == nil {
		now := time.Now()
		order.SetPaidAt(&now)
	}

	// 3. Match items with SKU matcher and compute item costs
	var totFil, totComp, totPack, totElec, totMaint, totDep, totCOGS float64
	items := order.Items()
	for i := range items {
		item := &items[i]

		var existingProdID *uuid.UUID
		if existingOrder != nil {
			for _, exIt := range existingOrder.Items() {
				if exIt.ItemID() == item.ItemID() && (item.ModelID() == 0 || exIt.ModelID() == item.ModelID()) {
					existingProdID = exIt.ProductID()
					break
				}
			}
		}

		matchRes := uc.skuMatcher.MatchItem(ctx, usecaseProduct.MatchItemInput{
			UserID:         uuid.Nil,
			ModelSKU:       item.ModelSKU(),
			ItemSKU:        item.ItemSKU(),
			ChannelItemID:  item.ItemID(),
			ChannelModelID: item.ModelID(),
			ItemName:       item.ItemName(),
			Quantity:       item.Quantity(),
			SellingPrice:   item.DiscountedPrice(),
			ExistingProdID: existingProdID,
		}, cache)

		item.SetCalculatedCosts(
			matchRes.ProductID,
			matchRes.MatchedSKU,
			matchRes.MappingStatus,
			matchRes.Costs.FilamentCost,
			matchRes.Costs.ComponentCost,
			matchRes.Costs.PackagingCost,
			matchRes.Costs.ElectricityCost,
			matchRes.Costs.MaintenanceCost,
			matchRes.Costs.DepreciationCost,
			matchRes.Costs.TotalCOGS,
		)

		totFil += matchRes.Costs.FilamentCost
		totComp += matchRes.Costs.ComponentCost
		totPack += matchRes.Costs.PackagingCost
		totElec += matchRes.Costs.ElectricityCost
		totMaint += matchRes.Costs.MaintenanceCost
		totDep += matchRes.Costs.DepreciationCost
		totCOGS += matchRes.Costs.TotalCOGS
	}
	order.SetItems(items)

	// 4. Allocate 8 pos kas using centralized domain service
	var alloc finance.FundAllocationResult
	if isRefunded {
		alloc = finance.FundAllocationResult{
			GrossAmount:      0,
			ChannelFee:       0,
			NetAmount:        0,
			COGSAmount:       totCOGS,
			NetProfit:        -totCOGS,
			FundFilament:     0,
			FundEmergency:    0,
			FundComponent:    0,
			FundPackaging:    0,
			FundElectricity:  0,
			FundMaintenance:  0,
			FundDepreciation: 0,
			FundNetProfit:    -totCOGS,
		}
	} else {
		alloc = finance.AllocateOrderFunds(finance.FundAllocationInput{
			GrossAmount:      grossAmount,
			ChannelFee:       channelFee,
			FilamentCost:     totFil,
			ComponentCost:    totComp,
			PackagingCost:    totPack,
			ElectricityCost:  totElec,
			MaintenanceCost:  totMaint,
			DepreciationCost: totDep,
		})
	}

	// 5. Update escrow entity if present
	if order.Escrow() != nil {
		order.Escrow().SetCalculatedFunds(
			alloc.COGSAmount,
			alloc.FundFilament,
			alloc.FundEmergency,
			alloc.FundComponent,
			alloc.FundPackaging,
			alloc.FundElectricity,
			alloc.FundMaintenance,
			alloc.FundDepreciation,
			alloc.FundNetProfit,
		)
	}

	// 6. Save through repository (pure persistence)
	if err := uc.repo.SaveOrder(ctx, order); err != nil {
		return err
	}

	// 7. Explicit Ledger Synchronizer (no hidden DB side effects!)
	if uc.ledgerSync != nil {
		_ = uc.ledgerSync.SyncRunningBalances(ctx)
	}

	return nil
}
