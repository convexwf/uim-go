// Copyright 2025 convexwf
//
// Project: uim-go
// File: conversation_service_test.go
// Description: Unit tests for conversation service

package service

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/convexwf/uim-go/internal/model"
	"github.com/convexwf/uim-go/internal/repository"
)

type mockConversationRepo struct {
	createErr              error
	getByIDConv            *model.Conversation
	getByIDErr             error
	listConvs              []*model.Conversation
	listErr                error
	addParticipantErr      error
	findOneOnOneConv       *model.Conversation
	findOneOnOneErr        error
	isParticipant          bool
	isParticipantErr       error
	getParticipantIDs      []uuid.UUID
	getParticipantIDsErr   error
}

func (m *mockConversationRepo) Create(conv *model.Conversation) error {
	if m.createErr != nil {
		return m.createErr
	}
	if conv.ConversationID == uuid.Nil {
		conv.ConversationID = uuid.MustParse("a0000000-0000-0000-0000-000000000001")
	}
	return nil
}
func (m *mockConversationRepo) GetByID(conversationID uuid.UUID) (*model.Conversation, error) {
	return m.getByIDConv, m.getByIDErr
}
func (m *mockConversationRepo) ListByUserID(userID uuid.UUID, limit, offset int) ([]*model.Conversation, error) {
	return m.listConvs, m.listErr
}
func (m *mockConversationRepo) AddParticipant(p *model.ConversationParticipant) error {
	return m.addParticipantErr
}
func (m *mockConversationRepo) FindOneOnOneBetween(userID1, userID2 uuid.UUID) (*model.Conversation, error) {
	return m.findOneOnOneConv, m.findOneOnOneErr
}
func (m *mockConversationRepo) IsParticipant(conversationID, userID uuid.UUID) (bool, error) {
	return m.isParticipant, m.isParticipantErr
}
func (m *mockConversationRepo) GetParticipantUserIDs(conversationID uuid.UUID) ([]uuid.UUID, error) {
	return m.getParticipantIDs, m.getParticipantIDsErr
}

type mockUserRepo struct {
	getByIDUser *model.User
	getByIDErr  error
}

func (m *mockUserRepo) Create(user *model.User) error { return nil }
func (m *mockUserRepo) GetByID(userID uuid.UUID) (*model.User, error) {
	return m.getByIDUser, m.getByIDErr
}
func (m *mockUserRepo) GetByUsername(username string) (*model.User, error) { return nil, nil }
func (m *mockUserRepo) GetByEmail(email string) (*model.User, error)       { return nil, nil }
func (m *mockUserRepo) Update(user *model.User) error                       { return nil }
func (m *mockUserRepo) Delete(userID uuid.UUID) error                       { return nil }

func TestConversationService_CreateOneOnOne_SameUser(t *testing.T) {
	uid := uuid.MustParse("b0000000-0000-0000-0000-000000000001")
	convRepo := &mockConversationRepo{}
	userRepo := &mockUserRepo{}
	svc := NewConversationService(convRepo, userRepo)
	_, err := svc.CreateOneOnOne(uid, uid)
	if err == nil {
		t.Fatal("expected error for same user")
	}
	if !errors.Is(err, ErrInvalidConversation) {
		t.Errorf("expected ErrInvalidConversation, got %v", err)
	}
}

func TestConversationService_CreateOneOnOne_OtherUserNotFound(t *testing.T) {
	creator := uuid.MustParse("b0000000-0000-0000-0000-000000000001")
	other := uuid.MustParse("b0000000-0000-0000-0000-000000000002")
	convRepo := &mockConversationRepo{}
	userRepo := &mockUserRepo{getByIDErr: errors.New("not found")}
	svc := NewConversationService(convRepo, userRepo)
	_, err := svc.CreateOneOnOne(creator, other)
	if err == nil {
		t.Fatal("expected error when other user not found")
	}
	if !errors.Is(err, ErrUserNotFound) {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}
}

func TestConversationService_CreateOneOnOne_Existing(t *testing.T) {
	creator := uuid.MustParse("b0000000-0000-0000-0000-000000000001")
	other := uuid.MustParse("b0000000-0000-0000-0000-000000000002")
	existingConv := &model.Conversation{ConversationID: uuid.MustParse("c0000000-0000-0000-0000-000000000001"), Type: model.ConversationTypeOneOnOne}
	convRepo := &mockConversationRepo{
		findOneOnOneConv: existingConv,
		findOneOnOneErr:  nil,
	}
	userRepo := &mockUserRepo{getByIDUser: &model.User{UserID: other}}
	svc := NewConversationService(convRepo, userRepo)
	conv, err := svc.CreateOneOnOne(creator, other)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if conv.ConversationID != existingConv.ConversationID {
		t.Errorf("expected existing conversation %v, got %v", existingConv.ConversationID, conv.ConversationID)
	}
}

func TestConversationService_CreateOneOnOne_New(t *testing.T) {
	creator := uuid.MustParse("b0000000-0000-0000-0000-000000000001")
	other := uuid.MustParse("b0000000-0000-0000-0000-000000000002")
	convRepo := &mockConversationRepo{
		findOneOnOneErr: errors.New("not found"),
	}
	userRepo := &mockUserRepo{getByIDUser: &model.User{UserID: other}}
	svc := NewConversationService(convRepo, userRepo)
	conv, err := svc.CreateOneOnOne(creator, other)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if conv.ConversationID == uuid.Nil {
		t.Error("expected new conversation to have id")
	}
	if conv.Type != model.ConversationTypeOneOnOne {
		t.Errorf("expected type one_on_one, got %s", conv.Type)
	}
}

func TestConversationService_GetByID_NotParticipant(t *testing.T) {
	convID := uuid.MustParse("c0000000-0000-0000-0000-000000000001")
	userID := uuid.MustParse("b0000000-0000-0000-0000-000000000001")
	convRepo := &mockConversationRepo{isParticipant: false}
	userRepo := &mockUserRepo{}
	svc := NewConversationService(convRepo, userRepo)
	_, err := svc.GetByID(convID, userID)
	if err == nil {
		t.Fatal("expected error when not participant")
	}
	if !errors.Is(err, ErrNotParticipant) {
		t.Errorf("expected ErrNotParticipant, got %v", err)
	}
}

func TestConversationService_GetByID_Success(t *testing.T) {
	convID := uuid.MustParse("c0000000-0000-0000-0000-000000000001")
	userID := uuid.MustParse("b0000000-0000-0000-0000-000000000001")
	expected := &model.Conversation{ConversationID: convID, Type: model.ConversationTypeOneOnOne}
	convRepo := &mockConversationRepo{isParticipant: true, getByIDConv: expected}
	userRepo := &mockUserRepo{}
	svc := NewConversationService(convRepo, userRepo)
	conv, err := svc.GetByID(convID, userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if conv.ConversationID != convID {
		t.Errorf("expected conv %v, got %v", convID, conv.ConversationID)
	}
}

func TestConversationService_ListByUserID(t *testing.T) {
	userID := uuid.MustParse("b0000000-0000-0000-0000-000000000001")
	list := []*model.Conversation{
		{ConversationID: uuid.MustParse("c0000000-0000-0000-0000-000000000001"), Type: model.ConversationTypeOneOnOne},
	}
	convRepo := &mockConversationRepo{listConvs: list}
	userRepo := &mockUserRepo{}
	svc := NewConversationService(convRepo, userRepo)
	convs, err := svc.ListByUserID(userID, 20, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(convs) != 1 || convs[0].ConversationID != list[0].ConversationID {
		t.Errorf("expected list %v, got %v", list, convs)
	}
}

// Ensure mockConversationRepo implements repository.ConversationRepository
var _ repository.ConversationRepository = (*mockConversationRepo)(nil)
var _ repository.UserRepository = (*mockUserRepo)(nil)
