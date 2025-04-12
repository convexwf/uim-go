// Copyright 2025 convexwf
//
// Project: uim-go
// File: conversation_service.go
// Email: convexwf@gmail.com
// Created: 2025-04-12
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// See the License for the full terms.
//
// Description: Conversation business logic

package service

import (
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/convexwf/uim-go/internal/model"
	"github.com/convexwf/uim-go/internal/repository"
)

var (
	ErrConversationNotFound = errors.New("conversation not found")
	ErrNotParticipant       = errors.New("user is not a participant")
	ErrInvalidConversation  = errors.New("invalid conversation")
)

// ConversationService defines conversation operations.
type ConversationService interface {
	CreateOneOnOne(creatorID, otherUserID uuid.UUID) (*model.Conversation, error)
	GetByID(conversationID, userID uuid.UUID) (*model.Conversation, error)
	ListByUserID(userID uuid.UUID, limit, offset int) ([]*model.Conversation, error)
	EnsureUserInConversation(conversationID, userID uuid.UUID) error
}

type conversationService struct {
	convRepo   repository.ConversationRepository
	userRepo   repository.UserRepository
}

// NewConversationService creates a new conversation service.
func NewConversationService(convRepo repository.ConversationRepository, userRepo repository.UserRepository) ConversationService {
	return &conversationService{
		convRepo: convRepo,
		userRepo: userRepo,
	}
}

// CreateOneOnOne creates or returns an existing one-on-one conversation between two users.
func (s *conversationService) CreateOneOnOne(creatorID, otherUserID uuid.UUID) (*model.Conversation, error) {
	if creatorID == otherUserID {
		return nil, fmt.Errorf("%w: cannot create conversation with self", ErrInvalidConversation)
	}
	_, err := s.userRepo.GetByID(otherUserID)
	if err != nil {
		return nil, ErrUserNotFound
	}
	existing, err := s.convRepo.FindOneOnOneBetween(creatorID, otherUserID)
	if err == nil {
		return existing, nil
	}
	conv := &model.Conversation{
		Type:      model.ConversationTypeOneOnOne,
		CreatedBy: creatorID,
	}
	if err := s.convRepo.Create(conv); err != nil {
		return nil, fmt.Errorf("create conversation: %w", err)
	}
	for _, uid := range []uuid.UUID{creatorID, otherUserID} {
		if err := s.convRepo.AddParticipant(&model.ConversationParticipant{
			ConversationID: conv.ConversationID,
			UserID:        uid,
			Role:          "member",
		}); err != nil {
			return nil, fmt.Errorf("add participant: %w", err)
		}
	}
	return conv, nil
}

// GetByID returns a conversation only if the user is a participant.
func (s *conversationService) GetByID(conversationID, userID uuid.UUID) (*model.Conversation, error) {
	ok, err := s.convRepo.IsParticipant(conversationID, userID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrNotParticipant
	}
	return s.convRepo.GetByID(conversationID)
}

// ListByUserID lists conversations for the user.
func (s *conversationService) ListByUserID(userID uuid.UUID, limit, offset int) ([]*model.Conversation, error) {
	return s.convRepo.ListByUserID(userID, limit, offset)
}

// EnsureUserInConversation returns nil only if the user is a participant (for access control).
func (s *conversationService) EnsureUserInConversation(conversationID, userID uuid.UUID) error {
	ok, err := s.convRepo.IsParticipant(conversationID, userID)
	if err != nil {
		return err
	}
	if !ok {
		return ErrNotParticipant
	}
	return nil
}
