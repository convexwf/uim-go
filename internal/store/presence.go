// Copyright 2025 convexwf
//
// Project: uim-go
// File: presence.go
// Email: convexwf@gmail.com
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// See the License for the full terms.
//
// Description: Presence (online/offline) store interface and Redis implementation

package store

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	presenceKeyPrefix    = "presence:"
	presenceChannel      = "presence:updates"
	presenceTTL          = 90 * time.Second
	presenceStatusOnline  = "online"
	presenceStatusOffline = "offline"
)

// PresenceStore manages user online/offline status in Redis.
// Implementations may be nil-safe; callers should check for nil.
type PresenceStore interface {
	SetOnline(ctx context.Context, userID uuid.UUID) error
	SetOffline(ctx context.Context, userID uuid.UUID) error
	Refresh(ctx context.Context, userID uuid.UUID) error
	GetStatus(ctx context.Context, userID uuid.UUID) (status string, lastSeen time.Time, err error)
	PublishUpdate(ctx context.Context, userID uuid.UUID, status string) error
}

// RedisPresenceStore implements PresenceStore using Redis.
type RedisPresenceStore struct {
	client redis.Cmdable
}

// NewRedisPresenceStore creates a presence store backed by Redis.
func NewRedisPresenceStore(client redis.Cmdable) *RedisPresenceStore {
	return &RedisPresenceStore{client: client}
}

func presenceKey(userID uuid.UUID) string {
	return presenceKeyPrefix + userID.String()
}

// SetOnline marks the user as online with TTL.
func (s *RedisPresenceStore) SetOnline(ctx context.Context, userID uuid.UUID) error {
	key := presenceKey(userID)
	if err := s.client.Set(ctx, key, presenceStatusOnline, presenceTTL).Err(); err != nil {
		return fmt.Errorf("presence set online: %w", err)
	}
	return nil
}

// SetOffline removes the user's presence key (or sets offline).
func (s *RedisPresenceStore) SetOffline(ctx context.Context, userID uuid.UUID) error {
	if err := s.client.Del(ctx, presenceKey(userID)).Err(); err != nil {
		return fmt.Errorf("presence set offline: %w", err)
	}
	return nil
}

// Refresh extends the TTL for the user's presence (e.g. on heartbeat).
func (s *RedisPresenceStore) Refresh(ctx context.Context, userID uuid.UUID) error {
	key := presenceKey(userID)
	if err := s.client.Expire(ctx, key, presenceTTL).Err(); err != nil {
		return fmt.Errorf("presence refresh: %w", err)
	}
	return nil
}

// GetStatus returns the user's presence status. If key is missing or expired, returns offline.
func (s *RedisPresenceStore) GetStatus(ctx context.Context, userID uuid.UUID) (status string, lastSeen time.Time, err error) {
	key := presenceKey(userID)
	val, err := s.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return presenceStatusOffline, time.Time{}, nil
		}
		return "", time.Time{}, fmt.Errorf("presence get: %w", err)
	}
	ttl, _ := s.client.TTL(ctx, key).Result()
	lastSeen = time.Now().Add(-(presenceTTL - ttl))
	if ttl < 0 {
		lastSeen = time.Now()
	}
	if val == presenceStatusOnline {
		return presenceStatusOnline, lastSeen, nil
	}
	return presenceStatusOffline, lastSeen, nil
}

// PublishUpdate publishes a presence update to the presence:updates channel.
// Payload should be JSON: {"user_id":"<uuid>","status":"online|offline"}.
func (s *RedisPresenceStore) PublishUpdate(ctx context.Context, userID uuid.UUID, status string) error {
	payload := fmt.Sprintf(`{"user_id":"%s","status":"%s"}`, userID.String(), status)
	if err := s.client.Publish(ctx, presenceChannel, payload).Err(); err != nil {
		return fmt.Errorf("presence publish: %w", err)
	}
	return nil
}
