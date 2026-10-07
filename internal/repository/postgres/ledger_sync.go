package postgres

import (
	"context"

	"gorm.io/gorm"
)

// SyncCashAccountLedger synchronizes running allocated_amount, spent_amount, and current_balance
// across all 7 cash accounts in the database from settled orders and recorded expenses.
func SyncCashAccountLedger(ctx context.Context, db *gorm.DB) error {
	syncSQL := `
		UPDATE cash_accounts SET 
			allocated_amount = COALESCE((SELECT SUM(fund_filament) FROM orders WHERE payment_status = 'PAID'), 0),
			spent_amount = COALESCE((SELECT SUM(amount) FROM finance_transactions WHERE type = 'EXPENSE' AND category = 'FILAMENT'), 0),
			current_balance = COALESCE((SELECT SUM(fund_filament) FROM orders WHERE payment_status = 'PAID'), 0) - COALESCE((SELECT SUM(amount) FROM finance_transactions WHERE type = 'EXPENSE' AND category = 'FILAMENT'), 0),
			updated_at = NOW()
		WHERE code = 'FILAMENT';

		UPDATE cash_accounts SET 
			allocated_amount = COALESCE((SELECT SUM(fund_component) FROM orders WHERE payment_status = 'PAID'), 0),
			spent_amount = COALESCE((SELECT SUM(amount) FROM finance_transactions WHERE type = 'EXPENSE' AND category = 'HARDWARE'), 0),
			current_balance = COALESCE((SELECT SUM(fund_component) FROM orders WHERE payment_status = 'PAID'), 0) - COALESCE((SELECT SUM(amount) FROM finance_transactions WHERE type = 'EXPENSE' AND category = 'HARDWARE'), 0),
			updated_at = NOW()
		WHERE code = 'COMPONENT';

		UPDATE cash_accounts SET 
			allocated_amount = COALESCE((SELECT SUM(fund_packaging) FROM orders WHERE payment_status = 'PAID'), 0),
			spent_amount = COALESCE((SELECT SUM(amount) FROM finance_transactions WHERE type = 'EXPENSE' AND category = 'PACKAGING'), 0),
			current_balance = COALESCE((SELECT SUM(fund_packaging) FROM orders WHERE payment_status = 'PAID'), 0) - COALESCE((SELECT SUM(amount) FROM finance_transactions WHERE type = 'EXPENSE' AND category = 'PACKAGING'), 0),
			updated_at = NOW()
		WHERE code = 'PACKAGING';

		UPDATE cash_accounts SET 
			allocated_amount = COALESCE((SELECT SUM(fund_electricity) FROM orders WHERE payment_status = 'PAID'), 0),
			spent_amount = COALESCE((SELECT SUM(amount) FROM finance_transactions WHERE type = 'EXPENSE' AND category = 'ELECTRICITY'), 0),
			current_balance = COALESCE((SELECT SUM(fund_electricity) FROM orders WHERE payment_status = 'PAID'), 0) - COALESCE((SELECT SUM(amount) FROM finance_transactions WHERE type = 'EXPENSE' AND category = 'ELECTRICITY'), 0),
			updated_at = NOW()
		WHERE code = 'ELECTRICITY';

		UPDATE cash_accounts SET 
			allocated_amount = COALESCE((SELECT SUM(fund_maintenance) FROM orders WHERE payment_status = 'PAID'), 0),
			spent_amount = COALESCE((SELECT SUM(amount) FROM finance_transactions WHERE type = 'EXPENSE' AND category = 'MACHINE_MAINTENANCE'), 0),
			current_balance = COALESCE((SELECT SUM(fund_maintenance) FROM orders WHERE payment_status = 'PAID'), 0) - COALESCE((SELECT SUM(amount) FROM finance_transactions WHERE type = 'EXPENSE' AND category = 'MACHINE_MAINTENANCE'), 0),
			updated_at = NOW()
		WHERE code = 'MAINTENANCE';

		UPDATE cash_accounts SET 
			allocated_amount = COALESCE((SELECT SUM(fund_depreciation) FROM orders WHERE payment_status = 'PAID'), 0),
			spent_amount = COALESCE((SELECT SUM(amount) FROM finance_transactions WHERE type = 'EXPENSE' AND category = 'MACHINE_PURCHASE'), 0),
			current_balance = COALESCE((SELECT SUM(fund_depreciation) FROM orders WHERE payment_status = 'PAID'), 0) - COALESCE((SELECT SUM(amount) FROM finance_transactions WHERE type = 'EXPENSE' AND category = 'MACHINE_PURCHASE'), 0),
			updated_at = NOW()
		WHERE code = 'DEPRECIATION';

		UPDATE cash_accounts SET 
			allocated_amount = COALESCE((SELECT SUM(fund_net_profit) FROM orders WHERE payment_status = 'PAID'), 0),
			spent_amount = COALESCE((SELECT SUM(amount) FROM finance_transactions WHERE type = 'EXPENSE' AND category NOT IN ('FILAMENT','HARDWARE','PACKAGING','ELECTRICITY','MACHINE_MAINTENANCE','MACHINE_PURCHASE')), 0),
			current_balance = COALESCE((SELECT SUM(fund_net_profit) FROM orders WHERE payment_status = 'PAID'), 0) - COALESCE((SELECT SUM(amount) FROM finance_transactions WHERE type = 'EXPENSE' AND category NOT IN ('FILAMENT','HARDWARE','PACKAGING','ELECTRICITY','MACHINE_MAINTENANCE','MACHINE_PURCHASE')), 0),
			updated_at = NOW()
		WHERE code = 'NET_PROFIT';
	`
	return db.WithContext(ctx).Exec(syncSQL).Error
}
