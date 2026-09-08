package database

import (
	"fmt"
	"log"

	"symetra-lab-backend/config"
	"symetra-lab-backend/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// InitDB menghubungkan backend ke Supabase PostgreSQL dan melakukan migrasi skema
func InitDB(cfg *config.Config) (*gorm.DB, error) {
	// 1. Validasi apakah DATABASE_URL sudah diisi di .env
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL belum diatur di file .env")
	}

	// 2. Konfigurasi logger GORM agar menampilkan query SQL di terminal
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	// 3. Buka koneksi ke PostgreSQL
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), gormConfig)
	if err != nil {
		// %w (wrap) membungkus error asli agar jejak errornya tidak hilang
		return nil, fmt.Errorf("gagal terhubung ke database: %w", err)
	}

	// 4. AutoMigrate: GORM akan otomatis membuat tabel "shops" di Supabase jika belum ada
	if err := db.AutoMigrate(&models.Shop{}); err != nil {
		return nil, fmt.Errorf("gagal migrasi tabel shop: %w", err)
	}

	log.Println("[INFO] Berhasil terhubung ke Supabase PostgreSQL & migrasi selesai")
	return db, nil
}
