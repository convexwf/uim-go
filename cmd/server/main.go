// Copyright 2025 convexwf
//
// Project: uim-go
// File: main.go
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
// Description: Application entry point and server initialization

package main

import (
	"log"

	"github.com/convexwf/uim-go/internal/api"
	"github.com/convexwf/uim-go/internal/config"
	"github.com/convexwf/uim-go/internal/middleware"
	"github.com/convexwf/uim-go/internal/model"
	"github.com/convexwf/uim-go/internal/pkg/jwt"
	"github.com/convexwf/uim-go/internal/repository"
	"github.com/convexwf/uim-go/internal/service"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database
	db, err := initDatabase(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Run migrations
	if err := runMigrations(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize JWT manager
	jwtManager := jwt.NewJWTManager(
		cfg.JWT.Secret,
		cfg.JWT.AccessExpiry,
		cfg.JWT.RefreshExpiry,
	)

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)

	// Initialize services
	authService := service.NewAuthService(userRepo, jwtManager)

	// Setup router
	router := api.SetupRouter(db, authService)

	// Apply middleware
	router.Use(middleware.CORSMiddleware(cfg))
	router.Use(middleware.LoggerMiddlewareSimple())
	router.Use(middleware.ErrorHandlerMiddleware())

	// Start server
	log.Printf("Server starting on port %s", cfg.App.Port)
	if err := router.Run(":" + cfg.App.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// initDatabase initializes the PostgreSQL database connection.
func initDatabase(cfg *config.Config) (*gorm.DB, error) {
	var logLevel logger.LogLevel
	switch cfg.App.LogLevel {
	case "debug":
		logLevel = logger.Info
	case "info":
		logLevel = logger.Warn
	default:
		logLevel = logger.Error
	}

	db, err := gorm.Open(postgres.Open(cfg.Database.DSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		return nil, err
	}

	return db, nil
}

// runMigrations runs database migrations using GORM AutoMigrate.
func runMigrations(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.User{},
		&model.Conversation{},
		&model.ConversationParticipant{},
		&model.Message{},
	)
}
