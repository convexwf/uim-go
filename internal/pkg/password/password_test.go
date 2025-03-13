// Copyright 2025 convexwf
//
// Project: uim-go
// File: password_test.go
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
// Description: Unit tests for password hashing and verification

package password

import "testing"

func TestHash(t *testing.T) {
	password := "testpassword123"
	hash, err := Hash(password)
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}
	if hash == "" {
		t.Error("Hash() returned empty string")
	}
	if hash == password {
		t.Error("Hash() returned plain password")
	}
}

func TestVerify(t *testing.T) {
	password := "testpassword123"
	hash, err := Hash(password)
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}

	if !Verify(password, hash) {
		t.Error("Verify() failed for correct password")
	}

	if Verify("wrongpassword", hash) {
		t.Error("Verify() succeeded for wrong password")
	}
}

func TestHashDifferentResults(t *testing.T) {
	password := "testpassword123"
	hash1, _ := Hash(password)
	hash2, _ := Hash(password)

	// Bcrypt hashes should be different due to salt
	if hash1 == hash2 {
		t.Error("Hash() returned same hash for same password (should be different due to salt)")
	}

	// But both should verify correctly
	if !Verify(password, hash1) {
		t.Error("First hash verification failed")
	}
	if !Verify(password, hash2) {
		t.Error("Second hash verification failed")
	}
}

func TestHashEmptyPassword(t *testing.T) {
	hash, err := Hash("")
	if err != nil {
		t.Fatalf("Hash() with empty password should not error, got: %v", err)
	}
	if hash == "" {
		t.Error("Hash() returned empty string for empty password")
	}
}

func TestVerifyEmptyPassword(t *testing.T) {
	hash, _ := Hash("testpassword")
	if Verify("", hash) {
		t.Error("Verify() should fail for empty password")
	}
}

func TestVerifyEmptyHash(t *testing.T) {
	if Verify("password", "") {
		t.Error("Verify() should fail for empty hash")
	}
}

func TestHashLongPassword(t *testing.T) {
	// Bcrypt has a 72-byte limit, so test with a password just under that limit
	longPassword := make([]byte, 70)
	for i := range longPassword {
		longPassword[i] = 'a'
	}
	
	hash, err := Hash(string(longPassword))
	if err != nil {
		t.Fatalf("Hash() with long password error = %v", err)
	}
	if hash == "" {
		t.Error("Hash() returned empty string for long password")
	}
	
	if !Verify(string(longPassword), hash) {
		t.Error("Verify() failed for long password")
	}
}

func TestHashPasswordExceedsBcryptLimit(t *testing.T) {
	// Bcrypt has a 72-byte limit, test that passwords exceeding this limit return an error
	longPassword := make([]byte, 100)
	for i := range longPassword {
		longPassword[i] = 'a'
	}
	
	_, err := Hash(string(longPassword))
	if err == nil {
		t.Error("Hash() should return error for password exceeding bcrypt 72-byte limit")
	}
}
