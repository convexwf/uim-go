// Copyright 2025 convexwf
//
// Project: uim-go
// File: websocket_handler.go
// Email: convexwf@gmail.com
// Created: 2025-04-12
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// See the License for the full terms.
//
// Description: WebSocket upgrade and message handling

package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	gorillawebsocket "github.com/gorilla/websocket"

	"github.com/convexwf/uim-go/internal/model"
	"github.com/convexwf/uim-go/internal/pkg/jwt"
	"github.com/convexwf/uim-go/internal/service"
	"github.com/convexwf/uim-go/internal/websocket"
)

const (
	wsWriteWait       = 10 * time.Second
	wsPongWait        = 60 * time.Second
	wsPingPeriod      = (wsPongWait * 9) / 10
	wsMaxMessageSize  = 64 * 1024
	wsRateLimitCount  = 60
	wsRateLimitWindow = time.Minute
)

var upgrader = gorillawebsocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// WebSocketHandler handles WebSocket connections.
type WebSocketHandler struct {
	jwtManager *jwt.JWTManager
	hub        *websocket.Hub
	msgSvc     service.MessageService
}

// NewWebSocketHandler creates a new WebSocket handler.
func NewWebSocketHandler(jwtManager *jwt.JWTManager, hub *websocket.Hub, msgSvc service.MessageService) *WebSocketHandler {
	return &WebSocketHandler{
		jwtManager: jwtManager,
		hub:        hub,
		msgSvc:     msgSvc,
	}
}

// ServeWS upgrades the HTTP connection to WebSocket and runs the connection loop.
// Token can be passed as query ?token=... or header Authorization: Bearer ...
func (h *WebSocketHandler) ServeWS(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		auth := c.GetHeader("Authorization")
		if strings.HasPrefix(auth, "Bearer ") {
			token = strings.TrimPrefix(auth, "Bearer ")
		}
	}
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
		return
	}
	claims, err := h.jwtManager.ValidateAccessToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}
	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	client := &websocket.Client{
		UserID: userID,
		Conn:   conn,
		Send:   make(chan []byte, 256),
		Hub:    h.hub,
	}
	h.hub.Register(client)

	go h.writePump(client)
	h.readPump(client)
}

func (h *WebSocketHandler) readPump(client *websocket.Client) {
	defer func() {
		h.hub.Unregister(client)
		_ = client.Conn.Close()
	}()
	client.Conn.SetReadLimit(wsMaxMessageSize)
	_ = client.Conn.SetReadDeadline(time.Now().Add(wsPongWait))
	client.Conn.SetPongHandler(func(string) error {
		_ = client.Conn.SetReadDeadline(time.Now().Add(wsPongWait))
		return nil
	})

	var rateMu sync.Mutex
	var rateCount int
	var rateWindowStart = time.Now()

	for {
		_, raw, err := client.Conn.ReadMessage()
		if err != nil {
			break
		}
		var msg websocket.WSClientMessage
		if err := json.Unmarshal(raw, &msg); err != nil {
			continue
		}
		if msg.Type != "send_message" {
			continue
		}

		// Rate limit
		rateMu.Lock()
		if time.Since(rateWindowStart) > wsRateLimitWindow {
			rateCount = 0
			rateWindowStart = time.Now()
		}
		rateCount++
		if rateCount > wsRateLimitCount {
			rateMu.Unlock()
			continue
		}
		rateMu.Unlock()

		convID, err := uuid.Parse(msg.ConversationID)
		if err != nil {
			continue
		}
		msgType := model.MessageTypeText
		if msg.Content == "" {
			continue
		}
		m, err := h.msgSvc.Create(convID, client.UserID, msg.Content, msgType)
		if err != nil {
			// Optionally send error back to client; for now skip
			continue
		}
		// Message is persisted and NotifyNewMessage already called by service -> hub broadcasts to all
		// So we don't need to send again from here unless we want an echo to sender
		_ = m
	}
}

func (h *WebSocketHandler) writePump(client *websocket.Client) {
	ticker := time.NewTicker(wsPingPeriod)
	defer func() {
		ticker.Stop()
		_ = client.Conn.Close()
	}()
	for {
		select {
		case payload, ok := <-client.Send:
			_ = client.Conn.SetWriteDeadline(time.Now().Add(wsWriteWait))
			if !ok {
				_ = client.Conn.WriteMessage(gorillawebsocket.CloseMessage, nil)
				return
			}
			if err := client.Conn.WriteMessage(gorillawebsocket.TextMessage, payload); err != nil {
				return
			}
		case <-ticker.C:
			_ = client.Conn.SetWriteDeadline(time.Now().Add(wsWriteWait))
			if err := client.Conn.WriteMessage(gorillawebsocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
