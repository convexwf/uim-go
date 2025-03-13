// Copyright 2025 convexwf
//
// Project: uim-go
// File: health_handler.go
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
// Description: HTTP handler for health check endpoint

package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// HealthHandler handles health check requests.
type HealthHandler struct {
	db *gorm.DB
}

// NewHealthHandler creates a new health check handler.
func NewHealthHandler(db *gorm.DB) *HealthHandler {
	return &HealthHandler{db: db}
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

	response := HealthResponse{
		Status: status,
		Checks: checks,
	}

	if status == "healthy" {
		c.JSON(http.StatusOK, response)
	} else {
		c.JSON(http.StatusServiceUnavailable, response)
	}
}
