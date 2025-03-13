// Copyright 2025 convexwf
//
// Project: uim-go
// File: conversation.go
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
// Description: Conversation and participant data models

package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ConversationType represents the type of conversation.
type ConversationType string

const (
	// ConversationTypeOneOnOne represents a one-on-one conversation between two users.
	ConversationTypeOneOnOne ConversationType = "one_on_one"
	// ConversationTypeGroup represents a group conversation with multiple participants.
	ConversationTypeGroup ConversationType = "group"
)

// Conversation represents a conversation (chat) in the system.
type Conversation struct {
	ConversationID uuid.UUID       `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"conversation_id"`
	Type           ConversationType `gorm:"type:varchar(20);not null" json:"type"`
	Name           string           `gorm:"type:varchar(255)" json:"name,omitempty"`
	CreatedBy      uuid.UUID        `gorm:"type:uuid;not null" json:"created_by"`
	CreatedAt      time.Time        `json:"created_at"`
	UpdatedAt      time.Time        `json:"updated_at"`
	DeletedAt      gorm.DeletedAt   `gorm:"index" json:"-"`

	// Relationships
	Participants []ConversationParticipant `gorm:"foreignKey:ConversationID" json:"participants,omitempty"`
	Messages     []Message                 `gorm:"foreignKey:ConversationID" json:"messages,omitempty"`
}

// TableName returns the database table name for the Conversation model.
func (Conversation) TableName() string {
	return "conversations"
}

// ConversationParticipant represents a user's participation in a conversation.
type ConversationParticipant struct {
	ConversationID    uuid.UUID `gorm:"type:uuid;primaryKey" json:"conversation_id"`
	UserID            uuid.UUID `gorm:"type:uuid;primaryKey" json:"user_id"`
	Role              string    `gorm:"type:varchar(20);default:'member'" json:"role"` // owner, admin, member
	JoinedAt          time.Time `json:"joined_at"`
	LastReadMessageID int64     `gorm:"type:bigint" json:"last_read_message_id"`
}

// TableName returns the database table name for the ConversationParticipant model.
func (ConversationParticipant) TableName() string {
	return "conversation_participants"
}
