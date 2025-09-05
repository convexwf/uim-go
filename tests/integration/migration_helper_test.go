package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gorm.io/gorm"
)

func applyAllMigrationsForTest(t *testing.T, db *gorm.DB) {
	t.Helper()
	paths, err := filepath.Glob("../../migrations/*.up.sql")
	if err != nil {
		t.Fatalf("glob migrations: %v", err)
	}
	if len(paths) == 0 {
		t.Fatalf("no migration files found")
	}
	raw, err := db.DB()
	if err != nil {
		t.Fatalf("db handle: %v", err)
	}
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read migration %s: %v", path, err)
		}
		for _, stmt := range strings.Split(string(data), ";") {
			stmt = stripSQLCommentsForTest(strings.TrimSpace(stmt))
			if stmt == "" {
				continue
			}
			if _, err := raw.Exec(stmt); err != nil {
				t.Fatalf("apply migration %s: %v", path, err)
			}
		}
	}
}

func stripSQLCommentsForTest(s string) string {
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
