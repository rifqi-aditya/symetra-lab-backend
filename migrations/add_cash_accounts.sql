-- ============================================================
-- Migration: Add cash_accounts table + link to finance_transactions
-- Run this in Supabase SQL Editor
-- ============================================================

-- 1. Buat tabel cash_accounts (7 pos kas)
CREATE TABLE IF NOT EXISTS public.cash_accounts (
  id          uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
  name        varchar(100) NOT NULL,
  description text,
  color       varchar(20)  NOT NULL DEFAULT '#6366f1',
  is_active   boolean      NOT NULL DEFAULT true,
  created_at  timestamptz  NOT NULL DEFAULT now(),
  updated_at  timestamptz  NOT NULL DEFAULT now()
);

-- 2. RLS
ALTER TABLE public.cash_accounts ENABLE ROW LEVEL SECURITY;

CREATE POLICY "Allow all for authenticated"
  ON public.cash_accounts
  FOR ALL TO authenticated
  USING (true)
  WITH CHECK (true);

-- 3. Seed 7 pos kas
INSERT INTO public.cash_accounts (name, description, color) VALUES
  ('Kas Filamen',    'Anggaran pembelian filamen dan bahan baku cetak',        '#f97316'),
  ('Kas Komponen',   'Anggaran komponen hardware (baut, bearing, magnet)',      '#8b5cf6'),
  ('Kas Kemasan',    'Anggaran bahan kemasan (kardus, bubble wrap, lakban)',    '#06b6d4'),
  ('Kas Listrik',    'Anggaran listrik workshop / token PLN & internet',        '#eab308'),
  ('Kas Maintenance','Anggaran perawatan dan sparepart mesin 3D printer',       '#ef4444'),
  ('Kas Mesin',      'Tabungan pembelian mesin 3D printer baru',               '#10b981'),
  ('Kas Laba',       'Tabungan laba / profit bersih usaha',                    '#3b82f6')
ON CONFLICT DO NOTHING;

-- 4. Tambah kolom cash_account_id ke finance_transactions (nullable → backward compat)
ALTER TABLE public.finance_transactions
  ADD COLUMN IF NOT EXISTS cash_account_id uuid
  REFERENCES public.cash_accounts(id) ON DELETE SET NULL;

-- 5. Optional index untuk query cepat
CREATE INDEX IF NOT EXISTS idx_finance_tx_cash_account
  ON public.finance_transactions(cash_account_id);
