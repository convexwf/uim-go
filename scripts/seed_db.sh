#!/usr/bin/env bash
# Copyright 2025 convexwf
# Idempotent seed: creates test users (alice, bob, test) with password "password123".
# Run after init_db.sh. Uses POSTGRES_* env or .env.
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$PROJECT_ROOT"

if [ -f .env ]; then
  set -a
  # shellcheck source=/dev/null
  source .env
  set +a
fi

echo "Seeding database (idempotent)..."
# Always build then run; do not use "go run" (see .cursor/rules/ and doc/feature/project-rules.md)
make -C "$PROJECT_ROOT" build-seed
./bin/seed
echo "Done."
