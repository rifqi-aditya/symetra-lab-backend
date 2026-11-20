-- ==============================================================================
-- MIGRATION: DROP REDUNDANT COLUMNS FROM ORDER_MARKETPLACE_DETAILS
-- - Drops channel (redundant with orders.channel)
-- - Drops total_marketplace_fee (redundant with orders.channel_fee)
-- ==============================================================================

DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_schema = 'public' 
          AND table_name = 'order_marketplace_details' 
          AND column_name = 'channel'
    ) THEN
        ALTER TABLE public.order_marketplace_details DROP COLUMN channel;
    END IF;

    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_schema = 'public' 
          AND table_name = 'order_marketplace_details' 
          AND column_name = 'total_marketplace_fee'
    ) THEN
        ALTER TABLE public.order_marketplace_details DROP COLUMN total_marketplace_fee;
    END IF;
END $$;
