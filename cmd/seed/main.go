// Copyright 2025 convexwf
//
// Project: uim-go
// File: main.go
// Description: Idempotent DB seed: creates test users for local/dev. Run after init_db.sh.
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/convexwf/uim-go/internal/config"
	"github.com/convexwf/uim-go/internal/model"
	"github.com/convexwf/uim-go/internal/pkg/password"
	"github.com/convexwf/uim-go/internal/repository"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// testUsers are created if they do not exist. Password for all: "password123"
var testUsers = []struct {
	Username    string
	Email       string
	DisplayName string
}{
	{"alice", "alice@example.com", "Alice"},
	{"bob", "bob@example.com", "Bob"},
	{"test", "test@example.com", "Test User"},
}

const seedPassword = "password123"

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	db, err := gorm.Open(postgres.Open(cfg.Database.DSN()), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	repo := repository.NewUserRepository(db)
	hash, err := password.Hash(seedPassword)
	if err != nil {
		log.Fatalf("Failed to hash seed password: %v", err)
	}
	for _, u := range testUsers {
		existing, _ := repo.GetByUsername(u.Username)
		if existing != nil {
			fmt.Fprintf(os.Stderr, "User %q already exists, skip\n", u.Username)
			continue
		}
		user := &model.User{
			Username:     u.Username,
			Email:        u.Email,
			PasswordHash: hash,
			DisplayName:  u.DisplayName,
		}
		if err := repo.Create(user); err != nil {
			log.Fatalf("Failed to create user %q: %v", u.Username, err)
		}
		fmt.Fprintf(os.Stderr, "Created user %q\n", u.Username)
	}
	fmt.Fprintln(os.Stderr, "Seed done.")
}
