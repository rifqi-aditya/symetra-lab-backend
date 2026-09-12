package routes

import (
	"net/http"

	"symetra-lab-backend/config"
	"symetra-lab-backend/handlers"
	"symetra-lab-backend/pkg/shopee"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SetupRouter mengatur semua rute HTTP dan middleware Gin
func SetupRouter(cfg *config.Config, db *gorm.DB, shopeeClient *shopee.Client) *gin.Engine {
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// Middleware CORS agar backend bisa dipanggil dari frontend web
	r.Use(func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if origin != "" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		} else {
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		}
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		reqHeaders := c.Request.Header.Get("Access-Control-Request-Headers")
		if reqHeaders != "" {
			c.Writer.Header().Set("Access-Control-Allow-Headers", reqHeaders)
		} else {
			c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-User-ID, x-user-id")
		}

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	// Endpoint health check
	r.GET("/health", func(c *gin.Context) {
		dbStatus := "connected"
		if db == nil {
			dbStatus = "disconnected / not configured"
		}
		c.JSON(http.StatusOK, gin.H{
			"status":      "ok",
			"database":    dbStatus,
			"environment": cfg.AppEnv,
		})
	})

	// Root endpoint
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"app":     "Symetra Lab Backend",
			"version": "1.0.0",
			"shopee":  "Open Platform v2 Ready",
		})
	})

	// Group route Shopee
	shopeeHandler := handlers.NewShopeeHandler(db, shopeeClient)
	orderHandler := handlers.NewOrderHandler(db, shopeeClient)
	logisticsHandler := handlers.NewLogisticsHandler(db, shopeeClient)
	filamentHandler := handlers.NewFilamentHandler(db)
	machineHandler := handlers.NewMachineHandler(db)
	componentHandler := handlers.NewComponentHandler(db)
	packagingHandler := handlers.NewPackagingHandler(db)
	productHandler := handlers.NewProductHandler(db)
	configHandler := handlers.NewConfigHandler(db)
	manualOrderHandler := handlers.NewManualOrderHandler(db)
	productionHandler := handlers.NewProductionHandler(db)
	authHandler := handlers.NewAuthHandler(cfg)
	categoryHandler := handlers.NewCategoryHandler(db)

	v1 := r.Group("/api/v1")
	{
		// Autentikasi Pengguna
		authRoutes := v1.Group("/auth")
		{
			authRoutes.POST("/login", authHandler.Login)
			authRoutes.GET("/me", authHandler.GetProfile)
			authRoutes.POST("/logout", authHandler.Logout)
		}

		// Kategori Produk
		categoryRoutes := v1.Group("/categories")
		{
			categoryRoutes.GET("", categoryHandler.GetAll)
			categoryRoutes.POST("", categoryHandler.Create)
			categoryRoutes.PUT("/:id", categoryHandler.Update)
			categoryRoutes.DELETE("/:id", categoryHandler.Delete)
		}

		shopeeRoutes := v1.Group("/shopee")
		{
			// Auth & Toko
			shopeeRoutes.GET("/auth-url", shopeeHandler.GetAuthURL)
			shopeeRoutes.GET("/callback", shopeeHandler.HandleCallback)
			shopeeRoutes.GET("/shops", shopeeHandler.GetShops)
			shopeeRoutes.POST("/shops/:shop_id/refresh", shopeeHandler.RefreshToken)

			// Pesanan & Keuangan (Escrow & Alokasi Kas Bengkel)
			shopeeRoutes.POST("/shops/:shop_id/sync-orders", orderHandler.SyncOrders)
			shopeeRoutes.GET("/shops/:shop_id/orders", orderHandler.GetOrders)
			shopeeRoutes.GET("/shops/:shop_id/raw-orders", orderHandler.GetRawOrders)
			shopeeRoutes.GET("/orders/:order_sn", orderHandler.GetOrderDetail)
			shopeeRoutes.POST("/orders/:order_sn/items/:item_id/link-sku", orderHandler.LinkItemSKU)
			shopeeRoutes.GET("/financial/cashflow-summary", orderHandler.GetCashflowSummary)
			shopeeRoutes.POST("/financial/recalculate", orderHandler.RecalculateFinances)

			// Logistik & Cetak Label Thermal
			shopeeRoutes.POST("/orders/:order_sn/ship", logisticsHandler.ShipOrder)
			shopeeRoutes.GET("/orders/:order_sn/shipping-label", logisticsHandler.DownloadShippingLabel)
		}

		// Inventori Bahan Baku: Filamen & Profil Teknis
		filamentRoutes := v1.Group("/filaments")
		{
			filamentRoutes.GET("", filamentHandler.GetAllFilaments)
			filamentRoutes.GET("/:id", filamentHandler.GetFilamentByID)
			filamentRoutes.POST("", filamentHandler.CreateFilament)
			filamentRoutes.PUT("/:id", filamentHandler.UpdateFilament)
			filamentRoutes.DELETE("/:id", filamentHandler.DeleteFilament)
			filamentRoutes.POST("/:id/sync-stock", filamentHandler.SyncStock)
		}
		v1.GET("/filament-profiles", filamentHandler.GetProfiles)

		// Mesin Cetak 3D & Suku Cadang Perawatan
		machineRoutes := v1.Group("/machines")
		{
			machineRoutes.GET("", machineHandler.GetAllMachines)
			machineRoutes.GET("/:id", machineHandler.GetMachineByID)
			machineRoutes.POST("", machineHandler.CreateMachine)
			machineRoutes.PUT("/:id", machineHandler.UpdateMachine)
			machineRoutes.PATCH("/:id/state", machineHandler.UpdateMachineState)
			machineRoutes.DELETE("/:id", machineHandler.DeleteMachine)

			// Sub-rute suku cadang perawatan (Machine Maintenance Parts)
			machineRoutes.GET("/:id/parts", machineHandler.GetPartsByMachine)
			machineRoutes.POST("/:id/parts", machineHandler.CreatePart)
			machineRoutes.PUT("/parts/:part_id", machineHandler.UpdatePart)
			machineRoutes.POST("/parts/:part_id/replace", machineHandler.ReplacePart)
			machineRoutes.DELETE("/parts/:part_id", machineHandler.DeletePart)
		}

		// Komponen Tambahan (Hardware & Aksesoris Non-3D Print)
		componentRoutes := v1.Group("/components")
		{
			componentRoutes.GET("", componentHandler.GetAllComponents)
			componentRoutes.GET("/:id", componentHandler.GetComponentByID)
			componentRoutes.POST("", componentHandler.CreateComponent)
			componentRoutes.PUT("/:id", componentHandler.UpdateComponent)
			componentRoutes.DELETE("/:id", componentHandler.DeleteComponent)
		}

		// Bahan Kemasan Individual (Packaging Items)
		packagingItemRoutes := v1.Group("/packaging-items")
		{
			packagingItemRoutes.GET("", packagingHandler.GetAllPackagingItems)
			packagingItemRoutes.GET("/:id", packagingHandler.GetPackagingItemByID)
			packagingItemRoutes.POST("", packagingHandler.CreatePackagingItem)
			packagingItemRoutes.PUT("/:id", packagingHandler.UpdatePackagingItem)
			packagingItemRoutes.DELETE("/:id", packagingHandler.DeletePackagingItem)
		}

		// Preset Kemasan / Bundel Packing (Packaging Presets)
		packagingPresetRoutes := v1.Group("/packaging-presets")
		{
			packagingPresetRoutes.GET("", packagingHandler.GetAllPresets)
			packagingPresetRoutes.GET("/:id", packagingHandler.GetPresetByID)
			packagingPresetRoutes.POST("", packagingHandler.CreatePreset)
			packagingPresetRoutes.PUT("/:id", packagingHandler.UpdatePreset)
			packagingPresetRoutes.DELETE("/:id", packagingHandler.DeletePreset)
		}

		// Master Produk, Resep BOM & SKU
		productRoutes := v1.Group("/products")
		{
			productRoutes.GET("", productHandler.ListProducts)
			productRoutes.GET("/:id", productHandler.GetProduct)
			productRoutes.GET("/by-sku/:sku", productHandler.GetProductBySKU)
			productRoutes.POST("", productHandler.CreateProduct)
			productRoutes.PUT("/:id", productHandler.UpdateProduct)
			productRoutes.DELETE("/:id", productHandler.DeleteProduct)
			productRoutes.POST("/generate-skus", productHandler.AutoGenerateDraftSKUs)
		}

		// Pengaturan Bengkel & Platform Marketplace
		configRoutes := v1.Group("/config")
		{
			configRoutes.GET("/shop", configHandler.GetShopConfig)
			configRoutes.PUT("/shop", configHandler.UpdateShopConfig)
			configRoutes.GET("/marketplaces", configHandler.GetMarketplaces)
			configRoutes.POST("/marketplaces", configHandler.CreateMarketplace)
			configRoutes.PUT("/marketplaces/:id", configHandler.UpdateMarketplace)
			configRoutes.DELETE("/marketplaces/:id", configHandler.DeleteMarketplace)
		}

		// Pesanan Manual / Offline (Direct Orders Hub)
		orderRoutes := v1.Group("/orders")
		{
			orderRoutes.GET("", manualOrderHandler.GetOrders)
			orderRoutes.GET("/:id", manualOrderHandler.GetOrderByID)
			orderRoutes.POST("", manualOrderHandler.CreateOrder)
			orderRoutes.PATCH("/:id/status", manualOrderHandler.UpdateOrderStatus)
			orderRoutes.PATCH("/:id/payment", manualOrderHandler.UpdatePaymentStatus)
			orderRoutes.DELETE("/:id", manualOrderHandler.DeleteOrder)
		}

		// Antrean Cetak Terpadu (Production Queue & Auto-Deduct)
		productionRoutes := v1.Group("/production")
		{
			productionRoutes.GET("/queue", productionHandler.GetQueue)
			productionRoutes.POST("/complete-job", productionHandler.CompleteJob)
		}
	}

	return r
}
