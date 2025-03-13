// Copyright 2025 convexwf
//
// Project: uim-go
// File: password.go
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
// Description: Password hashing and verification utilities using bcrypt

package password

import (
	"golang.org/x/crypto/bcrypt"
)

const (
	// DefaultCost is the default bcrypt cost factor used for password hashing.
	DefaultCost = 10
)

// Hash hashes a password using bcrypt with the default cost factor.
//
// The password is hashed using bcrypt's GenerateFromPassword function,
// which automatically generates a salt and includes it in the resulting hash.
//
// Parameters:
//   - password: The plain text password to hash
//
// Returns:
//   - string: The bcrypt hash of the password
//   - error: An error if hashing fails
func Hash(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// Verify verifies a password against a bcrypt hash.
//
// This function compares the provided password with the stored hash
// using bcrypt's CompareHashAndPassword function.
//
// Parameters:
//   - password: The plain text password to verify
//   - hash: The bcrypt hash to compare against
//
// Returns:
//   - bool: true if the password matches the hash, false otherwise
func Verify(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
