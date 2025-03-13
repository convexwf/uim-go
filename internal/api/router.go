// Copyright 2025 convexwf
//
// Project: uim-go
// File: router.go
// Email: convexwf@gmail.com
// Created: 2025-03-13
// Last modified: 2025-03-13
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

	"github.com/convexwf/uim-go/internal/service"
)

// SetupRouter configures and returns the HTTP router with all routes.
//
// It sets up health check endpoints, authentication endpoints,
// and prepares the structure for protected routes (to be added in Phase 2).
//
// Parameters:
//   - db: The database connection
//   - authService: The authentication service
//
// Returns:
//   - *gin.Engine: The configured Gin router
func SetupRouter(db *gorm.DB, authService service.AuthService) *gin.Engine {
	router := gin.Default()

	// Health check (no auth required)
	healthHandler := NewHealthHandler(db)
	router.GET("/health", healthHandler.Health)

	// API routes
	api := router.Group("/api")
	{
		// Auth routes (no auth required)
		authHandler := NewAuthHandler(authService)
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
		}

		// Protected routes will be added here in Phase 2
		// protected := api.Group("")
		// protected.Use(middleware.AuthMiddleware(jwtManager))
		// {
		//     // Conversation routes
		//     // Message routes
		// }
	}

	return router
}
