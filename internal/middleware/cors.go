// Copyright 2025 convexwf
//
// Project: uim-go
// File: cors.go
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
// Description: CORS (Cross-Origin Resource Sharing) middleware

package middleware

import (
	"github.com/gin-gonic/gin"

	"github.com/convexwf/uim-go/internal/config"
)

// CORSMiddleware creates a middleware that handles CORS headers.
//
// It checks if the request origin is in the allowed origins list
// and sets appropriate CORS headers. It also handles preflight OPTIONS requests.
//
// Parameters:
//   - cfg: The application configuration containing CORS settings
//
// Returns:
//   - gin.HandlerFunc: The CORS middleware handler
func CORSMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		
		// Check if origin is allowed
		allowed := false
		for _, allowedOrigin := range cfg.CORS.AllowedOrigins {
			if origin == allowedOrigin {
				allowed = true
				break
			}
		}

		if allowed {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		}

		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
