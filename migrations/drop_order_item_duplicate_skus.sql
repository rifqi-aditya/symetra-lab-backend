-- ==============================================================================
-- MIGRATION: DEDICATED SKU UNIFICATION & CLEANUP IN ORDER_ITEMS
-- - Ensures all rows have sku populated from matched_sku or item_sku
-- - Drops redundant legacy columns: matched_sku, item_sku
-- ==============================================================================

-- 1. Pastikan seluruh baris order_items memiliki kolom sku terisi
UPDATE public.order_items
SET sku = COALESCE(NULLIF(TRIM(sku), ''), NULLIF(TRIM(matched_sku), ''), NULLIF(TRIM(item_sku), ''))
WHERE (sku IS NULL OR TRIM(sku) = '') 
  AND (matched_sku IS NOT NULL OR item_sku IS NOT NULL);

-- 2. Hapus kolom duplikat matched_sku dan item_sku
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_schema = 'public' 
          AND table_name = 'order_items' 
          AND column_name = 'matched_sku'
    ) THEN
        ALTER TABLE public.order_items DROP COLUMN matched_sku;
    END IF;

    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_schema = 'public' 
          AND table_name = 'order_items' 
          AND column_name = 'item_sku'
    ) THEN
        ALTER TABLE public.order_items DROP COLUMN item_sku;
    END IF;
END $$;
