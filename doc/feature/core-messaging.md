# Core Messaging (One-on-One)

| Feature | Status | Date       |
| ------- | ------ | ---------- |
| Conversation & message backend | ✅ | 2025-03-27 |
| HTTP API (conversations, messages) | ✅ | 2025-03-27 |
| WebSocket (send_message, new_message) | ✅ | 2025-03-27 |
| Unit & integration tests | ✅ | 2025-03-27 |

---

## Table of Contents

- [Overview](#overview)
- [Backend Architecture](#backend-architecture)
- [HTTP API Endpoints](#http-api-endpoints)
  - [Conversation list response (with metadata)](#conversation-list-response-with-metadata)
  - [Mark read endpoint](#mark-read-endpoint)
- [WebSocket](#websocket)
  - [JSON Protocol](#json-protocol)
- [WebSocket Hub & Handler](#websocket-hub--handler)
- [Message Flow](#message-flow)
- [Testing](#testing)

---

## Overview

Core messaging provides **one-on-one conversations** and **text messages** with persistence and real-time delivery via WebSocket. There is **no web frontend** in this phase; behaviour is validated by **unit tests** (services with mocks) and **integration tests** (HTTP + WebSocket against a real DB).

---

## Backend Architecture

| Layer | Components |
| ----- | ---------- |
| **Repository** | `internal/repository/conversation_repository.go`, `message_repository.go` |
| **Service** | `internal/service/conversation_service.go`, `message_service.go` |
| **HTTP API** | `internal/api/conversation_handler.go`, `message_handler.go` |
| **WebSocket** | `internal/websocket/hub.go`, `internal/api/websocket_handler.go` |

- **ConversationRepository**: Create, GetByID, ListByUserID, AddParticipant, FindOneOnOneBetween, IsParticipant, GetParticipantUserIDs, UpdateParticipantLastRead, GetUnreadCounts, GetOtherParticipantUserIDsForOneOnOne.
- **MessageRepository**: Create, ListByConversationID (pagination, optional before_id), GetByID, GetLastMessagesByConversationIDs.
- **UserRepository**: GetByIDs (batch) used by conversation list with meta.
- **ConversationService**: CreateOneOnOne (reuse existing or create new), GetByID / ListByUserID / ListByUserIDWithMeta (access control, list with other_user/last_message/unread_count), MarkRead, EnsureUserInConversation.
- **MessageService**: Create (validation, persistence, optional broadcast via `MessageNotifier`), ListByConversationID (participant check). The **WebSocket hub** implements `MessageNotifier` and broadcasts `new_message` to conversation participants.

---

## HTTP API Endpoints

All messaging endpoints require **JWT** via `Authorization: Bearer <access_token>`.

| Method | Path | Description |
| ------ | ---- | ----------- |
| POST | `/api/conversations` | Create or return existing 1:1 conversation. Body: `{ "other_user_id": "<uuid>" }`. |
| GET | `/api/conversations` | List current user's conversations with metadata. Query: `limit`, `offset` (default 20, 0). Response includes `other_user`, `last_message`, `unread_count` per conversation. See [Conversation list response](#conversation-list-response-with-metadata). |
| POST | `/api/conversations/:id/read` | Update current user's last read message in the conversation. Body: `{ "last_read_message_id": <int64> }`. See [Mark read endpoint](#mark-read-endpoint). |
| GET | `/api/conversations/:id/messages` | List messages in a conversation (participant only). Query: `limit`, `offset`, optional `before_id` (cursor). |

Errors: 401 (missing/invalid token), 403 (not participant), 404 (user/conversation not found), 400 (invalid input).

#### Conversation list response (with metadata)

Each item in `GET /api/conversations` response `conversations` array includes:

- **Standard fields**: `conversation_id`, `type`, `name`, `created_by`, `created_at`, `updated_at`.
- **other_user** (optional): For 1:1 conversations only. Object with `user_id`, `username`, `display_name`, `avatar_url` of the other participant.
- **last_message** (optional): Object with `message_id`, `content`, `sender_id`, `created_at` of the latest message in the conversation. Omitted if there are no messages.
- **unread_count**: Number of messages in the conversation that the current user has not read (messages from others with `message_id` greater than the user's `last_read_message_id` for this conversation).

#### Mark read endpoint

`POST /api/conversations/:id/read` updates `conversation_participants.last_read_message_id` for the authenticated user in the given conversation. Request body must include `last_read_message_id` (integer). Returns 204 No Content on success. Requires the user to be a participant (403 otherwise).

---

## WebSocket

- **Endpoint**: `GET /ws`. Token via query `?token=<access_token>` or header `Authorization: Bearer <access_token>`.
- **Lifecycle**: Upgrade → validate JWT → register client in hub → read pump (handle `send_message`) and write pump (broadcast + ping/pong). On disconnect, hub unregisters the client and closes the send channel.

### JSON Protocol

**Client → Server**

- `send_message`: send a text message in a conversation.
  - `{ "type": "send_message", "conversation_id": "<uuid>", "content": "text" }`
  - Server persists via `MessageService.Create` and hub broadcasts `new_message` to all participants.

**Server → Client**

- `new_message`: new message in a conversation (broadcast to participants).
  - `{ "type": "new_message", "message": { "message_id", "conversation_id", "sender_id", "content", "type", "created_at", ... } }`

- **Rate limiting**: 60 messages per minute per connection (handler-level).
- **Ping/pong**: server sends ping; client should respond with pong to keep connection alive.

---

## WebSocket Hub & Handler

- **Hub** (`internal/websocket/hub.go`): Maps user ID → set of clients (one user, multiple connections). Implements `service.MessageNotifier`: on `NotifyNewMessage(conversationID, msg)` it resolves participant user IDs via `ConversationRepository.GetParticipantUserIDs` and sends the JSON `new_message` to each connected client for those users.
- **Handler** (`internal/api/websocket_handler.go`): Upgrades HTTP to WebSocket, validates JWT, registers client with hub, runs read pump (parse JSON `send_message` → call `MessageService.Create`) and write pump (send from hub + ping). Unregister and close on disconnect.

---

## Message Flow

1. **Send via WebSocket**: Client sends `send_message` → handler calls `MessageService.Create` → message persisted → hub `NotifyNewMessage` → all participants’ connections receive `new_message`.
2. **Send via HTTP** (future): Could add `POST /api/conversations/:id/messages` that calls `MessageService.Create`; hub would still broadcast to connected clients.
3. **History**: `GET /api/conversations/:id/messages` returns paginated messages (newest first; optional `before_id` cursor).

---

## Testing

- **Unit tests** (`internal/service/conversation_service_test.go`, `message_service_test.go`): Mock repositories and optional notifier; cover CreateOneOnOne (same user, other not found, existing, new), GetByID (not participant / success), ListByUserID; message Create (not participant, empty content, success + notifier called), ListByConversationID (not participant / success). Run: `go test ./internal/service/... -v`.
- **Integration tests** (`tests/integration/messaging_test.go`): Require DB and seed (`make init-db`, `make seed-db`). Use seed users (e.g. alice, bob). Tests: create 1:1 conversation (alice with bob), list conversations, list messages; WebSocket: connect with token, send `send_message`, verify message via `GET /api/conversations/:id/messages`. Run: `go test ./tests/integration/... -v -run TestConversationCreateAndList|TestMessagesList|TestWebSocketSendMessage` (or `make test-integration`).

See `doc/feature/testing.md` for project test layout and commands.
