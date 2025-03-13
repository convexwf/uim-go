// Copyright 2025 convexwf
//
// Project: uim-go
// File: jwt_test.go
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
// Description: Unit tests for JWT token generation and validation

package jwt

import (
	"testing"
	"time"
)

func TestJWTManager_GenerateAccessToken(t *testing.T) {
	manager := NewJWTManager("test-secret", 15*time.Minute, 168*time.Hour)
	userID := "test-user-id"

	token, err := manager.GenerateAccessToken(userID)
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}
	if token == "" {
		t.Error("GenerateAccessToken() returned empty token")
	}

	// Validate token
	claims, err := manager.ValidateAccessToken(token)
	if err != nil {
		t.Fatalf("ValidateAccessToken() error = %v", err)
	}
	if claims.UserID != userID {
		t.Errorf("ValidateAccessToken() userID = %v, want %v", claims.UserID, userID)
	}
	if claims.Type != "access" {
		t.Errorf("ValidateAccessToken() type = %v, want access", claims.Type)
	}
}

func TestJWTManager_GenerateRefreshToken(t *testing.T) {
	manager := NewJWTManager("test-secret", 15*time.Minute, 168*time.Hour)
	userID := "test-user-id"

	token, err := manager.GenerateRefreshToken(userID)
	if err != nil {
		t.Fatalf("GenerateRefreshToken() error = %v", err)
	}
	if token == "" {
		t.Error("GenerateRefreshToken() returned empty token")
	}

	// Validate token
	claims, err := manager.ValidateRefreshToken(token)
	if err != nil {
		t.Fatalf("ValidateRefreshToken() error = %v", err)
	}
	if claims.UserID != userID {
		t.Errorf("ValidateRefreshToken() userID = %v, want %v", claims.UserID, userID)
	}
	if claims.Type != "refresh" {
		t.Errorf("ValidateRefreshToken() type = %v, want refresh", claims.Type)
	}
}

func TestJWTManager_InvalidToken(t *testing.T) {
	manager := NewJWTManager("test-secret", 15*time.Minute, 168*time.Hour)

	_, err := manager.ValidateAccessToken("invalid-token")
	if err == nil {
		t.Error("ValidateAccessToken() should return error for invalid token")
	}
}

func TestJWTManager_WrongTokenType(t *testing.T) {
	manager := NewJWTManager("test-secret", 15*time.Minute, 168*time.Hour)
	userID := "test-user-id"

	refreshToken, _ := manager.GenerateRefreshToken(userID)

	// Try to validate refresh token as access token
	_, err := manager.ValidateAccessToken(refreshToken)
	if err == nil {
		t.Error("ValidateAccessToken() should return error for refresh token")
	}
}

func TestJWTManager_AccessTokenAsRefreshToken(t *testing.T) {
	manager := NewJWTManager("test-secret", 15*time.Minute, 168*time.Hour)
	userID := "test-user-id"

	accessToken, _ := manager.GenerateAccessToken(userID)

	// Try to validate access token as refresh token
	_, err := manager.ValidateRefreshToken(accessToken)
	if err == nil {
		t.Error("ValidateRefreshToken() should return error for access token")
	}
}

func TestJWTManager_DifferentSecrets(t *testing.T) {
	manager1 := NewJWTManager("secret1", 15*time.Minute, 168*time.Hour)
	manager2 := NewJWTManager("secret2", 15*time.Minute, 168*time.Hour)
	userID := "test-user-id"

	token, _ := manager1.GenerateAccessToken(userID)

	// Token signed with secret1 should not validate with secret2
	_, err := manager2.ValidateAccessToken(token)
	if err == nil {
		t.Error("ValidateAccessToken() should return error for token signed with different secret")
	}
}

func TestJWTManager_EmptyUserID(t *testing.T) {
	manager := NewJWTManager("test-secret", 15*time.Minute, 168*time.Hour)

	token, err := manager.GenerateAccessToken("")
	if err != nil {
		t.Fatalf("GenerateAccessToken() with empty userID should not error, got: %v", err)
	}

	claims, err := manager.ValidateAccessToken(token)
	if err != nil {
		t.Fatalf("ValidateAccessToken() error = %v", err)
	}
	if claims.UserID != "" {
		t.Errorf("ValidateAccessToken() userID = %v, want empty string", claims.UserID)
	}
}

func TestJWTManager_ExpiredToken(t *testing.T) {
	manager := NewJWTManager("test-secret", -1*time.Hour, 168*time.Hour) // Already expired
	userID := "test-user-id"

	token, err := manager.GenerateAccessToken(userID)
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}

	// Wait a bit to ensure token is expired
	time.Sleep(100 * time.Millisecond)

	_, err = manager.ValidateAccessToken(token)
	if err == nil {
		t.Error("ValidateAccessToken() should return error for expired token")
	}
}

func TestNewJWTManager(t *testing.T) {
	secret := "test-secret"
	accessExpiry := 15 * time.Minute
	refreshExpiry := 168 * time.Hour

	manager := NewJWTManager(secret, accessExpiry, refreshExpiry)
	if manager == nil {
		t.Fatal("NewJWTManager() returned nil")
	}
	if len(manager.secret) == 0 {
		t.Error("NewJWTManager() secret is empty")
	}
	if manager.accessExpiry != accessExpiry {
		t.Errorf("NewJWTManager() accessExpiry = %v, want %v", manager.accessExpiry, accessExpiry)
	}
	if manager.refreshExpiry != refreshExpiry {
		t.Errorf("NewJWTManager() refreshExpiry = %v, want %v", manager.refreshExpiry, refreshExpiry)
	}
}
