package postgres

import (
	"context"

	"gorm.io/gorm"
)

// SyncManualOrdersFinanceTransactions synchronizes paid manual/offline orders with finance_transactions.
// Paid orders create/update an INCOME record; unpaid/cancelled orders have their records removed.
func SyncManualOrdersFinanceTransactions(ctx context.Context, db *gorm.DB) error {
	manualSyncSQL := `
		-- 1. Upsert transaksi pemasukan untuk pesanan manual/offline yang berstatus PAID dan tidak CANCELLED
		INSERT INTO finance_transactions (
			id,
			type,
			category,
			amount,
			quantity,
			description,
			transaction_date,
			reference_type,
			reference_id,
			notes,
			created_at,
			updated_at
		)
		SELECT 
			gen_random_uuid(),
			'INCOME',
			'SALES_MANUAL',
			COALESCE(CASE WHEN o.gross_amount > 0 THEN o.gross_amount ELSE o.net_amount END, 0),
			1,
			'Penjualan Langsung - ' || o.order_number,
			o.created_at,
			'MANUAL_ORDER',
			o.id::text,
			'Pesanan langsung workshop lunas',
			NOW(),
			NOW()
		FROM orders o
		WHERE (o.channel = 'MANUAL' OR o.channel NOT IN ('SHOPEE', 'TOKOPEDIA'))
		  AND o.payment_status = 'PAID' 
		  AND o.status != 'CANCELLED'
		  AND (o.gross_amount > 0 OR o.net_amount > 0)
		ON CONFLICT (reference_type, reference_id) 
		WHERE reference_type IS NOT NULL AND reference_type != '' 
		  AND reference_id IS NOT NULL AND reference_id != ''
		DO UPDATE SET
			amount = EXCLUDED.amount,
			description = EXCLUDED.description,
			transaction_date = EXCLUDED.transaction_date,
			updated_at = NOW();

		-- 2. Hapus transaksi pemasukan untuk pesanan manual yang status pembayarannya bukan PAID atau dibatalkan
		DELETE FROM finance_transactions
		WHERE reference_type = 'MANUAL_ORDER'
		  AND reference_id IN (
			  SELECT o.id::text 
			  FROM orders o 
			  WHERE (o.channel = 'MANUAL' OR o.channel NOT IN ('SHOPEE', 'TOKOPEDIA'))
			    AND (o.payment_status != 'PAID' OR o.status = 'CANCELLED')
		  );

		-- 3. Hapus transaksi pemasukan dari pesanan manual yang sudah dihapus dari tabel orders
		DELETE FROM finance_transactions
		WHERE reference_type = 'MANUAL_ORDER'
		  AND reference_id NOT IN (
			  SELECT o.id::text FROM orders o 
			  WHERE (o.channel = 'MANUAL' OR o.channel NOT IN ('SHOPEE', 'TOKOPEDIA'))
		  );
	`
	return db.WithContext(ctx).Exec(manualSyncSQL).Error
}

// SyncCashAccountLedger synchronizes running allocated_amount, spent_amount, and current_balance
// across all 8 cash accounts in the database from settled orders and recorded expenses.
// It also ensures manual order transactions are automatically kept in sync in finance_transactions.
func SyncCashAccountLedger(ctx context.Context, db *gorm.DB) error {
	// 1. Pastikan transaksi pesanan manual selalu sinkron terlebih dahulu
	_ = SyncManualOrdersFinanceTransactions(ctx, db)

	// 2. Sinkronisasi buku besar 8 pos kas
	syncSQL := `
		UPDATE cash_accounts SET 
			allocated_amount = COALESCE((SELECT SUM(fund_filament) FROM orders WHERE payment_status = 'PAID' AND status != 'CANCELLED'), 0),
			spent_amount = COALESCE((SELECT SUM(amount) FROM finance_transactions WHERE type = 'EXPENSE' AND category = 'FILAMENT'), 0),
			current_balance = COALESCE((SELECT SUM(fund_filament) FROM orders WHERE payment_status = 'PAID' AND status != 'CANCELLED'), 0) - COALESCE((SELECT SUM(amount) FROM finance_transactions WHERE type = 'EXPENSE' AND category = 'FILAMENT'), 0),
			updated_at = NOW()
		WHERE code = 'FILAMENT';

		UPDATE cash_accounts SET 
			allocated_amount = COALESCE((SELECT SUM(fund_emergency) FROM orders WHERE payment_status = 'PAID' AND status != 'CANCELLED'), 0),
			spent_amount = COALESCE((SELECT SUM(amount) FROM finance_transactions WHERE type = 'EXPENSE' AND category = 'EMERGENCY_FUND'), 0),
			current_balance = COALESCE((SELECT SUM(fund_emergency) FROM orders WHERE payment_status = 'PAID' AND status != 'CANCELLED'), 0) - COALESCE((SELECT SUM(amount) FROM finance_transactions WHERE type = 'EXPENSE' AND category = 'EMERGENCY_FUND'), 0),
			updated_at = NOW()
		WHERE code = 'EMERGENCY';

		UPDATE cash_accounts SET 
			allocated_amount = COALESCE((SELECT SUM(fund_component) FROM orders WHERE payment_status = 'PAID' AND status != 'CANCELLED'), 0),
			spent_amount = COALESCE((SELECT SUM(amount) FROM finance_transactions WHERE type = 'EXPENSE' AND category = 'HARDWARE'), 0),
			current_balance = COALESCE((SELECT SUM(fund_component) FROM orders WHERE payment_status = 'PAID' AND status != 'CANCELLED'), 0) - COALESCE((SELECT SUM(amount) FROM finance_transactions WHERE type = 'EXPENSE' AND category = 'HARDWARE'), 0),
			updated_at = NOW()
		WHERE code = 'COMPONENT';

		UPDATE cash_accounts SET 
			allocated_amount = COALESCE((SELECT SUM(fund_packaging) FROM orders WHERE payment_status = 'PAID' AND status != 'CANCELLED'), 0),
			spent_amount = COALESCE((SELECT SUM(amount) FROM finance_transactions WHERE type = 'EXPENSE' AND category = 'PACKAGING'), 0),
			current_balance = COALESCE((SELECT SUM(fund_packaging) FROM orders WHERE payment_status = 'PAID' AND status != 'CANCELLED'), 0) - COALESCE((SELECT SUM(amount) FROM finance_transactions WHERE type = 'EXPENSE' AND category = 'PACKAGING'), 0),
			updated_at = NOW()
		WHERE code = 'PACKAGING';

		UPDATE cash_accounts SET 
			allocated_amount = COALESCE((SELECT SUM(fund_electricity) FROM orders WHERE payment_status = 'PAID' AND status != 'CANCELLED'), 0),
			spent_amount = COALESCE((SELECT SUM(amount) FROM finance_transactions WHERE type = 'EXPENSE' AND category = 'ELECTRICITY'), 0),
			current_balance = COALESCE((SELECT SUM(fund_electricity) FROM orders WHERE payment_status = 'PAID' AND status != 'CANCELLED'), 0) - COALESCE((SELECT SUM(amount) FROM finance_transactions WHERE type = 'EXPENSE' AND category = 'ELECTRICITY'), 0),
			updated_at = NOW()
		WHERE code = 'ELECTRICITY';

		UPDATE cash_accounts SET 
			allocated_amount = COALESCE((SELECT SUM(fund_maintenance) FROM orders WHERE payment_status = 'PAID' AND status != 'CANCELLED'), 0),
			spent_amount = COALESCE((SELECT SUM(amount) FROM finance_transactions WHERE type = 'EXPENSE' AND category = 'MACHINE_MAINTENANCE'), 0),
			current_balance = COALESCE((SELECT SUM(fund_maintenance) FROM orders WHERE payment_status = 'PAID' AND status != 'CANCELLED'), 0) - COALESCE((SELECT SUM(amount) FROM finance_transactions WHERE type = 'EXPENSE' AND category = 'MACHINE_MAINTENANCE'), 0),
			updated_at = NOW()
		WHERE code = 'MAINTENANCE';

		UPDATE cash_accounts SET 
			allocated_amount = COALESCE((SELECT SUM(fund_depreciation) FROM orders WHERE payment_status = 'PAID' AND status != 'CANCELLED'), 0),
			spent_amount = COALESCE((SELECT SUM(amount) FROM finance_transactions WHERE type = 'EXPENSE' AND category = 'MACHINE_PURCHASE'), 0),
			current_balance = COALESCE((SELECT SUM(fund_depreciation) FROM orders WHERE payment_status = 'PAID' AND status != 'CANCELLED'), 0) - COALESCE((SELECT SUM(amount) FROM finance_transactions WHERE type = 'EXPENSE' AND category = 'MACHINE_PURCHASE'), 0),
			updated_at = NOW()
		WHERE code = 'DEPRECIATION';

		UPDATE cash_accounts SET 
			allocated_amount = COALESCE((SELECT SUM(fund_net_profit) FROM orders WHERE payment_status = 'PAID' AND status != 'CANCELLED'), 0),
			spent_amount = COALESCE((SELECT SUM(amount) FROM finance_transactions WHERE type = 'EXPENSE' AND category NOT IN ('FILAMENT','HARDWARE','PACKAGING','ELECTRICITY','MACHINE_MAINTENANCE','MACHINE_PURCHASE','EMERGENCY_FUND','REIMBURSEMENT')), 0),
			current_balance = COALESCE((SELECT SUM(fund_net_profit) FROM orders WHERE payment_status = 'PAID' AND status != 'CANCELLED'), 0) - COALESCE((SELECT SUM(amount) FROM finance_transactions WHERE type = 'EXPENSE' AND category NOT IN ('FILAMENT','HARDWARE','PACKAGING','ELECTRICITY','MACHINE_MAINTENANCE','MACHINE_PURCHASE','EMERGENCY_FUND','REIMBURSEMENT')), 0),
			updated_at = NOW()
		WHERE code = 'NET_PROFIT';
	`
	return db.WithContext(ctx).Exec(syncSQL).Error
}
