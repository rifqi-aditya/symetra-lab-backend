-- Migration: Add attachment_url column to finance_transactions
ALTER TABLE public.finance_transactions
  ADD COLUMN IF NOT EXISTS attachment_url text;
