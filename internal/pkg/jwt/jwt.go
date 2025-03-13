// Copyright 2025 convexwf
//
// Project: uim-go
// File: jwt.go
// Email: convexwf@gmail.com
// Created: 2025-03-13
// Last modified: 2025-03-13
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
// Description: JWT token generation and validation utilities

package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims represents the JWT claims structure.
type Claims struct {
	UserID string `json:"user_id"`
	Type   string `json:"type"` // "access" or "refresh"
	jwt.RegisteredClaims
}

// JWTManager manages JWT token operations including generation and validation.
type JWTManager struct {
	secret        []byte
	accessExpiry  time.Duration
	refreshExpiry time.Duration
}

// NewJWTManager creates a new JWT manager with the specified configuration.
//
// Parameters:
//   - secret: The secret key used for signing tokens
//   - accessExpiry: The expiration time for access tokens
//   - refreshExpiry: The expiration time for refresh tokens
//
// Returns:
//   - *JWTManager: A new JWT manager instance
func NewJWTManager(secret string, accessExpiry, refreshExpiry time.Duration) *JWTManager {
	return &JWTManager{
		secret:        []byte(secret),
		accessExpiry:  accessExpiry,
		refreshExpiry: refreshExpiry,
	}
}

// GenerateAccessToken generates a JWT access token for the given user ID.
//
// The access token has a shorter expiration time and is used for
// authenticating API requests.
//
// Parameters:
//   - userID: The unique identifier of the user
//
// Returns:
//   - string: The signed JWT access token
//   - error: An error if token generation fails
func (m *JWTManager) GenerateAccessToken(userID string) (string, error) {
	claims := &Claims{
		UserID: userID,
		Type:   "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.accessExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

// GenerateRefreshToken generates a JWT refresh token for the given user ID.
//
// The refresh token has a longer expiration time and is used to
// obtain new access tokens without requiring user credentials.
//
// Parameters:
//   - userID: The unique identifier of the user
//
// Returns:
//   - string: The signed JWT refresh token
//   - error: An error if token generation fails
func (m *JWTManager) GenerateRefreshToken(userID string) (string, error) {
	claims := &Claims{
		UserID: userID,
		Type:   "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.refreshExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

// ValidateToken validates a JWT token and returns its claims.
//
// This is a generic validation function that checks the token signature
// and expiration. It does not check the token type.
//
// Parameters:
//   - tokenString: The JWT token string to validate
//
// Returns:
//   - *Claims: The token claims if valid
//   - error: An error if validation fails
func (m *JWTManager) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return m.secret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// ValidateAccessToken validates a JWT access token.
//
// This function validates the token and ensures it is specifically
// an access token (not a refresh token).
//
// Parameters:
//   - tokenString: The JWT access token string to validate
//
// Returns:
//   - *Claims: The token claims if valid
//   - error: An error if validation fails or token is not an access token
func (m *JWTManager) ValidateAccessToken(tokenString string) (*Claims, error) {
	claims, err := m.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	if claims.Type != "access" {
		return nil, errors.New("token is not an access token")
	}

	return claims, nil
}

// ValidateRefreshToken validates a JWT refresh token.
//
// This function validates the token and ensures it is specifically
// a refresh token (not an access token).
//
// Parameters:
//   - tokenString: The JWT refresh token string to validate
//
// Returns:
//   - *Claims: The token claims if valid
//   - error: An error if validation fails or token is not a refresh token
func (m *JWTManager) ValidateRefreshToken(tokenString string) (*Claims, error) {
	claims, err := m.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	if claims.Type != "refresh" {
		return nil, errors.New("token is not a refresh token")
	}

	return claims, nil
}
