// Copyright 2025 convexwf
//
// Project: uim-go
// File: user_repository.go
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
// Description: User repository for database operations

package repository

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/convexwf/uim-go/internal/model"
)

// UserRepository defines the interface for user data access operations.
type UserRepository interface {
	Create(user *model.User) error
	GetByID(userID uuid.UUID) (*model.User, error)
	GetByUsername(username string) (*model.User, error)
	GetByEmail(email string) (*model.User, error)
	Update(user *model.User) error
	Delete(userID uuid.UUID) error
}

type userRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new user repository instance.
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

// Create creates a new user in the database.
func (r *userRepository) Create(user *model.User) error {
	return r.db.Create(user).Error
}

// GetByID retrieves a user by their unique identifier.
func (r *userRepository) GetByID(userID uuid.UUID) (*model.User, error) {
	var user model.User
	err := r.db.Where("user_id = ?", userID).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByUsername retrieves a user by their username.
func (r *userRepository) GetByUsername(username string) (*model.User, error) {
	var user model.User
	err := r.db.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByEmail retrieves a user by their email address.
func (r *userRepository) GetByEmail(email string) (*model.User, error) {
	var user model.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Update updates an existing user in the database.
func (r *userRepository) Update(user *model.User) error {
	return r.db.Save(user).Error
}

// Delete soft deletes a user from the database.
func (r *userRepository) Delete(userID uuid.UUID) error {
	return r.db.Delete(&model.User{}, "user_id = ?", userID).Error
}
