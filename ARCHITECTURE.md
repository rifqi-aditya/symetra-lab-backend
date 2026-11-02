# Symetra Lab Backend — Architecture & System Design Documentation

Dokumen ini menjelaskan arsitektur sistem, struktur modular, aturan domain bisnis, dan panduan teknis backend **Symetra Lab** setelah refactoring Clean Architecture 100%.

---

## 1. Prinsip & Struktur Clean Architecture

Backend Symetra Lab mengadopsi prinsip **Clean Architecture** dan **Domain-Driven Design (DDD)** dengan batas lapisan (*layer boundaries*) yang tegas:

```
symetra-lab-backend/
├── cmd/
│   └── server/main.go            # Entrypoint utama aplikasi
├── config/                       # Konfigurasi aplikasi & parsing .env
├── internal/
│   ├── app/                      # Modular Dependency Injection Wiring
│   │   ├── app.go                # InitEchoApp (HTTP engine & routing)
│   │   ├── wire_product.go       # Wiring modul produk, material, mesin, kemasan
│   │   ├── wire_order.go         # Wiring modul pesanan workshop & manual
│   │   ├── wire_finance.go       # Wiring modul keuangan, transaksi, modal, pos kas
│   │   └── wire_marketplace.go   # Wiring modul marketplace (Shopee, Tokopedia/TikTok)
│   ├── delivery/http/            # Interface Adapters (Lapisan Transport HTTP)
│   │   ├── handler/              # HTTP Echo handlers (1 handler per domain)
│   │   ├── dto/                  # Request & Response Data Transfer Objects
│   │   ├── middleware/           # Auth JWT & validation middleware
│   │   └── router.go             # Registrasi endpoint API (/api/v1)
│   ├── domain/                   # Enterprise Core Business Rules (100% Pure Go)
│   │   ├── costing/              # Rumus HPP, kalkulator depresiasi & tarif listrik
│   │   ├── finance/              # Entitas kas, transaksi, pos kas, & pure allocator
│   │   ├── order/                # Entitas pesanan, kontrak StandardMarketplaceOrder
│   │   ├── product/              # Entitas master produk & BOM komponen
│   │   ├── shopee/               # Entitas toko & pesanan Shopee
│   │   └── shoptokopedia/        # Entitas toko & integrasi Shop Tokopedia
│   ├── repository/postgres/      # Infrastructure Persistence (GORM & Supabase PostgreSQL)
│   │   ├── finance_repo.go       # Query buku besar, transaksi, & pos kas
│   │   ├── order_repo.go         # Query pesanan, order items, & summary
│   │   ├── product_repo.go       # Query produk & relasi komponen/kemasan
│   │   └── shopee_repo.go        # Pure persistence pesanan Shopee (tanpa side-effects)
│   └── usecase/                  # Application Business Rules (Orkestrasi Alur Kerja)
│       ├── finance/              # UseCase pencatatan kas & sinkronisasi saldo buku besar
│       ├── order/                # UseCase pengelolaan pesanan workshop & manual
│       ├── product/              # UseCase produk, SKUMatcher, & CostCalculator
│       ├── shopee/               # UseCase sinkronisasi pesanan Shopee, penautan SKU
│       └── shoptokopedia/        # UseCase sinkronisasi Shop | Tokopedia / TikTok Shop
└── pkg/                          # Library SDK Eksternal Terisolasi
    ├── shopee/                   # Client SDK Shopee Open Platform v2
    └── shoptokopedia/            # Client SDK TikTok Shop / Shop Tokopedia Partner API
```

### Aturan Dependensi (The Dependency Rule)
* **Lapisan Domain** tidak boleh mengimpor package di luar domain (bebas dari GORM, Echo, HTTP, SQL).
* **Lapisan UseCase** hanya bergantung pada Domain dan interface Repository (Dependency Inversion).
* **Lapisan Repository** hanya menangani query basis data dan memetakan model database (GORM) ke Domain Entity.
* **Lapisan Delivery (Handler)** hanya menerima request HTTP, memvalidasi DTO, memanggil UseCase, dan mengembalikan response JSON seragam.

---

## 2. Formula 8 Pos Kas Workshop & Invarian Matematis

Sistem mengimplementasikan alokasi otomatis hasil penjualan (*Escrow Net Amount*) ke dalam **8 Pos Kas Khusus** menggunakan Domain Service murni:
[`finance.AllocateOrderFunds()`](file:///e:/Important/3D%20Printing/Software/symetra-lab-backend/internal/domain/finance/allocator.go#L36-L75).

### Rincian 8 Pos Kas
1. **Pos Kas Filamen (`FUND_FILAMENT`)**
   * Menyimpan modal pengadaan bahan baku murni (tanpa buffer kegagalan).
   * Rumus: $\text{Berat Cetak (gram)} \times \text{Tarif Bahan per Gram}$.
2. **Pos Kas Cadangan Risiko / Emergency (`FUND_EMERGENCY`)**
   * **Buffer Kegagalan Cetak:** $10\%$ dari biaya filamen murni yang digunakan (`FilamentFailureBufferPercent = 0.10`).
   * **Potongan Laba Bersih:** $5\%$ dari laba bersih jika pesanan menghasilkan laba positif (`NetProfitEmergencyCutPercent = 0.05`).
   * Rumus: $(10\% \times \text{FilamentCost}) + (5\% \times \text{PositiveNetProfit})$.
3. **Pos Kas Komponen Hardware (`FUND_COMPONENT`)**
   * Cadangan pembelian baut, bearing, magnet, insert brass, dan komponen mekanikal.
4. **Pos Kas Kemasan / Packaging (`FUND_PACKAGING`)**
   * Cadangan pembelian kardus, bubble wrap, lakban, dan stiker packing.
5. **Pos Kas Listrik & Daya (`FUND_ELECTRICITY`)**
   * Cadangan pembayaran tagihan PLN berbasis daya rata-rata printer (Watt) dan jam cetak.
6. **Pos Kas Pemeliharaan Mesin (`FUND_MAINTENANCE`)**
   * Cadangan servis printer, penggantian nozzle, belt, PTFE tube, dan pelumas.
7. **Pos Kas Depresiasi Mesin (`FUND_DEPRECIATION`)**
   * Tabungan pembelian unit 3D printer baru saat masa pakai printer lama habis.
8. **Pos Kas Laba Bersih (`FUND_NET_PROFIT`)**
   * Keuntungan murni yang menjadi hak workshop setelah disisihkan 5% ke Kas Cadangan Risiko.
   * Rumus: $\text{RawNetProfit} - (5\% \times \text{PositiveNetProfit})$.

### Pembuktian Invarian Matematis
Setiap kali pesanan diproses, sistem menjamin tidak ada uang yang hilang atau selisih Rp 1 pun:
$$\sum_{i=1}^{8} \text{Fund}_i = \text{NetAmount} = \text{GrossAmount} - \text{MarketplaceFees}$$

---

## 3. Pola Sinkronisasi Buku Besar Kas (Decoupled Ledger Sync)

### Masalah Desain Sebelumnya
Sebelumnya, pemanggilan fungsi `order_repo.Create()`, `UpdateStatus()`, atau `Delete()` secara otomatis menjalankan query SQL raw `SyncCashAccountLedger` di latar belakang (*hidden side-effect*). Hal ini membuat repository tidak testable secara unit test murni dan sulit dilacak.

### Solusi Bersih Saat Ini
1. Repository hanya melakukan CRUD murni pada tabel database.
2. Dibuat interface [`LedgerSynchronizer`](file:///e:/Important/3D%20Printing/Software/symetra-lab-backend/internal/usecase/order/manage_orders.go#L20-L24):
   ```go
   type LedgerSynchronizer interface {
       SyncRunningBalances(ctx context.Context) error
   }
   ```
3. Sinkronisasi saldo buku besar kas kini dipanggil secara eksplisit oleh UseCase terkait pada akhir transaksi:
   * Pembuatan/update status pesanan (`usecase/order`)
   * Pemrosesan pesanan marketplace (`usecase/shopee` & `usecase/shoptokopedia`)
   * Pencatatan transaksi manual, purchase order, dan modal (`usecase/finance`)

---

## 4. Engine Pencocokan SKU Multi-Marketplace (`SKUMatcher`)

Saat pesanan masuk dari marketplace (Shopee, Tokopedia, TikTok Shop), item pesanan belum tentu langsung terhubung ke produk internal workshop.

Sistem menggunakan [`SKUMatcher`](file:///e:/Important/3D%20Printing/Software/symetra-lab-backend/internal/usecase/product/sku_matcher.go#L40-L100) dengan strategi 3-tingkat:
1. **Pemeriksaan ID Produk Eksisting:** Jika item pesanan sudah pernah dipetakan sebelumnya, gunakan binding yang ada.
2. **Pencocokan Exact SKU:**
   * Diambil kandidat SKU: `ModelSKU` (variasi warna/ukuran) diutamakan dibanding `ItemSKU`.
   * Dicocokkan ke tabel master produk secara *case-insensitive* dan *trimmed*:
     `UPPER(TRIM(sku)) = UPPER(TRIM(?))`
3. **Pencocokan Riwayat Kanal (Fallback):**
   * Jika SKU kosong atau tidak cocok, sistem memeriksa riwayat tabel `order_items` untuk `channel_item_id` dan `channel_model_id` yang pernah ditautkan secara manual sebelumnya via interface `ChannelMappingFinder`.
4. **Kalkulasi Biaya Terpusat (`CostCalculator`):**
   * Jika produk ditemukan (`MATCHED`), biaya dihitung seketika menggunakan context-cache memory (harga filamen, komponen, kemasan, tarif listrik, dan depresiasi mesin) tanpa ratusan query berulang.

---

## 5. Arsitektur Adaptor Multi-Marketplace Terpadu

Untuk mencegah duplikasi kode saat mengintegrasikan marketplace baru, sistem menggunakan kontrak data universal:

### Kontrak Universal: `StandardMarketplaceOrder`
Semua adapter marketplace wajib menormalisasi datanya ke [`StandardMarketplaceOrder`](file:///e:/Important/3D%20Printing/Software/symetra-lab-backend/internal/domain/order/marketplace.go#L60-L80) di domain pesanan:
* `Channel`: `"SHOPEE"`, `"TOKOPEDIA"`, `"TIKTOK"`, `"MANUAL"`
* `OrderSN`: Nomor pesanan dari platform
* `OrderStatus`: Dinormalisasi via [`NormalizeMarketplaceStatus()`](file:///e:/Important/3D%20Printing/Software/symetra-lab-backend/internal/domain/order/marketplace.go#L82-L125) (`UNPAID`, `TO_SHIP`, `SHIPPED`, `COMPLETED`, `CANCELLED`)
* `Items`: Daftar item dengan kandidat SKU dan harga jual
* `Financials`: Rincian omzet kotor, total potongan platform, dan dana escrow bersih

### Integrasi TikTok Shop & Tokopedia (*Shop | Tokopedia*)
Sesuai merger platform, TikTok Shop dan Tokopedia menggunakan satu Open Platform backend: **TikTok Shop Partner API**.
* **Client SDK:** [`pkg/shoptokopedia/client.go`](file:///e:/Important/3D%20Printing/Software/symetra-lab-backend/pkg/shoptokopedia/client.go) menangani penandatanganan request HMAC-SHA256, OAuth token exchange, dan refresh token otomatis.
* **Order Parser:** [`pkg/shoptokopedia/orders.go`](file:///e:/Important/3D%20Printing/Software/symetra-lab-backend/pkg/shoptokopedia/orders.go) mengurai JSON payload TikTok Shop ke `StandardMarketplaceOrder`.
* **UseCase:** [`manage_shoptokopedia.go`](file:///e:/Important/3D%20Printing/Software/symetra-lab-backend/internal/usecase/shoptokopedia/manage_shoptokopedia.go) siap digunakan saat kredensial Partner Center dihubungkan.

---

## 6. Modular Dependency Injection Wiring

Fungsi `InitEchoApp()` di [`internal/app/app.go`](file:///e:/Important/3D%20Printing/Software/symetra-lab-backend/internal/app/app.go) kini ramping (108 baris) karena wiring didelegasikan ke 4 file modul terpisah:

| File Wire | Tanggung Jawab Modul |
| :--- | :--- |
| **`wire_product.go`** | Menginisialisasi repositori & usecase untuk Produk, Kategori, Bahan Filamen, Mesin 3D Printer, Komponen Mekanikal, Kemasan Packing, dan Konfigurasi Toko. |
| **`wire_order.go`** | Menginisialisasi usecase pesanan workshop manual, transisi status, dan histori pesanan. |
| **`wire_finance.go`** | Menginisialisasi transaksi keuangan, alokasi 8 pos kas, mutasi modal, purchase order, dan dasbor analitik. |
| **`wire_marketplace.go`** | Menginisialisasi klien Shopee, Shop Tokopedia, kalkulator biaya, matcher SKU, dan usecase pemrosesan pesanan marketplace. |

---

## 7. Menjalankan Server & Menjalankan Pengujian

```bash
# Masuk ke direktori backend
cd symetra-lab-backend

# Menjalankan seluruh rangkaian unit test
go test ./...

# Menjalankan unit test dengan verbose mode
go test -v ./internal/domain/... ./internal/usecase/...

# Mengompilasi binary server
go build ./cmd/server

# Menjalankan server secara lokal
go run cmd/server/main.go
```

