// Copyright 2025 convexwf
//
// Project: uim-go
// File: error_handler.go
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
// Description: Error handling middleware for HTTP requests

package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ErrorHandlerMiddleware creates a middleware that handles errors from handlers.
//
// It catches errors added to the context by handlers and returns
// them as JSON responses with appropriate HTTP status codes.
//
// Returns:
//   - gin.HandlerFunc: The error handling middleware handler
func ErrorHandlerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Check if there are any errors
		if len(c.Errors) > 0 {
			err := c.Errors.Last()
			
			// Default to 500 if no status code is set
			statusCode := http.StatusInternalServerError
			if c.Writer.Status() != 0 {
				statusCode = c.Writer.Status()
			}

			c.JSON(statusCode, gin.H{
				"error": err.Error(),
			})
		}
	}
}
