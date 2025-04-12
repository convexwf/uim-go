// Copyright 2025 convexwf
//
// Project: uim-go
// File: message_service_test.go
// Description: Unit tests for message service

package service

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/convexwf/uim-go/internal/model"
	"github.com/convexwf/uim-go/internal/repository"
)

type mockMessageRepo struct {
	createErr error
	listMsgs  []*model.Message
	listErr   error
}

func (m *mockMessageRepo) Create(msg *model.Message) error {
	if m.createErr != nil {
		return m.createErr
	}
	if msg.MessageID == 0 {
		msg.MessageID = 1
	}
	return nil
}
func (m *mockMessageRepo) ListByConversationID(convID uuid.UUID, limit, offset int, beforeID *int64) ([]*model.Message, error) {
	return m.listMsgs, m.listErr
}
func (m *mockMessageRepo) GetByID(messageID int64) (*model.Message, error) {
	return nil, nil
}

// mockConvServiceForMessage only implements EnsureUserInConversation behavior for message tests.
type mockConvServiceForMessage struct {
	ensureErr error
}

func (m *mockConvServiceForMessage) CreateOneOnOne(creatorID, otherUserID uuid.UUID) (*model.Conversation, error) {
	return nil, nil
}
func (m *mockConvServiceForMessage) GetByID(conversationID, userID uuid.UUID) (*model.Conversation, error) {
	return nil, nil
}
func (m *mockConvServiceForMessage) ListByUserID(userID uuid.UUID, limit, offset int) ([]*model.Conversation, error) {
	return nil, nil
}
func (m *mockConvServiceForMessage) EnsureUserInConversation(conversationID, userID uuid.UUID) error {
	return m.ensureErr
}

type mockNotifier struct {
	called  bool
	lastConv uuid.UUID
	lastMsg *model.Message
}

func (m *mockNotifier) NotifyNewMessage(conversationID uuid.UUID, msg *model.Message) {
	m.called = true
	m.lastConv = conversationID
	m.lastMsg = msg
}

func TestMessageService_Create_NotParticipant(t *testing.T) {
	convID := uuid.MustParse("c0000000-0000-0000-0000-000000000001")
	senderID := uuid.MustParse("b0000000-0000-0000-0000-000000000001")
	msgRepo := &mockMessageRepo{}
	convSvc := &mockConvServiceForMessage{ensureErr: ErrNotParticipant}
	svc := NewMessageService(msgRepo, convSvc, nil)
	_, err := svc.Create(convID, senderID, "hello", model.MessageTypeText)
	if err == nil {
		t.Fatal("expected error when not participant")
	}
	if !errors.Is(err, ErrNotParticipant) {
		t.Errorf("expected ErrNotParticipant, got %v", err)
	}
}

func TestMessageService_Create_EmptyContent(t *testing.T) {
	convID := uuid.MustParse("c0000000-0000-0000-0000-000000000001")
	senderID := uuid.MustParse("b0000000-0000-0000-0000-000000000001")
	msgRepo := &mockMessageRepo{}
	convSvc := &mockConvServiceForMessage{}
	svc := NewMessageService(msgRepo, convSvc, nil)
	_, err := svc.Create(convID, senderID, "   ", model.MessageTypeText)
	if err == nil {
		t.Fatal("expected error for empty content")
	}
	if !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got %v", err)
	}
}

func TestMessageService_Create_Success_NotifierCalled(t *testing.T) {
	convID := uuid.MustParse("c0000000-0000-0000-0000-000000000001")
	senderID := uuid.MustParse("b0000000-0000-0000-0000-000000000001")
	msgRepo := &mockMessageRepo{}
	convSvc := &mockConvServiceForMessage{}
	notifier := &mockNotifier{}
	svc := NewMessageService(msgRepo, convSvc, notifier)
	msg, err := svc.Create(convID, senderID, "hello", model.MessageTypeText)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg.MessageID == 0 {
		t.Error("expected message id to be set")
	}
	if msg.Content != "hello" {
		t.Errorf("expected content hello, got %s", msg.Content)
	}
	if !notifier.called {
		t.Error("expected notifier to be called")
	}
	if notifier.lastConv != convID || notifier.lastMsg != msg {
		t.Error("notifier should have received same conv and msg")
	}
}

func TestMessageService_ListByConversationID_NotParticipant(t *testing.T) {
	convID := uuid.MustParse("c0000000-0000-0000-0000-000000000001")
	userID := uuid.MustParse("b0000000-0000-0000-0000-000000000001")
	msgRepo := &mockMessageRepo{}
	convSvc := &mockConvServiceForMessage{ensureErr: ErrNotParticipant}
	svc := NewMessageService(msgRepo, convSvc, nil)
	_, err := svc.ListByConversationID(convID, userID, 50, 0, nil)
	if err == nil {
		t.Fatal("expected error when not participant")
	}
	if !errors.Is(err, ErrNotParticipant) {
		t.Errorf("expected ErrNotParticipant, got %v", err)
	}
}

func TestMessageService_ListByConversationID_Success(t *testing.T) {
	convID := uuid.MustParse("c0000000-0000-0000-0000-000000000001")
	userID := uuid.MustParse("b0000000-0000-0000-0000-000000000001")
	list := []*model.Message{
		{MessageID: 1, ConversationID: convID, SenderID: userID, Content: "hi", MessageType: model.MessageTypeText},
	}
	msgRepo := &mockMessageRepo{listMsgs: list}
	convSvc := &mockConvServiceForMessage{}
	svc := NewMessageService(msgRepo, convSvc, nil)
	msgs, err := svc.ListByConversationID(convID, userID, 50, 0, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(msgs) != 1 || msgs[0].Content != "hi" {
		t.Errorf("expected list with one message hi, got %v", msgs)
	}
}

var _ repository.MessageRepository = (*mockMessageRepo)(nil)
var _ ConversationService = (*mockConvServiceForMessage)(nil)
var _ MessageNotifier = (*mockNotifier)(nil)
