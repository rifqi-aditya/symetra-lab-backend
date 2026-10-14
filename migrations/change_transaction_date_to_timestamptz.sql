-- Migration: Change transaction_date column in finance_transactions from DATE to TIMESTAMPTZ
ALTER TABLE public.finance_transactions 
  ALTER COLUMN transaction_date TYPE TIMESTAMP WITH TIME ZONE 
  USING transaction_date::timestamptz;
