package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"

	"symetra-lab-backend-v2/config"
	"symetra-lab-backend-v2/internal/delivery/http/dto"
)

type AuthHandler struct {
	cfg *config.Config
}

func NewAuthHandler(cfg *config.Config) *AuthHandler {
	return &AuthHandler{cfg: cfg}
}

type supabaseTokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	User         struct {
		ID    string `json:"id"`
		Email string `json:"email"`
		Role  string `json:"role"`
	} `json:"user"`
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
	Message          string `json:"message"`
	Msg              string `json:"msg"`
}

func (h *AuthHandler) Login(c echo.Context) error {
	var req dto.LoginRequest
	if err := c.Bind(&req); err != nil || req.Email == "" || req.Password == "" {
		return c.JSON(http.StatusBadRequest, dto.Fail("Invalid credentials payload", "Email and password are required"))
	}

	endpoint := strings.TrimRight(h.cfg.SupabaseURL, "/") + "/auth/v1/token?grant_type=password"
	body, _ := json.Marshal(req)

	httpReq, err := http.NewRequestWithContext(c.Request().Context(), http.MethodPost, endpoint, bytes.NewBuffer(body))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.Fail("Auth request creation failed", err.Error()))
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("apikey", h.cfg.SupabaseAnonKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return c.JSON(http.StatusBadGateway, dto.Fail("Failed to reach auth provider", err.Error()))
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.Fail("Failed to read auth response", err.Error()))
	}

	var sbResp supabaseTokenResponse
	_ = json.Unmarshal(respBytes, &sbResp)

	if resp.StatusCode != http.StatusOK {
		msg := sbResp.ErrorDescription
		if msg == "" {
			msg = sbResp.Message
		}
		if msg == "" {
			msg = sbResp.Msg
		}
		if msg == "" {
			msg = "Invalid email or password"
		}
		return c.JSON(resp.StatusCode, dto.Fail("Authentication failed", msg))
	}

	return c.JSON(http.StatusOK, dto.Success(dto.AuthResponseData{
		AccessToken:  sbResp.AccessToken,
		RefreshToken: sbResp.RefreshToken,
		ExpiresIn:    sbResp.ExpiresIn,
		TokenType:    sbResp.TokenType,
		User: dto.AuthUserResponse{
			ID:    sbResp.User.ID,
			Email: sbResp.User.Email,
			Role:  sbResp.User.Role,
		},
	}))
}

func (h *AuthHandler) GetProfile(c echo.Context) error {
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		return c.JSON(http.StatusUnauthorized, dto.Fail("Authorization header required", "Missing token"))
	}

	endpoint := strings.TrimRight(h.cfg.SupabaseURL, "/") + "/auth/v1/user"
	httpReq, err := http.NewRequestWithContext(c.Request().Context(), http.MethodGet, endpoint, nil)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.Fail("Failed to create request", err.Error()))
	}

	httpReq.Header.Set("apikey", h.cfg.SupabaseAnonKey)
	httpReq.Header.Set("Authorization", authHeader)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return c.JSON(http.StatusBadGateway, dto.Fail("Failed to reach auth provider", err.Error()))
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.JSON(http.StatusUnauthorized, dto.Fail("Unauthorized", "Session expired or invalid token"))
	}

	var u struct {
		ID    string `json:"id"`
		Email string `json:"email"`
		Role  string `json:"role"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&u); err != nil {
		return c.JSON(http.StatusInternalServerError, dto.Fail("Failed to parse user data", err.Error()))
	}

	return c.JSON(http.StatusOK, dto.Success(dto.AuthUserResponse{
		ID:    u.ID,
		Email: u.Email,
		Role:  u.Role,
	}))
}

func (h *AuthHandler) RefreshToken(c echo.Context) error {
	var req dto.RefreshRequest
	if err := c.Bind(&req); err != nil || req.RefreshToken == "" {
		return c.JSON(http.StatusBadRequest, dto.Fail("Invalid payload", "refresh_token is required"))
	}

	endpoint := strings.TrimRight(h.cfg.SupabaseURL, "/") + "/auth/v1/token?grant_type=refresh_token"
	body, _ := json.Marshal(map[string]string{"refresh_token": req.RefreshToken})

	httpReq, err := http.NewRequestWithContext(c.Request().Context(), http.MethodPost, endpoint, bytes.NewBuffer(body))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.Fail("Request failed", err.Error()))
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("apikey", h.cfg.SupabaseAnonKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return c.JSON(http.StatusBadGateway, dto.Fail("Auth provider unreachable", err.Error()))
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.Fail("Failed to read response", err.Error()))
	}

	var sbResp supabaseTokenResponse
	_ = json.Unmarshal(respBytes, &sbResp)

	if resp.StatusCode != http.StatusOK {
		msg := sbResp.ErrorDescription
		if msg == "" {
			msg = "Session expired, please login again"
		}
		return c.JSON(resp.StatusCode, dto.Fail("Token refresh failed", msg))
	}

	return c.JSON(http.StatusOK, dto.Success(dto.AuthResponseData{
		AccessToken:  sbResp.AccessToken,
		RefreshToken: sbResp.RefreshToken,
		ExpiresIn:    sbResp.ExpiresIn,
		TokenType:    sbResp.TokenType,
		User: dto.AuthUserResponse{
			ID:    sbResp.User.ID,
			Email: sbResp.User.Email,
			Role:  sbResp.User.Role,
		},
	}))
}

func (h *AuthHandler) Logout(c echo.Context) error {
	return c.JSON(http.StatusOK, dto.Success(map[string]string{"message": "Logged out successfully"}))
}