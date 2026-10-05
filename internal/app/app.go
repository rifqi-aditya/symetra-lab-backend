package app

import (
	"log"
	"time"

	"github.com/labstack/echo/v4"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"symetra-lab-backend-v2/config"
	deliveryHTTP "symetra-lab-backend-v2/internal/delivery/http"
	"symetra-lab-backend-v2/internal/delivery/http/handler"
	postgresRepo "symetra-lab-backend-v2/internal/repository/postgres"
	categoryUC "symetra-lab-backend-v2/internal/usecase/category"
	componentUC "symetra-lab-backend-v2/internal/usecase/component"
	filamentUC "symetra-lab-backend-v2/internal/usecase/filament"
	financeUC "symetra-lab-backend-v2/internal/usecase/finance"
	machineUC "symetra-lab-backend-v2/internal/usecase/machine"
	orderUC "symetra-lab-backend-v2/internal/usecase/order"
	packagingUC "symetra-lab-backend-v2/internal/usecase/packaging"
	productUC "symetra-lab-backend-v2/internal/usecase/product"
	productionUC "symetra-lab-backend-v2/internal/usecase/production"
	shopeeUC "symetra-lab-backend-v2/internal/usecase/shopee"
	shopUC "symetra-lab-backend-v2/internal/usecase/shopconfig"
	pkgShopee "symetra-lab-backend-v2/pkg/shopee"
)

// InitEchoApp menginisialisasi Echo engine dan semua dependency injection
func InitEchoApp() (*echo.Echo, *config.Config, error) {
	cfg := config.Load()

	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  cfg.DatabaseURL,
		PreferSimpleProtocol: true,
	}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, nil, err
	}

	if sqlDB, err := db.DB(); err == nil {
		sqlDB.SetMaxIdleConns(5)
		sqlDB.SetMaxOpenConns(20)
		sqlDB.SetConnMaxLifetime(60 * time.Minute)
		sqlDB.SetConnMaxIdleTime(15 * time.Minute)
	}

	productRepo := postgresRepo.NewProductRepository(db)
	categoryRepo := postgresRepo.NewCategoryRepository(db)
	shopConfigRepo := postgresRepo.NewShopConfigRepository(db)
	orderRepo := postgresRepo.NewOrderRepository(db)
	filamentRepo := postgresRepo.NewFilamentRepository(db)
	machineRepo := postgresRepo.NewMachineRepository(db)
	componentRepo := postgresRepo.NewComponentRepository(db)
	packagingRepo := postgresRepo.NewPackagingRepository(db)
	shopeeRepo := postgresRepo.NewShopeeRepository(db)
	productionRepo := postgresRepo.NewProductionRepository(db)
	financeRepo := postgresRepo.NewFinanceRepository(db)

	shopeeClient := pkgShopee.NewClient(
		cfg.ShopeePartnerID,
		cfg.ShopeePartnerKey,
		cfg.ShopeeIsProduction,
		cfg.ShopeeRedirectURL,
	)
	shopeeUseCases := shopeeUC.NewShopeeUseCases(cfg, shopeeRepo, shopeeClient, financeRepo)

	listTransactionsUC := financeUC.NewListTransactionsUseCase(financeRepo)
	createTransactionUC := financeUC.NewCreateTransactionUseCase(financeRepo)
	deleteTransactionUC := financeUC.NewDeleteTransactionUseCase(financeRepo)
	getSummaryUC := financeUC.NewGetSummaryUseCase(financeRepo)
	createPurchaseOrderUC := financeUC.NewCreatePurchaseOrderUseCase(financeRepo)
	listPurchaseOrdersUC := financeUC.NewListPurchaseOrdersUseCase(financeRepo)
	getPurchaseOrderUC := financeUC.NewGetPurchaseOrderUseCase(financeRepo)
	deletePurchaseOrderUC := financeUC.NewDeletePurchaseOrderUseCase(financeRepo)
	createCapitalRecordUC := financeUC.NewCreateCapitalRecordUseCase(financeRepo)
	listCapitalRecordsUC := financeUC.NewListCapitalRecordsUseCase(financeRepo)
	getTotalCapitalUC := financeUC.NewGetTotalCapitalUseCase(financeRepo)
	getOrderAllocUC := financeUC.NewGetOrderAllocationUseCase(orderRepo)

	listProductsUC := productUC.NewListProductsUseCase(productRepo)
	getProductUC := productUC.NewGetProductUseCase(productRepo, shopConfigRepo, machineRepo, componentRepo, packagingRepo)
	createProductUC := productUC.NewCreateProductUseCase(productRepo, shopConfigRepo, machineRepo, componentRepo, packagingRepo)
	updateProductUC := productUC.NewUpdateProductUseCase(productRepo, shopConfigRepo, machineRepo, componentRepo, packagingRepo)
	deleteProductUC := productUC.NewDeleteProductUseCase(productRepo)
	autoGenSKUUC := productUC.NewAutoGenerateSKUsUseCase(productRepo)

	listCategoriesUC := categoryUC.NewListCategoriesUseCase(categoryRepo)
	createCategoryUC := categoryUC.NewCreateCategoryUseCase(categoryRepo)
	updateCategoryUC := categoryUC.NewUpdateCategoryUseCase(categoryRepo)
	deleteCategoryUC := categoryUC.NewDeleteCategoryUseCase(categoryRepo)

	listOrdersUC := orderUC.NewListOrdersUseCase(orderRepo)
	getOrderUC := orderUC.NewGetOrderUseCase(orderRepo)
	createOrderUC := orderUC.NewCreateOrderUseCase(orderRepo, productRepo)
	updateOrderStatusUC := orderUC.NewUpdateOrderStatusUseCase(orderRepo)
	updatePaymentStatusUC := orderUC.NewUpdatePaymentStatusUseCase(orderRepo)
	deleteOrderUC := orderUC.NewDeleteOrderUseCase(orderRepo)

	listFilamentsUC := filamentUC.NewListFilamentsUseCase(filamentRepo)
	getFilamentUC := filamentUC.NewGetFilamentUseCase(filamentRepo)
	createFilamentUC := filamentUC.NewCreateFilamentUseCase(filamentRepo)
	updateFilamentUC := filamentUC.NewUpdateFilamentUseCase(filamentRepo)
	deleteFilamentUC := filamentUC.NewDeleteFilamentUseCase(filamentRepo)
	syncStockUC := filamentUC.NewSyncStockUseCase(filamentRepo)
	listProfilesUC := filamentUC.NewListFilamentProfilesUseCase(filamentRepo)
	listRatesUC := filamentUC.NewListMaterialRatesUseCase(filamentRepo)
	createRateUC := filamentUC.NewCreateMaterialRateUseCase(filamentRepo)
	updateRateUC := filamentUC.NewUpdateMaterialRateUseCase(filamentRepo)
	deleteRateUC := filamentUC.NewDeleteMaterialRateUseCase(filamentRepo)

	listMachinesUC := machineUC.NewListMachinesUseCase(machineRepo)
	getMachineUC := machineUC.NewGetMachineUseCase(machineRepo)
	createMachineUC := machineUC.NewCreateMachineUseCase(machineRepo)
	updateMachineUC := machineUC.NewUpdateMachineUseCase(machineRepo)
	updateMachineStateUC := machineUC.NewUpdateMachineStateUseCase(machineRepo)
	deleteMachineUC := machineUC.NewDeleteMachineUseCase(machineRepo)
	listPartsUC := machineUC.NewListPartsUseCase(machineRepo)
	createPartUC := machineUC.NewCreatePartUseCase(machineRepo)
	updatePartUC := machineUC.NewUpdatePartUseCase(machineRepo)
	replacePartUC := machineUC.NewReplacePartUseCase(machineRepo)
	deletePartUC := machineUC.NewDeletePartUseCase(machineRepo)

	listComponentsUC := componentUC.NewListComponentsUseCase(componentRepo)
	getComponentUC := componentUC.NewGetComponentUseCase(componentRepo)
	createComponentUC := componentUC.NewCreateComponentUseCase(componentRepo)
	updateComponentUC := componentUC.NewUpdateComponentUseCase(componentRepo)
	deleteComponentUC := componentUC.NewDeleteComponentUseCase(componentRepo)

	listPackagingUC := packagingUC.NewListPackagingUseCase(packagingRepo)
	getPackagingUC := packagingUC.NewGetPackagingUseCase(packagingRepo)
	createPackagingUC := packagingUC.NewCreatePackagingItemUseCase(packagingRepo)
	updatePackagingUC := packagingUC.NewUpdatePackagingItemUseCase(packagingRepo)
	deletePackagingUC := packagingUC.NewDeletePackagingItemUseCase(packagingRepo)

	listPresetsUC := packagingUC.NewListPackagingPresetsUseCase(packagingRepo)
	createPresetUC := packagingUC.NewCreatePresetUseCase(packagingRepo)
	updatePresetUC := packagingUC.NewUpdatePresetUseCase(packagingRepo)
	deletePresetUC := packagingUC.NewDeletePresetUseCase(packagingRepo)

	getShopConfigUC := shopUC.NewGetShopConfigUseCase(shopConfigRepo)
	updateShopConfigUC := shopUC.NewUpdateShopConfigUseCase(shopConfigRepo)
	listMarketplacesUC := shopUC.NewListMarketplacesUseCase(shopConfigRepo)
	createMarketplaceUC := shopUC.NewCreateMarketplaceUseCase(shopConfigRepo)
	updateMarketplaceUC := shopUC.NewUpdateMarketplaceUseCase(shopConfigRepo)
	deleteMarketplaceUC := shopUC.NewDeleteMarketplaceUseCase(shopConfigRepo)

	getProductionQueueUC := productionUC.NewGetProductionQueueUseCase(productionRepo)
	completeJobUC := productionUC.NewCompleteJobUseCase(productionRepo)

	authHandler := handler.NewAuthHandler(cfg)
	productHandler := handler.NewProductHandler(
		listProductsUC,
		getProductUC,
		createProductUC,
		updateProductUC,
		deleteProductUC,
		autoGenSKUUC,
	)
	categoryHandler := handler.NewCategoryHandler(
		listCategoriesUC,
		createCategoryUC,
		updateCategoryUC,
		deleteCategoryUC,
	)
	orderHandler := handler.NewOrderHandler(
		listOrdersUC,
		getOrderUC,
		createOrderUC,
		updateOrderStatusUC,
		updatePaymentStatusUC,
		deleteOrderUC,
	)
	filamentHandler := handler.NewFilamentHandler(
		listFilamentsUC,
		getFilamentUC,
		createFilamentUC,
		updateFilamentUC,
		deleteFilamentUC,
		syncStockUC,
		listProfilesUC,
		listRatesUC,
		createRateUC,
		updateRateUC,
		deleteRateUC,
	)
	machineHandler := handler.NewMachineHandler(
		listMachinesUC,
		getMachineUC,
		createMachineUC,
		updateMachineUC,
		updateMachineStateUC,
		deleteMachineUC,
		listPartsUC,
		createPartUC,
		updatePartUC,
		replacePartUC,
		deletePartUC,
	)
	componentHandler := handler.NewComponentHandler(
		listComponentsUC,
		getComponentUC,
		createComponentUC,
		updateComponentUC,
		deleteComponentUC,
	)
	shopeeHandler := handler.NewShopeeHandler(cfg, shopeeUseCases)

	shopConfigHandler := handler.NewShopConfigHandler(
		getShopConfigUC,
		updateShopConfigUC,
		listMarketplacesUC,
		createMarketplaceUC,
		updateMarketplaceUC,
		deleteMarketplaceUC,
	)

	productionHandler := handler.NewProductionHandler(
		getProductionQueueUC,
		completeJobUC,
	)

	packagingHandler := handler.NewPackagingHandler(
		listPackagingUC,
		getPackagingUC,
		createPackagingUC,
		updatePackagingUC,
		deletePackagingUC,
		listPresetsUC,
		createPresetUC,
		updatePresetUC,
		deletePresetUC,
	)

	financeHandler := handler.NewFinanceHandler(
		listTransactionsUC,
		createTransactionUC,
		deleteTransactionUC,
		getSummaryUC,
		createPurchaseOrderUC,
		listPurchaseOrdersUC,
		getPurchaseOrderUC,
		deletePurchaseOrderUC,
		createCapitalRecordUC,
		listCapitalRecordsUC,
		getTotalCapitalUC,
		getOrderAllocUC,
	)

	e := echo.New()
	e.HideBanner = true

	// Pre-middleware untuk menangani rewrite path dari proxy/Vercel (query param path atau header)
	e.Pre(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			requestedPath := ""

			if qPath := req.URL.Query().Get("path"); qPath != "" {
				requestedPath = qPath
			} else if origPath := req.Header.Get("x-matched-path"); origPath != "" {
				requestedPath = origPath
			} else if origPath := req.Header.Get("x-vercel-matched-path"); origPath != "" {
				requestedPath = origPath
			}

			if requestedPath != "" {
				q := req.URL.Query()
				q.Del("path")
				req.URL.Path = requestedPath
				req.URL.RawPath = requestedPath
				req.URL.RawQuery = q.Encode()
				if q.Encode() != "" {
					req.RequestURI = requestedPath + "?" + q.Encode()
				} else {
					req.RequestURI = requestedPath
				}
			}
			return next(c)
		}
	})

	deliveryHTTP.SetupRouter(e, deliveryHTTP.RouterConfig{
		AppConfig:         cfg,
		AuthHandler:       authHandler,
		ProductHandler:    productHandler,
		CategoryHandler:   categoryHandler,
		OrderHandler:      orderHandler,
		FilamentHandler:   filamentHandler,
		MachineHandler:    machineHandler,
		ComponentHandler:  componentHandler,
		PackagingHandler:  packagingHandler,
		ShopeeHandler:     shopeeHandler,
		ShopConfigHandler: shopConfigHandler,
		ProductionHandler: productionHandler,
		FinanceHandler:    financeHandler,
	})

	log.Println("[Symetra Lab v2] Echo application initialized successfully.")
	return e, cfg, nil
}
