package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"symetra-lab-backend/models"
	"symetra-lab-backend/pkg/shopee"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ShopeeHandler struct {
	DB           *gorm.DB
	ShopeeClient *shopee.Client
}

func NewShopeeHandler(db *gorm.DB, client *shopee.Client) *ShopeeHandler {
	return &ShopeeHandler{
		DB:           db,
		ShopeeClient: client,
	}
}

// GetAuthURL mengembalikan link otorisasi Shopee untuk diarahkan ke seller
// GET /api/v1/shopee/auth-url
func (h *ShopeeHandler) GetAuthURL(c *gin.Context) {
	if h.ShopeeClient.PartnerID == 0 || h.ShopeeClient.PartnerKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "SHOPEE_PARTNER_ID atau SHOPEE_PARTNER_KEY belum diatur di konfigurasi",
		})
		return
	}

	authURL, err := h.ShopeeClient.BuildAuthURL()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": fmt.Sprintf("Gagal membuat auth URL: %v", err),
		})
		return
	}

	// Jika client meminta redirect langsung via query parameter ?redirect=true
	if c.Query("redirect") == "true" {
		c.Redirect(http.StatusTemporaryRedirect, authURL)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   "success",
		"auth_url": authURL,
		"message":  "Buka URL ini di browser untuk mengotorisasi toko Shopee Anda",
	})
}

// HandleCallback menerima redirect dari Shopee setelah seller menyetujui izin
// GET /api/v1/shopee/callback?code=...&shop_id=...
func (h *ShopeeHandler) HandleCallback(c *gin.Context) {
	code := c.Query("code")
	shopIDStr := c.Query("shop_id")

	if code == "" || shopIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Parameter code atau shop_id tidak ditemukan pada callback Shopee",
		})
		return
	}

	shopID, err := strconv.ParseUint(shopIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Format shop_id tidak valid",
		})
		return
	}

	// 1. Tukar authorization code ke Shopee untuk mendapatkan access token & refresh token
	tokenResp, err := h.ShopeeClient.GetAccessToken(code, shopID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": fmt.Sprintf("Gagal menukar code ke access token: %v", err),
		})
		return
	}

	// 2. Hitung waktu kedaluwarsa token
	accessTokenExpiry := time.Now().Add(time.Duration(tokenResp.ExpireIn) * time.Second)
	refreshTokenExpiry := time.Now().Add(30 * 24 * time.Hour) // Refresh token berlaku 30 hari

	shopRecord := models.Shop{
		ShopID:                shopID,
		AccessToken:           tokenResp.AccessToken,
		RefreshToken:          tokenResp.RefreshToken,
		AccessTokenExpiresAt:  accessTokenExpiry,
		RefreshTokenExpiresAt: refreshTokenExpiry,
	}

	// 3. Simpan / perbarui ke database Supabase PostgreSQL
	if h.DB != nil {
		err = h.DB.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "shop_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"access_token", "refresh_token", "access_token_expires_at", "refresh_token_expires_at", "updated_at"}),
		}).Create(&shopRecord).Error

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": fmt.Sprintf("Berhasil mendapat token tapi gagal simpan ke database: %v", err),
			})
			return
		}
	}

	// 4. Beri respon HTML ramah untuk pemilik toko
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, `
		<!DOCTYPE html>
		<html>
		<head>
			<title>Otorisasi Shopee Berhasil</title>
			<style>
				body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; display: flex; justify-content: center; align-items: center; min-height: 100vh; margin: 0; background-color: #f0fdf4; }
				.card { background: white; padding: 2.5rem; border-radius: 12px; box-shadow: 0 4px 6px -1px rgba(0,0,0,0.1); max-width: 480px; text-align: center; }
				h1 { color: #166534; font-size: 1.5rem; margin-bottom: 0.5rem; }
				p { color: #4b5563; line-height: 1.5; }
				.badge { display: inline-block; background: #dcfce7; color: #15803d; padding: 0.25rem 0.75rem; border-radius: 9999px; font-weight: 600; font-size: 0.875rem; margin-top: 1rem; }
			</style>
		</head>
		<body>
			<div class="card">
				<h1>Toko Berhasil Terhubung!</h1>
				<p>Akun Shopee Anda (Shop ID: <strong>`+shopIDStr+`</strong>) telah berhasil diotorisasi dan tersambung ke sistem backend.</p>
				<div class="badge">Status: Aktif</div>
			</div>
		</body>
		</html>
	`)
}

// GetShops mengambil daftar toko yang tersambung beserta status kedaluwarsanya
// GET /api/v1/shopee/shops
func (h *ShopeeHandler) GetShops(c *gin.Context) {
	if h.DB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":  "error",
			"message": "Database belum terhubung",
		})
		return
	}

	var shops []models.Shop
	if err := h.DB.Order("created_at desc").Find(&shops).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": fmt.Sprintf("Gagal mengambil data toko: %v", err),
		})
		return
	}

	type ShopDTO struct {
		ShopID               uint64    `json:"shop_id"`
		ShopName             string    `json:"shop_name"`
		IsTokenExpired       bool      `json:"is_token_expired"`
		AccessTokenExpiresAt time.Time `json:"access_token_expires_at"`
		ConnectedAt          time.Time `json:"connected_at"`
	}

	results := make([]ShopDTO, len(shops))
	for i, s := range shops {
		results[i] = ShopDTO{
			ShopID:               s.ShopID,
			ShopName:             s.ShopName,
			IsTokenExpired:       s.IsTokenExpired(),
			AccessTokenExpiresAt: s.AccessTokenExpiresAt,
			ConnectedAt:          s.CreatedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"total":  len(results),
		"data":   results,
	})
}

// RefreshToken memicu pembaruan access token secara manual
// POST /api/v1/shopee/shops/:shop_id/refresh
func (h *ShopeeHandler) RefreshToken(c *gin.Context) {
	if h.DB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":  "error",
			"message": "Database belum terhubung",
		})
		return
	}

	shopIDStr := c.Param("shop_id")
	shopID, err := strconv.ParseUint(shopIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Format shop_id tidak valid",
		})
		return
	}

	var shop models.Shop
	if err := h.DB.Where("shop_id = ?", shopID).First(&shop).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "error",
			"message": "Toko dengan shop_id tersebut belum terdaftar",
		})
		return
	}

	tokenResp, err := h.ShopeeClient.RefreshAccessToken(shop.RefreshToken, shop.ShopID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": fmt.Sprintf("Gagal me-refresh token Shopee: %v", err),
		})
		return
	}

	shop.AccessToken = tokenResp.AccessToken
	shop.RefreshToken = tokenResp.RefreshToken
	shop.AccessTokenExpiresAt = time.Now().Add(time.Duration(tokenResp.ExpireIn) * time.Second)
	shop.RefreshTokenExpiresAt = time.Now().Add(30 * 24 * time.Hour)

	if err := h.DB.Save(&shop).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": fmt.Sprintf("Gagal memperbarui token di database: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":     "success",
		"message":    "Access token Shopee berhasil diperbarui",
		"expires_at": shop.AccessTokenExpiresAt,
	})
}
