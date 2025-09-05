// Copyright 2025 convexwf
//
// Project: uim-go
// File: conversation_handler.go
// Email: convexwf@gmail.com
// Created: 2025-04-12
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// See the License for the full terms.
//
// Description: HTTP handlers for conversation endpoints

package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/convexwf/uim-go/internal/service"
)

// UserSummary is the API shape for a user in conversation list (other_user).
type UserSummary struct {
	UserID      string `json:"user_id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	AvatarURL   string `json:"avatar_url"`
}

// MessageSummary is the API shape for the last message in a conversation.
type MessageSummary struct {
	MessageID int64     `json:"message_id"`
	Content   string    `json:"content"`
	SenderID  string    `json:"sender_id"`
	CreatedAt time.Time `json:"created_at"`
}

// ConversationListItem is the API shape for one conversation in the list (with metadata).
type ConversationListItem struct {
	ConversationID string          `json:"conversation_id"`
	Type           string          `json:"type"`
	Name           string          `json:"name,omitempty"`
	CreatedBy      string          `json:"created_by"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
	OtherUser      *UserSummary    `json:"other_user,omitempty"`
	LastMessage    *MessageSummary `json:"last_message,omitempty"`
	UnreadCount    int             `json:"unread_count"`
}

// ConversationHandler handles conversation-related HTTP requests.
type ConversationHandler struct {
	convSvc service.ConversationService
}

// NewConversationHandler creates a new conversation handler.
func NewConversationHandler(convSvc service.ConversationService) *ConversationHandler {
	return &ConversationHandler{convSvc: convSvc}
}

// CreateOneOnOneRequest is the body for creating a 1:1 conversation.
type CreateOneOnOneRequest struct {
	OtherUserID string `json:"other_user_id" binding:"required"`
}

// MarkReadRequest is the body for marking messages as read.
type MarkReadRequest struct {
	LastReadMessageID *int64 `json:"last_read_message_id" binding:"required"`
}

// CreateOneOnOne creates or returns an existing one-on-one conversation.
// POST /api/conversations
func (h *ConversationHandler) CreateOneOnOne(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	var req CreateOneOnOneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}
	otherID, err := uuid.Parse(req.OtherUserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid other_user_id"})
		return
	}
	conv, err := h.convSvc.CreateOneOnOne(userID, otherID)
	if err != nil {
		switch {
		case err == service.ErrUserNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		case err == service.ErrInvalidConversation:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create conversation"})
		}
		return
	}
	c.JSON(http.StatusCreated, conv)
}

// List lists conversations for the current user with metadata (other_user, last_message, unread_count).
// GET /api/conversations?limit=20&offset=0
func (h *ConversationHandler) List(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if limit > 100 {
		limit = 100
	}
	convs, err := h.convSvc.ListByUserIDWithMeta(userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list conversations"})
		return
	}
	if convs == nil {
		convs = []*service.ConversationWithMeta{}
	}
	items := make([]ConversationListItem, len(convs))
	for i, m := range convs {
		items[i] = conversationWithMetaToListItem(m)
	}
	c.JSON(http.StatusOK, gin.H{"conversations": items})
}

func conversationWithMetaToListItem(m *service.ConversationWithMeta) ConversationListItem {
	item := ConversationListItem{
		ConversationID: m.Conv.ConversationID.String(),
		Type:           string(m.Conv.Type),
		Name:           m.Conv.Name,
		CreatedBy:      m.Conv.CreatedBy.String(),
		CreatedAt:      m.Conv.CreatedAt,
		UpdatedAt:      m.Conv.UpdatedAt,
		UnreadCount:    m.UnreadCount,
	}
	if m.OtherUser != nil {
		item.OtherUser = &UserSummary{
			UserID:      m.OtherUser.UserID.String(),
			Username:    m.OtherUser.Username,
			DisplayName: m.OtherUser.DisplayName,
			AvatarURL:   m.OtherUser.AvatarURL,
		}
	}
	if m.LastMessage != nil {
		item.LastMessage = &MessageSummary{
			MessageID: m.LastMessage.MessageID,
			Content:   m.LastMessage.Content,
			SenderID:  m.LastMessage.SenderID.String(),
			CreatedAt: m.LastMessage.CreatedAt,
		}
	}
	return item
}

// MarkRead updates the current user's last read message in the conversation.
// POST /api/conversations/:id/read
func (h *ConversationHandler) MarkRead(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	convIDStr := c.Param("id")
	if convIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing conversation id"})
		return
	}
	convID, err := uuid.Parse(convIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid conversation id"})
		return
	}
	var req MarkReadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}
	err = h.convSvc.MarkRead(convID, userID, *req.LastReadMessageID)
	if err != nil {
		if err == service.ErrNotParticipant {
			c.JSON(http.StatusForbidden, gin.H{"error": "not a participant"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to mark read"})
		return
	}
	c.Status(http.StatusNoContent)
}

// DeleteConversation removes a conversation for all participants.
// DELETE /api/conversations/:id
func (h *ConversationHandler) DeleteConversation(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	convIDStr := c.Param("id")
	if convIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing conversation id"})
		return
	}
	convID, err := uuid.Parse(convIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid conversation id"})
		return
	}
	err = h.convSvc.DeleteConversation(convID, userID)
	if err != nil {
		if err == service.ErrNotParticipant {
			c.JSON(http.StatusForbidden, gin.H{"error": "not a participant"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete conversation"})
		return
	}
	c.Status(http.StatusNoContent)
}
