-- 1. Drop redundant raw unix epoch ship_by_date from order_marketplace_details
DO $$ 
BEGIN
    IF EXISTS (
        SELECT 1 
        FROM information_schema.columns 
        WHERE table_name = 'order_marketplace_details' AND column_name = 'ship_by_date'
    ) THEN
        UPDATE order_marketplace_details
        SET ship_by_date_time = to_timestamp(ship_by_date)
        WHERE ship_by_date_time IS NULL AND ship_by_date > 0;

        ALTER TABLE order_marketplace_details DROP COLUMN ship_by_date;
    END IF;
END $$;

-- 2. Add buyer_cancel_reason to order_marketplace_details if not exists
DO $$ 
BEGIN
    IF NOT EXISTS (
        SELECT 1 
        FROM information_schema.columns 
        WHERE table_name = 'order_marketplace_details' AND column_name = 'buyer_cancel_reason'
    ) THEN
        ALTER TABLE order_marketplace_details ADD COLUMN buyer_cancel_reason text;
    END IF;
END $$;
