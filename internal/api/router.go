// Copyright 2025 convexwf
//
// Project: uim-go
// File: router.go
// Email: convexwf@gmail.com
// Created: 2025-03-13
// Last modified: 2025-04-12
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// Description: HTTP router setup and route configuration

package api

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/convexwf/uim-go/internal/middleware"
	"github.com/convexwf/uim-go/internal/pkg/jwt"
	"github.com/convexwf/uim-go/internal/service"
	"github.com/convexwf/uim-go/internal/websocket"
)

// SetupRouter configures and returns the HTTP router with all routes.
//
// Parameters:
//   - db: The database connection
//   - authService: The authentication service
//   - jwtManager: The JWT manager for protected routes
//   - convSvc: The conversation service (optional for messaging routes)
//   - msgSvc: The message service (optional for messaging routes)
//
// Returns:
//   - *gin.Engine: The configured Gin router
func SetupRouter(db *gorm.DB, authService service.AuthService, jwtManager *jwt.JWTManager, convSvc service.ConversationService, msgSvc service.MessageService, hub *websocket.Hub) *gin.Engine {
	router := gin.Default()

	// Health check (no auth required)
	healthHandler := NewHealthHandler(db)
	router.GET("/health", healthHandler.Health)

	// API routes
	apiGroup := router.Group("/api")
	{
		// Auth routes (no auth required)
		authHandler := NewAuthHandler(authService)
		auth := apiGroup.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
		}

		// Protected routes (messaging)
		protected := apiGroup.Group("")
		protected.Use(middleware.AuthMiddleware(jwtManager))
		{
			convHandler := NewConversationHandler(convSvc)
			protected.POST("/conversations", convHandler.CreateOneOnOne)
			protected.GET("/conversations", convHandler.List)

			msgHandler := NewMessageHandler(msgSvc)
			protected.GET("/conversations/:id/messages", msgHandler.ListByConversation)
		}
	}

	// WebSocket (token in query or Authorization header)
	wsHandler := NewWebSocketHandler(jwtManager, hub, msgSvc)
	router.GET("/ws", wsHandler.ServeWS)

	return router
}
