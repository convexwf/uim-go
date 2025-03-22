# Database Migrations Documentation

| Feature             | Status | Date       |
| ------------------- | ------ | ---------- |
| Migration Strategy  | ✅      | 2025-03-13 |
| SQL Migration Files | ✅      | 2025-03-13 |
| Index Consistency   | ✅      | 2025-03-13 |

---

## Overview

This document describes the database migration strategy for the UIM system, which uses a hybrid approach combining GORM AutoMigrate for development convenience and SQL migration files for production readiness.

## Migration Strategy

The project uses a hybrid approach for database schema management:

**Development Environment**:
- Uses GORM AutoMigrate for automatic schema synchronization
- Convenient for rapid development and testing
- Automatically creates tables and indexes based on model definitions

**Production Environment**:
- Uses explicit SQL migration files for version control and reproducibility
- Located in `migrations/` directory
- Follows naming convention: `{timestamp}_{description}.up.sql` and `.down.sql`

## GORM Model Index Definitions

All indexes in GORM models use explicit naming to ensure consistency with SQL migrations:

### User Model (`internal/model/user.go`)

```go
Username string `gorm:"type:varchar(50);uniqueIndex:idx_users_username;not null"`
Email    string `gorm:"type:varchar(255);uniqueIndex:idx_users_email;not null"`
DeletedAt gorm.DeletedAt `gorm:"index:idx_users_deleted_at"`
```

### Conversation Model (`internal/model/conversation.go`)

```go
DeletedAt gorm.DeletedAt `gorm:"index:idx_conversations_deleted_at"`
```

### Message Model (`internal/model/message.go`)

```go
ConversationID uuid.UUID `gorm:"type:uuid;not null;index:idx_messages_conversation_time"`
SenderID       uuid.UUID `gorm:"type:uuid;not null;index:idx_messages_sender"`
CreatedAt      time.Time `gorm:"index:idx_messages_conversation_time"`  // Composite index
DeletedAt      gorm.DeletedAt `gorm:"index:idx_messages_deleted_at"`
```

## SQL Migration Files

**Location**: `migrations/000001_initial_schema.up.sql`

The SQL migration file includes:
- Table creation with `CREATE TABLE IF NOT EXISTS`
- Index creation with `CREATE UNIQUE INDEX IF NOT EXISTS` and `CREATE INDEX IF NOT EXISTS`
- All index names match GORM model definitions exactly
- Partial indexes for `deleted_at` fields using `WHERE deleted_at IS NOT NULL`

### Index Naming Convention

- Format: `idx_{table_name}_{field_name}` or `idx_{table_name}_{purpose}`
- Examples:
  - `idx_users_username` - unique index on users.username
  - `idx_messages_conversation_time` - composite index on messages(conversation_id, created_at)
  - `idx_conversation_participants_user_id` - index on conversation_participants.user_id

### Migration File Structure

**Up Migration** (`000001_initial_schema.up.sql`):
- Creates all tables with `IF NOT EXISTS`
- Creates all indexes with `IF NOT EXISTS`
- Includes comments explaining the migration

**Down Migration** (`000001_initial_schema.down.sql`):
- Drops all indexes with `IF EXISTS`
- Drops all tables with `IF EXISTS`
- Drops in reverse dependency order to avoid foreign key constraint errors

## Index Consistency

### Index Consistency Table

| Table                     | Field/Purpose                | GORM Index Name                | SQL Index Name                        | Type              | Notes    |
| ------------------------- | ---------------------------- | ------------------------------ | ------------------------------------- | ----------------- | -------- |
| users                     | username                     | idx_users_username             | idx_users_username                    | UNIQUE            | ✓ Match  |
| users                     | email                        | idx_users_email                | idx_users_email                       | UNIQUE            | ✓ Match  |
| users                     | deleted_at                   | idx_users_deleted_at           | idx_users_deleted_at                  | INDEX             | ✓ Match  |
| conversations             | type                         | (N/A)                          | idx_conversations_type                | INDEX             | SQL only |
| conversations             | deleted_at                   | idx_conversations_deleted_at   | idx_conversations_deleted_at          | INDEX             | ✓ Match  |
| conversation_participants | user_id                      | (N/A)                          | idx_conversation_participants_user_id | INDEX             | SQL only |
| conversation_participants | conversation_id              | (N/A)                          | idx_conversation_participants_conv_id | INDEX             | SQL only |
| messages                  | conversation_id + created_at | idx_messages_conversation_time | idx_messages_conversation_time        | INDEX (composite) | ✓ Match  |
| messages                  | sender_id                    | idx_messages_sender            | idx_messages_sender                   | INDEX             | ✓ Match  |
| messages                  | deleted_at                   | idx_messages_deleted_at        | idx_messages_deleted_at               | INDEX             | ✓ Match  |

### Why Named Indexes?

Using explicit index names in both GORM models and SQL migrations ensures:
1. **No Conflicts**: GORM AutoMigrate can detect existing indexes by name and skip creation
2. **Consistency**: Both approaches use the same index names
3. **Predictability**: Index names follow a clear naming convention
4. **Maintainability**: Easy to track and manage indexes across code and SQL

## Why This Approach?

### Benefits

1. **No Conflicts**: Using `IF NOT EXISTS` in SQL and named indexes in GORM prevents duplicate index creation errors
2. **Development Speed**: AutoMigrate allows rapid iteration during development
3. **Production Safety**: SQL migrations provide explicit, versioned schema changes for production
4. **Consistency**: Named indexes ensure GORM and SQL use the same index names
5. **Flexibility**: Can use either approach depending on environment needs

### How It Works

- **When SQL migration runs first**: Creates tables and indexes with `IF NOT EXISTS`
- **When GORM AutoMigrate runs**: Detects existing indexes by name and skips creation (no error)
- **Result**: Both approaches work independently without conflicts

## Running Migrations

### Development (AutoMigrate)

Migrations run automatically on server startup:

```go
// cmd/server/main.go
db.AutoMigrate(
    &model.User{},
    &model.Conversation{},
    &model.ConversationParticipant{},
    &model.Message{},
)
```

**When to use**:
- Local development
- Testing environments
- Rapid prototyping

### Production (SQL Migrations)

Execute SQL migration files manually or via migration tool:

```bash
# Execute up migration
psql -U postgres -d uim_db -f migrations/000001_initial_schema.up.sql

# Or using environment variables
psql -h $POSTGRES_HOST -U $POSTGRES_USER -d $POSTGRES_DB -f migrations/000001_initial_schema.up.sql
```

**When to use**:
- Production deployments
- Staging environments
- CI/CD pipelines
- Database version control

### Rollback

Execute down migration to rollback schema changes:

```bash
# Execute down migration
psql -U postgres -d uim_db -f migrations/000001_initial_schema.down.sql
```

**Warning**: Rollback will drop all tables and indexes. Use with caution in production.

## Best Practices

1. **Always use named indexes** in GORM models to ensure consistency
2. **Use `IF NOT EXISTS`** in SQL migrations to avoid conflicts
3. **Test migrations** in development before applying to production
4. **Version control** all migration files
5. **Document changes** in migration file comments
6. **Never modify** existing migration files after they've been applied to production
7. **Create new migrations** for schema changes instead of modifying existing ones

## Future Enhancements

- [ ] Add migration tool integration (e.g., golang-migrate, migrate)
- [ ] Add migration version tracking table
- [ ] Add automated migration testing
- [ ] Add migration rollback verification

## Related Documentation

- [Initialization Feature Documentation](./initialization.md) - Overall initialization phase documentation
- [System Design v1.0](../design/v1.0/uim-system-design-v1.0.md) - System architecture and design
