-- ==============================================================================
-- MIGRATION: NORMALIZE ORDER ITEM UNIT COSTS (Symetra Lab)
-- - Converts legacy order_items where HPP and cost breakdowns were stored as
--   line totals instead of unit costs (affecting 63 orders / 82 items where quantity > 1).
-- - Ensures: oi.hpp * oi.quantity = o.cogs_amount (100% consistent across all orders).
-- - Prevents double multiplication in frontend OrderDetailItems and Micro-Ledger.
-- ==============================================================================

UPDATE public.order_items oi
SET 
    hpp               = ROUND(oi.hpp / oi.quantity, 2),
    filament_cost     = ROUND(oi.filament_cost / oi.quantity, 2),
    component_cost    = ROUND(oi.component_cost / oi.quantity, 2),
    packaging_cost    = ROUND(oi.packaging_cost / oi.quantity, 2),
    energy_cost       = ROUND(oi.energy_cost / oi.quantity, 2),
    maintenance_cost  = ROUND(oi.maintenance_cost / oi.quantity, 2),
    depreciation_cost = ROUND(oi.depreciation_cost / oi.quantity, 2),
    updated_at        = now()
FROM public.orders o
WHERE oi.order_id = o.id
  AND oi.quantity > 1
  AND oi.hpp > 0
  AND ABS((oi.hpp * oi.quantity) - o.cogs_amount) > 5.0
  AND (
      ABS(oi.hpp - o.cogs_amount) <= 5.0 
      OR EXISTS (
          SELECT 1 FROM public.order_items oi2 
          WHERE oi2.order_id = o.id 
          GROUP BY oi2.order_id 
          HAVING ABS(SUM(oi2.hpp) - o.cogs_amount) <= 5.0
      )
  );
