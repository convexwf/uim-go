// Copyright 2025 convexwf
//
// Project: uim-go
// File: auth_service_test.go
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
// Description: Unit tests for authentication service

package service

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/convexwf/uim-go/internal/model"
	"github.com/convexwf/uim-go/internal/pkg/jwt"
)

// mockUserRepository is a mock implementation of UserRepository for testing.
type mockUserRepository struct {
	users      map[uuid.UUID]*model.User
	usersByUsername map[string]*model.User
	usersByEmail    map[string]*model.User
}

func newMockUserRepository() *mockUserRepository {
	return &mockUserRepository{
		users:           make(map[uuid.UUID]*model.User),
		usersByUsername: make(map[string]*model.User),
		usersByEmail:    make(map[string]*model.User),
	}
}

func (m *mockUserRepository) Create(user *model.User) error {
	if _, exists := m.usersByUsername[user.Username]; exists {
		return errors.New("user already exists")
	}
	if _, exists := m.usersByEmail[user.Email]; exists {
		return errors.New("user already exists")
	}
	m.users[user.UserID] = user
	m.usersByUsername[user.Username] = user
	m.usersByEmail[user.Email] = user
	return nil
}

func (m *mockUserRepository) GetByID(userID uuid.UUID) (*model.User, error) {
	user, exists := m.users[userID]
	if !exists {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (m *mockUserRepository) GetByUsername(username string) (*model.User, error) {
	user, exists := m.usersByUsername[username]
	if !exists {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (m *mockUserRepository) GetByEmail(email string) (*model.User, error) {
	user, exists := m.usersByEmail[email]
	if !exists {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (m *mockUserRepository) Update(user *model.User) error {
	if _, exists := m.users[user.UserID]; !exists {
		return errors.New("user not found")
	}
	m.users[user.UserID] = user
	m.usersByUsername[user.Username] = user
	m.usersByEmail[user.Email] = user
	return nil
}

func (m *mockUserRepository) Delete(userID uuid.UUID) error {
	user, exists := m.users[userID]
	if !exists {
		return errors.New("user not found")
	}
	delete(m.users, userID)
	delete(m.usersByUsername, user.Username)
	delete(m.usersByEmail, user.Email)
	return nil
}

func TestAuthService_Register(t *testing.T) {
	userRepo := newMockUserRepository()
	jwtManager := jwt.NewJWTManager("test-secret", 15*time.Minute, 168*time.Hour)
	authService := NewAuthService(userRepo, jwtManager)

	user, accessToken, refreshToken, err := authService.Register("testuser", "test@example.com", "password123")
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if user == nil {
		t.Fatal("Register() returned nil user")
	}
	if user.Username != "testuser" {
		t.Errorf("Register() username = %v, want testuser", user.Username)
	}
	if accessToken == "" {
		t.Error("Register() returned empty access token")
	}
	if refreshToken == "" {
		t.Error("Register() returned empty refresh token")
	}
}

func TestAuthService_Register_DuplicateUsername(t *testing.T) {
	userRepo := newMockUserRepository()
	jwtManager := jwt.NewJWTManager("test-secret", 15*time.Minute, 168*time.Hour)
	authService := NewAuthService(userRepo, jwtManager)

	_, _, _, err := authService.Register("testuser", "test@example.com", "password123")
	if err != nil {
		t.Fatalf("First Register() error = %v", err)
	}

	_, _, _, err = authService.Register("testuser", "test2@example.com", "password123")
	if err != ErrUserExists {
		t.Errorf("Register() with duplicate username error = %v, want ErrUserExists", err)
	}
}

func TestAuthService_Register_DuplicateEmail(t *testing.T) {
	userRepo := newMockUserRepository()
	jwtManager := jwt.NewJWTManager("test-secret", 15*time.Minute, 168*time.Hour)
	authService := NewAuthService(userRepo, jwtManager)

	_, _, _, err := authService.Register("testuser", "test@example.com", "password123")
	if err != nil {
		t.Fatalf("First Register() error = %v", err)
	}

	_, _, _, err = authService.Register("testuser2", "test@example.com", "password123")
	if err != ErrUserExists {
		t.Errorf("Register() with duplicate email error = %v, want ErrUserExists", err)
	}
}

func TestAuthService_Register_InvalidInput(t *testing.T) {
	userRepo := newMockUserRepository()
	jwtManager := jwt.NewJWTManager("test-secret", 15*time.Minute, 168*time.Hour)
	authService := NewAuthService(userRepo, jwtManager)

	testCases := []struct {
		name     string
		username string
		email    string
		password string
	}{
		{"empty username", "", "test@example.com", "password123"},
		{"empty email", "testuser", "", "password123"},
		{"empty password", "testuser", "test@example.com", ""},
		{"short username", "ab", "test@example.com", "password123"},
		{"long username", string(make([]byte, 51)), "test@example.com", "password123"},
		{"short password", "testuser", "test@example.com", "12345"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, _, _, err := authService.Register(tc.username, tc.email, tc.password)
			if err != ErrInvalidInput && !errors.Is(err, ErrInvalidInput) {
				t.Errorf("Register() error = %v, want ErrInvalidInput or wrapped", err)
			}
		})
	}
}

func TestAuthService_Login(t *testing.T) {
	userRepo := newMockUserRepository()
	jwtManager := jwt.NewJWTManager("test-secret", 15*time.Minute, 168*time.Hour)
	authService := NewAuthService(userRepo, jwtManager)

	// Register first
	_, _, _, err := authService.Register("testuser", "test@example.com", "password123")
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	// Login
	user, accessToken, refreshToken, err := authService.Login("testuser", "password123")
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if user == nil {
		t.Fatal("Login() returned nil user")
	}
	if accessToken == "" {
		t.Error("Login() returned empty access token")
	}
	if refreshToken == "" {
		t.Error("Login() returned empty refresh token")
	}
}

func TestAuthService_Login_InvalidCredentials(t *testing.T) {
	userRepo := newMockUserRepository()
	jwtManager := jwt.NewJWTManager("test-secret", 15*time.Minute, 168*time.Hour)
	authService := NewAuthService(userRepo, jwtManager)

	// Register first
	_, _, _, err := authService.Register("testuser", "test@example.com", "password123")
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	testCases := []struct {
		name     string
		username string
		password string
	}{
		{"wrong username", "wronguser", "password123"},
		{"wrong password", "testuser", "wrongpassword"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, _, _, err := authService.Login(tc.username, tc.password)
			if err != ErrInvalidCredentials {
				t.Errorf("Login() error = %v, want ErrInvalidCredentials", err)
			}
		})
	}
}

func TestAuthService_RefreshToken(t *testing.T) {
	userRepo := newMockUserRepository()
	jwtManager := jwt.NewJWTManager("test-secret", 15*time.Minute, 168*time.Hour)
	authService := NewAuthService(userRepo, jwtManager)

	// Register and get refresh token
	_, _, refreshToken, err := authService.Register("testuser", "test@example.com", "password123")
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	// Refresh token
	accessToken, newRefreshToken, err := authService.RefreshToken(refreshToken)
	if err != nil {
		t.Fatalf("RefreshToken() error = %v", err)
	}
	if accessToken == "" {
		t.Error("RefreshToken() returned empty access token")
	}
	if newRefreshToken == "" {
		t.Error("RefreshToken() returned empty refresh token")
	}
}

func TestAuthService_RefreshToken_InvalidToken(t *testing.T) {
	userRepo := newMockUserRepository()
	jwtManager := jwt.NewJWTManager("test-secret", 15*time.Minute, 168*time.Hour)
	authService := NewAuthService(userRepo, jwtManager)

	_, _, err := authService.RefreshToken("invalid-token")
	if err != ErrInvalidCredentials {
		t.Errorf("RefreshToken() error = %v, want ErrInvalidCredentials", err)
	}
}

func TestNewAuthService(t *testing.T) {
	userRepo := newMockUserRepository()
	jwtManager := jwt.NewJWTManager("test-secret", 15*time.Minute, 168*time.Hour)

	authService := NewAuthService(userRepo, jwtManager)
	if authService == nil {
		t.Fatal("NewAuthService() returned nil")
	}
}
