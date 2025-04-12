// Copyright 2025 convexwf
//
// Project: uim-go
// File: context_helper.go
// Email: convexwf@gmail.com
// Created: 2025-04-12
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// See the License for the full terms.
//
// Description: Helpers for API context (e.g. auth user ID)

package api

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// getUserIDFromContext returns the authenticated user's UUID from the request context.
// AuthMiddleware must have run and set "user_id" (string).
func getUserIDFromContext(c *gin.Context) (uuid.UUID, error) {
	v, ok := c.Get("user_id")
	if !ok {
		return uuid.Nil, errors.New("authorization required")
	}
	s, ok := v.(string)
	if !ok {
		return uuid.Nil, errors.New("invalid user context")
	}
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil, errors.New("invalid user id")
	}
	return id, nil
}
