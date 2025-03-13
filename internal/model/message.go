// Copyright 2025 convexwf
//
// Project: uim-go
// File: message.go
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
// Description: Message data model and database schema

package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MessageType represents the type of message.
type MessageType string

const (
	// MessageTypeText represents a text message.
	MessageTypeText MessageType = "text"
)

// Message represents a message in a conversation.
type Message struct {
	MessageID      int64          `gorm:"primaryKey;autoIncrement" json:"message_id"`
	ConversationID uuid.UUID      `gorm:"type:uuid;not null;index:idx_conversation_time" json:"conversation_id"`
	SenderID       uuid.UUID      `gorm:"type:uuid;not null;index" json:"sender_id"`
	Content        string         `gorm:"type:text;not null" json:"content"`
	MessageType    MessageType    `gorm:"type:varchar(20);default:'text'" json:"type"`
	CreatedAt      time.Time      `gorm:"index:idx_conversation_time" json:"created_at"`
	Metadata       string         `gorm:"type:jsonb" json:"metadata,omitempty"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Conversation Conversation `gorm:"foreignKey:ConversationID" json:"conversation,omitempty"`
	Sender       User         `gorm:"foreignKey:SenderID" json:"sender,omitempty"`
}

// TableName returns the database table name for the Message model.
func (Message) TableName() string {
	return "messages"
}
