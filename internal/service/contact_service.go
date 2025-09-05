// Copyright 2025 convexwf
//
// Project: uim-go
// File: contact_service.go
// Email: convexwf@gmail.com
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// See the License for the full terms.
//
// Description: Contact business logic

package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/convexwf/uim-go/internal/repository"
	"github.com/convexwf/uim-go/internal/store"
)

var (
	ErrInvalidContact = errors.New("invalid contact")
)

// ContactWithPresence is the service shape for one contact row.
type ContactWithPresence struct {
	UserID      uuid.UUID
	Username    string
	DisplayName string
	AvatarURL   string
	CreatedAt   time.Time
	Status      string
	LastSeen    time.Time
}

// UserSearchResult is the service shape for exact username lookup.
type UserSearchResult struct {
	UserID       uuid.UUID
	Username     string
	DisplayName  string
	AvatarURL    string
	AlreadyAdded bool
}

// ContactService defines contact operations.
type ContactService interface {
	ListContacts(ownerUserID uuid.UUID, limit, offset int) ([]*ContactWithPresence, error)
	AddContact(ownerUserID, contactUserID uuid.UUID) (bool, error)
	DeleteContact(ownerUserID, contactUserID uuid.UUID) (bool, error)
	SearchByUsername(ownerUserID uuid.UUID, username string) ([]*UserSearchResult, error)
}

type contactService struct {
	contactRepo   repository.ContactRepository
	userRepo      repository.UserRepository
	presenceStore store.PresenceStore
}

// NewContactService creates a new contact service.
func NewContactService(contactRepo repository.ContactRepository, userRepo repository.UserRepository, presenceStore store.PresenceStore) ContactService {
	return &contactService{
		contactRepo:   contactRepo,
		userRepo:      userRepo,
		presenceStore: presenceStore,
	}
}

// ListContacts lists one-way contacts for the current user.
func (s *contactService) ListContacts(ownerUserID uuid.UUID, limit, offset int) ([]*ContactWithPresence, error) {
	rows, err := s.contactRepo.ListByOwner(ownerUserID, limit, offset)
	if err != nil {
		return nil, err
	}
	out := make([]*ContactWithPresence, len(rows))
	for i, row := range rows {
		item := &ContactWithPresence{
			UserID:      row.ContactUserID,
			Username:    row.Username,
			DisplayName: row.DisplayName,
			AvatarURL:   row.AvatarURL,
			CreatedAt:   row.CreatedAt,
			Status:      "offline",
		}
		if s.presenceStore != nil {
			status, lastSeen, err := s.presenceStore.GetStatus(context.Background(), row.ContactUserID)
			if err != nil {
				return nil, err
			}
			item.Status = status
			item.LastSeen = lastSeen
		}
		out[i] = item
	}
	return out, nil
}

// AddContact creates a one-way contact relation.
func (s *contactService) AddContact(ownerUserID, contactUserID uuid.UUID) (bool, error) {
	if ownerUserID == contactUserID {
		return false, ErrInvalidContact
	}
	if _, err := s.userRepo.GetByID(contactUserID); err != nil {
		return false, ErrUserNotFound
	}
	exists, err := s.contactRepo.Exists(ownerUserID, contactUserID)
	if err != nil {
		return false, err
	}
	if err := s.contactRepo.Add(ownerUserID, contactUserID); err != nil {
		return false, err
	}
	return !exists, nil
}

// DeleteContact removes a one-way contact relation for the current user.
func (s *contactService) DeleteContact(ownerUserID, contactUserID uuid.UUID) (bool, error) {
	if ownerUserID == contactUserID {
		return false, ErrInvalidContact
	}
	return s.contactRepo.Delete(ownerUserID, contactUserID)
}

// SearchByUsername finds at most one user by exact username and marks whether it is already added.
func (s *contactService) SearchByUsername(ownerUserID uuid.UUID, username string) ([]*UserSearchResult, error) {
	trimmed := strings.TrimSpace(username)
	if trimmed == "" {
		return nil, ErrInvalidInput
	}
	user, err := s.userRepo.GetByUsername(trimmed)
	if err != nil {
		return []*UserSearchResult{}, nil
	}
	if user.UserID == ownerUserID {
		return []*UserSearchResult{}, nil
	}
	exists, err := s.contactRepo.Exists(ownerUserID, user.UserID)
	if err != nil {
		return nil, err
	}
	return []*UserSearchResult{{
		UserID:       user.UserID,
		Username:     user.Username,
		DisplayName:  user.DisplayName,
		AvatarURL:    user.AvatarURL,
		AlreadyAdded: exists,
	}}, nil
}
