// Copyright 2025 convexwf
//
// Project: uim-go
// File: contact.go
// Email: convexwf@gmail.com
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// See the License for the full terms.
//
// Description: Contact data model

package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserContact represents a one-way contact relation: owner_user_id -> contact_user_id.
type UserContact struct {
	OwnerUserID   uuid.UUID      `gorm:"type:uuid;primaryKey" json:"owner_user_id"`
	ContactUserID uuid.UUID      `gorm:"type:uuid;primaryKey" json:"contact_user_id"`
	Source        string         `gorm:"type:varchar(32);not null;default:'manual'" json:"source"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index:idx_user_contacts_deleted_at" json:"-"`
}

// TableName returns the database table name for the UserContact model.
func (UserContact) TableName() string {
	return "user_contacts"
}
