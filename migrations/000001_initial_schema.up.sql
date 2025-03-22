-- Migration: 000001_initial_schema
-- Description: Initial database schema with tables and indexes
-- Created: 2025-03-13
--
-- This migration creates the initial database schema for the UIM system.
-- All indexes use IF NOT EXISTS to avoid conflicts with GORM AutoMigrate.
-- Index names match those defined in GORM models to ensure consistency.

-- Users table
CREATE TABLE IF NOT EXISTS users (
    user_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(50) NOT NULL,
    email VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    display_name VARCHAR(100),
    avatar_url TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP
);

-- Users indexes
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users(deleted_at) WHERE deleted_at IS NOT NULL;

-- Conversations table
CREATE TABLE IF NOT EXISTS conversations (
    conversation_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type VARCHAR(20) NOT NULL,
    name VARCHAR(255),
    created_by UUID NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP
);

-- Conversations indexes
CREATE INDEX IF NOT EXISTS idx_conversations_type ON conversations(type);
CREATE INDEX IF NOT EXISTS idx_conversations_deleted_at ON conversations(deleted_at) WHERE deleted_at IS NOT NULL;

-- Conversation participants table
CREATE TABLE IF NOT EXISTS conversation_participants (
    conversation_id UUID NOT NULL,
    user_id UUID NOT NULL,
    role VARCHAR(20) DEFAULT 'member',
    joined_at TIMESTAMP DEFAULT NOW(),
    last_read_message_id BIGINT,
    PRIMARY KEY (conversation_id, user_id)
);

-- Conversation participants indexes
CREATE INDEX IF NOT EXISTS idx_conversation_participants_user_id ON conversation_participants(user_id);
CREATE INDEX IF NOT EXISTS idx_conversation_participants_conv_id ON conversation_participants(conversation_id);

-- Messages table
CREATE TABLE IF NOT EXISTS messages (
    message_id BIGSERIAL PRIMARY KEY,
    conversation_id UUID NOT NULL,
    sender_id UUID NOT NULL,
    content TEXT NOT NULL,
    message_type VARCHAR(20) DEFAULT 'text',
    created_at TIMESTAMP DEFAULT NOW(),
    metadata JSONB,
    deleted_at TIMESTAMP
);

-- Messages indexes
CREATE INDEX IF NOT EXISTS idx_messages_conversation_time ON messages(conversation_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_messages_sender ON messages(sender_id);
CREATE INDEX IF NOT EXISTS idx_messages_deleted_at ON messages(deleted_at) WHERE deleted_at IS NOT NULL;
