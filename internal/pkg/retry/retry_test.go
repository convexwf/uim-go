// Copyright 2025 convexwf
//
// Project: uim-go
// File: retry_test.go
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// See the License for the full terms.
//
// Description: Unit tests for retry helper

package retry

import (
	"errors"
	"testing"
)

func TestIsRetryableError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"timeout", errors.New("dial tcp: i/o timeout"), true},
		{"connection refused", errors.New("connection refused"), true},
		{"connection reset", errors.New("connection reset by peer"), true},
		{"validation", errors.New("invalid input"), false},
		{"not found", errors.New("record not found"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsRetryableError(tt.err); got != tt.want {
				t.Errorf("IsRetryableError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDo_Success(t *testing.T) {
	attempts := 0
	err := Do(3, 0, func() error {
		attempts++
		return nil
	})
	if err != nil {
		t.Errorf("Do() = %v", err)
	}
	if attempts != 1 {
		t.Errorf("expected 1 attempt, got %d", attempts)
	}
}

func TestDo_RetryThenSuccess(t *testing.T) {
	attempts := 0
	err := Do(3, 0, func() error {
		attempts++
		if attempts < 2 {
			return errors.New("i/o timeout")
		}
		return nil
	})
	if err != nil {
		t.Errorf("Do() = %v", err)
	}
	if attempts != 2 {
		t.Errorf("expected 2 attempts, got %d", attempts)
	}
}

func TestDo_NonRetryableFailsImmediately(t *testing.T) {
	attempts := 0
	err := Do(3, 0, func() error {
		attempts++
		return errors.New("invalid input")
	})
	if err == nil {
		t.Error("Do() expected error")
	}
	if attempts != 1 {
		t.Errorf("expected 1 attempt (no retry), got %d", attempts)
	}
}
