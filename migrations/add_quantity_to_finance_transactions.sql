-- Migration: Add quantity column to finance_transactions
ALTER TABLE public.finance_transactions
  ADD COLUMN IF NOT EXISTS quantity numeric DEFAULT 1;
