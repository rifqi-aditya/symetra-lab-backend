package main

import (
	"context"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"symetra-lab-backend-v2/config"
	postgresRepo "symetra-lab-backend-v2/internal/repository/postgres"
)

func main() {
	cfg := config.Load()

	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  cfg.DatabaseURL,
		PreferSimpleProtocol: true,
	}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("[Backfill] Connected to database.")

	// 1. Backfill Shopee Escrow Released
	sqlShopee := `
		INSERT INTO finance_transactions (
			type,
			category,
			amount,
			description,
			transaction_date,
			reference_type,
			reference_id,
			notes,
			created_at,
			updated_at
		)
		SELECT 
			'INCOME',
			'SALES_SHOPEE',
			m.escrow_amount,
			'Pencairan Escrow Shopee - Order #' || m.order_sn,
			COALESCE(m.ship_by_date_time::date, o.created_at::date),
			'SHOPEE_ESCROW',
			m.order_sn,
			'Dana bersih masuk setelah potongan biaya marketplace',
			NOW(),
			NOW()
		FROM order_marketplace_details m
		JOIN orders o ON o.id = m.order_id
		WHERE o.status = 'COMPLETED' AND m.financial_status = 'RELEASED'
		ON CONFLICT (reference_type, reference_id) 
		WHERE reference_type IS NOT NULL AND reference_type != '' 
		  AND reference_id IS NOT NULL AND reference_id != ''
		DO NOTHING;
	`
	resShopee := db.Exec(sqlShopee)
	if resShopee.Error != nil {
		log.Fatalf("Failed to backfill Shopee escrows: %v", resShopee.Error)
	}
	log.Printf("[Backfill] Inserted/Verified Shopee escrow records: %d rows affected.", resShopee.RowsAffected)

	// 2. Backfill Paid Direct/Manual Orders
	sqlManual := `
		INSERT INTO finance_transactions (
			type,
			category,
			amount,
			description,
			transaction_date,
			reference_type,
			reference_id,
			notes,
			created_at,
			updated_at
		)
		SELECT 
			'INCOME',
			'SALES_MANUAL',
			o.gross_amount,
			'Penjualan Langsung - ' || o.order_number,
			o.created_at::date,
			'MANUAL_ORDER',
			o.id::text,
			'Pesanan langsung workshop lunas',
			NOW(),
			NOW()
		FROM orders o
		WHERE o.channel = 'MANUAL' AND o.payment_status = 'PAID' AND o.gross_amount > 0
		ON CONFLICT (reference_type, reference_id) 
		WHERE reference_type IS NOT NULL AND reference_type != '' 
		  AND reference_id IS NOT NULL AND reference_id != ''
		DO NOTHING;
	`
	resManual := db.Exec(sqlManual)
	if resManual.Error != nil {
		log.Fatalf("Failed to backfill Manual orders: %v", resManual.Error)
	}
	log.Printf("[Backfill] Inserted/Verified Manual orders: %d rows affected.", resManual.RowsAffected)

	// 3. Sync Cash Account Ledger and Manual Order Transactions
	if err := postgresRepo.SyncCashAccountLedger(context.Background(), db); err != nil {
		log.Printf("[Backfill] Warning: SyncCashAccountLedger failed: %v", err)
	}

	// 4. Count total in finance_transactions
	var totalTx int64
	db.Raw("SELECT COUNT(*) FROM finance_transactions").Scan(&totalTx)
	log.Printf("[Backfill] Completed! Total finance_transactions in DB now: %d", totalTx)
}
