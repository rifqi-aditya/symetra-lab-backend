package main

import (
	"context"
	"fmt"
	"net"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"symetra-lab-backend-v2/config"
	"symetra-lab-backend-v2/internal/app"
)

func main() {
	net.DefaultResolver = &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{Timeout: 5 * time.Second}
			return d.DialContext(ctx, "udp", "8.8.8.8:53")
		},
	}

	cfg := config.Load()
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  cfg.DatabaseURL,
		PreferSimpleProtocol: true,
	}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		fmt.Printf("DB error: %v\n", err)
		return
	}

	shopeeUC := app.WireShopeeService(db, cfg)

	ctx := context.Background()
	fmt.Println("Memulai sinkronisasi seluruh pesanan Shopee 14 hari terakhir (update_time & create_time)...")
	count, err := shopeeUC.SyncShopeeOrders(ctx)
	if err != nil {
		fmt.Printf("SyncShopeeOrders error: %v\n", err)
		return
	}

	fmt.Printf("BERHASIL! Sebanyak %d pesanan Shopee telah disinkronkan ke database.\n", count)
}
