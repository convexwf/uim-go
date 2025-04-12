// Copyright 2025 convexwf
//
// Project: uim-go
// File: hub.go
// Email: convexwf@gmail.com
// Created: 2025-04-12
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// See the License for the full terms.
//
// Description: WebSocket connection hub and message broadcaster

package websocket

import (
	"encoding/json"
	"sync"

	"github.com/google/uuid"
	gorillawebsocket "github.com/gorilla/websocket"

	"github.com/convexwf/uim-go/internal/model"
	"github.com/convexwf/uim-go/internal/repository"
)

// Hub maintains active WebSocket connections and broadcasts messages to conversation participants.
type Hub struct {
	convRepo repository.ConversationRepository
	// userID -> set of clients (one user can have multiple connections)
	clients map[uuid.UUID]map[*Client]struct{}
	mu      sync.RWMutex
}

// Client represents a single WebSocket connection with its user ID.
type Client struct {
	UserID uuid.UUID
	Conn   *gorillawebsocket.Conn
	Send   chan []byte
	Hub    *Hub
}

// NewHub creates a new WebSocket hub.
func NewHub(convRepo repository.ConversationRepository) *Hub {
	return &Hub{
		convRepo: convRepo,
		clients:  make(map[uuid.UUID]map[*Client]struct{}),
	}
}

// NotifyNewMessage implements service.MessageNotifier. It broadcasts the message to all participants of the conversation.
func (h *Hub) NotifyNewMessage(conversationID uuid.UUID, msg *model.Message) {
	userIDs, err := h.convRepo.GetParticipantUserIDs(conversationID)
	if err != nil {
		return
	}
	payload, err := json.Marshal(WSMessage{
		Type:    "new_message",
		Message: msg,
	})
	if err != nil {
		return
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, uid := range userIDs {
		if conns, ok := h.clients[uid]; ok {
			for c := range conns {
				select {
				case c.Send <- payload:
				default:
					// skip if send buffer full
				}
			}
		}
	}
}

// Register adds a client to the hub.
func (h *Hub) Register(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.clients[client.UserID] == nil {
		h.clients[client.UserID] = make(map[*Client]struct{})
	}
	h.clients[client.UserID][client] = struct{}{}
}

// Unregister removes a client from the hub.
func (h *Hub) Unregister(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if conns, ok := h.clients[client.UserID]; ok {
		delete(conns, client)
		if len(conns) == 0 {
			delete(h.clients, client.UserID)
		}
	}
	close(client.Send)
}

// WSMessage is the JSON envelope for WebSocket messages.
type WSMessage struct {
	Type    string        `json:"type"`
	Message *model.Message `json:"message,omitempty"`
}

// WSClientMessage is the JSON format for client-to-server messages.
type WSClientMessage struct {
	Type            string `json:"type"`
	ConversationID  string `json:"conversation_id"`
	Content         string `json:"content"`
}
