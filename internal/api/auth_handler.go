// Copyright 2025 convexwf
//
// Project: uim-go
// File: auth_handler.go
// Email: convexwf@gmail.com
// Created: 2025-03-13
// Last modified: 2025-04-28
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// Description: HTTP handlers for authentication endpoints

package api

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/convexwf/uim-go/internal/service"
)

// AuthHandler handles authentication-related HTTP requests.
type AuthHandler struct {
	authService service.AuthService
}

// NewAuthHandler creates a new authentication handler.
func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// RegisterRequest represents a user registration request.
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// LoginRequest represents a user login request.
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// RefreshTokenRequest represents a token refresh request.
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// AuthResponse represents an authentication response with user and tokens.
type AuthResponse struct {
	User         interface{} `json:"user"`
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
}

// Register handles user registration.
//
// It validates the registration request, creates a new user account,
// and returns the user information along with access and refresh tokens.
//
// @Summary Register a new user
// @Description Register a new user with username, email, and password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Registration request"
// @Success 201 {object} AuthResponse
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Router /api/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[AUTH] register invalid request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	log.Printf("[AUTH] register attempt username=%s", req.Username)

	user, accessToken, refreshToken, err := h.authService.Register(req.Username, req.Email, req.Password)
	if err != nil {
		switch err {
		case service.ErrUserExists:
			log.Printf("[AUTH] register failed username=%s reason=user_exists", req.Username)
			c.JSON(http.StatusConflict, gin.H{"error": "user already exists"})
		case service.ErrInvalidInput:
			log.Printf("[AUTH] register failed username=%s reason=invalid_input %v", req.Username, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			log.Printf("[AUTH] register failed username=%s reason=internal %v", req.Username, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}
	log.Printf("[AUTH] register success username=%s", req.Username)

	c.JSON(http.StatusCreated, AuthResponse{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}

// Login handles user login.
//
// It validates the login credentials and returns the user information
// along with access and refresh tokens if authentication succeeds.
//
// @Summary Login user
// @Description Login with username and password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login request"
// @Success 200 {object} AuthResponse
// @Failure 401 {object} map[string]string
// @Router /api/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[AUTH] login invalid request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	log.Printf("[AUTH] login attempt username=%s", req.Username)

	user, accessToken, refreshToken, err := h.authService.Login(req.Username, req.Password)
	if err != nil {
		switch err {
		case service.ErrInvalidCredentials:
			log.Printf("[AUTH] login failed username=%s reason=invalid_credentials", req.Username)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		default:
			log.Printf("[AUTH] login failed username=%s reason=internal %v", req.Username, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}
	log.Printf("[AUTH] login success username=%s", req.Username)

	c.JSON(http.StatusOK, AuthResponse{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}

// RefreshToken handles token refresh.
//
// It validates the refresh token and returns new access and refresh tokens.
//
// @Summary Refresh access token
// @Description Refresh access token using refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RefreshTokenRequest true "Refresh token request"
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[AUTH] refresh invalid request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	log.Printf("[AUTH] refresh attempt")

	accessToken, refreshToken, err := h.authService.RefreshToken(req.RefreshToken)
	if err != nil {
		log.Printf("[AUTH] refresh failed reason=invalid_token")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
		return
	}
	log.Printf("[AUTH] refresh success")

	c.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}
