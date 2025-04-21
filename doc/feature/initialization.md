# Initialization Feature Documentation

| Feature             | Status | Date       |
| ------------------- | ------ | ---------- |
| Project Setup       | ✅      | 2025-03-13 |
| Database Schema     | ✅      | 2025-03-15 |
| User Authentication | ✅      | 2025-03-17 |
| API Server          | ✅      | 2025-03-19 |
| Documentation       | ✅      | 2025-03-20 |

---

## Table of Contents

- [Overview](#overview)
- [Project Structure](#project-structure)
- [Database Schema](#database-schema)
  - [Users Table](#users-table)
  - [Conversations Table](#conversations-table)
  - [Conversation Participants Table](#conversation-participants-table)
  - [Messages Table](#messages-table)
- [Database Migrations](#database-migrations)
- [Authentication Flow](#authentication-flow)
  - [Registration](#registration)
  - [Login](#login)
  - [Token Refresh](#token-refresh)
  - [JWT Token Structure](#jwt-token-structure)
- [API Endpoints](#api-endpoints)
  - [Health Check](#health-check)
  - [Authentication Endpoints](#authentication-endpoints)
- [Middleware](#middleware)
  - [CORS Middleware](#cors-middleware)
  - [Logger Middleware](#logger-middleware)
  - [Error Handler Middleware](#error-handler-middleware)
  - [Authentication Middleware](#authentication-middleware)
- [Configuration](#configuration)
- [Development Setup](#development-setup)
  - [Prerequisites](#prerequisites)
  - [Setup Steps](#setup-steps)
- [Testing](#testing)
  - [Unit Tests](#unit-tests)
  - [Integration Tests](#integration-tests)
- [Deliverables](#deliverables)
- [Next Steps](#next-steps)
- [Notes](#notes)

---

## Overview

This document describes the initialization phase of the UIM system, which includes project setup, database schema design, user authentication, and basic API server infrastructure.

## Project Structure

```txt
uim-go/
├── cmd/
│   └── server/          # Application entry point
├── internal/
│   ├── api/             # HTTP API handlers
│   ├── config/          # Configuration management
│   ├── middleware/      # HTTP middleware
│   ├── model/           # Data models
│   ├── repository/      # Data access layer
│   └── service/         # Business logic layer
│   └── pkg/             # Internal utility packages
│   │   ├── jwt/         # JWT token management
│   │   └── password/    # Password hashing utilities
├── migrations/          # SQL database migration files
├── scripts/             # Utility scripts (init_db.sh, seed_db.sh)
├── tests/               # Integration tests (not unit tests)
│   └── integration/    # DB and auth endpoint integration tests
├── docker/              # Docker-related files
├── docker-compose.yml   # Local development environment
└── doc/                 # Documentation
```

## Database Schema

### Users Table

```sql
CREATE TABLE users (
    user_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    display_name VARCHAR(100),
    avatar_url TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP
);

CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_email ON users(email);
```

**Fields**:
- `user_id`: Primary key, UUID v4
- `username`: Unique, 3-50 characters
- `email`: Unique, RFC 5322 compliant
- `password_hash`: bcrypt hash with cost factor 10
- `display_name`: Optional display name
- `avatar_url`: Optional avatar URL

### Conversations Table

```sql
CREATE TABLE conversations (
    conversation_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type VARCHAR(20) NOT NULL, -- 'one_on_one' or 'group'
    name VARCHAR(255), -- For group chats
    created_by UUID NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP
);

CREATE INDEX idx_conversations_type ON conversations(type);
```

**Fields**:
- `conversation_id`: Primary key, UUID v4
- `type`: Conversation type (one_on_one or group)
- `name`: Optional name for group chats
- `created_by`: User ID who created the conversation

### Conversation Participants Table

```sql
CREATE TABLE conversation_participants (
    conversation_id UUID NOT NULL,
    user_id UUID NOT NULL,
    role VARCHAR(20) DEFAULT 'member', -- owner, admin, member
    joined_at TIMESTAMP DEFAULT NOW(),
    last_read_message_id BIGINT,
    PRIMARY KEY (conversation_id, user_id),
    FOREIGN KEY (conversation_id) REFERENCES conversations(conversation_id),
    FOREIGN KEY (user_id) REFERENCES users(user_id)
);

CREATE INDEX idx_conversation_participants_user_id ON conversation_participants(user_id);
CREATE INDEX idx_conversation_participants_conv_id ON conversation_participants(conversation_id);
```

**Fields**:
- `conversation_id`: Foreign key to conversations
- `user_id`: Foreign key to users
- `role`: Participant role (owner, admin, member)
- `last_read_message_id`: Last read message ID for unread tracking

### Messages Table

```sql
CREATE TABLE messages (
    message_id BIGSERIAL PRIMARY KEY,
    conversation_id UUID NOT NULL,
    sender_id UUID NOT NULL,
    content TEXT NOT NULL,
    message_type VARCHAR(20) DEFAULT 'text',
    created_at TIMESTAMP DEFAULT NOW(),
    metadata JSONB,
    deleted_at TIMESTAMP,
    FOREIGN KEY (conversation_id) REFERENCES conversations(conversation_id),
    FOREIGN KEY (sender_id) REFERENCES users(user_id)
);

CREATE INDEX idx_messages_conversation_time ON messages(conversation_id, created_at DESC);
CREATE INDEX idx_messages_sender ON messages(sender_id);
```

**Fields**:
- `message_id`: Primary key, auto-incrementing bigint
- `conversation_id`: Foreign key to conversations
- `sender_id`: Foreign key to users
- `content`: Message content (text, up to 10,000 characters)
- `message_type`: Message type (currently only 'text')
- `metadata`: Optional JSON metadata

**Note**: Friendships table is deferred to Phase 2 (not required for MVP).

## Database Migrations

Schema is **SQL-first**: run **`scripts/init_db.sh`** before starting the service to create tables and indexes (idempotent). The app does **not** run migrations by default; it only **checks** that the schema exists. GORM is used for this check and, optionally in non-production, as a fallback that applies the same SQL file. See [Database Migrations Documentation](./database-migrations.md) for details.

**Seed and testing**: After `init_db.sh`, run **`make seed-db`** (or `./scripts/seed_db.sh`) to create test users (alice, bob, test / password `password123`). Integration tests live in **`tests/integration/`** and run with **`make test-integration`** (requires DB; run seed-db for the DB performance test).

## Authentication Flow

### Registration

1. Client sends `POST /api/auth/register` with username, email, password
2. Server validates input (username 3-50 chars, valid email, password min 6 chars)
3. Server checks if user already exists (by username or email)
4. Server hashes password using bcrypt (cost factor 10)
5. Server creates user record in database
6. Server generates JWT access token (15 min expiry) and refresh token (7 days expiry)
7. Server returns user object and tokens

### Login

1. Client sends `POST /api/auth/login` with username and password
2. Server retrieves user by username
3. Server verifies password using bcrypt
4. Server generates JWT access token and refresh token
5. Server returns user object and tokens

### Token Refresh

1. Client sends `POST /api/auth/refresh` with refresh token
2. Server validates refresh token
3. Server retrieves user by user ID from token
4. Server generates new access token and refresh token
5. Server returns new tokens

### JWT Token Structure

**Access Token**:
- Type: "access"
- Expiry: 15 minutes
- Contains: user_id

**Refresh Token**:
- Type: "refresh"
- Expiry: 7 days
- Contains: user_id

## API Endpoints

### Health Check

**GET** `/health`

Returns system health status including database connectivity.

**Response**:
```json
{
  "status": "healthy",
  "checks": {
    "database": "healthy"
  }
}
```

### Authentication Endpoints

**POST** `/api/auth/register`

Register a new user.

**Request Body**:
```json
{
  "username": "johndoe",
  "email": "john@example.com",
  "password": "password123"
}
```

**Response** (201 Created):
```json
{
  "user": {
    "user_id": "uuid",
    "username": "johndoe",
    "email": "john@example.com",
    "display_name": "johndoe",
    "created_at": "2025-02-10T00:00:00Z"
  },
  "access_token": "jwt-token",
  "refresh_token": "jwt-token"
}
```

**POST** `/api/auth/login`

Login with username and password.

**Request Body**:
```json
{
  "username": "johndoe",
  "password": "password123"
}
```

**Response** (200 OK):
```json
{
  "user": {
    "user_id": "uuid",
    "username": "johndoe",
    "email": "john@example.com",
    "display_name": "johndoe"
  },
  "access_token": "jwt-token",
  "refresh_token": "jwt-token"
}
```

**POST** `/api/auth/refresh`

Refresh access token using refresh token.

**Request Body**:
```json
{
  "refresh_token": "jwt-refresh-token"
}
```

**Response** (200 OK):
```json
{
  "access_token": "new-jwt-token",
  "refresh_token": "new-jwt-token"
}
```

## Middleware

### CORS Middleware

Handles Cross-Origin Resource Sharing. Allows requests from configured origins.

**Configuration**: Set via `CORS_ALLOWED_ORIGINS` environment variable (comma-separated).

### Logger Middleware

Logs all HTTP requests with method, path, IP, and latency.

### Error Handler Middleware

Catches and formats errors, returns appropriate HTTP status codes.

### Authentication Middleware

Validates JWT access tokens for protected routes. Extracts user ID and sets it in request context.

**Usage** (for future protected routes):
```go
protected := api.Group("")
protected.Use(middleware.AuthMiddleware(jwtManager))
```

## Configuration

Configuration is loaded from environment variables with defaults. See `.env.example` for all available options.

**Key Configuration**:
- `APP_PORT`: Server port (default: 8080)
- `POSTGRES_HOST`: Database host (default: localhost)
- `POSTGRES_DB`: Database name (default: uim_db)
- `JWT_SECRET`: Secret key for JWT signing (MUST change in production)
- `JWT_ACCESS_EXPIRY`: Access token expiry (default: 15m)
- `JWT_REFRESH_EXPIRY`: Refresh token expiry (default: 168h)

## Development Setup

### Prerequisites

- Go 1.21+
- Docker and Docker Compose
- PostgreSQL 15 (via Docker)
- Redis 7 (via Docker)

### Setup Steps

1. **Clone repository**:
   ```bash
   git clone https://github.com/convexwf/uim-go.git
   cd uim-go
   ```

2. **Copy environment file**:
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

3. **Start Docker services**:
   ```bash
   docker-compose up -d
   ```

4. **Install dependencies**:
   ```bash
   go mod download
   ```

5. **Run migrations**:
   - **Development**: Migrations run automatically on startup using GORM AutoMigrate
   - **Production**: See [Database Migrations Documentation](./database-migrations.md) for SQL migration instructions

6. **Start server**:
   ```bash
   go run cmd/server/main.go
   ```

7. **Verify health**:
   ```bash
   curl http://localhost:8080/health
   ```

## Testing

### Unit Tests

Run unit tests

```bash
go test ./pkg/...
```

**Test Coverage**

- Password hashing and verification
- JWT token generation and validation

### Integration Tests

Integration tests for authentication endpoints will be added in Phase 2.

## Deliverables

✅ **Project Structure**: Complete Go project structure with proper layering  
✅ **Docker Compose**: PostgreSQL and Redis services configured  
✅ **Database Schema**: Users, conversations, participants, messages tables  
✅ **Database Migrations**: GORM AutoMigrate + SQL migration files (see [Database Migrations Documentation](./database-migrations.md))  
✅ **User Authentication**: Registration, login, token refresh  
✅ **JWT Implementation**: Access and refresh tokens  
✅ **API Server**: Gin framework with middleware  
✅ **Health Check**: Database connectivity check  
✅ **Configuration Management**: Environment-based configuration  
✅ **Documentation**: This feature documentation

## Next Steps

**Week 3-4: Core Messaging**

- Implement WebSocket connections
- Implement message sending/receiving
- Implement message persistence
- Create conversation management APIs

## Notes

- Friendships table is intentionally deferred to Phase 2 (not required for MVP)
- CI/CD pipeline setup is deferred to Phase 3 (Production Readiness)
- Database migrations use hybrid approach: GORM AutoMigrate for development, SQL migration files for production
- See [Database Migrations Documentation](./database-migrations.md) for detailed migration strategy and index management
- All passwords are hashed using bcrypt with cost factor 10
- JWT tokens use HS256 signing algorithm
