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
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

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

	v1 := r.Group("/api/v1")
	{
		shopeeRoutes := v1.Group("/shopee")
		{
			// Auth & Toko
			shopeeRoutes.GET("/auth-url", shopeeHandler.GetAuthURL)
			shopeeRoutes.GET("/callback", shopeeHandler.HandleCallback)
			shopeeRoutes.GET("/shops", shopeeHandler.GetShops)
			shopeeRoutes.POST("/shops/:shop_id/refresh", shopeeHandler.RefreshToken)

			// Pesanan & Keuangan (Escrow)
			shopeeRoutes.POST("/shops/:shop_id/sync-orders", orderHandler.SyncOrders)
			shopeeRoutes.GET("/shops/:shop_id/orders", orderHandler.GetOrders)
			shopeeRoutes.GET("/orders/:order_sn", orderHandler.GetOrderDetail)

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
	}

	return r
}
