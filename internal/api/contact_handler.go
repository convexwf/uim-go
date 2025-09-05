// Copyright 2025 convexwf
//
// Project: uim-go
// File: contact_handler.go
// Email: convexwf@gmail.com
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// See the License for the full terms.
//
// Description: HTTP handlers for contact endpoints

package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/convexwf/uim-go/internal/service"
)

// ContactPresenceResponse is the API shape for contact presence.
type ContactPresenceResponse struct {
	Status   string `json:"status"`
	LastSeen string `json:"last_seen,omitempty"`
}

// ContactListItem is the API shape for one contact row.
type ContactListItem struct {
	UserID      string                   `json:"user_id"`
	Username    string                   `json:"username"`
	DisplayName string                   `json:"display_name"`
	AvatarURL   string                   `json:"avatar_url"`
	CreatedAt   time.Time                `json:"created_at"`
	Presence    *ContactPresenceResponse `json:"presence,omitempty"`
}

// UserSearchItem is the API shape for user search results.
type UserSearchItem struct {
	UserID       string `json:"user_id"`
	Username     string `json:"username"`
	DisplayName  string `json:"display_name"`
	AvatarURL    string `json:"avatar_url"`
	AlreadyAdded bool   `json:"already_added"`
}

// ContactHandler handles contact-related HTTP requests.
type ContactHandler struct {
	contactSvc service.ContactService
}

// NewContactHandler creates a new contact handler.
func NewContactHandler(contactSvc service.ContactService) *ContactHandler {
	return &ContactHandler{contactSvc: contactSvc}
}

// AddContactRequest is the body for adding a contact.
type AddContactRequest struct {
	ContactUserID string `json:"contact_user_id" binding:"required"`
}

// ListContacts lists one-way contacts for the current user.
// GET /api/contacts?limit=20&offset=0
func (h *ContactHandler) ListContacts(c *gin.Context) {
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
	contacts, err := h.contactSvc.ListContacts(userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list contacts"})
		return
	}
	items := make([]ContactListItem, len(contacts))
	for i, item := range contacts {
		items[i] = contactToListItem(item)
	}
	c.JSON(http.StatusOK, gin.H{"contacts": items})
}

// AddContact adds a one-way contact for the current user.
// POST /api/contacts
func (h *ContactHandler) AddContact(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	var req AddContactRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}
	contactUserID, err := uuid.Parse(req.ContactUserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid contact_user_id"})
		return
	}
	created, err := h.contactSvc.AddContact(userID, contactUserID)
	if err != nil {
		switch err {
		case service.ErrInvalidContact:
			c.JSON(http.StatusBadRequest, gin.H{"error": "cannot add self as contact"})
		case service.ErrUserNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add contact"})
		}
		return
	}
	status := http.StatusOK
	if created {
		status = http.StatusCreated
	}
	c.JSON(status, gin.H{"contact_user_id": contactUserID.String()})
}

// DeleteContact removes a one-way contact for the current user.
// DELETE /api/contacts/:id
func (h *ContactHandler) DeleteContact(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	contactUserID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid contact user id"})
		return
	}
	deleted, err := h.contactSvc.DeleteContact(userID, contactUserID)
	if err != nil {
		if err == service.ErrInvalidContact {
			c.JSON(http.StatusBadRequest, gin.H{"error": "cannot delete self as contact"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete contact"})
		return
	}
	if !deleted {
		c.Status(http.StatusNoContent)
		return
	}
	c.Status(http.StatusNoContent)
}

// SearchUsers searches by exact username for adding contacts.
// GET /api/users/search?q=username
func (h *ContactHandler) SearchUsers(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	query := c.Query("q")
	results, err := h.contactSvc.SearchByUsername(userID, query)
	if err != nil {
		if err == service.ErrInvalidInput {
			c.JSON(http.StatusBadRequest, gin.H{"error": "query required"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to search users"})
		return
	}
	items := make([]UserSearchItem, len(results))
	for i, item := range results {
		items[i] = UserSearchItem{
			UserID:       item.UserID.String(),
			Username:     item.Username,
			DisplayName:  item.DisplayName,
			AvatarURL:    item.AvatarURL,
			AlreadyAdded: item.AlreadyAdded,
		}
	}
	c.JSON(http.StatusOK, gin.H{"users": items})
}

func contactToListItem(item *service.ContactWithPresence) ContactListItem {
	out := ContactListItem{
		UserID:      item.UserID.String(),
		Username:    item.Username,
		DisplayName: item.DisplayName,
		AvatarURL:   item.AvatarURL,
		CreatedAt:   item.CreatedAt,
	}
	if item.Status != "" {
		p := &ContactPresenceResponse{Status: item.Status}
		if !item.LastSeen.IsZero() {
			p.LastSeen = item.LastSeen.UTC().Format(time.RFC3339)
		}
		out.Presence = p
	}
	return out
}
