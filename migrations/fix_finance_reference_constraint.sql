-- ============================================================
-- Migration: Fix uq_finance_reference constraint for manual expenses
-- ============================================================

-- 1. Normalisasi data lama yang bernilai string kosong menjadi NULL
UPDATE public.finance_transactions
SET reference_type = NULL
WHERE reference_type = '';

UPDATE public.finance_transactions
SET reference_id = NULL
WHERE reference_id = '';

-- 2. Hapus unique index lama yang kaku
DROP INDEX IF EXISTS public.uq_finance_reference;

-- 3. Buat partial unique index baru:
-- Hanya menerapkan aturan keunikan JIKA reference_type dan reference_id benar-benar ada (tidak NULL dan tidak kosong).
-- Transaksi manual yang tidak punya referensi tidak akan pernah bertabrakan lagi.
CREATE UNIQUE INDEX uq_finance_reference 
ON public.finance_transactions (reference_type, reference_id) 
WHERE reference_type IS NOT NULL AND reference_type != '' 
  AND reference_id IS NOT NULL AND reference_id != '';
