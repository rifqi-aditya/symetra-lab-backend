package app

import (
	"context"
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
	shopeeUC "symetra-lab-backend-v2/internal/usecase/shopee"
)

// InitEchoApp initializes the Echo engine and wires all modular dependencies.
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

	// 1. Wire Core Domain & Application Modules
	orderRepo := postgresRepo.NewOrderRepository(db)

	productMod := wireProductModule(db, cfg)
	financeMod := wireFinanceModule(db, orderRepo)
	orderMod := wireOrderModule(orderRepo, productMod.ProductRepo, financeMod.FinanceRepo)
	marketplaceMod := wireMarketplaceModule(db, cfg, productMod, financeMod.FinanceRepo)

	authHandler := handler.NewAuthHandler(cfg)

	e := echo.New()
	e.HideBanner = true

	// Pre-middleware for Vercel/proxy path rewriting
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

	// 2. Register HTTP Routes
	deliveryHTTP.SetupRouter(e, deliveryHTTP.RouterConfig{
		AppConfig:         cfg,
		AuthHandler:       authHandler,
		ProductHandler:    productMod.ProductHandler,
		CategoryHandler:   productMod.CategoryHandler,
		OrderHandler:      orderMod.OrderHandler,
		FilamentHandler:   productMod.FilamentHandler,
		MachineHandler:    productMod.MachineHandler,
		ComponentHandler:  productMod.ComponentHandler,
		PackagingHandler:  productMod.PackagingHandler,
		ShopeeHandler:          marketplaceMod.ShopeeHandler,
		MarketplaceCredHandler: marketplaceMod.MarketplaceCredHandler,
		ShopConfigHandler:      productMod.ShopConfigHandler,
		FinanceHandler:    financeMod.FinanceHandler,
		DashboardHandler:  financeMod.DashboardHandler,
	})

	// 3. Initial Ledger Verification
	if err := postgresRepo.SyncCashAccountLedger(context.Background(), db); err != nil {
		log.Printf("[Symetra Lab v2] Warning: Initial SyncCashAccountLedger failed: %v", err)
	}

	log.Println("[Symetra Lab v2] Echo application initialized successfully.")
	return e, cfg, nil
}

// WireShopeeService merakit modul ShopeeUseCases untuk keperluan CLI dan task background
func WireShopeeService(db *gorm.DB, cfg *config.Config) *shopeeUC.ShopeeUseCases {
	productMod := wireProductModule(db, cfg)
	orderRepo := postgresRepo.NewOrderRepository(db)
	financeMod := wireFinanceModule(db, orderRepo)
	shopeeUC, _ := WireShopeeUseCases(db, cfg, productMod, financeMod.FinanceRepo)
	return shopeeUC
}
