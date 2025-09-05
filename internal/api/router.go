// Copyright 2025 convexwf
//
// Project: uim-go
// File: router.go
// Email: convexwf@gmail.com
// Created: 2025-03-13
// Last modified: 2025-09-05
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
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/convexwf/uim-go/internal/config"
	"github.com/convexwf/uim-go/internal/middleware"
	"github.com/convexwf/uim-go/internal/pkg/jwt"
	"github.com/convexwf/uim-go/internal/service"
	"github.com/convexwf/uim-go/internal/store"
	"github.com/convexwf/uim-go/internal/websocket"
)

// SetupRouter configures and returns the HTTP router with all routes.
//
// redisClient may be nil (offline queue and presence disabled, health check skips Redis).
// offlineQueue and presenceStore may be nil (offline messages dropped, presence returns offline).
func SetupRouter(cfg *config.Config, db *gorm.DB, authService service.AuthService, jwtManager *jwt.JWTManager, convSvc service.ConversationService, msgSvc service.MessageService, hub *websocket.Hub, redisClient redis.Cmdable, offlineQueue store.OfflineQueue, presenceStore store.PresenceStore) *gin.Engine {
	// Use New + Recovery only: gin.Default() also attaches gin.Logger() writing to
	// gin.DefaultWriter (often stderr), which does not follow log.SetOutput(UIM_LOG_FILE).
	// cmd/server applies LoggerMiddlewareSimple so [HTTP] lines share the same log sink as [AUTH]/[DB].
	router := gin.New()
	router.Use(gin.Recovery())
	// Register middleware before routes so normal handlers and preflight requests both
	// receive CORS / logging / error handling consistently.
	router.Use(
		middleware.CORSMiddleware(cfg),
		middleware.LoggerMiddlewareSimple(),
		middleware.ErrorHandlerMiddleware(),
	)

	// Health check (no auth required)
	healthHandler := NewHealthHandler(db, redisClient)
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

		authProtected := apiGroup.Group("/auth")
		authProtected.Use(middleware.AuthMiddleware(jwtManager))
		{
			authProtected.GET("/me", authHandler.Me)
		}

		// Protected routes (messaging)
		protected := apiGroup.Group("")
		protected.Use(middleware.AuthMiddleware(jwtManager))
		{
			convHandler := NewConversationHandler(convSvc)
			protected.POST("/conversations", convHandler.CreateOneOnOne)
			protected.GET("/conversations", convHandler.List)
			protected.POST("/conversations/:id/read", convHandler.MarkRead)

			msgHandler := NewMessageHandler(msgSvc)
			protected.GET("/conversations/:id/messages", msgHandler.ListByConversation)

			presenceHandler := NewPresenceHandler(presenceStore)
			protected.GET("/users/:id/presence", presenceHandler.GetPresence)
		}
	}

	// WebSocket (token in query or Authorization header)
	wsHandler := NewWebSocketHandler(jwtManager, hub, msgSvc, offlineQueue, presenceStore)
	router.GET("/ws", wsHandler.ServeWS)

	return router
}
