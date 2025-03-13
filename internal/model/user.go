// Copyright 2025 convexwf
//
// Project: uim-go
// File: user.go
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
// Description: User data model and database schema

package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User represents a user in the system.
type User struct {
	UserID       uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"user_id"`
	Username     string         `gorm:"type:varchar(50);uniqueIndex;not null" json:"username"`
	Email        string         `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	PasswordHash string         `gorm:"type:varchar(255);not null" json:"-"`
	DisplayName  string         `gorm:"type:varchar(100)" json:"display_name"`
	AvatarURL    string         `gorm:"type:text" json:"avatar_url"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName returns the database table name for the User model.
func (User) TableName() string {
	return "users"
}
