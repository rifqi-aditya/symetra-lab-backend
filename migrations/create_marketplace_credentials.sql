-- Migration: create_marketplace_credentials.sql
-- Menampung konfigurasi kredensial tingkat partner (Shopee, Tokopedia, TikTok Shop)
-- Kolom partner_key dienkripsi menggunakan AES-256-GCM pada level backend.

CREATE TABLE IF NOT EXISTS marketplace_credentials (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    platform VARCHAR(50) NOT NULL UNIQUE, -- 'SHOPEE', 'TOKOPEDIA', 'TIKTOK'
    partner_id VARCHAR(100) NOT NULL,
    partner_key TEXT NOT NULL,           -- AES-256-GCM encrypted ciphertext hex
    redirect_url TEXT,
    is_production BOOLEAN DEFAULT false,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

-- Indeks untuk pencarian cepat berdasarkan platform
CREATE INDEX IF NOT EXISTS idx_marketplace_credentials_platform ON marketplace_credentials(platform);

