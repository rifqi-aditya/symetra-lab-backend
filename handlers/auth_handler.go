package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"symetra-lab-backend/config"

	"github.com/gin-gonic/gin"
)

// AuthHandler menangani operasi otentikasi (login, me, logout) via Supabase Auth REST API
type AuthHandler struct {
	cfg *config.Config
}

func NewAuthHandler(cfg *config.Config) *AuthHandler {
	return &AuthHandler{cfg: cfg}
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type SupabaseAuthResponse struct {
	AccessToken  string       `json:"access_token"`
	TokenType    string       `json:"token_type"`
	ExpiresIn    int          `json:"expires_in"`
	RefreshToken string       `json:"refresh_token"`
	User         SupabaseUser `json:"user"`
	Error        string       `json:"error,omitempty"`
	ErrorDesc    string       `json:"error_description,omitempty"`
	Message      string       `json:"message,omitempty"`
	Msg          string       `json:"msg,omitempty"`
}

type SupabaseUser struct {
	ID        string    `json:"id"`
	Aud       string    `json:"aud"`
	Role      string    `json:"role"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Login melakukan verifikasi kredensial ke Supabase Auth
// POST /api/v1/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Email dan password wajib diisi secara valid",
		})
		return
	}

	supabaseURL := strings.TrimRight(h.cfg.SupabaseURL, "/")
	endpoint := supabaseURL + "/auth/v1/token?grant_type=password"

	reqBody, _ := json.Marshal(req)
	httpReq, err := http.NewRequestWithContext(c.Request.Context(), http.MethodPost, endpoint, bytes.NewBuffer(reqBody))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Gagal membuat request auth: " + err.Error()})
		return
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("apikey", h.cfg.SupabaseAnonKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"status":  "error",
			"message": "Gagal terhubung ke layanan otentikasi: " + err.Error(),
		})
		return
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Gagal membaca response auth"})
		return
	}

	if resp.StatusCode != http.StatusOK {
		var errResp SupabaseAuthResponse
		_ = json.Unmarshal(respBody, &errResp)
		msg := errResp.ErrorDesc
		if msg == "" {
			msg = errResp.Message
		}
		if msg == "" {
			msg = errResp.Msg
		}
		if msg == "" {
			msg = "Email atau kata sandi tidak valid"
		}
		c.JSON(resp.StatusCode, gin.H{
			"status":  "error",
			"message": msg,
		})
		return
	}

	var authResp SupabaseAuthResponse
	if err := json.Unmarshal(respBody, &authResp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Format data auth tidak valid"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"access_token":  authResp.AccessToken,
			"refresh_token": authResp.RefreshToken,
			"expires_in":    authResp.ExpiresIn,
			"token_type":    authResp.TokenType,
			"user": gin.H{
				"id":    authResp.User.ID,
				"email": authResp.User.Email,
				"role":  authResp.User.Role,
			},
		},
	})
}

// GetProfile mengambil data user dari token Authorization
// GET /api/v1/auth/me
func (h *AuthHandler) GetProfile(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"status": "error", "message": "Token otorisasi diperlukan"})
		return
	}

	supabaseURL := strings.TrimRight(h.cfg.SupabaseURL, "/")
	endpoint := supabaseURL + "/auth/v1/user"

	httpReq, err := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, endpoint, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Gagal membuat request profil: " + err.Error()})
		return
	}

	httpReq.Header.Set("apikey", h.cfg.SupabaseAnonKey)
	httpReq.Header.Set("Authorization", authHeader)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": "Gagal memverifikasi token: " + err.Error()})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusUnauthorized, gin.H{"status": "error", "message": "Sesi telah kedaluwarsa atau token tidak valid"})
		return
	}

	var user SupabaseUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Gagal membaca data pengguna"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"id":    user.ID,
			"email": user.Email,
			"role":  user.Role,
		},
	})
}

// Logout membersihkan sesi di server jika diperlukan
// POST /api/v1/auth/logout
func (h *AuthHandler) Logout(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Berhasil logout",
	})
}
