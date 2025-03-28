// Copyright 2025 convexwf
//
// Project: uim-go
// File: auth_test.go
// Description: Integration tests for auth endpoints. Requires DB (run init_db.sh first).
// Skip when -short or when DB is not available.
package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/convexwf/uim-go/internal/api"
	"github.com/convexwf/uim-go/internal/config"
	"github.com/joho/godotenv"
	"github.com/convexwf/uim-go/internal/middleware"
	"github.com/convexwf/uim-go/internal/pkg/jwt"
	"github.com/convexwf/uim-go/internal/repository"
	"github.com/convexwf/uim-go/internal/service"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestAuthEndpoints_RegisterLoginRefresh(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping auth integration test in short mode")
	}
	// Load .env from project root so we use the same DB as seed_db (go test cwd is tests/integration)
	_ = godotenv.Load("../../.env")
	cfg, err := config.Load()
	if err != nil {
		t.Skipf("config not loadable: %v", err)
	}
	db, err := gorm.Open(postgres.Open(cfg.Database.DSN()), &gorm.Config{})
	if err != nil {
		t.Skipf("DB not available: %v", err)
	}
	jwtMgr := jwt.NewJWTManager(cfg.JWT.Secret, cfg.JWT.AccessExpiry, cfg.JWT.RefreshExpiry)
	userRepo := repository.NewUserRepository(db)
	authSvc := service.NewAuthService(userRepo, jwtMgr)
	router := api.SetupRouter(db, authSvc)
	router.Use(middleware.CORSMiddleware(cfg))
	router.Use(middleware.LoggerMiddlewareSimple())
	router.Use(middleware.ErrorHandlerMiddleware())

	// Register (idempotent: if user exists, we'll login instead)
	regBody := map[string]string{
		"username": "inttest_user",
		"email":    "inttest@example.com",
		"password": "password123",
	}
	regJSON, _ := json.Marshal(regBody)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(regJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	var regResp api.AuthResponse
	_ = json.NewDecoder(w.Body).Decode(&regResp)
	// Accept 201/200 (created) or 4xx (e.g. conflict if already exists)
	if w.Code != http.StatusCreated && w.Code != http.StatusOK {
		// User may already exist from previous run; proceed to login
		if w.Code < 400 || w.Code >= 500 {
			t.Fatalf("register: unexpected status %d, body %s", w.Code, w.Body.String())
		}
	}

	// Login (works for newly registered or existing user)
	loginBody := map[string]string{"username": "inttest_user", "password": "password123"}
	loginJSON, _ := json.Marshal(loginBody)
	req2 := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(loginJSON))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("login: status %d, body %s", w2.Code, w2.Body.String())
	}
	var loginResp api.AuthResponse
	if err := json.NewDecoder(w2.Body).Decode(&loginResp); err != nil {
		t.Fatalf("login response decode: %v", err)
	}
	if loginResp.AccessToken == "" || loginResp.RefreshToken == "" {
		t.Fatalf("login: missing tokens")
	}

	// Refresh
	refreshBody := map[string]string{"refresh_token": loginResp.RefreshToken}
	refreshJSON, _ := json.Marshal(refreshBody)
	req3 := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", bytes.NewReader(refreshJSON))
	req3.Header.Set("Content-Type", "application/json")
	w3 := httptest.NewRecorder()
	router.ServeHTTP(w3, req3)
	if w3.Code != http.StatusOK {
		t.Fatalf("refresh: status %d, body %s", w3.Code, w3.Body.String())
	}
	var refreshResp api.AuthResponse
	if err := json.NewDecoder(w3.Body).Decode(&refreshResp); err != nil {
		t.Fatalf("refresh response decode: %v", err)
	}
	if refreshResp.AccessToken == "" || refreshResp.RefreshToken == "" {
		t.Fatalf("refresh: missing tokens")
	}
}
