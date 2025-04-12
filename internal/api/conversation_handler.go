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

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/convexwf/uim-go/internal/model"
	"github.com/convexwf/uim-go/internal/service"
)

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

// List lists conversations for the current user.
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
	convs, err := h.convSvc.ListByUserID(userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list conversations"})
		return
	}
	if convs == nil {
		convs = []*model.Conversation{}
	}
	c.JSON(http.StatusOK, gin.H{"conversations": convs})
}
