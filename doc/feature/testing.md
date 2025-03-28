# Testing Documentation

| Feature | Status | Date       |
| ------- | ------ | ---------- |
| Unit tests | ✅ | 2025-03-13 |
| Integration tests | ✅ | 2025-03-13 |
| Test layout | ✅ | 2025-03-13 |

---

## Overview

The project uses the **standard Go testing package** (`testing`). No extra test framework is required. Unit tests live next to code in `internal/`; integration tests live in **`tests/integration/`**.

---

## Layout

| Type | Location | Run with |
| ---- | -------- | -------- |
| **Unit tests** | `internal/*/*_test.go` (next to code) | `make test` |
| **Integration tests** | `tests/integration/*_test.go` | `make test-integration` |

**Convention**:
- **Unit tests**: Same package as the code (e.g. `package api` in `internal/api/auth_handler_test.go`), or `package foo_test` for black-box tests. Use mocks where needed; no DB required. Skipped when `go test -short` is used (so `make test` runs only unit tests).
- **Integration tests**: Package `integration` under `tests/integration/`. They import `internal/*` and hit real DB or HTTP. They skip when `-short` or when DB is unavailable.

---

## Commands

- **`make test`** – Unit tests only: `go test ./internal/... -short -v`. Fast, no DB.
- **`make test-integration`** – Integration tests only: `go test ./tests/... -v`. Requires DB; run `make init-db` and optionally `make seed-db` first.
- **`make test-cover`** – Unit tests with coverage: `go test ./internal/... -cover -coverprofile=coverage.out` and HTML report.

---

## Integration Tests

**Location**: `tests/integration/`

**Contents**:
- **`auth_test.go`** – Auth endpoints: register, login, refresh (HTTP + real DB). Skips if `-short` or DB not available.
- **`db_performance_test.go`** – Light DB check: `GetByUsername` × 20 with sample data. Requires DB and seed (e.g. user `alice`). Skips if `-short` or DB/seed missing.

**Requirements**:
- PostgreSQL available (e.g. `docker-compose up -d postgres`).
- Schema created: `make init-db`.
- For DB performance test: `make seed-db` (creates alice, bob, test).

**Same DB as seed_db**: Integration tests load `.env` from the **project root** (`../../.env`), because `go test` runs with cwd = `tests/integration`. That way they use the same `POSTGRES_*` as `make seed-db` (which runs from project root). If you had run `make seed-db` but the performance test still skipped with "alice not found", it was likely because tests were loading a different config; this is now fixed.

**Skipping**: Integration tests call `t.Skip(...)` when config or DB is missing, so `make test-integration` does not fail in environments without DB.

---

## Unit Tests

**Location**: `internal/*/*_test.go`

**Examples**:
- `internal/pkg/password/password_test.go` – Hash/verify
- `internal/pkg/jwt/jwt_test.go` – Token generation/validation
- `internal/service/auth_service_test.go` – Auth logic with mock repository

**Running**: `make test` or `go test ./internal/... -short -v`.

---

## Related

- [Project rules](./project-rules.md) – Binaries in `bin/`, etc.
- [Initialization](./initialization.md) – Setup and seed
- [Database migrations](./database-migrations.md) – Schema and init-db
