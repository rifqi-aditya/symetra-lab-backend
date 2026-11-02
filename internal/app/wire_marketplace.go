package app

import (
	"gorm.io/gorm"

	"symetra-lab-backend-v2/config"
	"symetra-lab-backend-v2/internal/delivery/http/handler"
	"symetra-lab-backend-v2/internal/domain/finance"
	postgresRepo "symetra-lab-backend-v2/internal/repository/postgres"
	marketplacecredUC "symetra-lab-backend-v2/internal/usecase/marketplacecred"
	productUC "symetra-lab-backend-v2/internal/usecase/product"
	shopeeUC "symetra-lab-backend-v2/internal/usecase/shopee"
	pkgShopee "symetra-lab-backend-v2/pkg/shopee"
)

type MarketplaceModule struct {
	ShopeeHandler          *handler.ShopeeHandler
	MarketplaceCredHandler *handler.MarketplaceCredHandler
}

func wireMarketplaceModule(
	db *gorm.DB,
	cfg *config.Config,
	productModule *ProductModule,
	financeRepo finance.Repository,
) *MarketplaceModule {
	shopeeRepo := postgresRepo.NewShopeeRepository(db)

	shopeeClient := pkgShopee.NewClient(
		cfg.ShopeePartnerID,
		cfg.ShopeePartnerKey,
		cfg.ShopeeIsProduction,
		cfg.ShopeeRedirectURL,
	)

	costCalculator := productUC.NewCostCalculator(
		productModule.ShopConfigRepo,
		productModule.MachineRepo,
		productModule.ComponentRepo,
		productModule.PackagingRepo,
	)
	skuMatcher := productUC.NewSKUMatcher(productModule.ProductRepo, costCalculator, shopeeRepo)
	processShopeeOrderUC := shopeeUC.NewProcessShopeeOrderUseCase(shopeeRepo, skuMatcher, financeRepo)
	linkShopeeSKUUC := shopeeUC.NewLinkShopeeSKUUseCase(shopeeRepo, financeRepo)
	recalcShopeeFinancesUC := shopeeUC.NewRecalculateFinancesUseCase(shopeeRepo, financeRepo)

	shopeeUseCases := shopeeUC.NewShopeeUseCases(
		cfg,
		shopeeRepo,
		shopeeClient,
		financeRepo,
		processShopeeOrderUC,
		linkShopeeSKUUC,
		recalcShopeeFinancesUC,
	)

	credRepo := postgresRepo.NewMarketplaceCredentialRepository(db, cfg.EncryptionKey)
	credUC := marketplacecredUC.NewUseCase(credRepo, cfg)
	credHandler := handler.NewMarketplaceCredHandler(credUC)

	shopeeUseCases.WithCredUseCase(credUC)

	shopeeHandler := handler.NewShopeeHandler(cfg, shopeeUseCases)
	shopeeHandler.WithCredUseCase(credUC)

	return &MarketplaceModule{
		ShopeeHandler:          shopeeHandler,
		MarketplaceCredHandler: credHandler,
	}
}

