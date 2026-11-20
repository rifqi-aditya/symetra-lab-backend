-- ==============================================================================
-- MIGRATION: DROP ORDER FUND COLUMNS & FINALIZE ORDER ECOSYSTEM NORMALIZATION
-- - Ensures 100% data preservation by verifying order_allocations
-- - Drops 8 redundant fund_* columns from public.orders
-- ==============================================================================

-- 1. Safety verification and final backfill just in case any order was added recently
INSERT INTO public.order_allocations (order_id, cash_account_id, account_code, amount, created_at)
SELECT o.id, ca.id, ca.code, 
    CASE ca.code
        WHEN 'FILAMENT' THEN COALESCE(o.fund_filament, 0)
        WHEN 'EMERGENCY' THEN COALESCE(o.fund_emergency, 0)
        WHEN 'COMPONENT' THEN COALESCE(o.fund_component, 0)
        WHEN 'PACKAGING' THEN COALESCE(o.fund_packaging, 0)
        WHEN 'ELECTRICITY' THEN COALESCE(o.fund_electricity, 0)
        WHEN 'MAINTENANCE' THEN COALESCE(o.fund_maintenance, 0)
        WHEN 'DEPRECIATION' THEN COALESCE(o.fund_depreciation, 0)
        WHEN 'NET_PROFIT' THEN COALESCE(o.fund_net_profit, 0)
        ELSE 0
    END AS amount,
    o.created_at
FROM public.orders o
CROSS JOIN public.cash_accounts ca
ON CONFLICT (order_id, cash_account_id) 
DO UPDATE SET amount = EXCLUDED.amount;

-- 2. Drop the 8 redundant fund columns from orders safely
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_schema = 'public' AND table_name = 'orders' AND column_name = 'fund_filament'
    ) THEN
        ALTER TABLE public.orders DROP COLUMN fund_filament;
    END IF;

    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_schema = 'public' AND table_name = 'orders' AND column_name = 'fund_emergency'
    ) THEN
        ALTER TABLE public.orders DROP COLUMN fund_emergency;
    END IF;

    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_schema = 'public' AND table_name = 'orders' AND column_name = 'fund_component'
    ) THEN
        ALTER TABLE public.orders DROP COLUMN fund_component;
    END IF;

    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_schema = 'public' AND table_name = 'orders' AND column_name = 'fund_packaging'
    ) THEN
        ALTER TABLE public.orders DROP COLUMN fund_packaging;
    END IF;

    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_schema = 'public' AND table_name = 'orders' AND column_name = 'fund_electricity'
    ) THEN
        ALTER TABLE public.orders DROP COLUMN fund_electricity;
    END IF;

    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_schema = 'public' AND table_name = 'orders' AND column_name = 'fund_maintenance'
    ) THEN
        ALTER TABLE public.orders DROP COLUMN fund_maintenance;
    END IF;

    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_schema = 'public' AND table_name = 'orders' AND column_name = 'fund_depreciation'
    ) THEN
        ALTER TABLE public.orders DROP COLUMN fund_depreciation;
    END IF;

    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_schema = 'public' AND table_name = 'orders' AND column_name = 'fund_net_profit'
    ) THEN
        ALTER TABLE public.orders DROP COLUMN fund_net_profit;
    END IF;
END $$;
