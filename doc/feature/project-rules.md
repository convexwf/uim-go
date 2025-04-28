# Project Rules

| Rule                              | Status | Date       |
| --------------------------------- | ------ | ---------- |
| Binaries in bin/                  | ✅      | 2025-03-13 |
| Build then run (all runnables)    | ✅      | 2025-03-13 |
| Long-term vs temporary            | ✅      | 2025-03-13 |
| Backup rules in project-rules.md  | ✅      | 2025-03-13 |
| Table of Contents (doc/, to ####) | ✅      | 2025-03-27 |
| Diagrams in doc/ use Mermaid      | ✅      | 2025-04-26 |
| Directory tree format in doc/     | ✅      | 2025-04-26 |
| Backend logging convention        | ✅      | 2026-02-04 |

---

**Authoritative rules for the AI are in `.cursor/rules/`** (Cursor rules). We use **two rule files**: `project-rules.mdc` (build, backup, maintainability) and `doc-rules.mdc` (TOC, Mermaid, directory tree). **This doc is the backup:** every rule in `.cursor/rules/` must have a corresponding section or summary here; when you add or change a rule, update this doc so it stays in sync.

---

## Table of Contents

- [Project Rules](#project-rules)
  - [Table of Contents](#table-of-contents)
  - [Binaries Must Go in `bin/`](#binaries-must-go-in-bin)
  - [Build Then Run (All runnables, no `go run`)](#build-then-run-all-runnables-no-go-run)
  - [Long-term vs temporary](#long-term-vs-temporary)
  - [Backup rules in project-rules.md](#backup-rules-in-project-rulesmd)
  - [Table of Contents in documentation](#table-of-contents-in-documentation)
  - [Diagrams in doc/ use Mermaid](#diagrams-in-doc-use-mermaid)
  - [Directory tree format in doc/](#directory-tree-format-in-doc)
  - [Backend logging convention](#backend-logging-convention)
  - [Related](#related)

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

*(Backup of `project-rules.mdc` – Maintainability.)*

---

## Backup rules in project-rules.md

**Rule**: Every rule added or changed in `.cursor/rules/` must be backed up in this doc (`doc/feature/project-rules.md`). When adding a new .mdc under .cursor/rules/, add a corresponding section here; when editing an existing rule, update the backup here so the doc stays in sync.

*(Backup of `project-rules.mdc` – Backup.)*

---

## Table of Contents in documentation

**Rule**: Every Markdown document under `doc/` (design and feature docs) must include a `## Table of Contents` section. The TOC must list headings down to the 4th level (`##`, `###`, `####`) where such headings exist, with indentation and markdown links to heading anchors. For very long docs (e.g. 50+ sections), TOC may list to `###` with key `####` subsections. Place the TOC after the title/metadata and first `---`, before the first content section. Use GitHub-style anchors (lowercase, spaces to hyphens). When adding or restructuring a doc under `doc/`, add or update the TOC so it stays accurate to level 4.

*(Backup of `doc-rules.mdc` – TOC.)*

---

## Diagrams in doc/ use Mermaid

**Rule**: All flow diagrams and interaction diagrams in documentation under `doc/` must be written in Mermaid (e.g. `sequenceDiagram`, `flowchart`, `stateDiagram`). Do not use ASCII-art or other non-Mermaid formats for flows or interactions. Use Mermaid code blocks so diagrams render consistently. When adding or updating docs that describe a process or interaction, add or update the corresponding Mermaid diagram.

*(Backup of `doc-rules.mdc` – Diagrams.)*

---

## Directory tree format in doc/

**Rule**: When showing directory or file tree structure in any doc under `doc/`, use a **tree-style format with lines**: `├──` for an item with siblings below, `└──` for the last item at that level, `│` for the vertical line from a parent. Root can be `.`. Do not use plain indentation-only trees.

*(Backup of `doc-rules.mdc` – Directory tree.)*

---

## Backend logging convention

**Rule**: Backend logs use a **tag prefix** so that HTTP、auth and DB logs are distinguishable and grep-friendly. Do not log secrets (passwords, tokens) or full request bodies.

- **HTTP request**: `[HTTP] METHOD path status client_ip latency` (e.g. `[HTTP] POST /api/auth/login 200 192.168.1.1 12ms`). Implemented in `internal/middleware/logger.go`.
- **Auth**: `[AUTH] action ...` (e.g. `[AUTH] login attempt username=alice`, `[AUTH] login success username=alice`, `[AUTH] login failed username=alice reason=invalid_credentials`). Implemented in `internal/api/auth_handler.go`.
- **Database (GORM)**: Each GORM log line is prefixed with `[DB] ` (e.g. `[DB] [157.857ms] [rows:1] SELECT ...`). Implemented in `cmd/server/main.go` via custom writer for `logger.New`.

Previously the database query log had no tag; it was just `[157.857ms] [rows:1] SELECT ...`, which did not match the convention. It is now wrapped so all DB output starts with `[DB] `.

---

## Related

- [Testing](./testing.md) – Unit and integration test layout and commands
- [Initialization](./initialization.md) – Project structure and setup
