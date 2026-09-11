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
| `GET` | `/api/v1/filaments` | Mengambil seluruh roll filamen beserta profil teknisnya (*flattened*) |
| `GET` | `/api/v1/filaments/:id` | Mengambil rincian 1 roll filamen beserta profil teknis |
| `POST` | `/api/v1/filaments` | Menambahkan roll filamen baru (otomatis membuat/mengaitkan profil) |
| `PUT` | `/api/v1/filaments/:id` | Memperbarui data roll filamen dan profil teknis |
| `DELETE` | `/api/v1/filaments/:id` | Menghapus roll filamen |
| `POST` | `/api/v1/filaments/:id/sync-stock` | Mencatat hasil timbangan fisik roll (kalkulasi sisa gramasi bersih) |
| `GET` | `/api/v1/filament-profiles` | Mengambil daftar master profil teknis bahan filamen |
| `GET` | `/api/v1/machines` | Mengambil seluruh mesin printer beserta suku cadang perawatannya |
| `GET` | `/api/v1/machines/:id` | Mengambil rincian 1 mesin printer |
| `POST` | `/api/v1/machines` | Mendaftarkan 3D printer baru ke bengkel |
| `PUT` | `/api/v1/machines/:id` | Memperbarui parameter printer (daya watt, harga beli, jam pakai) |
| `PATCH` | `/api/v1/machines/:id/state` | Mengubah status mesin (`IDLE`, `PRINTING`, `MAINTENANCE`, `OFFLINE`) |
| `DELETE` | `/api/v1/machines/:id` | Menghapus printer beserta suku cadangnya |
| `GET` | `/api/v1/machines/:id/parts` | Mengambil daftar suku cadang perawatan untuk printer |
| `POST` | `/api/v1/machines/:id/parts` | Menambahkan suku cadang baru ke printer |
| `PUT` | `/api/v1/machines/parts/:part_id` | Memperbarui suku cadang (stok cadangan, biaya unit) |
| `POST` | `/api/v1/machines/parts/:part_id/replace` | Penggantian suku cadang (reset jam pakai ke 0 & kurangi stok) |
| `DELETE` | `/api/v1/machines/parts/:part_id` | Menghapus suku cadang |
| `GET` | `/api/v1/components` | Mengambil seluruh komponen tambahan (filter search `?q=`, sort) |
| `GET` | `/api/v1/components/:id` | Mengambil rincian 1 komponen tambahan |
| `POST` | `/api/v1/components` | Menambahkan komponen tambahan baru |
| `PUT` | `/api/v1/components/:id` | Memperbarui data komponen (harga pokok, markup, deskripsi) |
| `DELETE` | `/api/v1/components/:id` | Menghapus komponen (dilengkapi safety check resep produk) |
| `GET` | `/api/v1/packaging-items` | Mengambil seluruh bahan kemasan (filter `category`, search `?q=`, sort) |
| `GET` | `/api/v1/packaging-items/:id` | Mengambil rincian 1 bahan kemasan |
| `POST` | `/api/v1/packaging-items` | Menambahkan bahan kemasan baru (otomatis hitung `unit_cost`) |
| `PUT` | `/api/v1/packaging-items/:id` | Memperbarui data bahan kemasan |
| `DELETE` | `/api/v1/packaging-items/:id` | Menghapus bahan kemasan (dengan safety check relasi) |
| `GET` | `/api/v1/packaging-presets` | Mengambil seluruh preset kemasan beserta komponen & total biaya terhitung |
| `GET` | `/api/v1/packaging-presets/:id` | Mengambil rincian 1 preset kemasan |
| `POST` | `/api/v1/packaging-presets` | Membuat preset kemasan baru beserta susunan item & kuantitas |
| `PUT` | `/api/v1/packaging-presets/:id` | Memperbarui preset kemasan |
| `DELETE` | `/api/v1/packaging-presets/:id` | Menghapus preset kemasan |

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

---

## 🧵 Modul 5: Inventori Bahan Baku (Filaments & Profiles)

Modul ini mengelola data gulungan (*spool*) filamen fisik dan profil teknis pencetakan (suhu nozzle/bed, retraksi, flow ratio, bobot spool kosong). Format respon `GET` otomatis digabungkan (*flattened*) sehingga siap dikonsumsi langsung oleh komponen frontend.

### 1. `GET /api/v1/filaments`
Mengambil seluruh gulungan filamen milik user beserta profil teknisnya.

* **Query Parameters (Opsional)**:
  - `brand` (string): Filter berdasarkan brand (contoh: `Sunlu`, `eSUN`).
  - `material_type` (string): Filter berdasarkan tipe material (contoh: `PLA`, `PETG`, `ABS`).
  - `user_id` (string): ID user spesifik (default menggunakan admin Supabase).

* **Contoh Request**:
  ```bash
  curl -X GET "https://symetra-lab-backend.vercel.app/api/v1/filaments"
  ```

* **Contoh Response (200 OK)**:
  ```json
  {
    "status": "success",
    "total": 19,
    "data": [
      {
        "id": "67324391-da1d-4001-a96a-0498305c4125",
        "user_id": "4aecd5c5-c0a9-4f52-ba2f-b6f4272218b4",
        "profile_id": "e81881ef-1fe5-412d-b0ad-ec82a1fc8375",
        "brand": "Sunlu",
        "material_type": "PLA+",
        "diameter_mm": 1.75,
        "nozzle_temp": 210,
        "bed_temp": 60,
        "retraction_length": 0.8,
        "flow_ratio": 0.98,
        "pressure_advance": 0.025,
        "cooling_fan_percent": 100,
        "max_volumetric_speed": 15.0,
        "empty_spool_weight_grams": 220.0,
        "spool_weight_grams": 1000.0,
        "color_name": "Matte Navy Blue",
        "color_hex": "#1B263B",
        "sku": "SNLU-PLA-NVY",
        "price_per_roll": 145000,
        "current_stock_grams": 850.0,
        "low_stock_threshold_grams": 200.0,
        "last_weighed_grams": 1070.0,
        "last_weighed_at": "2026-09-09T18:00:00Z",
        "created_at": "2026-09-01T10:00:00Z",
        "updated_at": "2026-09-09T18:00:00Z"
      }
    ]
  }
  ```

---

### 2. `POST /api/v1/filaments`
Menambahkan roll filamen baru. Jika profil teknis dengan kombinasi `(brand, material_type)` belum ada di database, profil baru akan otomatis dibuatkan.

* **Contoh Request Body**:
  ```json
  {
    "brand": "eSUN",
    "material_type": "PLA+",
    "nozzle_temp": 215,
    "bed_temp": 60,
    "empty_spool_weight_grams": 230,
    "spool_weight_grams": 1000,
    "color_name": "Cold White",
    "color_hex": "#FFFFFF",
    "sku": "ESUN-PLA-WHT",
    "price_per_roll": 150000,
    "current_stock_grams": 1000,
    "low_stock_threshold_grams": 250
  }
  ```

---

### 3. `POST /api/v1/filaments/:id/sync-stock`
Fitur pencatatan timbangan fisik spool di bengkel. Backend secara otomatis menghitung sisa filamen bersih:
$$\text{Calculated Net Grams} = \text{Gross Weight} - \text{Empty Spool Weight}$$

* **Contoh Request Body**:
  ```json
  {
    "gross_weight_grams": 730.0
  }
  ```

* **Contoh Response (200 OK)**:
  ```json
  {
    "status": "success",
    "message": "Stok berhasil disinkronkan dari timbangan",
    "data": {
      "filament_id": "67324391-da1d-4001-a96a-0498305c4125",
      "gross_weight_grams": 730.0,
      "empty_spool_weight_grams": 230.0,
      "calculated_net_grams": 500.0,
      "last_weighed_at": "2026-09-10T00:50:00Z"
    }
  }
  ```

---

### 4. `GET /api/v1/filament-profiles`
Mengambil daftar master profil teknis bahan filamen.

---

## ⚙️ Modul 6: Mesin 3D Printer & Suku Cadang (Machines & Maintenance Parts)

Modul ini mengelola unit mesin cetak 3D di bengkel Symetra Lab, parameter biaya investasi, daya listrik, depresiasi jam terbang, serta monitoring suku cadang *wear-and-tear* (nozzle, build plate, belt).

### 1. `GET /api/v1/machines`
Mengambil seluruh daftar mesin 3D printer beserta relasi komponen suku cadang perawatannya.

* **Query Parameters (Opsional)**:
  - `state` (string): Filter berdasarkan status mesin (`IDLE`, `PRINTING`, `MAINTENANCE`, `OFFLINE`).

* **Contoh Response (200 OK)**:
  ```json
  {
    "status": "success",
    "total": 1,
    "data": [
      {
        "id": "e4933924-f4aa-49ee-b9b5-fcfb9195b058",
        "user_id": "4aecd5c5-c0a9-4f52-ba2f-b6f4272218b4",
        "name": "A1 Semoga Ga Libur",
        "brand": "Bambu Lab",
        "total_purchase_cost": 7500000,
        "salvage_value": 1500000,
        "lifespan_hours": 5000,
        "total_hours_used": 1420.5,
        "avg_power_watts": 350,
        "current_state": "IDLE",
        "maintenance_parts": [
          {
            "id": "180373fe-aef4-4f01-8b2b-426c117b3f94",
            "machine_id": "e4933924-f4aa-49ee-b9b5-fcfb9195b058",
            "part_name": "Hardened Steel Nozzle 0.4mm",
            "cost_idr": 185000,
            "lifespan_hours": 800,
            "stock_quantity": 2,
            "hours_used_current": 150.0,
            "last_replaced_at": "2026-08-15T09:00:00Z"
          }
        ]
      }
    ]
  }
  ```

---

### 2. `PATCH /api/v1/machines/:id/state`
Mengubah status mesin secara cepat (*lightweight*). Sangat berguna saat memulai pencetakan atau saat mesin masuk jadwal servis.

* **Pilihan State Valid**: `IDLE`, `PRINTING`, `MAINTENANCE`, `OFFLINE`.
* **Contoh Request Body**:
  ```json
  {
    "state": "PRINTING"
  }
  ```

---

### 3. `POST /api/v1/machines/parts/:part_id/replace`
Mencatat penggantian suku cadang yang aus. Secara otomatis:
1. Mereset `hours_used_current` menjadi `0`.
2. Mengurangi `stock_quantity` suku cadang sebesar `1`.
3. Memperbarui `last_replaced_at` ke waktu sekarang (*timestamp* saat ini).

* **Contoh Response (200 OK)**:
  ```json
  {
    "status": "success",
    "message": "Penggantian suku cadang berhasil dicatat",
    "data": {
      "part_id": "180373fe-aef4-4f01-8b2b-426c117b3f94",
      "part_name": "Hardened Steel Nozzle 0.4mm",
      "hours_used_current": 0,
      "remaining_stock": 1,
      "last_replaced_at": "2026-09-10T00:50:30Z"
    }
  }
  ```

---

## 🔩 Modul 7: Komponen Tambahan (Hardware & Aksesoris Non-3D Print)

Modul ini mengelola data inventori komponen non-cetak yang disematkan ke dalam produk akhir bengkel Symetra Lab (seperti baut, mur, magnet, gantungan kunci, bearing, modul TP4056, saklar, lampu LED, baterai, serta ongkos jasa perakitan). Respon API otomatis menghitung estimasi harga jual yang disarankan (`calculated_selling_price`) berdasarkan persentase markup.

### 1. `GET /api/v1/components`
Mengambil daftar seluruh komponen tambahan milik user.

* **Query Parameters (Opsional)**:
  - `q` (string): Pencarian kata kunci pada nama atau deskripsi komponen (contoh: `switch`, `baut`, `bearing`).
  - `sort` (string): Urutkan berdasarkan `name`, `price`, atau `created_at` (default: `name`).
  - `order` (string): `ASC` atau `DESC` (default: `ASC`).

* **Contoh Request**:
  ```bash
  curl -X GET "https://symetra-lab-backend.vercel.app/api/v1/components?q=switch"
  ```

* **Contoh Response (200 OK)**:
  ```json
  {
    "status": "success",
    "total": 1,
    "data": [
      {
        "id": "828352ce-c0e2-48b5-a55d-af8d50d33fd3",
        "user_id": "4aecd5c5-c0a9-4f52-ba2f-b6f4272218b4",
        "name": "Blue Switch Keyboard",
        "price_per_unit": 1600,
        "default_markup_percent": 25,
        "calculated_selling_price": 2000,
        "description": "Blue Switch Keyboard Tactile Clicky",
        "created_at": "2026-07-04T12:00:00Z",
        "updated_at": "2026-07-04T12:00:00Z"
      }
    ]
  }
  ```

---

### 2. `GET /api/v1/components/:id`
Mengambil rincian 1 komponen spesifik berdasarkan UUID.

---

### 3. `POST /api/v1/components`
Menambahkan komponen hardware baru ke inventori bengkel.

* **Contoh Request Body**:
  ```json
  {
    "name": "Magnet Neodymium 5x2mm",
    "price_per_unit": 450,
    "default_markup_percent": 30,
    "description": "Magnet bulat kuat untuk engsel box dan miniatur"
  }
  ```

* **Contoh Response (201 Created)**:
  ```json
  {
    "status": "success",
    "message": "Komponen berhasil ditambahkan",
    "data": {
      "id": "3a7b91d2-9982-4fa1-b1e2-9b2f34918e9a",
      "user_id": "4aecd5c5-c0a9-4f52-ba2f-b6f4272218b4",
      "name": "Magnet Neodymium 5x2mm",
      "price_per_unit": 450,
      "default_markup_percent": 30,
      "calculated_selling_price": 585,
      "description": "Magnet bulat kuat untuk engsel box dan miniatur",
      "created_at": "2026-09-11T18:05:00Z",
      "updated_at": "2026-09-11T18:05:00Z"
    }
  }
  ```

---

### 4. `PUT /api/v1/components/:id`
Memperbarui data komponen (harga pokok beli, target markup %, atau catatan spesifikasi).

* **Contoh Request Body**:
  ```json
  {
    "price_per_unit": 500,
    "default_markup_percent": 40
  }
  ```

---

### 5. `DELETE /api/v1/components/:id`
Menghapus komponen dari inventori. Endpoint ini memiliki fitur **Safe Delete**: jika komponen sedang tercatat dalam resep produk (`product_components`), sistem akan menolak penghapusan dengan status `409 Conflict` guna menjaga keutuhan resep produk.

* **Query Parameters**:
  - `force=true` (opsional): Memaksa penghapusan meskipun komponen terhubung ke resep produk.

* **Contoh Response Penolakan (409 Conflict)**:
  ```json
  {
    "status": "error",
    "message": "Komponen tidak dapat dihapus karena sedang digunakan dalam 4 resep produk. Gunakan ?force=true jika tetap ingin menghapus.",
    "usage_count": 4
  }
  ```

* **Contoh Response Berhasil (200 OK)**:
  ```json
  {
    "status": "success",
    "message": "Komponen berhasil dihapus"
  }
  ```

---

## 📦 Modul 8: Bahan Kemasan & Preset Packing (Packaging System)

Modul ini mengelola inventori bahan kemasan pengiriman (kardus diecut, plastik clip hologram, polymailer, sticker label, bubble wrap) dan bundel kemasan standar (*Packaging Presets*). Backend secara otomatis menghitung biaya satuan bahan pokok (`unit_cost = purchase_price / purchase_quantity`) serta total estimasi biaya kemasan per pesanan/produk.

### 1. `GET /api/v1/packaging-items`
Mengambil daftar seluruh bahan kemasan individual milik user.

* **Query Parameters (Opsional)**:
  - `category` (string): Filter jenis kemasan (`BOX`, `PLASTIC`, `LABEL`, `OTHER`).
  - `q` (string): Pencarian kata kunci pada nama kemasan (contoh: `clip`, `mailer`, `diecut`).
  - `sort` (string): Urutkan berdasarkan `name`, `unit_cost`, `stock`, `created_at` (default: `name`).
  - `order` (string): `ASC` atau `DESC` (default: `ASC`).

* **Contoh Request**:
  ```bash
  curl -X GET "https://symetra-lab-backend.vercel.app/api/v1/packaging-items"
  ```

* **Contoh Response (200 OK)**:
  ```json
  {
    "status": "success",
    "total": 6,
    "data": [
      {
        "id": "1aa4dcce-5599-43bd-9a5f-95148ed3fc0a",
        "user_id": "4aecd5c5-c0a9-4f52-ba2f-b6f4272218b4",
        "name": "Diecut Kecil",
        "category": "BOX",
        "unit_type": "PCS",
        "purchase_price": 515,
        "purchase_quantity": 1,
        "unit_cost": 515,
        "stock_quantity": 50,
        "created_at": "2026-08-01T10:00:00Z",
        "updated_at": "2026-08-01T10:00:00Z"
      }
    ]
  }
  ```

---

### 2. `POST /api/v1/packaging-items`
Menambahkan bahan kemasan baru. Biaya per unit dihitung otomatis oleh sistem.

* **Contoh Request Body**:
  ```json
  {
    "name": "Kardus Diecut Sedang 15x15x5",
    "category": "BOX",
    "unit_type": "PCS",
    "purchase_price": 42500,
    "purchase_quantity": 50,
    "stock_quantity": 50
  }
  ```

---

### 3. `GET /api/v1/packaging-presets`
Mengambil seluruh bundel preset kemasan beserta komponen item penyusunnya dan kalkulasi otomatis `total_cost`.

* **Contoh Response (200 OK)**:
  ```json
  {
    "status": "success",
    "total": 2,
    "data": [
      {
        "id": "00caeb44-4409-4224-bfb2-88bae9e41268",
        "user_id": "4aecd5c5-c0a9-4f52-ba2f-b6f4272218b4",
        "name": "Keychain",
        "description": "Paket kemasan gantungan kunci",
        "total_cost": 1509,
        "items": [
          {
            "id": "4b721e90-2134-4bc1-9c60-e837f1912a7e",
            "packaging_item_id": "e277a177-cd84-4813-b8b9-9ebaca2a6d99",
            "item_name": "Plastik Clip Hologram 10x18",
            "category": "PLASTIC",
            "unit_type": "PCS",
            "quantity_used": 1,
            "unit_cost": 464,
            "subtotal_cost": 464
          },
          {
            "id": "8f812b10-2134-4bc1-9c60-e837f1912a7f",
            "packaging_item_id": "520e932a-f8ae-455f-aee1-fc017be5b7b8",
            "item_name": "Mailer Putih",
            "category": "PLASTIC",
            "unit_type": "PCS",
            "quantity_used": 1,
            "unit_cost": 895,
            "subtotal_cost": 895
          },
          {
            "id": "9a123c10-2134-4bc1-9c60-e837f1912a80",
            "packaging_item_id": "8cd8ebf2-b3a5-4951-a33d-be176f88b48f",
            "item_name": "Sticker Symetra",
            "category": "LABEL",
            "unit_type": "PCS",
            "quantity_used": 1,
            "unit_cost": 150,
            "subtotal_cost": 150
          }
        ],
        "created_at": "2026-08-01T10:00:00Z",
        "updated_at": "2026-08-01T10:00:00Z"
      }
    ]
  }
  ```

---

### 4. `POST /api/v1/packaging-presets`
Membuat preset bundel kemasan baru secara transaksional.

* **Contoh Request Body**:
  ```json
  {
    "name": "Box Miniatur Hemat",
    "description": "Dus diecut kecil + bubble wrap + sticker logo",
    "items": [
      {
        "packaging_item_id": "1aa4dcce-5599-43bd-9a5f-95148ed3fc0a",
        "quantity_used": 1
      },
      {
        "packaging_item_id": "8cd8ebf2-b3a5-4951-a33d-be176f88b48f",
        "quantity_used": 1
      }
    ]
  }
  ```

---

### 5. `DELETE /api/v1/packaging-presets/:id`
Menghapus preset kemasan beserta daftar item di dalamnya (cascade delete).




