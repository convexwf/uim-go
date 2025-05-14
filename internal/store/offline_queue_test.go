// Copyright 2025 convexwf
//
// Project: uim-go
// File: offline_queue_test.go
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// See the License for the full terms.
//
// Description: Unit tests for Redis offline queue

package store

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

func setupRedisOfflineQueue(t *testing.T) (*RedisOfflineQueue, *miniredis.Miniredis) {
	t.Helper()
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis: %v", err)
	}
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	q := NewRedisOfflineQueue(client)
	t.Cleanup(func() { mr.Close(); _ = client.Close() })
	return q, mr
}

func TestOfflineQueue_Push_PopAll(t *testing.T) {
	q, _ := setupRedisOfflineQueue(t)
	ctx := context.Background()
	userID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")

	// Empty queue
	got, err := q.PopAll(ctx, userID)
	if err != nil {
		t.Fatalf("PopAll empty: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("PopAll empty: got %d messages", len(got))
	}

	// Push one
	if err := q.Push(ctx, userID, []byte(`{"type":"new_message"}`)); err != nil {
		t.Fatalf("Push: %v", err)
	}
	got, err = q.PopAll(ctx, userID)
	if err != nil {
		t.Fatalf("PopAll one: %v", err)
	}
	if len(got) != 1 || string(got[0]) != `{"type":"new_message"}` {
		t.Errorf("PopAll one: got %v", got)
	}

	// After PopAll queue is empty
	got, _ = q.PopAll(ctx, userID)
	if len(got) != 0 {
		t.Errorf("PopAll after clear: got %d", len(got))
	}
}

func TestOfflineQueue_Order(t *testing.T) {
	q, _ := setupRedisOfflineQueue(t)
	ctx := context.Background()
	userID := uuid.New()

	// Push oldest first (simulate messages arriving in order)
	if err := q.Push(ctx, userID, []byte("msg1")); err != nil {
		t.Fatal(err)
	}
	if err := q.Push(ctx, userID, []byte("msg2")); err != nil {
		t.Fatal(err)
	}
	if err := q.Push(ctx, userID, []byte("msg3")); err != nil {
		t.Fatal(err)
	}

	// PopAll returns oldest first (msg1, msg2, msg3)
	got, err := q.PopAll(ctx, userID)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 {
		t.Fatalf("got %d messages", len(got))
	}
	if string(got[0]) != "msg1" || string(got[1]) != "msg2" || string(got[2]) != "msg3" {
		t.Errorf("order: got %q %q %q", got[0], got[1], got[2])
	}
}

func TestOfflineQueue_Len(t *testing.T) {
	q, _ := setupRedisOfflineQueue(t)
	ctx := context.Background()
	userID := uuid.New()

	n, err := q.Len(ctx, userID)
	if err != nil || n != 0 {
		t.Errorf("Len empty: n=%d err=%v", n, err)
	}
	_ = q.Push(ctx, userID, []byte("a"))
	_ = q.Push(ctx, userID, []byte("b"))
	n, err = q.Len(ctx, userID)
	if err != nil || n != 2 {
		t.Errorf("Len 2: n=%d err=%v", n, err)
	}
}
