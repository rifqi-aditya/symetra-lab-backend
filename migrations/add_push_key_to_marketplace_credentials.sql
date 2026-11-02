-- Migration: add_push_key_to_marketplace_credentials.sql
-- Menambahkan kolom push_key untuk penanganan Shopee Webhook Push Key (AES-256-GCM encrypted di layer Go)

ALTER TABLE marketplace_credentials ADD COLUMN IF NOT EXISTS push_key TEXT;
