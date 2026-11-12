-- ==============================================================================
-- ADD PARENT_SKU TO PRODUCTS TABLE FOR MULTI-VARIANT GROUPING
-- ==============================================================================

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_schema = 'public' AND table_name = 'products' AND column_name = 'parent_sku'
    ) THEN
        ALTER TABLE public.products ADD COLUMN parent_sku VARCHAR(255);
        CREATE INDEX IF NOT EXISTS idx_products_parent_sku ON public.products(parent_sku);
    END IF;
END $$;

