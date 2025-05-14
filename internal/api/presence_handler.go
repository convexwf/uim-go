// Copyright 2025 convexwf
//
// Project: uim-go
// File: presence_handler.go
// Email: convexwf@gmail.com
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// See the License for the full terms.
//
// Description: HTTP handler for user presence (online/offline) API

package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/convexwf/uim-go/internal/store"
)

// PresenceHandler handles presence API requests.
type PresenceHandler struct {
	presence store.PresenceStore
}

// NewPresenceHandler creates a new presence handler. presence may be nil (returns offline for all).
func NewPresenceHandler(presence store.PresenceStore) *PresenceHandler {
	return &PresenceHandler{presence: presence}
}

// PresenceResponse is the JSON response for GET /api/users/:id/presence.
type PresenceResponse struct {
	UserID   string `json:"user_id"`
	Status   string `json:"status"`
	LastSeen string `json:"last_seen,omitempty"`
}

// GetPresence returns the presence status for a user.
// GET /api/users/:id/presence
func (h *PresenceHandler) GetPresence(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user id required"})
		return
	}
	userID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	if h.presence == nil {
		c.JSON(http.StatusOK, PresenceResponse{
			UserID: userID.String(),
			Status: "offline",
		})
		return
	}

	status, lastSeen, err := h.presence.GetStatus(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get presence"})
		return
	}

	resp := PresenceResponse{
		UserID: userID.String(),
		Status: status,
	}
	if !lastSeen.IsZero() {
		resp.LastSeen = lastSeen.UTC().Format(time.RFC3339)
	}
	c.JSON(http.StatusOK, resp)
}
