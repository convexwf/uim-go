// Copyright 2025 convexwf
//
// Project: uim-go
// File: message_repository.go
// Email: convexwf@gmail.com
// Created: 2025-04-12
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// See the License for the full terms.
//
// Description: Message repository for database operations

package repository

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/convexwf/uim-go/internal/model"
)

// MessageRepository defines the interface for message data access.
type MessageRepository interface {
	Create(msg *model.Message) error
	ListByConversationID(conversationID uuid.UUID, limit, offset int, beforeID *int64) ([]*model.Message, error)
	GetByID(messageID int64) (*model.Message, error)
}

type messageRepository struct {
	db *gorm.DB
}

// NewMessageRepository creates a new message repository.
func NewMessageRepository(db *gorm.DB) MessageRepository {
	return &messageRepository{db: db}
}

// Create creates a new message.
func (r *messageRepository) Create(msg *model.Message) error {
	return r.db.Create(msg).Error
}

// ListByConversationID lists messages in a conversation, newest first.
// If beforeID is set, returns messages older than that ID (cursor-based pagination).
func (r *messageRepository) ListByConversationID(conversationID uuid.UUID, limit, offset int, beforeID *int64) ([]*model.Message, error) {
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	q := r.db.Where("conversation_id = ?", conversationID).Order("created_at DESC")
	if beforeID != nil {
		q = q.Where("message_id < ?", *beforeID)
	}
	var msgs []*model.Message
	err := q.Limit(limit).Offset(offset).Find(&msgs).Error
	if err != nil {
		return nil, err
	}
	return msgs, nil
}

// GetByID retrieves a message by ID.
func (r *messageRepository) GetByID(messageID int64) (*model.Message, error) {
	var msg model.Message
	err := r.db.Where("message_id = ?", messageID).First(&msg).Error
	if err != nil {
		return nil, err
	}
	return &msg, nil
}
