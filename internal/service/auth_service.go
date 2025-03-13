// Copyright 2025 convexwf
//
// Project: uim-go
// File: auth_service.go
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
// Description: Authentication service

package service

import (
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/convexwf/uim-go/internal/model"
	"github.com/convexwf/uim-go/internal/pkg/jwt"
	pwd "github.com/convexwf/uim-go/internal/pkg/password"
	"github.com/convexwf/uim-go/internal/repository"
)

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserExists         = errors.New("user already exists")
	ErrInvalidInput       = errors.New("invalid input")
)

type AuthService interface {
	Register(username, email, password string) (*model.User, string, string, error)
	Login(username, password string) (*model.User, string, string, error)
	RefreshToken(refreshToken string) (string, string, error)
}

type authService struct {
	userRepo   repository.UserRepository
	jwtManager *jwt.JWTManager
}

func NewAuthService(userRepo repository.UserRepository, jwtManager *jwt.JWTManager) AuthService {
	return &authService{
		userRepo:   userRepo,
		jwtManager: jwtManager,
	}
}

func (s *authService) Register(username, email, password string) (*model.User, string, string, error) {
	// Validate input
	if username == "" || email == "" || password == "" {
		return nil, "", "", ErrInvalidInput
	}

	if len(username) < 3 || len(username) > 50 {
		return nil, "", "", fmt.Errorf("%w: username must be between 3 and 50 characters", ErrInvalidInput)
	}

	if len(password) < 6 {
		return nil, "", "", fmt.Errorf("%w: password must be at least 6 characters", ErrInvalidInput)
	}

	// Check if user already exists
	_, err := s.userRepo.GetByUsername(username)
	if err == nil {
		return nil, "", "", ErrUserExists
	}

	_, err = s.userRepo.GetByEmail(email)
	if err == nil {
		return nil, "", "", ErrUserExists
	}

	// Hash password
	passwordHash, err := pwd.Hash(password)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &model.User{
		Username:     username,
		Email:        email,
		PasswordHash: passwordHash,
		DisplayName:  username,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, "", "", fmt.Errorf("failed to create user: %w", err)
	}

	// Generate tokens
	accessToken, err := s.jwtManager.GenerateAccessToken(user.UserID.String())
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.jwtManager.GenerateRefreshToken(user.UserID.String())
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return user, accessToken, refreshToken, nil
}

func (s *authService) Login(username, password string) (*model.User, string, string, error) {
	// Get user
	user, err := s.userRepo.GetByUsername(username)
	if err != nil {
		return nil, "", "", ErrInvalidCredentials
	}

	// Verify password
	if !pwd.Verify(password, user.PasswordHash) {
		return nil, "", "", ErrInvalidCredentials
	}

	// Generate tokens
	accessToken, err := s.jwtManager.GenerateAccessToken(user.UserID.String())
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.jwtManager.GenerateRefreshToken(user.UserID.String())
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return user, accessToken, refreshToken, nil
}

func (s *authService) RefreshToken(refreshToken string) (string, string, error) {
	// Validate refresh token
	claims, err := s.jwtManager.ValidateRefreshToken(refreshToken)
	if err != nil {
		return "", "", ErrInvalidCredentials
	}

	// Get user
	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return "", "", ErrInvalidCredentials
	}

	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return "", "", ErrUserNotFound
	}

	// Generate new tokens
	accessToken, err := s.jwtManager.GenerateAccessToken(user.UserID.String())
	if err != nil {
		return "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	newRefreshToken, err := s.jwtManager.GenerateRefreshToken(user.UserID.String())
	if err != nil {
		return "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return accessToken, newRefreshToken, nil
}
