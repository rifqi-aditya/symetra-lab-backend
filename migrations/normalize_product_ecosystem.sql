-- ==============================================================================
-- MIGRATION: NORMALIZE PRODUCT ECOSYSTEM (Symetra Lab)
-- - Creates parent_products table (master etalase & 3D model listing)
-- - Synchronizes category names into product_categories
-- - Backfills 103 existing products into parent_products
-- - Adds parent_id and variant_name to products table
-- - Links products to parent_products
-- - Backfills any missing packaging preset items into product_packaging_items
-- - Drops deprecated packaging_preset tables and foreign keys
-- - Safely drops redundant columns on products (category, parent_sku, times_ordered,
--   packing_fee_idr, packaging_preset_id, description, thumbnail_url, design_link)
-- ==============================================================================

-- 1. Pastikan tabel product_categories memiliki semua kategori unik dari products
INSERT INTO public.product_categories (id, user_id, name, created_at)
SELECT 
    gen_random_uuid(),
    p.user_id,
    TRIM(p.category),
    now()
FROM public.products p
WHERE p.category IS NOT NULL 
  AND TRIM(p.category) != ''
  AND NOT EXISTS (
      SELECT 1 FROM public.product_categories pc 
      WHERE LOWER(TRIM(pc.name)) = LOWER(TRIM(p.category))
  )
GROUP BY p.user_id, TRIM(p.category);

-- 2. Buat tabel parent_products jika belum ada
CREATE TABLE IF NOT EXISTS public.parent_products (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       UUID NOT NULL,
    parent_sku    VARCHAR(100) NOT NULL UNIQUE,
    name          VARCHAR(255) NOT NULL,
    category_id   UUID REFERENCES public.product_categories(id) ON DELETE SET NULL,
    description   TEXT,
    thumbnail_url TEXT,
    design_link   TEXT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_parent_products_user_id ON public.parent_products(user_id);
CREATE INDEX IF NOT EXISTS idx_parent_products_parent_sku ON public.parent_products(parent_sku);
CREATE INDEX IF NOT EXISTS idx_parent_products_category_id ON public.parent_products(category_id);

-- 3. Tambahkan kolom parent_id dan variant_name ke tabel products jika belum ada
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_schema = 'public' AND table_name = 'products' AND column_name = 'parent_id'
    ) THEN
        ALTER TABLE public.products ADD COLUMN parent_id UUID;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_schema = 'public' AND table_name = 'products' AND column_name = 'variant_name'
    ) THEN
        ALTER TABLE public.products ADD COLUMN variant_name VARCHAR(100) DEFAULT 'Standard';
    END IF;
END $$;

-- 4. Backfill parent_products dan hubungkan products.parent_id
DO $$
BEGIN
    -- Hanya lakukan backfill jika ada produk yang belum memiliki parent_id
    IF EXISTS (SELECT 1 FROM public.products WHERE parent_id IS NULL) THEN
        WITH inserted AS (
            INSERT INTO public.parent_products (
                id,
                user_id,
                parent_sku,
                name,
                category_id,
                description,
                thumbnail_url,
                design_link,
                created_at,
                updated_at
            )
            SELECT 
                gen_random_uuid(),
                p.user_id,
                COALESCE(
                    NULLIF(TRIM(p.parent_sku), ''), 
                    NULLIF(TRIM(p.sku), ''), 
                    'PAR-' || UPPER(SUBSTRING(p.id::text, 1, 8))
                ) AS parent_sku,
                p.name,
                pc.id AS category_id,
                p.description,
                p.thumbnail_url,
                p.design_link,
                COALESCE(p.created_at, now()),
                COALESCE(p.updated_at, now())
            FROM public.products p
            LEFT JOIN public.product_categories pc 
                ON LOWER(TRIM(pc.name)) = LOWER(TRIM(p.category))
            WHERE p.parent_id IS NULL
            RETURNING id, parent_sku
        )
        UPDATE public.products p
        SET parent_id = ins.id,
            variant_name = COALESCE(NULLIF(p.variant_name, ''), 'Standard')
        FROM inserted ins
        WHERE p.parent_id IS NULL
          AND ins.parent_sku = COALESCE(
                NULLIF(TRIM(p.parent_sku), ''), 
                NULLIF(TRIM(p.sku), ''), 
                'PAR-' || UPPER(SUBSTRING(p.id::text, 1, 8))
          );
    END IF;
END $$;

-- 5. Set NOT NULL dan Foreign Key constraint pada products.parent_id
DO $$
BEGIN
    ALTER TABLE public.products ALTER COLUMN parent_id SET NOT NULL;

    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'fk_products_parent_id' AND table_name = 'products'
    ) THEN
        ALTER TABLE public.products 
        ADD CONSTRAINT fk_products_parent_id 
        FOREIGN KEY (parent_id) REFERENCES public.parent_products(id) ON DELETE CASCADE;
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_products_parent_id ON public.products(parent_id);

-- 6. Migrasi data packaging presets ke product_packaging_items
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'packaging_preset_items') THEN
        INSERT INTO public.product_packaging_items (id, product_id, packaging_item_id, quantity_used, created_at)
        SELECT 
            gen_random_uuid(),
            p.id,
            ppi.packaging_item_id,
            ppi.quantity_used,
            now()
        FROM public.products p
        JOIN public.packaging_preset_items ppi ON ppi.preset_id = p.packaging_preset_id
        WHERE NOT EXISTS (
            SELECT 1 FROM public.product_packaging_items curr
            WHERE curr.product_id = p.id AND curr.packaging_item_id = ppi.packaging_item_id
        );
    END IF;
END $$;

-- 7. Hapus FK dari products ke packaging_presets sebelum drop tabel preset
DO $$
DECLARE
    r RECORD;
BEGIN
    FOR r IN (
        SELECT constraint_name 
        FROM information_schema.table_constraints 
        WHERE table_name = 'products' AND constraint_name LIKE '%packaging_preset%'
    ) LOOP
        EXECUTE 'ALTER TABLE public.products DROP CONSTRAINT ' || quote_ident(r.constraint_name);
    END LOOP;
END $$;

-- 8. Drop tabel packaging_preset_items dan packaging_presets
DROP TABLE IF EXISTS public.packaging_preset_items CASCADE;
DROP TABLE IF EXISTS public.packaging_presets CASCADE;

-- 9. Bersihkan kolom-kolom redundan pada tabel products
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'products' AND column_name = 'category') THEN
        ALTER TABLE public.products DROP COLUMN category;
    END IF;

    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'products' AND column_name = 'parent_sku') THEN
        ALTER TABLE public.products DROP COLUMN parent_sku;
    END IF;

    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'products' AND column_name = 'times_ordered') THEN
        ALTER TABLE public.products DROP COLUMN times_ordered;
    END IF;

    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'products' AND column_name = 'packing_fee_idr') THEN
        ALTER TABLE public.products DROP COLUMN packing_fee_idr;
    END IF;

    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'products' AND column_name = 'packaging_preset_id') THEN
        ALTER TABLE public.products DROP COLUMN packaging_preset_id;
    END IF;

    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'products' AND column_name = 'description') THEN
        ALTER TABLE public.products DROP COLUMN description;
    END IF;

    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'products' AND column_name = 'thumbnail_url') THEN
        ALTER TABLE public.products DROP COLUMN thumbnail_url;
    END IF;

    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'products' AND column_name = 'design_link') THEN
        ALTER TABLE public.products DROP COLUMN design_link;
    END IF;
END $$;
