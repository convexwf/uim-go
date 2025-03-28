// Copyright 2025 convexwf
//
// Project: uim-go
// File: db_performance_test.go
// Description: Lightweight DB performance check with sample data. Run after seed_db.sh.
// Skip when -short or when DB is not available (e.g. unset POSTGRES_HOST).
package integration

import (
	"testing"
	"time"

	"github.com/convexwf/uim-go/internal/config"
	"github.com/convexwf/uim-go/internal/repository"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// TestDatabasePerformanceWithSampleData runs key queries and asserts they complete in reasonable time.
// Requires DB and sample data (run init_db.sh and seed_db.sh first). Skip with -short.
func TestDatabasePerformanceWithSampleData(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping DB performance test in short mode")
	}
	// Load .env from project root so we use the same DB as seed_db (go test cwd is tests/integration)
	_ = godotenv.Load("../../.env")
	cfg, err := config.Load()
	if err != nil {
		t.Skipf("config not loadable (missing .env?): %v", err)
	}
	db, err := gorm.Open(postgres.Open(cfg.Database.DSN()), &gorm.Config{})
	if err != nil {
		t.Skipf("DB not available: %v", err)
	}
	repo := repository.NewUserRepository(db)
	const iterations = 20
	const maxTotalMs = 2000 // 20 queries, ~100ms each acceptable
	start := time.Now()
	for i := 0; i < iterations; i++ {
		_, err := repo.GetByUsername("alice")
		if err != nil {
			t.Skipf("sample user 'alice' not found (run scripts/seed_db.sh first): %v", err)
		}
	}
	elapsed := time.Since(start)
	if elapsed > maxTotalMs*time.Millisecond {
		t.Errorf("GetByUsername x%d took %v (max %dms total)", iterations, elapsed, maxTotalMs)
	}
	t.Logf("GetByUsername x%d: %v", iterations, elapsed)
}
