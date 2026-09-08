package handler

import (
	"log"
	"net/http"
	"sync"

	"symetra-lab-backend/config"
	"symetra-lab-backend/database"
	"symetra-lab-backend/pkg/shopee"
	"symetra-lab-backend/routes"

	"github.com/gin-gonic/gin"
)

var (
	router *gin.Engine
	once   sync.Once
)

func initApp() {
	cfg := config.LoadConfig()

	db, err := database.InitDB(cfg)
	if err != nil {
		log.Printf("[WARN] Koneksi database belum aktif: %v", err)
	}

	shopeeClient := shopee.NewClient(
		cfg.ShopeePartnerID,
		cfg.ShopeePartnerKey,
		cfg.ShopeeIsProduction,
		cfg.ShopeeRedirectURL,
	)

	router = routes.SetupRouter(cfg, db, shopeeClient)
}

// Handler adalah fungsi gerbang utama untuk Vercel Serverless Function
func Handler(w http.ResponseWriter, r *http.Request) {
	// sync.Once memastikan inisialisasi hanya berjalan 1 kali saat serverless instance pertama kali bangun
	once.Do(initApp)
	router.ServeHTTP(w, r)
}
