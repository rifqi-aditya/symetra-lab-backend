package main

import (
	"log"

	"symetra-lab-backend/config"
	"symetra-lab-backend/database"
	"symetra-lab-backend/pkg/shopee"
	"symetra-lab-backend/routes"
)

func main() {
	log.Println("[INFO] Memulai server Symetra Lab Backend...")

	// 1. Muat konfigurasi
	cfg := config.LoadConfig()

	// 2. Hubungkan ke Supabase PostgreSQL
	db, err := database.InitDB(cfg)
	if err != nil {
		log.Printf("[WARN] Koneksi database belum aktif: %v", err)
		log.Println("[INFO] Masukkan DATABASE_URL di file .env jika ingin menyambungkan database")
	}

	// 3. Siapkan Client Shopee
	shopeeClient := shopee.NewClient(
		cfg.ShopeePartnerID,
		cfg.ShopeePartnerKey,
		cfg.ShopeeIsProduction,
		cfg.ShopeeRedirectURL,
	)

	// 4. Siapkan Router Gin
	r := routes.SetupRouter(cfg, db, shopeeClient)

	// 5. Jalankan server lokal
	log.Printf("[INFO] Server Symetra Lab Backend aktif di port :%s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Gagal menjalankan server: %v", err)
	}
}