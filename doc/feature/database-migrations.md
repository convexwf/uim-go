# Database Migrations Documentation

| Feature             | Status | Date       |
| ------------------- | ------ | ---------- |
| Migration Strategy  | ✅      | 2025-03-13 |
| init_db.sh          | ✅      | 2025-03-13 |
| Schema Check        | ✅      | 2025-03-13 |
| Optional Fallback   | ✅      | 2025-03-13 |

---

## Table of Contents

- [Overview](#overview)
- [Roles](#roles)
- [Primary Path: init_db.sh](#primary-path-init_dbsh)
- [Application Startup: Check Only](#application-startup-check-only)
- [Optional Fallback (Non-Production Only)](#optional-fallback-non-production-only)
- [GORM and Index Naming](#gorm-and-index-naming)
- [SQL Migration Files](#sql-migration-files)
- [Running Migrations in Practice](#running-migrations-in-practice)
  - [Before first start (or after schema changes)](#before-first-start-or-after-schema-changes)
  - [Without init_db.sh (dev only, optional)](#without-init_dbsh-dev-only-optional)
  - [Production](#production)
- [Related Documentation](#related-documentation)

---

## Overview

Schema is **SQL-first**: tables and indexes are created by running **`scripts/init_db.sh`** before starting the service. The application does **not** run migrations by default; it only **checks** that the schema exists. GORM is used for **assistant check** and, in non-production, an **optional fallback** when the schema is missing.

## Roles

| Role              | Responsibility                          | When |
| ----------------- | --------------------------------------- | ---- |
| **SQL + init_db.sh** | Create/update tables and indexes (single source of truth) | Before first start or after pulling new migrations |
| **Application**   | Check that required tables exist        | Every startup |
| **GORM**          | No AutoMigrate by default; fallback only when explicitly enabled | Only if `AUTO_MIGRATE_FALLBACK=1` and schema missing and not production |

## Primary Path: init_db.sh

**Run before starting the service** (locally or in deployment):

```bash
# From project root; uses POSTGRES_* from env or .env
./scripts/init_db.sh
```

- **Idempotent**: Safe to run multiple times. Uses `CREATE TABLE IF NOT EXISTS` and `CREATE INDEX IF NOT EXISTS` in `migrations/000001_initial_schema.up.sql`.
- **Connection**: Uses same env as the app: `POSTGRES_HOST`, `POSTGRES_PORT`, `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB`. Loads `.env` from project root if present.
- **Requirement**: `psql` (PostgreSQL client) must be installed. If Postgres runs in Docker and the host has no `psql`, use **`make init-db`**: it runs the migration inside the `uim-postgres` container.

## Application Startup: Check Only

On startup the server:

1. Connects to the database.
2. **Checks** that required schema exists (e.g. `users` table in `information_schema.tables`).
3. If check passes → continues and serves traffic.
4. If check fails → exits with: **"schema not ready: run scripts/init_db.sh before starting the service"**.

No SQL is executed and no GORM AutoMigrate runs in the normal path.

## Optional Fallback (Non-Production Only)

If you **do not** run `init_db.sh` (e.g. quick local try) and still want the app to create the schema:

- Set **`AUTO_MIGRATE_FALLBACK=1`** (or `true` or `yes`).
- **Do not** set this in production; fallback is disabled when `APP_ENV=production`.

When fallback is enabled and the schema check fails, the app will:

1. Run the same SQL file (`migrations/000001_initial_schema.up.sql`) from disk (same statements as `init_db.sh`).
2. Re-run the schema check; if it still fails, the app exits with an error.

So the fallback is **SQL execution from the app**, not GORM AutoMigrate. This keeps a single source of truth (the SQL file) and avoids GORM type issues (e.g. `Message.conversation_id` as UUID).

## GORM and Index Naming

GORM models use **named indexes** so that, if in the future AutoMigrate is used in any form, it would not conflict with indexes created by SQL:

- Same index names in SQL and GORM (e.g. `idx_users_username`, `idx_messages_conversation_time`).
- SQL uses `IF NOT EXISTS`; GORM would skip creating indexes that already exist by name.

Currently the app does **not** call `db.AutoMigrate(...)` in the normal or fallback path; the fallback runs the raw SQL file only.

## SQL Migration Files

- **Location**: `migrations/000001_initial_schema.up.sql` (and `.down.sql` for rollback).
- **Content**: Tables and indexes with `IF NOT EXISTS`; index names match the names used in GORM models where applicable.
- **Running**: Prefer **`./scripts/init_db.sh`**; fallback path runs the same `.up.sql` from the app when `AUTO_MIGRATE_FALLBACK=1`.

## Running Migrations in Practice

### Before first start (or after schema changes)

```bash
# Ensure Postgres is up (e.g. docker-compose up -d postgres)
# Optionally copy .env.example to .env and set POSTGRES_*
./scripts/init_db.sh
# Then start the app (e.g. make run or docker-compose up -d uim-server)
```

### Without init_db.sh (dev only, optional)

```bash
export AUTO_MIGRATE_FALLBACK=1
export APP_ENV=development
make run
# App will apply migrations/000001_initial_schema.up.sql if schema is missing
```

### Production

- Always run **`init_db.sh`** (or equivalent SQL) as part of deployment before starting the service.
- Do **not** set `AUTO_MIGRATE_FALLBACK` in production.

## Related Documentation

- [Initialization Feature](./initialization.md) – Project setup and schema overview
- [System Design v1.0](../design/v1.0/uim-system-design-v1.0.md) – Architecture
