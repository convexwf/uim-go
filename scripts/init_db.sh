#!/usr/bin/env bash
# Copyright 2025 convexwf
# Idempotent database init: creates tables and indexes from SQL migrations.
# Run before starting the service. Uses POSTGRES_* env (or .env in project root).
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$PROJECT_ROOT"

# Load .env if present (same vars as app)
if [ -f .env ]; then
  set -a
  # shellcheck source=/dev/null
  source .env
  set +a
fi

export PGHOST="${POSTGRES_HOST:-localhost}"
export PGPORT="${POSTGRES_PORT:-5432}"
export PGUSER="${POSTGRES_USER:-uim_user}"
export PGPASSWORD="${POSTGRES_PASSWORD:-uim_password}"
export PGDATABASE="${POSTGRES_DB:-uim_db}"

MIGRATION_FILE="migrations/000001_initial_schema.up.sql"
if [ ! -f "$MIGRATION_FILE" ]; then
  echo "Error: $MIGRATION_FILE not found (run from project root or set correct path)" >&2
  exit 1
fi

if ! command -v psql &>/dev/null; then
  echo "Error: psql not found. Install PostgreSQL client." >&2
  exit 1
fi

echo "Initializing database at $PGHOST:$PGPORT/$PGDATABASE (idempotent)..."
psql -v ON_ERROR_STOP=1 -f "$MIGRATION_FILE"
echo "Done."
