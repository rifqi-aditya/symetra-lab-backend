package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	_ = godotenv.Load()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgresql://postgres.gllqnxgvllxkmrywrnbu:aDoTmR9nXSj1U7SC@aws-1-ap-southeast-2.pooler.supabase.com:6543/postgres?sslmode=require"
	}

	email := "admin@symetralab.com"
	newPassword := "admin12345"

	if len(os.Args) > 1 && os.Args[1] != "" {
		newPassword = os.Args[1]
	}
	if len(os.Args) > 2 && os.Args[2] != "" {
		email = os.Args[2]
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Gagal terhubung ke database: %v", err)
	}

	query := `
		UPDATE auth.users
		SET encrypted_password = crypt(?, gen_salt('bf')),
		    updated_at = NOW()
		WHERE email = ?;
	`
	res := db.Exec(query, newPassword, email)
	if res.Error != nil {
		log.Fatalf("Gagal mengubah password: %v", res.Error)
	}

	if res.RowsAffected == 0 {
		fmt.Printf("User dengan email '%s' tidak ditemukan di database.\n", email)
		return
	}

	fmt.Println("==================================================")
	fmt.Println("   PASSWORD BERHASIL DIRESET!")
	fmt.Println("==================================================")
	fmt.Printf("Email        : %s\n", email)
	fmt.Printf("Password Baru: %s\n", newPassword)
	fmt.Println("==================================================")
	fmt.Println("Silakan gunakan kredensial ini untuk login di frontend.")
}
