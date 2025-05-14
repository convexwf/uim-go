// Copyright 2025 convexwf
//
// Project: uim-go
// File: retry.go
// Email: convexwf@gmail.com
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// See the License for the full terms.
//
// Description: Simple retry helper for transient failures

package retry

import (
	"strings"
	"time"
)

// IsRetryableError returns true for errors that may succeed on retry (e.g. connection timeout, temporary failure).
func IsRetryableError(err error) bool {
	if err == nil {
		return false
	}
	s := strings.ToLower(err.Error())
	for _, sub := range []string{"timeout", "connection refused", "connection reset", "temporary failure", "i/o timeout", "dial tcp", "connection closed", "broken pipe"} {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}

// Do retries fn up to maxAttempts times with exponential backoff (base, 2*base, 4*base).
// Only retries when IsRetryableError returns true for the error.
func Do(maxAttempts int, baseDelay time.Duration, fn func() error) error {
	var lastErr error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		lastErr = fn()
		if lastErr == nil {
			return nil
		}
		if attempt == maxAttempts-1 {
			return lastErr
		}
		if !IsRetryableError(lastErr) {
			return lastErr
		}
		time.Sleep(baseDelay)
		baseDelay *= 2
	}
	return lastErr
}
