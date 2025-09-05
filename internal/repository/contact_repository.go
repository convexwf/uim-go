// Copyright 2025 convexwf
//
// Project: uim-go
// File: contact_repository.go
// Email: convexwf@gmail.com
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// See the License for the full terms.
//
// Description: Contact repository for database operations

package repository

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/convexwf/uim-go/internal/model"
)

// ContactListRow is the joined row shape for listing contacts with user fields.
type ContactListRow struct {
	ContactUserID uuid.UUID `gorm:"column:contact_user_id"`
	Username      string    `gorm:"column:username"`
	DisplayName   string    `gorm:"column:display_name"`
	AvatarURL     string    `gorm:"column:avatar_url"`
	CreatedAt     time.Time `gorm:"column:created_at"`
}

// ContactRepository defines contact data access operations.
type ContactRepository interface {
	ListByOwner(ownerUserID uuid.UUID, limit, offset int) ([]*ContactListRow, error)
	Exists(ownerUserID, contactUserID uuid.UUID) (bool, error)
	Add(ownerUserID, contactUserID uuid.UUID) error
	Delete(ownerUserID, contactUserID uuid.UUID) (bool, error)
}

type contactRepository struct {
	db *gorm.DB
}

// NewContactRepository creates a new contact repository instance.
func NewContactRepository(db *gorm.DB) ContactRepository {
	return &contactRepository{db: db}
}

// ListByOwner lists contacts for the given owner, newest first.
func (r *contactRepository) ListByOwner(ownerUserID uuid.UUID, limit, offset int) ([]*ContactListRow, error) {
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	var rows []*ContactListRow
	err := r.db.Table("user_contacts").
		Select(
			"user_contacts.contact_user_id, users.username, users.display_name, users.avatar_url, user_contacts.created_at",
		).
		Joins("INNER JOIN users ON users.user_id = user_contacts.contact_user_id").
		Where("user_contacts.owner_user_id = ? AND user_contacts.deleted_at IS NULL AND users.deleted_at IS NULL", ownerUserID).
		Order("user_contacts.created_at DESC").
		Limit(limit).Offset(offset).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}

// Exists returns true if the contact relation exists and is not soft-deleted.
func (r *contactRepository) Exists(ownerUserID, contactUserID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&model.UserContact{}).
		Where("owner_user_id = ? AND contact_user_id = ? AND deleted_at IS NULL", ownerUserID, contactUserID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// Add inserts or restores a one-way contact relation.
func (r *contactRepository) Add(ownerUserID, contactUserID uuid.UUID) error {
	now := time.Now()
	rec := &model.UserContact{
		OwnerUserID:   ownerUserID,
		ContactUserID: contactUserID,
		Source:        "manual",
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	return r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "owner_user_id"},
			{Name: "contact_user_id"},
		},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"source":     "manual",
			"updated_at": now,
			"deleted_at": nil,
		}),
	}).Create(rec).Error
}

// Delete soft-deletes a one-way contact relation. Returns whether any row changed.
func (r *contactRepository) Delete(ownerUserID, contactUserID uuid.UUID) (bool, error) {
	tx := r.db.Model(&model.UserContact{}).
		Where("owner_user_id = ? AND contact_user_id = ? AND deleted_at IS NULL", ownerUserID, contactUserID).
		Updates(map[string]interface{}{
			"deleted_at": time.Now(),
			"updated_at": time.Now(),
		})
	if tx.Error != nil {
		return false, tx.Error
	}
	return tx.RowsAffected > 0, nil
}
