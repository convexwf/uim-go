# Reliability & Polish (Week 5-6)

| Feature | Status | Date       |
| ------- | ------ | ---------- |
| Offline message queue (Redis) | Done | 2025-05 |
| Online presence (Redis + API) | Done | 2025-05 |
| Health check (DB + Redis) | Done | 2025-05 |
| Error handling and retry | Done | 2025-05 |
| Unit and integration tests | Done | 2025-05 |

---

## Table of Contents

- [Overview](#overview)
- [Offline Message Queue](#offline-message-queue)
- [Online Presence](#online-presence)
- [Health Check](#health-check)
- [Error Handling and Retry](#error-handling-and-retry)
- [Testing](#testing)
- [Common Commands](#common-commands)

---

## Overview

This phase adds **offline message delivery**, **online presence**, **health checks** (including Redis), and **retry/graceful degradation** so the system can run when Redis is unavailable.

- **Offline queue**: Messages for offline users are stored in Redis (`offline:user:{user_id}`); on WebSocket connect, queued messages are delivered and the queue is cleared. TTL 24 hours.
- **Presence**: User online/offline state is stored in Redis (`presence:{user_id}`) with 90s TTL, updated on connect/disconnect/heartbeat. `GET /api/users/:id/presence` returns status.
- **Health**: `GET /health` checks database and Redis; if Redis is down, response is `"degraded"` but the app keeps running (no offline queue or presence).
- **Retry**: Message persistence and offline queue push use limited retries for transient errors.

---

## Offline Message Queue

- **Storage**: Redis list key `offline:user:{user_id}`, TTL 24 hours.
- **Flow**: When a message is sent, the hub delivers to connected participants; for each participant not in the hub, the message is pushed to their offline queue. On WebSocket connect, the server calls `PopAll` and sends each message to the client, then deletes the key.
- **Code**: `internal/store/offline_queue.go` (interface + Redis implementation), `internal/websocket/hub.go` (push when user offline), `internal/api/websocket_handler.go` (deliver on connect).

---

## Online Presence

- **Storage**: Redis key `presence:{user_id}` value `"online"`, TTL 90 seconds. Refreshed on each heartbeat (pong).
- **Updates**: On WebSocket connect: `SetOnline` + `PublishUpdate(..., "online")`. On disconnect: `SetOffline` + `PublishUpdate(..., "offline")`. On pong: `Refresh` (extend TTL).
- **API**: `GET /api/users/:id/presence` (authenticated). Response: `{"user_id":"...","status":"online|offline","last_seen":"..."}`.
- **Code**: `internal/store/presence.go`, `internal/api/presence_handler.go`, `internal/api/websocket_handler.go`.

---

## Health Check

- **Endpoint**: `GET /health` (no auth).
- **Checks**: Database ping; Redis ping (if client is configured).
- **Response**: `{"status":"healthy"|"degraded"|"unhealthy","checks":{"database":"healthy", "redis":"healthy"|"unhealthy: ..."}}`. Returns 503 only when status is `"unhealthy"` (e.g. DB down).

---

## Error Handling and Retry

- **Message create**: `internal/repository/message_repository.go` uses `retry.Do(3, 100ms, ...)` for transient DB errors (timeout, connection refused, etc.). See `internal/pkg/retry/retry.go`.
- **Offline queue push**: `internal/store/offline_queue.go` uses `retry.Do(2, 50ms, ...)` for Push. On failure the hub logs and continues (message is already persisted).
- **Graceful degradation**: If Redis is not available at startup, the server starts without offline queue and presence; health reports Redis as unhealthy.

---

## Testing

- **Unit**: `internal/store/offline_queue_test.go`, `internal/store/presence_test.go` (miniredis), `internal/pkg/retry/retry_test.go`.
- **Integration**: `tests/integration/messaging_test.go` — `TestOfflineMessageDelivery` (alice sends while bob offline, bob connects and receives), `TestPresenceAPI` (GET presence before/after connect and after disconnect). Use `setupMessagingRouterWithRedis(t)` which starts miniredis and wires offline queue and presence.

**Run tests**:
```bash
go test ./internal/store/... ./internal/pkg/retry/... -v
go test ./tests/integration/... -v -run 'TestOfflineMessageDelivery|TestPresenceAPI'
make test-integration   # if available
```

---

## Common Commands

```bash
# Build server (binary in bin/)
go build -o bin/uim-server ./cmd/server

# Run server (requires PostgreSQL; Redis optional)
./bin/uim-server

# Run all tests
go test ./...

# Run integration tests (need DB + seed)
make init-db && make seed-db
go test ./tests/integration/... -v
```

---

## References

- Design: [doc/design/v1.0/uim-system-design-v1.0.md](../design/v1.0/uim-system-design-v1.0.md) (§5.2.2, §5.3, §5.4.3, §6.1, §6.4)
- Deployment: [doc/design/v1.0/deployment-guide-v1.0.md](../design/v1.0/deployment-guide-v1.0.md) (§6.1 Health Checks)
