// Copyright 2025 convexwf
//
// Project: uim-go
// File: message_handler.go
// Email: convexwf@gmail.com
// Created: 2025-04-12
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// See the License for the full terms.
//
// Description: HTTP handlers for message endpoints

package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/convexwf/uim-go/internal/model"
	"github.com/convexwf/uim-go/internal/service"
)

// MessageHandler handles message-related HTTP requests.
type MessageHandler struct {
	msgSvc service.MessageService
}

// NewMessageHandler creates a new message handler.
func NewMessageHandler(msgSvc service.MessageService) *MessageHandler {
	return &MessageHandler{msgSvc: msgSvc}
}

// ListByConversation returns paginated messages for a conversation.
// GET /api/conversations/:id/messages?limit=50&offset=0&before_id=123
func (h *MessageHandler) ListByConversation(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	convIDStr := c.Param("id")
	convID, err := uuid.Parse(convIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid conversation id"})
		return
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if limit > 100 {
		limit = 100
	}
	var beforeID *int64
	if b := c.Query("before_id"); b != "" {
		id, err := strconv.ParseInt(b, 10, 64)
		if err == nil {
			beforeID = &id
		}
	}
	msgs, err := h.msgSvc.ListByConversationID(convID, userID, limit, offset, beforeID)
	if err != nil {
		if err == service.ErrNotParticipant {
			c.JSON(http.StatusForbidden, gin.H{"error": "not a participant"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list messages"})
		return
	}
	if msgs == nil {
		msgs = []*model.Message{}
	}
	c.JSON(http.StatusOK, gin.H{"messages": msgs})
}
