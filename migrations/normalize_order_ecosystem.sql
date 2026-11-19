-- ==============================================================================
-- MIGRATION: NORMALIZE ORDER ECOSYSTEM (Symetra Lab)
-- - Creates order_allocations table (relational 8-bucket cash accounts)
-- - Backfills historical allocation data from orders.fund_* (761+ orders safe)
-- - Prepares order_items (adds unified sku column, backfills from item_sku/matched_sku)
-- - Prepares order_marketplace_details (ensures ship_by_date_time is populated)
-- ==============================================================================

-- 1. Buat tabel order_allocations jika belum ada
CREATE TABLE IF NOT EXISTS public.order_allocations (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id        UUID NOT NULL REFERENCES public.orders(id) ON DELETE CASCADE,
    cash_account_id UUID NOT NULL REFERENCES public.cash_accounts(id) ON DELETE RESTRICT,
    account_code    VARCHAR(50) NOT NULL,
    amount          NUMERIC NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT uq_order_allocations_order_account UNIQUE (order_id, cash_account_id)
);

CREATE INDEX IF NOT EXISTS idx_order_allocations_order_id ON public.order_allocations(order_id);
CREATE INDEX IF NOT EXISTS idx_order_allocations_account_code ON public.order_allocations(account_code);

-- 2. Backfill alokasi data historis dari orders ke order_allocations
--    (Menjaga 100% integritas data pesanan lama)

-- 2.1 FILAMENT
INSERT INTO public.order_allocations (order_id, cash_account_id, account_code, amount, created_at)
SELECT o.id, ca.id, 'FILAMENT', COALESCE(o.fund_filament, 0), o.created_at
FROM public.orders o
CROSS JOIN public.cash_accounts ca
WHERE ca.code = 'FILAMENT'
ON CONFLICT (order_id, cash_account_id) DO UPDATE SET amount = EXCLUDED.amount;

-- 2.2 EMERGENCY (Kas Cadangan Risiko)
INSERT INTO public.order_allocations (order_id, cash_account_id, account_code, amount, created_at)
SELECT o.id, ca.id, 'EMERGENCY', COALESCE(o.fund_emergency, 0), o.created_at
FROM public.orders o
CROSS JOIN public.cash_accounts ca
WHERE ca.code = 'EMERGENCY'
ON CONFLICT (order_id, cash_account_id) DO UPDATE SET amount = EXCLUDED.amount;

-- 2.3 COMPONENT (Kas Komponen)
INSERT INTO public.order_allocations (order_id, cash_account_id, account_code, amount, created_at)
SELECT o.id, ca.id, 'COMPONENT', COALESCE(o.fund_component, 0), o.created_at
FROM public.orders o
CROSS JOIN public.cash_accounts ca
WHERE ca.code = 'COMPONENT'
ON CONFLICT (order_id, cash_account_id) DO UPDATE SET amount = EXCLUDED.amount;

-- 2.4 PACKAGING (Kas Kemasan)
INSERT INTO public.order_allocations (order_id, cash_account_id, account_code, amount, created_at)
SELECT o.id, ca.id, 'PACKAGING', COALESCE(o.fund_packaging, 0), o.created_at
FROM public.orders o
CROSS JOIN public.cash_accounts ca
WHERE ca.code = 'PACKAGING'
ON CONFLICT (order_id, cash_account_id) DO UPDATE SET amount = EXCLUDED.amount;

-- 2.5 ELECTRICITY (Kas Listrik)
INSERT INTO public.order_allocations (order_id, cash_account_id, account_code, amount, created_at)
SELECT o.id, ca.id, 'ELECTRICITY', COALESCE(o.fund_electricity, 0), o.created_at
FROM public.orders o
CROSS JOIN public.cash_accounts ca
WHERE ca.code = 'ELECTRICITY'
ON CONFLICT (order_id, cash_account_id) DO UPDATE SET amount = EXCLUDED.amount;

-- 2.6 MAINTENANCE (Kas Maintenance)
INSERT INTO public.order_allocations (order_id, cash_account_id, account_code, amount, created_at)
SELECT o.id, ca.id, 'MAINTENANCE', COALESCE(o.fund_maintenance, 0), o.created_at
FROM public.orders o
CROSS JOIN public.cash_accounts ca
WHERE ca.code = 'MAINTENANCE'
ON CONFLICT (order_id, cash_account_id) DO UPDATE SET amount = EXCLUDED.amount;

-- 2.7 DEPRECIATION (Kas Mesin)
INSERT INTO public.order_allocations (order_id, cash_account_id, account_code, amount, created_at)
SELECT o.id, ca.id, 'DEPRECIATION', COALESCE(o.fund_depreciation, 0), o.created_at
FROM public.orders o
CROSS JOIN public.cash_accounts ca
WHERE ca.code = 'DEPRECIATION'
ON CONFLICT (order_id, cash_account_id) DO UPDATE SET amount = EXCLUDED.amount;

-- 2.8 NET_PROFIT (Kas Laba Bersih)
INSERT INTO public.order_allocations (order_id, cash_account_id, account_code, amount, created_at)
SELECT o.id, ca.id, 'NET_PROFIT', COALESCE(o.fund_net_profit, 0), o.created_at
FROM public.orders o
CROSS JOIN public.cash_accounts ca
WHERE ca.code = 'NET_PROFIT'
ON CONFLICT (order_id, cash_account_id) DO UPDATE SET amount = EXCLUDED.amount;

-- 3. Rampingkan order_items: Tambah kolom `sku` dan migrasi dari matched_sku / item_sku
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_schema = 'public' AND table_name = 'order_items' AND column_name = 'sku'
    ) THEN
        ALTER TABLE public.order_items ADD COLUMN sku VARCHAR(100);
    END IF;
END $$;

UPDATE public.order_items
SET sku = COALESCE(matched_sku, item_sku)
WHERE (sku IS NULL OR sku = '') AND (matched_sku IS NOT NULL OR item_sku IS NOT NULL);

-- 4. Rampingkan order_marketplace_details: Pastikan ship_by_date_time terisi
UPDATE public.order_marketplace_details
SET ship_by_date_time = to_timestamp(ship_by_date)
WHERE ship_by_date_time IS NULL AND ship_by_date IS NOT NULL AND ship_by_date > 0;
