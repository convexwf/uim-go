-- Migration: 000002_user_contacts
-- Description: Add one-way user contacts table
-- Created: 2026-04-23

CREATE TABLE IF NOT EXISTS user_contacts (
    owner_user_id UUID NOT NULL,
    contact_user_id UUID NOT NULL,
    source VARCHAR(32) NOT NULL DEFAULT 'manual',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP,
    PRIMARY KEY (owner_user_id, contact_user_id),
    CONSTRAINT chk_user_contacts_no_self CHECK (owner_user_id <> contact_user_id)
);

CREATE INDEX IF NOT EXISTS idx_user_contacts_owner_created
    ON user_contacts(owner_user_id, created_at DESC)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_user_contacts_contact
    ON user_contacts(contact_user_id)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_user_contacts_deleted_at
    ON user_contacts(deleted_at)
    WHERE deleted_at IS NOT NULL;
