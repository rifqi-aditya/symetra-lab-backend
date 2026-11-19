-- ==============================================================================
-- MIGRATION: SIMPLIFY COMPONENTS & PACKAGING ITEMS
-- - Drops markup_percent from product_components (centralized in components.default_markup_percent)
-- - Drops category, purchase_price, purchase_quantity from packaging_items
--   (centralized in packaging_items.unit_cost as single source of truth)
-- ==============================================================================

-- 1. Hapus markup_percent dari product_components jika masih ada
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_schema = 'public' AND table_name = 'product_components' AND column_name = 'markup_percent'
    ) THEN
        ALTER TABLE public.product_components DROP COLUMN markup_percent;
    END IF;
END $$;

-- 2. Pastikan unit_cost di packaging_items terisi sebelum drop purchase_price/quantity
UPDATE public.packaging_items
SET unit_cost = CASE 
    WHEN unit_cost > 0 THEN unit_cost
    WHEN purchase_quantity > 0 AND purchase_price > 0 THEN ROUND((purchase_price / purchase_quantity)::numeric, 2)
    ELSE 0
END
WHERE unit_cost IS NULL OR unit_cost = 0;

-- 3. Hapus kolom-kolom redundan dari packaging_items
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_schema = 'public' AND table_name = 'packaging_items' AND column_name = 'category'
    ) THEN
        ALTER TABLE public.packaging_items DROP COLUMN category;
    END IF;

    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_schema = 'public' AND table_name = 'packaging_items' AND column_name = 'purchase_price'
    ) THEN
        ALTER TABLE public.packaging_items DROP COLUMN purchase_price;
    END IF;

    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_schema = 'public' AND table_name = 'packaging_items' AND column_name = 'purchase_quantity'
    ) THEN
        ALTER TABLE public.packaging_items DROP COLUMN purchase_quantity;
    END IF;
END $$;
