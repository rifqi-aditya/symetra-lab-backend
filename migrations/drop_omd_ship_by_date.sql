-- Drop redundant raw unix epoch ship_by_date from order_marketplace_details
-- The standardized column is ship_by_date_time (TIMESTAMPTZ)

DO $$ 
BEGIN
    -- Backfill ship_by_date_time from ship_by_date if ship_by_date_time is null but ship_by_date is present
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
