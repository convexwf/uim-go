// Copyright 2025 convexwf
//
// Project: uim-go
// File: message_service.go
// Email: convexwf@gmail.com
// Created: 2025-04-12
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// See the License for the full terms.
//
// Description: Message business logic and optional real-time notification

package service

import (
	"errors"
	"fmt"
	"unicode/utf8"

	"github.com/google/uuid"

	"github.com/convexwf/uim-go/internal/model"
	"github.com/convexwf/uim-go/internal/repository"
)

var (
	ErrMessageNotFound = errors.New("message not found")
)

const (
	MaxMessageContentLength = 64 * 1024 // 64KB
)

// MessageNotifier is called after a message is persisted (e.g. to broadcast via WebSocket).
// Implementations can be nil-safe; the service will only call if non-nil.
type MessageNotifier interface {
	NotifyNewMessage(conversationID uuid.UUID, msg *model.Message)
}

// MessageService defines message operations.
type MessageService interface {
	Create(conversationID, senderID uuid.UUID, content string, msgType model.MessageType) (*model.Message, error)
	ListByConversationID(conversationID, userID uuid.UUID, limit, offset int, beforeID *int64) ([]*model.Message, error)
}

type messageService struct {
	msgRepo   repository.MessageRepository
	convSvc   ConversationService
	notifier  MessageNotifier
}

// NewMessageService creates a new message service. notifier can be nil.
func NewMessageService(msgRepo repository.MessageRepository, convSvc ConversationService, notifier MessageNotifier) MessageService {
	return &messageService{
		msgRepo:  msgRepo,
		convSvc:  convSvc,
		notifier: notifier,
	}
}

// Create validates, persists a message, and optionally notifies (e.g. WebSocket broadcast).
func (s *messageService) Create(conversationID, senderID uuid.UUID, content string, msgType model.MessageType) (*model.Message, error) {
	if err := s.convSvc.EnsureUserInConversation(conversationID, senderID); err != nil {
		return nil, err
	}
	content = trimContent(content)
	if content == "" {
		return nil, fmt.Errorf("%w: message content required", ErrInvalidInput)
	}
	if utf8.RuneCountInString(content) > MaxMessageContentLength {
		return nil, fmt.Errorf("%w: message too long", ErrInvalidInput)
	}
	if msgType == "" {
		msgType = model.MessageTypeText
	}
	msg := &model.Message{
		ConversationID: conversationID,
		SenderID:       senderID,
		Content:        content,
		MessageType:    msgType,
	}
	if err := s.msgRepo.Create(msg); err != nil {
		return nil, fmt.Errorf("create message: %w", err)
	}
	if s.notifier != nil {
		s.notifier.NotifyNewMessage(conversationID, msg)
	}
	return msg, nil
}

// ListByConversationID returns messages for a conversation if the user is a participant.
func (s *messageService) ListByConversationID(conversationID, userID uuid.UUID, limit, offset int, beforeID *int64) ([]*model.Message, error) {
	if err := s.convSvc.EnsureUserInConversation(conversationID, userID); err != nil {
		return nil, err
	}
	return s.msgRepo.ListByConversationID(conversationID, limit, offset, beforeID)
}

func trimContent(s string) string {
	const cutset = " \t\n\r"
	start := 0
	for start < len(s) {
		r, w := utf8.DecodeRuneInString(s[start:])
		if r == utf8.RuneError || (r != ' ' && r != '\t' && r != '\n' && r != '\r') {
			break
		}
		start += w
	}
	end := len(s)
	for end > start {
		r, w := utf8.DecodeLastRuneInString(s[start:end])
		if r == utf8.RuneError || (r != ' ' && r != '\t' && r != '\n' && r != '\r') {
			break
		}
		end -= w
	}
	return s[start:end]
}
