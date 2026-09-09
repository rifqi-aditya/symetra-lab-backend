# Dokumentasi API - Symetra Lab Backend

Backend manajemen toko yang terintegrasi dengan **Shopee Open Platform (Open API v2)** dan database **Supabase PostgreSQL**.

---

## 🌐 Base URL

| Environment | URL |
| :--- | :--- |
| **Production (Vercel)** | `https://symetra-lab-backend.vercel.app` |
| **Local Development** | `http://localhost:8080` |

---

## 📋 Daftar Endpoint

| Method | Endpoint | Deskripsi |
| :--- | :--- | :--- |
| `GET` | `/` | Informasi versi dan aplikasi |
| `GET` | `/health` | Health check server & koneksi database |
| `GET` | `/api/v1/shopee/auth-url` | Generate URL otorisasi toko Shopee |
| `GET` | `/api/v1/shopee/callback` | Callback redirect OAuth Shopee |
| `GET` | `/api/v1/shopee/shops` | Mengambil daftar toko yang terhubung |
| `POST` | `/api/v1/shopee/shops/:shop_id/refresh` | Manual refresh access token toko |
| `POST` | `/api/v1/shopee/shops/:shop_id/sync-orders` | Sinkronisasi pesanan, item, dan rincian escrow dari Shopee |
| `GET` | `/api/v1/shopee/shops/:shop_id/orders` | Mengambil daftar pesanan toko (filter status, search, deadline SLA) |
| `GET` | `/api/v1/shopee/orders/:order_sn` | Mengambil detail lengkap 1 pesanan (item, custom note, potongan fee) |
| `POST` | `/api/v1/shopee/orders/:order_sn/ship` | Konfirmasi "Atur Pengiriman" pesanan & terbitkan no resi kurir |
| `GET` | `/api/v1/shopee/orders/:order_sn/shipping-label` | Download / Cetak langsung PDF Label Resi Pengiriman Thermal (100x150 mm) |

---

## 📖 Rincian Endpoint

### 1. Root & Health Check

#### `GET /`
Mengembalikan informasi dasar aplikasi.

* **Request**:
  ```bash
  curl -X GET https://symetra-lab-backend.vercel.app/
  ```
* **Response (200 OK)**:
  ```json
  {
    "app": "Symetra Lab Backend",
    "version": "1.0.0",
    "shopee": "Open Platform v2 Ready"
  }
  ```

---

#### `GET /health`
Mengecek kondisi kesehatan server backend dan status koneksi ke Supabase PostgreSQL.

* **Request**:
  ```bash
  curl -X GET https://symetra-lab-backend.vercel.app/health
  ```
* **Response (200 OK)**:
  ```json
  {
    "status": "ok",
    "database": "connected",
    "environment": "production"
  }
  ```

---

### 2. Modul Shopee Open Platform OAuth v2

#### `GET /api/v1/shopee/auth-url`
Menghasilkan URL otorisasi resmi Shopee dengan tanda tangan **HMAC-SHA256** dan *timestamp* yang valid.

* **Query Parameters (Opsional)**:
  | Parameter | Tipe | Default | Keterangan |
  | :--- | :--- | :--- | :--- |
  | `redirect` | `string` (`true`/`false`) | `false` | Jika diisi `true`, browser akan langsung dialihkan (*307 Temporary Redirect*) ke halaman login Shopee. |

* **Request Contoh (JSON Mode)**:
  ```bash
  curl -X GET https://symetra-lab-backend.vercel.app/api/v1/shopee/auth-url
  ```
* **Response (200 OK)**:
  ```json
  {
    "status": "success",
    "auth_url": "https://partner.shopeemobile.com/api/v2/shop/auth_partner?partner_id=1243796&redirect=https%3A%2F%2Fsymetra-lab-backend.vercel.app%2Fapi%2Fv1%2Fshopee%2Fcallback&sign=99b2eb...&timestamp=1788959645",
    "message": "Buka URL ini di browser untuk mengotorisasi toko Shopee Anda"
  }
  ```
* **Request Contoh (Direct Redirect Mode)**:
  ```text
  Buka di browser: https://symetra-lab-backend.vercel.app/api/v1/shopee/auth-url?redirect=true
  ```

---

#### `GET /api/v1/shopee/callback`
Endpoint yang dipanggil secara otomatis oleh Shopee (*Redirect URI*) setelah pemilik toko menyetujui izin aplikasi.

> [!NOTE]
> Endpoint ini dipanggil via browser oleh Shopee, bukan dipanggil manual oleh frontend Anda.

* **Query Parameters dari Shopee**:
  | Parameter | Tipe | Keterangan |
  | :--- | :--- | :--- |
  | `code` | `string` | Authorization code sementara dari Shopee |
  | `shop_id` | `uint64` | ID Toko Shopee yang berhasil diotorisasi |

* **Alur Eksekusi**:
  1. Backend menukar `code` dan `shop_id` ke Shopee endpoint `/api/v2/auth/token/get`.
  2. Shopee mengembalikan `access_token` (berlaku 4 jam) dan `refresh_token` (berlaku 30 hari).
  3. Backend melakukan **Upsert** (Insert or Update on Conflict) ke tabel `shops` di Supabase PostgreSQL.
  4. Menampilkan halaman sukses HTML kepada pengguna.

---

#### `GET /api/v1/shopee/shops`
Mengambil daftar seluruh toko Shopee yang sudah berhasil tersambung ke backend beserta status masa berlaku tokennya.

* **Request**:
  ```bash
  curl -X GET https://symetra-lab-backend.vercel.app/api/v1/shopee/shops
  ```
* **Response (200 OK)**:
  ```json
  {
    "status": "success",
    "total": 1,
    "data": [
      {
        "shop_id": 12345678,
        "shop_name": "Toko 3D Printing Saya",
        "is_token_expired": false,
        "access_token_expires_at": "2026-09-09T23:59:59Z",
        "connected_at": "2026-09-09T20:00:00Z"
      }
    ]
  }
  ```
  *(Catatan: `access_token` dan `refresh_token` sengaja disembunyikan `json:"-"` pada response demi keamanan data).*

---

#### `POST /api/v1/shopee/shops/:shop_id/refresh`
Memicu pembaruan (*refresh*) `access_token` toko secara manual menggunakan `refresh_token` yang tersimpan di database.

* **URL Parameters**:
  | Parameter | Tipe | Keterangan |
  | :--- | :--- | :--- |
  | `shop_id` | `uint64` | ID Toko Shopee yang ingin diperbarui tokennya |

* **Request**:
  ```bash
  curl -X POST https://symetra-lab-backend.vercel.app/api/v1/shopee/shops/12345678/refresh
  ```
* **Response (200 OK)**:
  ```json
  {
    "status": "success",
    "message": "Access token Shopee berhasil diperbarui",
    "expires_at": "2026-09-10T04:00:00Z"
  }
  ```
* **Response Error (404 Not Found)**:
  ```json
  {
    "status": "error",
    "message": "Toko dengan shop_id tersebut belum terdaftar"
  }
  ```

---

### 3. Modul Pesanan & Keuangan (Orders & Escrow)

#### `POST /api/v1/shopee/shops/:shop_id/sync-orders`
Menarik daftar pesanan dari Shopee, mengambil rincian varian produk & catatan pembeli, mengkalkulasi rincian potongan admin/layanan Shopee, dan menyimpannya (upsert) ke Supabase PostgreSQL.

* **Path Parameter**:
  | Parameter | Tipe | Deskripsi |
  | :--- | :--- | :--- |
  | `shop_id` | `uint64` | ID Toko Shopee yang ingin disinkronkan pesanannya |

* **Query Parameters (Opsional)**:
  | Parameter | Tipe | Default | Keterangan |
  | :--- | :--- | :--- | :--- |
  | `days` | `int` | `15` | Rentang hari ke belakang untuk penarikan pesanan |
  | `status` | `string` | *(Semua)* | Filter status: `READY_TO_SHIP`, `PROCESSED`, `SHIPPED`, `COMPLETED`, `CANCELLED` |

* **Request Contoh**:
  ```bash
  curl -X POST "https://symetra-lab-backend.vercel.app/api/v1/shopee/shops/227895003/sync-orders?days=30&status=READY_TO_SHIP"
  ```
* **Response (200 OK)**:
  ```json
  {
    "status": "success",
    "message": "Berhasil sinkronisasi 1 pesanan dari Shopee",
    "synced_count": 1,
    "data": [
      {
        "order_sn": "260909K7J0045Q",
        "shop_id": 227895003,
        "order_status": "READY_TO_SHIP",
        "buyer_user_id": 12345,
        "buyer_username": "buyer_symetra",
        "message_to_seller": "Tolong warna hitam doff infill 30% ya kak",
        "ship_by_date": 1789123200,
        "ship_by_date_time": "2026-09-11T12:00:00Z",
        "shipping_carrier": "SPX Standard",
        "total_amount": 150000.00,
        "items": [
          {
            "item_id": 98765,
            "item_name": "Stand Handphone Articulated 3D Print",
            "model_name": "Matte Black - PLA+",
            "model_sku": "SKU-STAND-BLK",
            "quantity": 2,
            "original_price": 75000.00,
            "discounted_price": 75000.00
          }
        ],
        "escrow": {
          "order_sn": "260909K7J0045Q",
          "escrow_amount": 131800.00,
          "selling_price": 150000.00,
          "commission_fee": 9750.00,
          "commission_rule_name": "Biaya Komisi Kategori Hobi & Koleksi",
          "commission_percentage": 6.5,
          "service_fee": 6000.00,
          "service_rule_name": "Program Gratis Ongkir XTRA",
          "service_percentage": 4.0,
          "seller_transaction_fee": 6000.00,
          "seller_order_processing_fee": 1250.00,
          "seller_voucher_discount": 0.00
        }
      }
    ]
  }
  ```

---

#### `GET /api/v1/shopee/shops/:shop_id/orders`
Mengambil daftar pesanan dari database lokal Supabase untuk ditampilkan di dashboard frontend (dilengkapi fitur pencarian catatan pembeli, filter status, dan sorting deadline SLA terdekat).

* **Query Parameters (Opsional)**:
  | Parameter | Tipe | Default | Keterangan |
  | :--- | :--- | :--- | :--- |
  | `status` | `string` | *(Semua)* | Filter status pesanan (`READY_TO_SHIP`, `COMPLETED`, dll) |
  | `search` | `string` | - | Cari berdasarkan `order_sn`, username pembeli, atau kata kunci di **catatan pembeli** |
  | `sort_by` | `string` | `newest` | Opsi: `deadline` (prioritas batas kirim paling dekat), `newest` (terbaru), `amount` (nominal tertinggi) |
  | `page` | `int` | `1` | Halaman data |
  | `page_size` | `int` | `20` | Jumlah data per halaman (maks 100) |

* **Request Contoh (Cari pesanan yang deadline-nya paling mendesak)**:
  ```bash
  curl -X GET "https://symetra-lab-backend.vercel.app/api/v1/shopee/shops/227895003/orders?status=READY_TO_SHIP&sort_by=deadline"
  ```
* **Response (200 OK)**:
  ```json
  {
    "status": "success",
    "data": {
      "orders": [ ... ],
      "total": 1,
      "page": 1,
      "page_size": 20,
      "total_pages": 1
    }
  }
  ```

---

#### `GET /api/v1/shopee/orders/:order_sn`
Mengambil detail 1 pesanan secara spesifik beserta varian item dan seluruh transparansi biaya admin Shopee.

* **Request Contoh**:
  ```bash
  curl -X GET https://symetra-lab-backend.vercel.app/api/v1/shopee/orders/260909K7J0045Q
  ```

---

### 4. Modul Logistik & Pengiriman (Logistics & Shipping Labels)

#### `POST /api/v1/shopee/orders/:order_sn/ship`
Melakukan aksi **"Atur Pengiriman"** untuk pesanan yang berstatus `READY_TO_SHIP`. Backend secara otomatis memeriksa parameter pengiriman (apakah `dropoff` ke gerai kurir terdekat atau `pickup` ke workshop), mengirim konfirmasi siap kirim ke Shopee, mengambil nomor resi resmi, dan mengupdate status pesanan di database menjadi `PROCESSED`.

* **Path Parameter**:
  | Parameter | Tipe | Deskripsi |
  | :--- | :--- | :--- |
  | `order_sn` | `string` | Nomor pesanan Shopee yang ingin dikirim |

* **Request Contoh**:
  ```bash
  curl -X POST https://symetra-lab-backend.vercel.app/api/v1/shopee/orders/260909K7J0045Q/ship
  ```
* **Response (200 OK)**:
  ```json
  {
    "status": "success",
    "message": "Pengiriman berhasil diatur",
    "order_sn": "260909K7J0045Q",
    "order_status": "PROCESSED",
    "tracking_number": "SPXID04829103948"
  }
  ```

---

#### `GET /api/v1/shopee/orders/:order_sn/shipping-label`
Mengambil file biner **PDF Label Resi Pengiriman Thermal (100x150 mm)** resmi dari Shopee dan langsung menyajikannya ke browser atau dashboard frontend (*inline streaming*).

* **Fitur Utama**:
  - Format standar thermal 100x150 mm (cocok untuk printer thermal seperti Xprinter, Kassen, Phomemo, dll).
  - Browser dapat langsung merender PDF di tab baru atau di dalam `<iframe>` untuk pencetakan sekali klik.
  - Header otomatis: `Content-Type: application/pdf` dan `Content-Disposition: inline; filename="resi_260909K7J0045Q.pdf"`.

* **Request Contoh**:
  ```bash
  # Buka langsung di browser atau panggil via cURL
  curl -X GET https://symetra-lab-backend.vercel.app/api/v1/shopee/orders/260909K7J0045Q/shipping-label --output label_resi.pdf
  ```

---

## 🗄️ Skema Database Supabase PostgreSQL

Tabel otomatis dikelola via **GORM AutoMigrate** dan terisolasi dari tabel internal 3D printing:

```sql
-- 1. Tabel Toko (Kredensial & Token)
CREATE TABLE "shops" (
    "id" bigserial PRIMARY KEY,
    "shop_id" bigint NOT NULL UNIQUE,
    "shop_name" varchar(255),
    "region" varchar(10),
    "access_token" text NOT NULL,
    "refresh_token" text NOT NULL,
    "access_token_expires_at" timestamptz,
    "refresh_token_expires_at" timestamptz,
    "created_at" timestamptz,
    "updated_at" timestamptz
);

-- 2. Tabel Pesanan Shopee
CREATE TABLE "shopee_orders" (
    "order_sn" varchar(64) PRIMARY KEY,
    "shop_id" bigint NOT NULL,
    "order_status" varchar(32) NOT NULL,
    "buyer_user_id" bigint,
    "buyer_username" varchar(128),
    "message_to_seller" text,
    "ship_by_date" bigint,
    "ship_by_date_time" timestamptz,
    "shipping_carrier" varchar(64),
    "tracking_number" varchar(64),
    "total_amount" numeric(15,2),
    "buyer_cancel_reason" varchar(255),
    "create_time_shopee" bigint,
    "update_time_shopee" bigint,
    "created_at" timestamptz,
    "updated_at" timestamptz
);

-- 3. Tabel Item/Varian Pesanan Shopee
CREATE TABLE "shopee_order_items" (
    "id" bigserial PRIMARY KEY,
    "order_sn" varchar(64) NOT NULL REFERENCES "shopee_orders"("order_sn") ON DELETE CASCADE,
    "item_id" bigint,
    "item_name" varchar(255),
    "item_sku" varchar(128),
    "model_id" bigint,
    "model_name" varchar(255),
    "model_sku" varchar(128),
    "quantity" integer,
    "original_price" numeric(15,2),
    "discounted_price" numeric(15,2),
    "created_at" timestamptz,
    "updated_at" timestamptz
);

-- 4. Tabel Transparansi Keuangan & Escrow Shopee
CREATE TABLE "shopee_order_escrows" (
    "order_sn" varchar(64) PRIMARY KEY REFERENCES "shopee_orders"("order_sn") ON DELETE CASCADE,
    "escrow_amount" numeric(15,2),
    "selling_price" numeric(15,2),
    "commission_fee" numeric(15,2),
    "commission_rule_name" varchar(255),
    "commission_percentage" numeric(5,2),
    "service_fee" numeric(15,2),
    "service_rule_name" varchar(255),
    "service_percentage" numeric(5,2),
    "seller_transaction_fee" numeric(15,2),
    "seller_order_processing_fee" numeric(15,2),
    "seller_voucher_discount" numeric(15,2),
    "created_at" timestamptz,
    "updated_at" timestamptz
);
```

---

## 🔐 Manajemen Siklus Token Shopee

1. **Access Token**
   - Masa berlaku: **4 Jam** (14.400 detik).
   - Fungsi pembantu: `shop.IsTokenExpired()` memberikan toleransi buffer **5 menit** lebih awal sebelum expired sebenarnya untuk memastikan request API tidak gagal di tengah jalan.
   - **Auto-Refresh**: Handler pesanan secara otomatis memperbarui access token di database jika mendeteksi token sudah kedaluwarsa.
2. **Refresh Token**
   - Masa berlaku: **30 Hari**.
   - Digunakan untuk meminta `access_token` baru tanpa mengharuskan seller login ulang.

