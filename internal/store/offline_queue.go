// Copyright 2025 convexwf
//
// Project: uim-go
// File: offline_queue.go
// Email: convexwf@gmail.com
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// See the License for the full terms.
//
// Description: Offline message queue interface and Redis implementation

package store

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/convexwf/uim-go/internal/pkg/retry"
)

const (
	offlineKeyPrefix = "offline:user:"
	offlineQueueTTL  = 24 * time.Hour
)

// OfflineQueue stores messages for delivery when a user reconnects.
// Implementations may be nil-safe; callers should check for nil.
type OfflineQueue interface {
	Push(ctx context.Context, userID uuid.UUID, message []byte) error
	PopAll(ctx context.Context, userID uuid.UUID) ([][]byte, error)
	Len(ctx context.Context, userID uuid.UUID) (int64, error)
}

// RedisOfflineQueue implements OfflineQueue using Redis lists.
type RedisOfflineQueue struct {
	client redis.Cmdable
}

// NewRedisOfflineQueue creates an offline queue backed by Redis.
func NewRedisOfflineQueue(client redis.Cmdable) *RedisOfflineQueue {
	return &RedisOfflineQueue{client: client}
}

func offlineKey(userID uuid.UUID) string {
	return offlineKeyPrefix + userID.String()
}

// Push appends a message to the user's offline queue and refreshes TTL. Retries once on transient failure.
func (q *RedisOfflineQueue) Push(ctx context.Context, userID uuid.UUID, message []byte) error {
	return retry.Do(2, 50*time.Millisecond, func() error {
		key := offlineKey(userID)
		pipe := q.client.Pipeline()
		pipe.LPush(ctx, key, message)
		pipe.Expire(ctx, key, offlineQueueTTL)
		_, err := pipe.Exec(ctx)
		if err != nil {
			return fmt.Errorf("offline queue push: %w", err)
		}
		return nil
	})
}

// PopAll returns all messages for the user and removes them from the queue.
// Messages are returned in FIFO order (oldest first). The key is deleted after read.
func (q *RedisOfflineQueue) PopAll(ctx context.Context, userID uuid.UUID) ([][]byte, error) {
	key := offlineKey(userID)
	// LRANGE 0 -1 returns all; list is LPUSH so index 0 is newest, -1 is oldest. We want oldest first for delivery order.
	vals, err := q.client.LRange(ctx, key, 0, -1).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, fmt.Errorf("offline queue lrange: %w", err)
	}
	if len(vals) == 0 {
		return nil, nil
	}
	// Reverse so oldest (last in list) is first in result
	result := make([][]byte, len(vals))
	for i := range vals {
		result[len(vals)-1-i] = []byte(vals[i])
	}
	if err := q.client.Del(ctx, key).Err(); err != nil {
		return nil, fmt.Errorf("offline queue del: %w", err)
	}
	return result, nil
}

// Len returns the number of messages in the user's offline queue.
func (q *RedisOfflineQueue) Len(ctx context.Context, userID uuid.UUID) (int64, error) {
	n, err := q.client.LLen(ctx, offlineKey(userID)).Result()
	if err != nil {
		if err == redis.Nil {
			return 0, nil
		}
		return 0, fmt.Errorf("offline queue llen: %w", err)
	}
	return n, nil
}
