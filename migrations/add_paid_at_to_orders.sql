-- Migration: Add paid_at column to orders table
-- Description: Supports Shopee and marketplace payment timestamp tracking

ALTER TABLE public.orders 
ADD COLUMN IF NOT EXISTS paid_at TIMESTAMP WITH TIME ZONE;
