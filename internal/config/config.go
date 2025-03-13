// Copyright 2025 convexwf
//
// Project: uim-go
// File: config.go
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
// Description: Configuration management and environment variable loading

package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all application configuration.
type Config struct {
	App       AppConfig
	Database  DatabaseConfig
	Redis     RedisConfig
	JWT       JWTConfig
	CORS      CORSConfig
	RateLimit RateLimitConfig
}

// AppConfig holds application-level configuration.
type AppConfig struct {
	Env      string
	Port     string
	WSPort   string
	LogLevel string
}

// DatabaseConfig holds PostgreSQL database configuration.
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// DSN returns the PostgreSQL data source name (connection string).
func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.DBName, d.SSLMode)
}

// RedisConfig holds Redis configuration.
type RedisConfig struct {
	Host     string
	Port     string
	Password string
}

// Addr returns the Redis server address in "host:port" format.
func (r RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%s", r.Host, r.Port)
}

// JWTConfig holds JWT token configuration.
type JWTConfig struct {
	Secret        string
	AccessExpiry  time.Duration
	RefreshExpiry time.Duration
}

// CORSConfig holds CORS (Cross-Origin Resource Sharing) configuration.
type CORSConfig struct {
	AllowedOrigins []string
}

// RateLimitConfig holds rate limiting configuration.
type RateLimitConfig struct {
	Messages int
	Requests int
}

// Load loads configuration from environment variables.
//
// It attempts to load a .env file if present, then reads configuration
// from environment variables with sensible defaults.
//
// Returns:
//   - *Config: The loaded configuration
//   - error: An error if configuration loading fails
func Load() (*Config, error) {
	// Load .env file if it exists (ignore error if not found)
	_ = godotenv.Load()

	accessExpiry, err := parseDuration(getEnv("JWT_ACCESS_EXPIRY", "15m"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_ACCESS_EXPIRY: %w", err)
	}

	refreshExpiry, err := parseDuration(getEnv("JWT_REFRESH_EXPIRY", "168h"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_REFRESH_EXPIRY: %w", err)
	}

	allowedOrigins := getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000")
	corsOrigins := splitString(allowedOrigins, ",")

	return &Config{
		App: AppConfig{
			Env:      getEnv("APP_ENV", "development"),
			Port:     getEnv("APP_PORT", "8080"),
			WSPort:   getEnv("WS_PORT", "8081"),
			LogLevel: getEnv("LOG_LEVEL", "debug"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("POSTGRES_HOST", "localhost"),
			Port:     getEnv("POSTGRES_PORT", "5432"),
			User:     getEnv("POSTGRES_USER", "uim_user"),
			Password: getEnv("POSTGRES_PASSWORD", "uim_password"),
			DBName:   getEnv("POSTGRES_DB", "uim_db"),
			SSLMode:  getEnv("POSTGRES_SSLMODE", "disable"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
		},
		JWT: JWTConfig{
			Secret:        getEnv("JWT_SECRET", "your-secret-key-change-this-in-production"),
			AccessExpiry:  accessExpiry,
			RefreshExpiry: refreshExpiry,
		},
		CORS: CORSConfig{
			AllowedOrigins: corsOrigins,
		},
		RateLimit: RateLimitConfig{
			Messages: getEnvAsInt("RATE_LIMIT_MESSAGES", 50),
			Requests: getEnvAsInt("RATE_LIMIT_REQUESTS", 100),
		},
	}, nil
}

// getEnv retrieves an environment variable or returns a default value.
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt retrieves an environment variable as an integer or returns a default value.
func getEnvAsInt(key string, defaultValue int) int {
	value := getEnv(key, fmt.Sprintf("%d", defaultValue))
	var result int
	fmt.Sscanf(value, "%d", &result)
	return result
}

// parseDuration parses a duration string (e.g., "15m", "1h").
func parseDuration(s string) (time.Duration, error) {
	return time.ParseDuration(s)
}

// splitString splits a string by separator and trims whitespace from each part.
func splitString(s, sep string) []string {
	if s == "" {
		return []string{}
	}
	parts := strings.Split(s, sep)
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
