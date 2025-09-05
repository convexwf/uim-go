// Copyright 2025 convexwf
//
// Project: uim-go
// File: main.go
// Email: convexwf@gmail.com
// Created: 2025-03-13
// Last modified: 2025-09-05
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
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/convexwf/uim-go/internal/api"
	"github.com/convexwf/uim-go/internal/config"
	"github.com/convexwf/uim-go/internal/pkg/jwt"
	"github.com/convexwf/uim-go/internal/repository"
	"github.com/convexwf/uim-go/internal/service"
	"github.com/convexwf/uim-go/internal/store"
	"github.com/convexwf/uim-go/internal/websocket"
	"github.com/redis/go-redis/v9"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// openLogOutput returns a writer for the standard library log and DB logger.
// If UIM_LOG_FILE is set, logs are duplicated to that file (and its parent dirs are created).
func openLogOutput() (io.Writer, func(), error) {
	path := strings.TrimSpace(os.Getenv("UIM_LOG_FILE"))
	if path == "" {
		return os.Stdout, func() {}, nil
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, nil, fmt.Errorf("create log dir: %w", err)
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, nil, fmt.Errorf("open log file: %w", err)
	}
	mw := io.MultiWriter(os.Stdout, f)
	return mw, func() { _ = f.Close() }, nil
}

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	logOut, closeLog, err := openLogOutput()
	if err != nil {
		log.Fatalf("Failed to set up logging: %v", err)
	}
	defer closeLog()
	log.SetOutput(logOut)

	// Initialize database
	db, err := initDatabase(cfg, logOut)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Schema is created by scripts/init_db.sh (SQL first). Here we only check; optional fallback.
	if err := ensureSchema(db, cfg); err != nil {
		log.Fatalf("Schema check failed: %v", err)
	}

	// Initialize JWT manager
	jwtManager := jwt.NewJWTManager(
		cfg.JWT.Secret,
		cfg.JWT.AccessExpiry,
		cfg.JWT.RefreshExpiry,
	)

	// Initialize Redis (optional: on failure, offline queue and presence are disabled)
	var redisClient *redis.Client
	redisOpts := &redis.Options{
		Addr:     cfg.Redis.Addr(),
		Password: cfg.Redis.Password,
		DB:       0,
	}
	redisClient = redis.NewClient(redisOpts)
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		log.Printf("Redis not available (offline queue and presence disabled): %v", err)
		redisClient = nil
	}

	var offlineQueue store.OfflineQueue
	var presenceStore store.PresenceStore
	if redisClient != nil {
		offlineQueue = store.NewRedisOfflineQueue(redisClient)
		presenceStore = store.NewRedisPresenceStore(redisClient)
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	convRepo := repository.NewConversationRepository(db)
	msgRepo := repository.NewMessageRepository(db)

	// Initialize services
	authService := service.NewAuthService(userRepo, jwtManager)
	convSvc := service.NewConversationService(convRepo, userRepo, msgRepo)
	hub := websocket.NewHub(convRepo, offlineQueue)
	msgSvc := service.NewMessageService(msgRepo, convSvc, hub)
	router := api.SetupRouter(cfg, db, authService, jwtManager, convSvc, msgSvc, hub, redisClient, offlineQueue, presenceStore)

	// Start server
	log.Printf("Server starting on port %s", cfg.App.Port)
	if err := router.Run(":" + cfg.App.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// dbLogWriter prefixes each log line with [DB] for consistent log format.
type dbLogWriter struct {
	w      io.Writer
	prefix string
}

func (d *dbLogWriter) Write(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}
	_, err = d.w.Write(append([]byte(d.prefix), p...))
	return len(p), err
}

// initDatabase initializes the PostgreSQL database connection.
func initDatabase(cfg *config.Config, logOut io.Writer) (*gorm.DB, error) {
	var logLevel logger.LogLevel
	switch cfg.App.LogLevel {
	case "debug":
		logLevel = logger.Info
	case "info":
		logLevel = logger.Warn
	default:
		logLevel = logger.Error
	}

	dbLogger := logger.New(
		log.New(&dbLogWriter{w: logOut, prefix: "[DB] "}, "", log.LstdFlags),
		logger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  logLevel,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)

	db, err := gorm.Open(postgres.Open(cfg.Database.DSN()), &gorm.Config{
		Logger: dbLogger,
	})
	if err != nil {
		return nil, err
	}

	return db, nil
}

// ensureSchema checks that required tables exist. If not, exits with a message to run init_db.sh,
// unless AUTO_MIGRATE_FALLBACK=1 (and not production), in which case it runs the SQL migration as fallback.
func ensureSchema(db *gorm.DB, cfg *config.Config) error {
	if checkSchemaExists(db) {
		return nil
	}
	fallback := os.Getenv("AUTO_MIGRATE_FALLBACK")
	if cfg.App.Env == "production" || fallback == "" || (fallback != "1" && fallback != "true" && fallback != "yes") {
		return fmt.Errorf("schema not ready: run scripts/init_db.sh before starting the service")
	}
	log.Print("Schema missing; applying fallback migration (AUTO_MIGRATE_FALLBACK is set)...")
	if err := runSQLMigration(db); err != nil {
		return fmt.Errorf("fallback migration failed: %w", err)
	}
	if !checkSchemaExists(db) {
		return fmt.Errorf("schema still missing after fallback migration")
	}
	log.Print("Fallback migration completed.")
	return nil
}

// checkSchemaExists returns true if required tables exist (e.g. users).
func checkSchemaExists(db *gorm.DB) bool {
	var exists bool
	err := db.Raw("SELECT EXISTS(SELECT 1 FROM information_schema.tables WHERE table_schema = CURRENT_SCHEMA() AND table_name = 'users')").Scan(&exists).Error
	return err == nil && exists
}

// runSQLMigration runs the initial schema SQL file (used only as fallback when AUTO_MIGRATE_FALLBACK=1).
func runSQLMigration(db *gorm.DB) error {
	const path = "migrations/000001_initial_schema.up.sql"
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("migration file not found: %s (run init_db.sh instead)", path)
		}
		return err
	}
	raw, err := db.DB()
	if err != nil {
		return err
	}
	for _, s := range strings.Split(string(data), ";") {
		s = stripSQLComments(strings.TrimSpace(s))
		if s == "" {
			continue
		}
		if _, err := raw.Exec(s); err != nil {
			return err
		}
	}
	return nil
}

func stripSQLComments(s string) string {
	var b strings.Builder
	for _, line := range strings.Split(s, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" && !strings.HasPrefix(trimmed, "--") {
			b.WriteString(line)
			b.WriteString("\n")
		}
	}
	return strings.TrimSpace(b.String())
}
