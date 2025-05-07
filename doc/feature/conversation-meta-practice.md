# Conversation List Metadata and Mark-Read (Practice)

| Feature | Status | Date       |
| ------- | ------ | ---------- |
| Mark-read API | Done | 2026-03-02 |
| Conversation list with other_user, last_message, unread_count | Done | 2026-03-02 |

---

## Table of Contents

- [Overview](#overview)
- [Commands](#commands)
- [API usage](#api-usage)
- [References](#references)

---

## Overview

This doc records **successful practices** for the Phase 2 completion: mark-read endpoint and conversation list metadata (other_user, last_message, unread_count for 1:1). See [core-messaging.md](./core-messaging.md) for full API and architecture.

---

## Commands

- **Build server**: `go build -o bin/uim-server ./cmd/server` or `make build`
- **Run server**: `./bin/uim-server` (requires DB and optional Redis; see [initialization.md](./initialization.md))
- **Run unit tests**: `go test ./internal/service/... -v`
- **Run integration tests**: `go test ./tests/integration/... -v` (requires `make init-db` and `make seed-db`)

---

## API usage

**List conversations (with metadata)**

```bash
curl -s -H "Authorization: Bearer <access_token>" "http://localhost:8080/api/conversations?limit=20&offset=0"
```

Response `conversations[]` items include: `conversation_id`, `type`, `name`, `created_by`, `created_at`, `updated_at`, `other_user` (for 1:1), `last_message` (if any), `unread_count`.

**Mark read**

```bash
curl -s -X POST -H "Authorization: Bearer <access_token>" \
  -H "Content-Type: application/json" \
  -d '{"last_read_message_id": 123}' \
  "http://localhost:8080/api/conversations/<conversation_id>/read"
```

Expect 204 No Content on success; 403 if not a participant.

---

## References

- [core-messaging.md](./core-messaging.md) – HTTP endpoints, list response shape, mark-read
- [roadmap-v1.0.md](../design/v1.0/roadmap-v1.0.md) – Phase 2 Core Messaging tasks
