-- Migration: 000001_initial_schema (rollback)
-- Description: Drop initial database schema
-- Created: 2025-03-13
--
-- This migration rolls back the initial schema by dropping all tables and indexes.
-- Tables are dropped in reverse dependency order to avoid foreign key constraint errors.

-- Drop messages table and indexes
DROP INDEX IF EXISTS idx_messages_deleted_at;
DROP INDEX IF EXISTS idx_messages_sender;
DROP INDEX IF EXISTS idx_messages_conversation_time;
DROP TABLE IF EXISTS messages;

-- Drop conversation_participants table and indexes
DROP INDEX IF EXISTS idx_conversation_participants_conv_id;
DROP INDEX IF EXISTS idx_conversation_participants_user_id;
DROP TABLE IF EXISTS conversation_participants;

-- Drop conversations table and indexes
DROP INDEX IF EXISTS idx_conversations_deleted_at;
DROP INDEX IF EXISTS idx_conversations_type;
DROP TABLE IF EXISTS conversations;

-- Drop users table and indexes
DROP INDEX IF EXISTS idx_users_deleted_at;
DROP INDEX IF EXISTS idx_users_email;
DROP INDEX IF EXISTS idx_users_username;
DROP TABLE IF EXISTS users;
