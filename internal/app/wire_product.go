package app

import (
	"gorm.io/gorm"

	"symetra-lab-backend-v2/config"
	"symetra-lab-backend-v2/internal/delivery/http/handler"
	"symetra-lab-backend-v2/internal/domain/component"
	"symetra-lab-backend-v2/internal/domain/machine"
	"symetra-lab-backend-v2/internal/domain/packaging"
	"symetra-lab-backend-v2/internal/domain/product"
	"symetra-lab-backend-v2/internal/domain/shopconfig"
	postgresRepo "symetra-lab-backend-v2/internal/repository/postgres"
	categoryUC "symetra-lab-backend-v2/internal/usecase/category"
	componentUC "symetra-lab-backend-v2/internal/usecase/component"
	filamentUC "symetra-lab-backend-v2/internal/usecase/filament"
	machineUC "symetra-lab-backend-v2/internal/usecase/machine"
	packagingUC "symetra-lab-backend-v2/internal/usecase/packaging"
	productUC "symetra-lab-backend-v2/internal/usecase/product"
	shopUC "symetra-lab-backend-v2/internal/usecase/shopconfig"
)

type ProductModule struct {
	ProductRepo    product.Repository
	ShopConfigRepo shopconfig.Repository
	MachineRepo    machine.Repository
	ComponentRepo  component.Repository
	PackagingRepo  packaging.Repository

	ProductHandler    *handler.ProductHandler
	CategoryHandler   *handler.CategoryHandler
	FilamentHandler   *handler.FilamentHandler
	MachineHandler    *handler.MachineHandler
	ComponentHandler  *handler.ComponentHandler
	PackagingHandler  *handler.PackagingHandler
	ShopConfigHandler *handler.ShopConfigHandler
}

func wireProductModule(db *gorm.DB, cfg *config.Config) *ProductModule {
	productRepo := postgresRepo.NewProductRepository(db)
	categoryRepo := postgresRepo.NewCategoryRepository(db)
	shopConfigRepo := postgresRepo.NewShopConfigRepository(db)
	filamentRepo := postgresRepo.NewFilamentRepository(db)
	machineRepo := postgresRepo.NewMachineRepository(db)
	componentRepo := postgresRepo.NewComponentRepository(db)
	packagingRepo := postgresRepo.NewPackagingRepository(db)

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
	shopConfigHandler := handler.NewShopConfigHandler(
		getShopConfigUC,
		updateShopConfigUC,
		listMarketplacesUC,
		createMarketplaceUC,
		updateMarketplaceUC,
		deleteMarketplaceUC,
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

	return &ProductModule{
		ProductRepo:       productRepo,
		ShopConfigRepo:    shopConfigRepo,
		MachineRepo:       machineRepo,
		ComponentRepo:     componentRepo,
		PackagingRepo:     packagingRepo,
		ProductHandler:    productHandler,
		CategoryHandler:   categoryHandler,
		FilamentHandler:   filamentHandler,
		MachineHandler:    machineHandler,
		ComponentHandler:  componentHandler,
		PackagingHandler:  packagingHandler,
		ShopConfigHandler: shopConfigHandler,
	}
}

