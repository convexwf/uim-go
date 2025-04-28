// Copyright 2025 convexwf
//
// Project: uim-go
// File: logger.go
// Email: convexwf@gmail.com
// Created: 2025-03-13
// Last modified: 2025-04-28
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
// Description: HTTP request logging middleware

package middleware

import (
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// LoggerMiddleware creates a middleware that logs HTTP requests using Gin's default formatter.
//
// Returns:
//   - gin.HandlerFunc: The logging middleware handler
func LoggerMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC1123),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	})
}

// LoggerMiddlewareSimple creates a simple logging middleware that logs request method, path, IP, and latency.
//
// This is a simpler alternative to LoggerMiddleware that provides basic request logging.
//
// Returns:
//   - gin.HandlerFunc: The simple logging middleware handler
func LoggerMiddlewareSimple() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		log.Printf("[HTTP] %s %s %d %s %v", method, path, status, c.ClientIP(), latency)
	}
}
