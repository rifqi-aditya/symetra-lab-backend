-- ============================================================
-- Migration: Add Emergency Fund (Kas Cadangan Risiko)
-- Run this in Supabase SQL Editor (staging first, then production)
-- Branch: feature/emergency-fund-allocation
-- ============================================================

-- 1. Tambah kolom code (unique identifier) ke cash_accounts jika belum ada
ALTER TABLE public.cash_accounts
  ADD COLUMN IF NOT EXISTS code varchar(50);

-- 2. Isi kolom code untuk 7 pos kas yang sudah ada (berdasarkan urutan insert awal)
UPDATE public.cash_accounts SET code = 'FILAMENT'    WHERE name = 'Kas Filamen'      AND (code IS NULL OR code = '');
UPDATE public.cash_accounts SET code = 'COMPONENT'   WHERE name = 'Kas Komponen'     AND (code IS NULL OR code = '');
UPDATE public.cash_accounts SET code = 'PACKAGING'   WHERE name = 'Kas Kemasan'      AND (code IS NULL OR code = '');
UPDATE public.cash_accounts SET code = 'ELECTRICITY' WHERE name = 'Kas Listrik'      AND (code IS NULL OR code = '');
UPDATE public.cash_accounts SET code = 'MAINTENANCE' WHERE name = 'Kas Maintenance'  AND (code IS NULL OR code = '');
UPDATE public.cash_accounts SET code = 'DEPRECIATION'WHERE name = 'Kas Mesin'        AND (code IS NULL OR code = '');
UPDATE public.cash_accounts SET code = 'NET_PROFIT'  WHERE name = 'Kas Laba'         AND (code IS NULL OR code = '');

-- Pastikan ada UNIQUE constraint pada code
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'cash_accounts_code_key'
    ) THEN
        ALTER TABLE public.cash_accounts ADD CONSTRAINT cash_accounts_code_key UNIQUE (code);
    END IF;
END $$;

-- Tambah kolom allocated_amount, spent_amount, current_balance jika belum ada
ALTER TABLE public.cash_accounts
  ADD COLUMN IF NOT EXISTS allocated_amount float8 NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS spent_amount     float8 NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS current_balance  float8 NOT NULL DEFAULT 0;

-- 3. Tambah pos kas baru: Kas Cadangan Risiko
INSERT INTO public.cash_accounts (name, code, description, color, allocated_amount, spent_amount, current_balance)
VALUES (
  'Kas Cadangan Risiko',
  'EMERGENCY',
  'Dana cadangan untuk menanggung kerugian akibat kegagalan cetak dan risiko operasional 3D printing',
  '#dc2626',
  0,
  0,
  0
)
ON CONFLICT (code) DO NOTHING;

-- 4. Tambah kolom fund_emergency di tabel orders
ALTER TABLE public.orders
  ADD COLUMN IF NOT EXISTS fund_emergency float8 NOT NULL DEFAULT 0;

-- 5. Retroactive backfill:
--    a. Pisahkan nilai buffer 10% dari fund_filament historis
--       Formula: buffer_share = fund_filament_lama * (0.10 / 1.10) = fund_filament * 0.090909...
UPDATE public.orders
SET
  fund_emergency = ROUND((fund_filament * 0.090909)::numeric, 2),
  fund_filament  = ROUND((fund_filament * 0.909091)::numeric, 2),
  updated_at     = NOW()
WHERE
  payment_status = 'PAID'
  AND fund_emergency = 0
  AND fund_filament > 0;

--    b. Alokasikan 5% dari laba bersih pesanan ke fund_emergency secara otomatis (95% ke fund_net_profit)
UPDATE public.orders
SET 
  cogs_amount = ROUND((fund_filament + ROUND(fund_filament * 0.10, 2) + fund_component + fund_packaging + fund_electricity + fund_maintenance + fund_depreciation)::numeric, 2),
  net_profit  = ROUND((net_amount - (fund_filament + ROUND(fund_filament * 0.10, 2) + fund_component + fund_packaging + fund_electricity + fund_maintenance + fund_depreciation))::numeric, 2),
  fund_emergency = ROUND(fund_filament * 0.10, 2) + 
    CASE 
      WHEN (net_amount - (fund_filament + ROUND(fund_filament * 0.10, 2) + fund_component + fund_packaging + fund_electricity + fund_maintenance + fund_depreciation)) > 0 
      THEN ROUND(((net_amount - (fund_filament + ROUND(fund_filament * 0.10, 2) + fund_component + fund_packaging + fund_electricity + fund_maintenance + fund_depreciation)) * 0.05)::numeric, 2)
      ELSE 0 
    END,
  fund_net_profit = (net_amount - (fund_filament + ROUND(fund_filament * 0.10, 2) + fund_component + fund_packaging + fund_electricity + fund_maintenance + fund_depreciation)) -
    CASE 
      WHEN (net_amount - (fund_filament + ROUND(fund_filament * 0.10, 2) + fund_component + fund_packaging + fund_electricity + fund_maintenance + fund_depreciation)) > 0 
      THEN ROUND(((net_amount - (fund_filament + ROUND(fund_filament * 0.10, 2) + fund_component + fund_packaging + fund_electricity + fund_maintenance + fund_depreciation)) * 0.05)::numeric, 2)
      ELSE 0 
    END,
  updated_at = NOW()
WHERE status NOT IN ('CANCELLED', 'UNPAID');

-- 7. Sync ledger cash_accounts setelah backfill
UPDATE public.cash_accounts SET
  allocated_amount = COALESCE((SELECT SUM(fund_filament)    FROM orders WHERE payment_status = 'PAID'), 0),
  spent_amount     = COALESCE((SELECT SUM(amount) FROM finance_transactions WHERE type = 'EXPENSE' AND category = 'FILAMENT'), 0),
  current_balance  = COALESCE((SELECT SUM(fund_filament)    FROM orders WHERE payment_status = 'PAID'), 0)
                   - COALESCE((SELECT SUM(amount) FROM finance_transactions WHERE type = 'EXPENSE' AND category = 'FILAMENT'), 0),
  updated_at       = NOW()
WHERE code = 'FILAMENT';

UPDATE public.cash_accounts SET
  allocated_amount = COALESCE((SELECT SUM(fund_emergency) FROM orders WHERE payment_status = 'PAID'), 0),
  spent_amount     = COALESCE((SELECT SUM(amount) FROM finance_transactions WHERE type = 'EXPENSE' AND category = 'EMERGENCY_FUND'), 0),
  current_balance  = COALESCE((SELECT SUM(fund_emergency) FROM orders WHERE payment_status = 'PAID'), 0)
                   - COALESCE((SELECT SUM(amount) FROM finance_transactions WHERE type = 'EXPENSE' AND category = 'EMERGENCY_FUND'), 0),
  updated_at       = NOW()
WHERE code = 'EMERGENCY';

UPDATE public.cash_accounts SET
  allocated_amount = COALESCE((SELECT SUM(fund_net_profit) FROM orders WHERE payment_status = 'PAID'), 0),
  spent_amount     = COALESCE((SELECT SUM(amount) FROM finance_transactions WHERE type = 'EXPENSE'
                               AND category NOT IN ('FILAMENT','HARDWARE','PACKAGING','ELECTRICITY','MACHINE_MAINTENANCE','MACHINE_PURCHASE','EMERGENCY_FUND','REIMBURSEMENT')), 0),
  current_balance  = COALESCE((SELECT SUM(fund_net_profit) FROM orders WHERE payment_status = 'PAID'), 0)
                   - COALESCE((SELECT SUM(amount) FROM finance_transactions WHERE type = 'EXPENSE'
                               AND category NOT IN ('FILAMENT','HARDWARE','PACKAGING','ELECTRICITY','MACHINE_MAINTENANCE','MACHINE_PURCHASE','EMERGENCY_FUND','REIMBURSEMENT')), 0),
  updated_at       = NOW()
WHERE code = 'NET_PROFIT';
