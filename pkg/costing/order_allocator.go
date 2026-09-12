package costing

import (
	"fmt"
	"math"
	"strings"

	"symetra-lab-backend/models"

	"gorm.io/gorm"
)

// AllocateShopeeOrderFinances menghitung alokasi HPP, matching SKU, dan 5 ember kas bengkel untuk 1 pesanan Shopee
func AllocateShopeeOrderFinances(order *models.ShopeeOrder, db *gorm.DB) error {
	if order == nil {
		return fmt.Errorf("order tidak boleh nil")
	}

	// 1. Ambil ShopConfig default
	var cfg models.ShopConfig
	_ = db.First(&cfg).Error

	var totalFilamentCost float64
	var totalHardwarePackagingCost float64
	var totalMachineCost float64
	var totalOrderHPP float64

	// 2. Iterasi setiap item dalam pesanan untuk dicocokkan ke Master Produk
	for i := range order.Items {
		item := &order.Items[i]

		// Cari kandidat SKU dari Shopee: prioritaskan model_sku (varian), fallback ke item_sku
		candidateSKU := strings.TrimSpace(item.ModelSKU)
		if candidateSKU == "" {
			candidateSKU = strings.TrimSpace(item.ItemSKU)
		}

		var matchedProduct *models.Product

		// A. Coba cari berdasarkan SKU di database
		if candidateSKU != "" {
			var prod models.Product
			err := db.Preload("Filaments.Filament.Profile").
				Preload("Components.Component").
				Preload("PackagingPreset.Items.PackagingItem").
				Preload("PackagingItems.PackagingItem").
				Preload("DefaultMachine").
				Where("LOWER(sku) = LOWER(?) OR LOWER(parent_sku) = LOWER(?)", candidateSKU, candidateSKU).
				First(&prod).Error

			if err == nil && prod.ID != "" {
				matchedProduct = &prod
				if prod.SKU != nil {
					item.MatchedSKU = *prod.SKU
				}
				item.MappingStatus = "MATCHED"
			}
		}

		// B. Jika belum ketemu lewat SKU, cek apakah item sudah pernah di-link manual via ProductID
		if matchedProduct == nil && item.ProductID != nil && *item.ProductID != "" {
			var prod models.Product
			err := db.Preload("Filaments.Filament.Profile").
				Preload("Components.Component").
				Preload("PackagingPreset.Items.PackagingItem").
				Preload("PackagingItems.PackagingItem").
				Preload("DefaultMachine").
				Where("id = ?", *item.ProductID).
				First(&prod).Error

			if err == nil && prod.ID != "" {
				matchedProduct = &prod
				if prod.SKU != nil {
					item.MatchedSKU = *prod.SKU
				}
				item.MappingStatus = "MANUAL_LINKED"
			}
		}

		// C. Jika produk fisik ditemukan, hitung HPP riil per unit dan kalikan dengan quantity
		if matchedProduct != nil {
			item.ProductID = &matchedProduct.ID
			breakdown := CalculateCostBreakdown(matchedProduct, &cfg, nil)

			qty := float64(item.Quantity)
			if qty <= 0 {
				qty = 1
			}

			item.FilamentCost = math.Round(breakdown.FilamentCost*qty*100) / 100
			item.HardwareCost = math.Round(breakdown.HardwareCost*qty*100) / 100
			item.PackagingCost = math.Round(breakdown.PackagingCost*qty*100) / 100
			item.MachineCost = math.Round((breakdown.ElectricityCost+breakdown.DepreciationCost)*qty*100) / 100
			item.BaseHPP = math.Round(breakdown.BaseHPP*qty*100) / 100

			totalFilamentCost += item.FilamentCost
			totalHardwarePackagingCost += (item.HardwareCost + item.PackagingCost)
			totalMachineCost += item.MachineCost
			totalOrderHPP += item.BaseHPP
		} else {
			// Belum terpetakan ke master produk fisik
			item.MappingStatus = "UNMAPPED"
			item.FilamentCost = 0
			item.HardwareCost = 0
			item.PackagingCost = 0
			item.MachineCost = 0
			item.BaseHPP = 0
		}
	}

	// 3. Hitung alokasi kas pada Escrow (jika data escrow tersedia)
	if order.Escrow != nil {
		order.Escrow.TotalHPP = math.Round(totalOrderHPP*100) / 100
		order.Escrow.TotalFilamentCost = math.Round(totalFilamentCost*100) / 100
		order.Escrow.TotalHardwarePackagingCost = math.Round(totalHardwarePackagingCost*100) / 100
		order.Escrow.TotalMachineCost = math.Round(totalMachineCost*100) / 100

		// Laba Bersih = Dana Bersih Escrow - Total HPP (Filamen + Hardware + Mesin)
		netProfit := order.Escrow.EscrowAmount - totalOrderHPP
		order.Escrow.NetProfit = math.Round(netProfit*100) / 100

		if order.Escrow.SellingPrice > 0 {
			order.Escrow.ProfitMarginPercent = math.Round((netProfit/order.Escrow.SellingPrice*100)*100) / 100
		}

		// Distribusikan laba bersih ke masing-masing item secara proporsional terhadap harga jual
		totalSellingPrice := order.Escrow.SellingPrice
		for i := range order.Items {
			item := &order.Items[i]
			itemRevenue := item.DiscountedPrice * float64(item.Quantity)
			if itemRevenue <= 0 {
				itemRevenue = item.OriginalPrice * float64(item.Quantity)
			}

			if totalSellingPrice > 0 && item.BaseHPP > 0 {
				proportionalEscrow := (itemRevenue / totalSellingPrice) * order.Escrow.EscrowAmount
				item.NetProfit = math.Round((proportionalEscrow-item.BaseHPP)*100) / 100
			} else {
				item.NetProfit = 0
			}
		}
	}

	return nil
}
