// Copyright 2025 convexwf
//
// Project: uim-go
// File: presence_test.go
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// See the License for the full terms.
//
// Description: Unit tests for Redis presence store

package store

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

func setupRedisPresenceStore(t *testing.T) (*RedisPresenceStore, *miniredis.Miniredis) {
	t.Helper()
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis: %v", err)
	}
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	s := NewRedisPresenceStore(client)
	t.Cleanup(func() { mr.Close(); _ = client.Close() })
	return s, mr
}

func TestPresence_SetOnline_GetStatus(t *testing.T) {
	s, _ := setupRedisPresenceStore(t)
	ctx := context.Background()
	userID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")

	status, _, err := s.GetStatus(ctx, userID)
	if err != nil {
		t.Fatal(err)
	}
	if status != presenceStatusOffline {
		t.Errorf("initial: got status %q", status)
	}

	if err := s.SetOnline(ctx, userID); err != nil {
		t.Fatal(err)
	}
	status, _, err = s.GetStatus(ctx, userID)
	if err != nil {
		t.Fatal(err)
	}
	if status != presenceStatusOnline {
		t.Errorf("after SetOnline: got %q", status)
	}
}

func TestPresence_SetOffline(t *testing.T) {
	s, _ := setupRedisPresenceStore(t)
	ctx := context.Background()
	userID := uuid.New()
	_ = s.SetOnline(ctx, userID)
	if err := s.SetOffline(ctx, userID); err != nil {
		t.Fatal(err)
	}
	status, _, _ := s.GetStatus(ctx, userID)
	if status != presenceStatusOffline {
		t.Errorf("after SetOffline: got %q", status)
	}
}

func TestPresence_PublishUpdate(t *testing.T) {
	s, _ := setupRedisPresenceStore(t)
	ctx := context.Background()
	userID := uuid.New()
	if err := s.PublishUpdate(ctx, userID, presenceStatusOnline); err != nil {
		t.Errorf("PublishUpdate: %v", err)
	}
}
