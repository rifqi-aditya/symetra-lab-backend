package database

import (
	"fmt"
	"log"
	"os"
	"time"

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

	// 2. Konfigurasi logger GORM dengan ambang batas slow query 2 detik (menyesuaikan latensi cloud Supabase)
	customLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             2 * time.Second,
			LogLevel:                  logger.Warn,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)

	gormConfig := &gorm.Config{
		Logger: customLogger,
	}

	// 3. Buka koneksi ke PostgreSQL
	// Catatan: PreferSimpleProtocol: true WAJIB digunakan untuk Supabase Transaction Pooler (PgBouncer)
	log.Println("[INFO] Menghubungkan ke Supabase PostgreSQL...")
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  cfg.DatabaseURL,
		PreferSimpleProtocol: true,
	}), gormConfig)
	if err != nil {
		// %w (wrap) membungkus error asli agar jejak errornya tidak hilang
		return nil, fmt.Errorf("gagal terhubung ke database: %w", err)
	}

	// 4. AutoMigrate: Hanya jalankan DDL jika tabel belum ada atau jika diminta via AUTO_MIGRATE=true.
	// Ini membuat startup server instan (<1 detik) tanpa harus menunggu ratusan query DDL ke cloud Sydney.
	shouldMigrate := os.Getenv("AUTO_MIGRATE") == "true" || !db.Migrator().HasTable(&models.Product{})
	if shouldMigrate && os.Getenv("VERCEL") == "" {
		log.Println("[INFO] Memeriksa & memigrasi skema database...")
		if err := db.AutoMigrate(
			&models.Shop{},
			&models.ShopeeOrder{},
			&models.ShopeeOrderItem{},
			&models.ShopeeOrderEscrow{},
			&models.FilamentProfile{},
			&models.Filament{},
			&models.Machine{},
			&models.MachineMaintenancePart{},
			&models.Component{},
			&models.PackagingItem{},
			&models.PackagingPreset{},
			&models.PackagingPresetItem{},
			&models.ProductCategory{},
			&models.Product{},
			&models.ProductFilament{},
			&models.ProductComponent{},
			&models.ProductPackagingItem{},
			&models.ShopConfig{},
			&models.MarketplacePlatform{},
			&models.Order{},
			&models.OrderItem{},
			&models.OrderFilament{},
			&models.OrderComponent{},
		); err != nil {
			return nil, fmt.Errorf("gagal migrasi database: %w", err)
		}
		log.Println("[INFO] Migrasi skema database selesai")
	}

	log.Println("[INFO] Berhasil terhubung ke Supabase PostgreSQL & database siap")
	return db, nil
}
