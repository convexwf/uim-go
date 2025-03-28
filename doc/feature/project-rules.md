# Project Rules

| Rule | Status | Date       |
| ---- | ------ | ---------- |
| Binaries in bin/ | ✅ | 2025-03-13 |
| Build then run (all runnables) | ✅ | 2025-03-13 |
| Long-term vs temporary | ✅ | 2025-03-13 |
| Backup rules in project-rules.md | ✅ | 2025-03-13 |

---

**Authoritative rules for the AI are in `.cursor/rules/`** (Cursor rules). **This doc is the backup:** every rule in `.cursor/rules/` must have a corresponding section or summary here; when you add or change a rule in .cursor/rules/, update this doc so it stays in sync.

---

## Binaries Must Go in `bin/`

**Rule**: All compiled binaries must be built into the **`bin/`** directory. No binaries are allowed in the project root or other source directories.

**Examples**:
- **Server**: `go build -o bin/uim-server ./cmd/server` (Makefile: `make build`)
- **Seed**: `go build -o bin/seed ./cmd/seed` (Makefile: `make build-seed`)

**Enforcement**:
- `.gitignore` ignores root-level `/seed` and `/server` so they are never committed.
- **Do not** run `go build ./cmd/...` without `-o bin/...`; that leaves a binary in the current directory.

---

## Build Then Run (All runnables, no `go run`)

**Rule**: For **every** runnable (server, seed, or any future cmd), scripts, Makefile, and documentation must **build then run**. Do not use `go run ./cmd/...` anywhere.

**Correct**: Build (e.g. `make build` / `make build-seed` or `go build -o bin/<name> ./cmd/<name>`), then run `./bin/<name>`.

**Enforcement**:
- `scripts/seed_db.sh` runs `make build-seed` then `./bin/seed`.
- Any other script or doc that runs a cmd must follow the same pattern (build to bin/, then execute bin/).

---

## Long-term vs temporary

**Rule**: Do not add temporary or one-off fixes to long-term maintained files (e.g. Makefile, core config). Keep Makefile and similar files for stable, lasting targets only. Fix the root cause or document the rule instead of special-case cleanup.

*(Backup of `.cursor/rules/maintainability.mdc`)*

---

## Backup rules in project-rules.md

**Rule**: Every rule added or changed in `.cursor/rules/` must be backed up in this doc (`doc/feature/project-rules.md`). When adding a new .mdc under .cursor/rules/, add a corresponding section here; when editing an existing rule, update the backup here so the doc stays in sync.

*(Backup of `.cursor/rules/backup-rules.mdc`)*

---

## Related

- [Testing](./testing.md) – Unit and integration test layout and commands
- [Initialization](./initialization.md) – Project structure and setup
