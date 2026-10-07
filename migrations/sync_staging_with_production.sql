-- ==============================================================================
-- SYNC STAGING DATABASE TO MATCH PRODUCTION SCHEMA
-- Generated from direct inspection of Production Database
-- ==============================================================================

-- 1. Perbaikan tabel `products` (Hapus parent_sku, pastikan sku & packing_fee_idr)
DO $$
BEGIN
    -- Hapus parent_sku jika masih ada di staging
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_schema = 'public' AND table_name = 'products' AND column_name = 'parent_sku'
    ) THEN
        ALTER TABLE public.products DROP COLUMN parent_sku;
    END IF;

    -- Hapus kolom lama `packing_fee_id_r` yang typo
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_schema = 'public' AND table_name = 'products' AND column_name = 'packing_fee_id_r'
    ) THEN
        IF EXISTS (
            SELECT 1 FROM information_schema.columns 
            WHERE table_schema = 'public' AND table_name = 'products' AND column_name = 'packing_fee_idr'
        ) THEN
            -- Salin nilai jika packing_fee_idr kosong/0
            UPDATE public.products 
            SET packing_fee_idr = COALESCE(packing_fee_idr, packing_fee_id_r, 0)
            WHERE packing_fee_idr IS NULL OR packing_fee_idr = 0;

            -- Hapus kolom lama
            ALTER TABLE public.products DROP COLUMN packing_fee_id_r;
        ELSE
            -- Jika packing_fee_idr belum ada sama sekali, rename
            ALTER TABLE public.products RENAME COLUMN packing_fee_id_r TO packing_fee_idr;
        END IF;
    ELSIF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_schema = 'public' AND table_name = 'products' AND column_name = 'packing_fee_idr'
    ) THEN
        ALTER TABLE public.products ADD COLUMN packing_fee_idr BIGINT DEFAULT 0;
    END IF;

    -- Pastikan kolom sku ada di products
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_schema = 'public' AND table_name = 'products' AND column_name = 'sku'
    ) THEN
        ALTER TABLE public.products ADD COLUMN sku VARCHAR;
    END IF;

    -- Pastikan material_type ada di products
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_schema = 'public' AND table_name = 'products' AND column_name = 'material_type'
    ) THEN
        ALTER TABLE public.products ADD COLUMN material_type VARCHAR DEFAULT 'PLA';
    END IF;
END $$;

-- 2. Pastikan tabel `shops` untuk Shopee integration sama persis
CREATE TABLE IF NOT EXISTS public.shops (
    id BIGSERIAL PRIMARY KEY,
    shop_id BIGINT NOT NULL,
    shop_name VARCHAR,
    region VARCHAR,
    access_token TEXT NOT NULL,
    refresh_token TEXT NOT NULL,
    access_token_expires_at TIMESTAMPTZ,
    refresh_token_expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- 3. Pastikan kolom-kolom baru di order_items (Shopee SKU linking & COGS breakdown)
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'order_items' AND column_name = 'item_sku') THEN
        ALTER TABLE public.order_items ADD COLUMN item_sku VARCHAR;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'order_items' AND column_name = 'channel_item_id') THEN
        ALTER TABLE public.order_items ADD COLUMN channel_item_id BIGINT;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'order_items' AND column_name = 'channel_model_id') THEN
        ALTER TABLE public.order_items ADD COLUMN channel_model_id BIGINT;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'order_items' AND column_name = 'mapping_status') THEN
        ALTER TABLE public.order_items ADD COLUMN mapping_status VARCHAR DEFAULT 'MATCHED';
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'order_items' AND column_name = 'matched_sku') THEN
        ALTER TABLE public.order_items ADD COLUMN matched_sku VARCHAR;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'order_items' AND column_name = 'updated_at') THEN
        ALTER TABLE public.order_items ADD COLUMN updated_at TIMESTAMPTZ DEFAULT NOW();
    END IF;
END $$;
