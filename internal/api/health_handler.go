// Copyright 2025 convexwf
//
// Project: uim-go
// File: health_handler.go
// Email: convexwf@gmail.com
// Created: 2025-03-13
// Last modified: 2025-05-14
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
// Description: HTTP handler for health check endpoint

package api

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// HealthHandler handles health check requests.
type HealthHandler struct {
	db    *gorm.DB
	redis redis.Cmdable
}

// NewHealthHandler creates a new health check handler. redis may be nil (Redis check skipped).
func NewHealthHandler(db *gorm.DB, redis redis.Cmdable) *HealthHandler {
	return &HealthHandler{db: db, redis: redis}
}

// HealthResponse represents the health check response.
type HealthResponse struct {
	Status string            `json:"status"`
	Checks map[string]string `json:"checks"`
}

// Health handles health check requests.
//
// It checks the database connectivity and returns the system health status.
//
// @Summary Health check
// @Description Check system health status
// @Tags health
// @Produce json
// @Success 200 {object} HealthResponse
// @Router /health [get]
func (h *HealthHandler) Health(c *gin.Context) {
	checks := make(map[string]string)
	status := "healthy"

	// Check database connection
	sqlDB, err := h.db.DB()
	if err != nil {
		status = "unhealthy"
		checks["database"] = "error: " + err.Error()
	} else {
		if err := sqlDB.Ping(); err != nil {
			status = "unhealthy"
			checks["database"] = "error: " + err.Error()
		} else {
			checks["database"] = "healthy"
		}
	}

	// Check Redis if configured
	if h.redis != nil {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()
		if err := h.redis.Ping(ctx).Err(); err != nil {
			checks["redis"] = "unhealthy: " + err.Error()
			if status == "healthy" {
				status = "degraded"
			}
		} else {
			checks["redis"] = "healthy"
		}
	}

	response := HealthResponse{
		Status: status,
		Checks: checks,
	}

	if status == "healthy" || status == "degraded" {
		c.JSON(http.StatusOK, response)
	} else {
		c.JSON(http.StatusServiceUnavailable, response)
	}
}
