# Symetra Lab Backend API Documentation

Welcome to the **Symetra Lab Backend** API documentation. This backend powers the 3D printing workshop management platform, handling everything from technical inventory (filaments, 3D printers, maintenance spare parts, hardware components, packaging), BOM cost calculations using the **7 Pos Kas Workshop Formula**, Direct/Manual orders, Shopee Open Platform v2 integration, and an automated Priority Print Queue with inventory auto-deductions.

---

## 1. Overview & Architecture

- **Framework:** Go (v1.23+) with **Echo v4**
- **Architecture:** Pure Domain-Driven Design (DDD) & Clean Architecture
- **Design Principles:**
  - **Pure Domain Layer:** Domain entities have private fields, explicit constructors (`New...`), and reconstitution factories (`Reconstruct...`), with 0 external framework dependencies.
  - **Strict SRP:** 1 Handler = 1 Domain. 1 Use Case = 1 Business Action.
  - **Manual Dependency Injection:** Pure Go wiring in `cmd/server/main.go`.
- **Default Server Port:** `8080` (Configurable via `PORT` environment variable)
- **Base URL Prefix:** `http://localhost:8080/api/v1` *(Dual-compatible with `/api/v2`)*
- **Database:** Supabase PostgreSQL with PgBouncer Transaction Pooler (`PreferSimpleProtocol: true`)

---

## 2. Authentication & Authorization

All protected endpoints require user identification. You can authenticate using either method:

### A. Supabase JWT Bearer Token (Standard Production)
Pass the access token received from login:
```http
Authorization: Bearer <supabase_jwt_access_token>
```

### B. Direct User ID Header (Development / Testing)
For local testing or microservice integrations:
```http
X-User-ID: 4aecd5c5-c0a9-4f52-ba2f-b6f4272218b4
```

---

## 3. Standard Response Format

All API responses follow a uniform JSON structure:

### Success Response
```json
{
  "success": true,
  "status": "success",
  "message": "Optional descriptive message",
  "data": { ... }
}
```

### Paginated Response
```json
{
  "success": true,
  "status": "success",
  "data": [ ... ],
  "total": 120,
  "page": 1,
  "limit": 20,
  "total_pages": 6
}
```

### Error Response
```json
{
  "success": false,
  "status": "error",
  "message": "Error description",
  "error": "Detailed validation or technical error"
}
```

---

## 4. Endpoints Quick Index

| Module | Base Route | Description |
| :--- | :--- | :--- |
| **System** | `/`, `/health` | Server heartbeat & health status |
| **Auth** | `/auth` | Login, profile check, token refresh, logout |
| **Categories** | `/categories` | Workshop product categories |
| **Products** | `/products` | Catalog, BOM components, 7 Pos Kas HPP, SKU generator |
| **Filaments** | `/filaments`, `/filament-profiles`, `/filament-material-rates` | Spools, color codes, weight tracking, material rates |
| **Machines** | `/machines` | 3D Printer fleet, maintenance parts, hours tracking |
| **Components** | `/components` | Screws, bearings, switches, hardware markup |
| **Packaging** | `/packaging-items`, `/packaging-presets` | Boxes, bubble wrap, tape, preset bundles |
| **Orders** | `/orders` | Manual orders hub, status transitions, payments |
| **Config** | `/config/shop`, `/config/marketplaces` | Electricity tariffs, machine lifespan, platform fee % |
| **Shopee** | `/shopee` | OAuth, order sync, SKU link, label PDF, 7 Pos Kas cashflow |
| **Production** | `/production/queue`, `/production/complete-job` | Priority print queue & stock/machine auto-deduct |

---

## 5. Detailed Endpoint Reference

### 5.1 System & Health

#### `GET /`
Returns root application metadata.
- **Auth:** Public
- **Response:**
  ```json
  {
    "app": "Symetra Lab Backend",
    "version": "2.0.0",
    "status": "online"
  }
  ```

#### `GET /health`
Returns health check status for load balancers.
- **Auth:** Public
- **Response:**
  ```json
  {
    "status": "healthy",
    "version": "2.0.0"
  }
  ```

---

### 5.2 Authentication (`/api/v1/auth`)

#### `POST /api/v1/auth/login`
Authenticates a user via Supabase Auth and returns JWT tokens.
- **Auth:** Public
- **Request Body:**
  ```json
  {
    "email": "user@example.com",
    "password": "password123"
  }
  ```
- **Response:**
  ```json
  {
    "success": true,
    "status": "success",
    "data": {
      "access_token": "eyJhbGciOi...",
      "refresh_token": "...",
      "expires_in": 3600,
      "user": {
        "id": "4aecd5c5-c0a9-4f52-ba2f-b6f4272218b4",
        "email": "user@example.com"
      }
    }
  }
  ```

#### `GET /api/v1/auth/me`
Retrieves the authenticated user profile.
- **Auth:** Bearer Token

#### `POST /api/v1/auth/refresh`
Refreshes an expired access token using a refresh token.
- **Auth:** Public
- **Request Body:**
  ```json
  {
    "refresh_token": "..."
  }
  ```

#### `POST /api/v1/auth/logout`
Terminates the active session.
- **Auth:** Protected

---

### 5.3 Product Categories (`/api/v1/categories`)

- `GET /api/v1/categories` — List all categories
- `POST /api/v1/categories` — Create category:
  ```json
  { "name": "Mechanical Parts" }
  ```
- `PUT /api/v1/categories/:id` — Update category
- `DELETE /api/v1/categories/:id` — Delete category

---

### 5.4 Products & 7 Pos Kas Costing (`/api/v1/products`)

Products feature real-time calculation of **7 Pos Kas**:
1. `filament_cost`
2. `hardware_cost`
3. `packaging_cost`
4. `electricity_cost`
5. `maintenance_cost`
6. `depreciation_cost`
7. `target_profit_idr`

#### `GET /api/v1/products`
Returns paginated list of products with calculated HPP and total units sold.
- **Query Params:** `page` (default: 1), `limit` (default: 20), `category` (optional), `search` (optional)

#### `GET /api/v1/products/:id`
Retrieves product details including Bill of Materials (BOM) components and packaging items.

#### `GET /api/v1/products/by-sku/:sku`
Look up a product variant by SKU code.

#### `POST /api/v1/products`
Creates a product with recipe/BOM items.
- **Request Body:**
  ```json
  {
    "name": "Heavy Duty Filament Spool Holder",
    "category": "Accessories",
    "material_type": "PLA",
    "default_weight_grams": 120.0,
    "default_print_time_hours": 3.5,
    "default_machine_id": "07f899c6-248b-44da-8f4d-7e329ea787de",
    "target_margin_percent": 35,
    "base_selling_price": 65000,
    "components": [
      {
        "component_id": "uuid-here",
        "quantity": 2
      }
    ]
  }
  ```

#### `PUT /api/v1/products/:id`
Updates product specifications and BOM structure.

#### `DELETE /api/v1/products/:id`
Deletes a product and cleans up BOM relations.

#### `POST /api/v1/products/auto-generate-skus`
Scans unassigned variants and automatically assigns standardized SKUs.

---

### 5.5 Filaments & Material Rates (`/api/v1/filaments`)

#### `GET /api/v1/filaments`
Lists active filament spools with preloaded technical profiles.

#### `POST /api/v1/filaments`
Registers a new spool. If profile does not exist, automatically creates default profile matching `(brand, material_type)`.
- **Request Body:**
  ```json
  {
    "brand": "eSUN",
    "material_type": "PLA+",
    "color_name": "Cold White",
    "color_hex": "#F4F4F4",
    "price_per_roll": 175000,
    "current_stock_grams": 1000,
    "low_stock_threshold_grams": 200
  }
  ```

#### `PATCH /api/v1/filaments/:id/stock`
Syncs or records physical scale weighing for a filament spool.
- **Request Body:**
  ```json
  {
    "current_stock_grams": 845.5,
    "is_weighed": true
  }
  ```

#### `GET /api/v1/filament-profiles`
Returns technical profiles (print temperature, bed temperature, density, speed limits).

#### `GET /api/v1/filament-material-rates`
Retrieves workshop default price-per-gram rates per material.

---

### 5.6 3D Printers & Maintenance Fleet (`/api/v1/machines`)

- `GET /api/v1/machines` — List all 3D printers with spare parts preloaded.
- `POST /api/v1/machines` — Register a new 3D printer:
  ```json
  {
    "name": "Bambu Lab P1S #1",
    "brand": "Bambu Lab",
    "total_purchase_cost": 11500000,
    "lifespan_hours": 10000,
    "avg_power_watts": 220
  }
  ```
- `PATCH /api/v1/machines/:id/state` — Update machine state (`IDLE`, `PRINTING`, `MAINTENANCE`, `OFFLINE`).
- `GET /api/v1/machines/:id/parts` — Maintenance parts assigned to machine.
- `POST /api/v1/machines/:id/parts` — Add replacement part (nozzle, belt, extruder, PTFE tube).
- `POST /api/v1/machines/parts/:part_id/replace` — Resets part operating hours counter upon physical replacement.

---

### 5.7 Hardware Components (`/api/v1/components`)

Manages non-printed hardware items (screws, heat inserts, bearings, magnets).
- `GET /api/v1/components`
- `POST /api/v1/components`:
  ```json
  {
    "name": "Brass Heat Insert M3x4x5",
    "price_per_unit": 450,
    "markup_percent": 25,
    "description": "Standard Voron / 3D print threaded insert"
  }
  ```

---

### 5.8 Packaging Items & Presets (`/api/v1/packaging-items`)

- `GET /api/v1/packaging-items` — Raw packaging items (box sizes, polymailers, bubble wrap).
- `GET /api/v1/packaging-presets` — Pre-bundled presets (e.g., "Small Box + 1m Bubble Wrap").

---

### 5.9 Direct & Manual Orders Hub (`/api/v1/orders`)

Manages direct offline, WhatsApp, or custom client orders.

#### `GET /api/v1/orders`
Lists all manual orders with line items.

#### `POST /api/v1/orders`
Creates a manual order. Automatically generates format `ORD-YYYYMMDD-XXXX` and calculates revenue, HPP, and profit.
- **Request Body:**
  ```json
  {
    "customer_name": "Pak Hendra Studio",
    "customer_contact": "081234567890",
    "source": "DIRECT_WHATSAPP",
    "items": [
      {
        "product_name": "Custom Drone Frame",
        "quantity": 2,
        "selling_price": 75000,
        "weight_grams": 60.0,
        "print_time_hours": 2.5,
        "machine_id": "07f899c6-248b-44da-8f4d-7e329ea787de"
      }
    ]
  }
  ```

#### `PATCH /api/v1/orders/:id/status`
Updates status (`PENDING`, `IN_PRODUCTION`, `COMPLETED`, `DELIVERED`, `CANCELLED`).

#### `PATCH /api/v1/orders/:id/payment`
Updates payment (`PAID`, `UNPAID`).

---

### 5.10 Workshop Configuration (`/api/v1/config`)

- `GET /api/v1/config/shop` — Retrieve electricity tariff, printer price, lifespan, failure buffer %.
- `PUT /api/v1/config/shop` — Update parameters:
  ```json
  {
    "electricity_tariff_per_kwh": 1750,
    "failure_buffer_percent": 10
  }
  ```
- `GET /api/v1/config/marketplaces` — List platform fee structures (Shopee, Tokopedia).
- `POST /api/v1/config/marketplaces` — Add marketplace fee structure:
  ```json
  {
    "name": "Shopee Star+",
    "commission_percent": 6.5,
    "promo_fee_percent": 3.0,
    "free_shipping_percent": 4.0,
    "order_fee_idr": 1000,
    "is_active": true
  }
  ```

---

### 5.11 Shopee Open Platform & 7 Pos Kas Finance (`/api/v1/shopee`)

#### `GET /api/v1/shopee/auth-url`
Generates OAuth2 authorize URL signed with HMAC-SHA256 for linking a seller shop.

#### `POST /api/v1/shopee/sync-shopee`
Synchronizes recent orders, escrow financial releases, and order items from Shopee.

#### `GET /api/v1/shopee/orders`
Lists online Shopee orders with calculated 7 Pos Kas financial distribution.

#### `POST /api/v1/shopee/link-sku`
Maps a Shopee item/variation ID to internal workshop product SKU:
```json
{
  "item_id": 123456789,
  "model_id": 987654321,
  "matched_sku": "SKU-HOLDER-BLK"
}
```

#### `POST /api/v1/shopee/orders/:order_sn/ship`
Arranges pickup or dropoff shipment on Shopee logistics.

#### `GET /api/v1/shopee/orders/:order_sn/shipping-label`
Streams shipping airway bill PDF directly for thermal printing.

#### `GET /api/v1/shopee/financial/cashflow-summary`
Aggregates workshop cash balances grouped by **7 Pos Kas**:
```json
{
  "success": true,
  "status": "success",
  "data": {
    "total_orders": 256,
    "total_gross_sales": 15420000,
    "total_marketplace_fees": 1542000,
    "total_escrow_net_in": 13878000,
    "total_hpp": 5200000,
    "kas_filamen": 2800000,
    "kas_komponen": 850000,
    "kas_packing": 600000,
    "kas_listrik": 350000,
    "kas_maintenance": 300000,
    "kas_depresiasi": 300000,
    "kas_laba_bersih": 8678000,
    "unmapped_items_count": 0
  }
}
```

#### `POST /api/v1/shopee/financial/recalculate`
Executes single-query vectorized recalculation of all escrow records.

---

### 5.12 Production Queue & Auto-Deduct (`/api/v1/production`)

#### `GET /api/v1/production/queue`
Combines online Shopee orders (`READY_TO_SHIP`, `PROCESSED`) and manual orders (`PENDING`, `IN_PRODUCTION`) into a unified priority print board.
- Automatically flags `is_urgent: true` when deadline < 24 hours.
- Returns `total_jobs`, `total_hours_waiting`, `urgent_jobs_count`.

#### `POST /api/v1/production/complete-job`
Marks 1 job complete. Automatically:
1. Deducts spool filament weight (`current_stock_grams`).
2. Increments printer operating hours (`total_hours_used`).
3. Resets printer state to `IDLE`.
4. Checks maintenance part lifespan warnings.
5. Transitions order status to `COMPLETED`.

- **Request Body:**
  ```json
  {
    "source": "MANUAL",
    "job_id": "0aedbcc9-64f5-40cf-8768-d32063195bba",
    "machine_id": "07f899c6-248b-44da-8f4d-7e329ea787de",
    "filament_id": "1e18b2b4-e0e7-4e4c-9995-22b192a94329"
  }
  ```

- **Response:**
  ```json
  {
    "success": true,
    "status": "success",
    "data": {
      "status": "success",
      "message": "Pekerjaan cetak selesai. Filamen terpotong untuk 2 pcs dan mesin beroperasi +3.00 jam.",
      "deducted_filaments": [
        "White: -90.00g (Sisa: 910.00g)"
      ],
      "added_machine_hours": 3.0,
      "low_stock_warning": false,
      "maintenance_warning": false
    }
  }
  ```

---

## 6. Finance Module

The Finance module tracks cashflow, procurement expenses, capital investments, and automatically records escrow income when Shopee orders are completed.

### A. Summary & Cashflow
- **`GET /api/v1/finance/summary?date_from=2026-01-01&date_to=2026-12-31`**
  Returns total income, total expense, total capital in, net cashflow (`total_income - total_expense`), and totals broken down by category.

### B. Cashflow Transactions
- **`GET /api/v1/finance/transactions?type=EXPENSE&category=FILAMENT&date_from=2026-01-01&date_to=2026-12-31`**
  Filter by type (`INCOME`, `EXPENSE`, `CAPITAL_IN`), category, or date range.
- **`POST /api/v1/finance/transactions`**
  Record manual transaction.
  ```json
  {
    "type": "EXPENSE",
    "category": "ELECTRICITY",
    "amount": 250000,
    "description": "Tagihan listrik workshop bulan ini",
    "transaction_date": "2026-10-04",
    "notes": "Token PLN"
  }
  ```
- **`DELETE /api/v1/finance/transactions/:id`**
  Delete transaction.

### C. Purchase Orders (Procurement & Auto-Expense)
- **`GET /api/v1/finance/purchase-orders`** - List all purchase orders with items.
- **`GET /api/v1/finance/purchase-orders/:id`** - Get purchase order detail.
- **`POST /api/v1/finance/purchase-orders`**
  Creates a purchase order and **automatically generates a linked `EXPENSE` transaction** in `finance_transactions`.
  ```json
  {
    "supplier_name": "Sunlu Official Store",
    "purchase_date": "2026-10-04",
    "notes": "Restock filamen",
    "items": [
      {
        "item_type": "FILAMENT",
        "item_name": "PLA+ Black 1kg",
        "quantity": 3,
        "unit": "roll",
        "unit_price": 135000
      }
    ]
  }
  ```
- **`DELETE /api/v1/finance/purchase-orders/:id`** - Delete purchase order.

### D. Capital Management (Modal & Auto-Capital In)
- **`GET /api/v1/finance/capital`** - List all capital records.
- **`POST /api/v1/finance/capital`**
  Records capital injection (`INITIAL` or `ADDITION`) and **automatically creates a linked `CAPITAL_IN` transaction**.
  ```json
  {
    "type": "INITIAL",
    "amount": 10000000,
    "description": "Modal awal pembukaan workshop",
    "record_date": "2026-10-04",
    "notes": "Setoran dari rekening pribadi"
  }
  ```
- **`GET /api/v1/finance/capital/total`** - Get total injected capital.

---

## 7. Running Locally

```bash
# 1. Navigate to backend directory
cd symetra-lab-backend-v2

# 2. Run with Go
go run cmd/server/main.go
```
