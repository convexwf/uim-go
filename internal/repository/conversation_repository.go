// Copyright 2025 convexwf
//
// Project: uim-go
// File: conversation_repository.go
// Email: convexwf@gmail.com
// Created: 2025-04-12
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// See the License for the full terms.
//
// Description: Conversation repository for database operations

package repository

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/convexwf/uim-go/internal/model"
)

// ConversationRepository defines the interface for conversation data access.
type ConversationRepository interface {
	Create(conv *model.Conversation) error
	GetByID(conversationID uuid.UUID) (*model.Conversation, error)
	ListByUserID(userID uuid.UUID, limit, offset int) ([]*model.Conversation, error)
	AddParticipant(p *model.ConversationParticipant) error
	FindOneOnOneBetween(userID1, userID2 uuid.UUID) (*model.Conversation, error)
	IsParticipant(conversationID, userID uuid.UUID) (bool, error)
	GetParticipantUserIDs(conversationID uuid.UUID) ([]uuid.UUID, error)
	UpdateParticipantLastRead(conversationID, userID uuid.UUID, lastReadMessageID int64) error
	GetUnreadCounts(userID uuid.UUID, conversationIDs []uuid.UUID) (map[uuid.UUID]int, error)
	GetOtherParticipantUserIDsForOneOnOne(currentUserID uuid.UUID, conversationIDs []uuid.UUID) (map[uuid.UUID]uuid.UUID, error)
}

type conversationRepository struct {
	db *gorm.DB
}

// NewConversationRepository creates a new conversation repository.
func NewConversationRepository(db *gorm.DB) ConversationRepository {
	return &conversationRepository{db: db}
}

// Create creates a new conversation.
func (r *conversationRepository) Create(conv *model.Conversation) error {
	return r.db.Create(conv).Error
}

// GetByID retrieves a conversation by ID.
func (r *conversationRepository) GetByID(conversationID uuid.UUID) (*model.Conversation, error) {
	var conv model.Conversation
	err := r.db.Where("conversation_id = ?", conversationID).First(&conv).Error
	if err != nil {
		return nil, err
	}
	return &conv, nil
}

// ListByUserID lists conversations for a user (via participants), ordered by updated_at desc.
func (r *conversationRepository) ListByUserID(userID uuid.UUID, limit, offset int) ([]*model.Conversation, error) {
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	var convs []*model.Conversation
	err := r.db.Table("conversations").
		Joins("INNER JOIN conversation_participants ON conversation_participants.conversation_id = conversations.conversation_id").
		Where("conversation_participants.user_id = ?", userID).
		Order("conversations.updated_at DESC").
		Limit(limit).Offset(offset).
		Find(&convs).Error
	if err != nil {
		return nil, err
	}
	return convs, nil
}

// AddParticipant adds a participant to a conversation.
func (r *conversationRepository) AddParticipant(p *model.ConversationParticipant) error {
	return r.db.Create(p).Error
}

// FindOneOnOneBetween finds an existing one-on-one conversation between two users.
// It looks up conversations of type one_on_one that have exactly these two participants.
func (r *conversationRepository) FindOneOnOneBetween(userID1, userID2 uuid.UUID) (*model.Conversation, error) {
	var conv model.Conversation
	err := r.db.Table("conversations").
		Joins("INNER JOIN conversation_participants cp1 ON cp1.conversation_id = conversations.conversation_id AND cp1.user_id = ?", userID1).
		Joins("INNER JOIN conversation_participants cp2 ON cp2.conversation_id = conversations.conversation_id AND cp2.user_id = ?", userID2).
		Where("conversations.type = ?", model.ConversationTypeOneOnOne).
		First(&conv).Error
	if err != nil {
		return nil, err
	}
	return &conv, nil
}

// IsParticipant returns true if the user is a participant of the conversation.
func (r *conversationRepository) IsParticipant(conversationID, userID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&model.ConversationParticipant{}).
		Where("conversation_id = ? AND user_id = ?", conversationID, userID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetParticipantUserIDs returns all user IDs that participate in the conversation.
func (r *conversationRepository) GetParticipantUserIDs(conversationID uuid.UUID) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	err := r.db.Model(&model.ConversationParticipant{}).
		Where("conversation_id = ?", conversationID).
		Pluck("user_id", &ids).Error
	if err != nil {
		return nil, err
	}
	return ids, nil
}

// UpdateParticipantLastRead sets the last_read_message_id for the participant.
func (r *conversationRepository) UpdateParticipantLastRead(conversationID, userID uuid.UUID, lastReadMessageID int64) error {
	return r.db.Model(&model.ConversationParticipant{}).
		Where("conversation_id = ? AND user_id = ?", conversationID, userID).
		Update("last_read_message_id", lastReadMessageID).Error
}

// GetUnreadCounts returns the count of messages (from others) not yet read by the user per conversation.
// Unread = messages where message_id > participant's last_read_message_id and sender_id != userID.
func (r *conversationRepository) GetUnreadCounts(userID uuid.UUID, conversationIDs []uuid.UUID) (map[uuid.UUID]int, error) {
	if len(conversationIDs) == 0 {
		return map[uuid.UUID]int{}, nil
	}
	type row struct {
		ConversationID uuid.UUID `gorm:"column:conversation_id"`
		Cnt            int       `gorm:"column:cnt"`
	}
	var rows []row
	err := r.db.Table("messages").
		Select("messages.conversation_id, COUNT(*) AS cnt").
		Joins("INNER JOIN conversation_participants cp ON cp.conversation_id = messages.conversation_id AND cp.user_id = ?", userID).
		Where("messages.conversation_id IN ? AND messages.sender_id != ? AND messages.message_id > COALESCE(cp.last_read_message_id, 0)", conversationIDs, userID).
		Group("messages.conversation_id").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make(map[uuid.UUID]int, len(rows))
	for _, rw := range rows {
		out[rw.ConversationID] = rw.Cnt
	}
	return out, nil
}

// GetOtherParticipantUserIDsForOneOnOne returns the other participant's user_id for each one_on_one conversation.
// Only includes conversations that are type one_on_one and have exactly one other participant.
func (r *conversationRepository) GetOtherParticipantUserIDsForOneOnOne(currentUserID uuid.UUID, conversationIDs []uuid.UUID) (map[uuid.UUID]uuid.UUID, error) {
	if len(conversationIDs) == 0 {
		return map[uuid.UUID]uuid.UUID{}, nil
	}
	var results []struct {
		ConversationID uuid.UUID
		UserID         uuid.UUID
	}
	err := r.db.Table("conversation_participants").
		Select("conversation_participants.conversation_id, conversation_participants.user_id").
		Joins("INNER JOIN conversations ON conversations.conversation_id = conversation_participants.conversation_id").
		Where("conversations.conversation_id IN ? AND conversations.type = ? AND conversation_participants.user_id != ?",
			conversationIDs, model.ConversationTypeOneOnOne, currentUserID).
		Find(&results).Error
	if err != nil {
		return nil, err
	}
	out := make(map[uuid.UUID]uuid.UUID, len(results))
	for _, rw := range results {
		out[rw.ConversationID] = rw.UserID
	}
	return out, nil
}
