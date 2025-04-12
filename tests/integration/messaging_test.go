// Copyright 2025 convexwf
//
// Project: uim-go
// File: messaging_test.go
// Description: Integration tests for conversation and message endpoints. Requires DB and seed (make init-db, make seed-db).
package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/convexwf/uim-go/internal/api"
	"github.com/convexwf/uim-go/internal/config"
	"github.com/convexwf/uim-go/internal/middleware"
	"github.com/convexwf/uim-go/internal/pkg/jwt"
	"github.com/convexwf/uim-go/internal/repository"
	"github.com/convexwf/uim-go/internal/service"
	"github.com/convexwf/uim-go/internal/websocket"
	"github.com/joho/godotenv"
	gorillawebsocket "github.com/gorilla/websocket"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupMessagingRouter(t *testing.T) (http.Handler, string) {
	t.Helper()
	_ = godotenv.Load("../../.env")
	cfg, err := config.Load()
	if err != nil {
		t.Skipf("config not loadable: %v", err)
	}
	db, err := gorm.Open(postgres.Open(cfg.Database.DSN()), &gorm.Config{})
	if err != nil {
		t.Skipf("DB not available: %v", err)
	}
	jwtMgr := jwt.NewJWTManager(cfg.JWT.Secret, cfg.JWT.AccessExpiry, cfg.JWT.RefreshExpiry)
	userRepo := repository.NewUserRepository(db)
	convRepo := repository.NewConversationRepository(db)
	msgRepo := repository.NewMessageRepository(db)
	authSvc := service.NewAuthService(userRepo, jwtMgr)
	convSvc := service.NewConversationService(convRepo, userRepo)
	hub := websocket.NewHub(convRepo)
	msgSvc := service.NewMessageService(msgRepo, convSvc, hub)
	router := api.SetupRouter(db, authSvc, jwtMgr, convSvc, msgSvc, hub)
	router.Use(middleware.CORSMiddleware(cfg))
	router.Use(middleware.LoggerMiddlewareSimple())
	router.Use(middleware.ErrorHandlerMiddleware())

	// Login as alice to get token (seed user: alice / password123)
	loginBody := map[string]string{"username": "alice", "password": "password123"}
	loginJSON, _ := json.Marshal(loginBody)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(loginJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Skipf("login failed (seed may be missing): %d %s", w.Code, w.Body.String())
	}
	var loginResp api.AuthResponse
	if err := json.NewDecoder(w.Body).Decode(&loginResp); err != nil {
		t.Skipf("login response decode: %v", err)
	}
	userMap, ok := loginResp.User.(map[string]interface{})
	if !ok {
		t.Skipf("login user not map: %T", loginResp.User)
	}
	userID, _ := userMap["user_id"].(string)
	if userID == "" {
		t.Skip("login response missing user_id")
	}
	return router, loginResp.AccessToken
}

// getBobUserID returns bob's user_id by logging in as bob. Requires seed.
func getBobUserID(t *testing.T, router http.Handler) string {
	t.Helper()
	loginBody := map[string]string{"username": "bob", "password": "password123"}
	loginJSON, _ := json.Marshal(loginBody)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(loginJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Skipf("bob login failed: %d %s", w.Code, w.Body.String())
	}
	var loginResp api.AuthResponse
	_ = json.NewDecoder(w.Body).Decode(&loginResp)
	userMap, ok := loginResp.User.(map[string]interface{})
	if !ok {
		return ""
	}
	bobID, _ := userMap["user_id"].(string)
	return bobID
}

func TestConversationCreateAndList(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping messaging integration test in short mode")
	}
	router, token := setupMessagingRouter(t)
	bobID := getBobUserID(t, router)
	if bobID == "" {
		t.Skip("could not get bob user_id")
	}

	// Create 1:1 conversation (alice with bob)
	createBody := map[string]string{"other_user_id": bobID}
	createJSON, _ := json.Marshal(createBody)
	req := httptest.NewRequest(http.MethodPost, "/api/conversations", bytes.NewReader(createJSON))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("create conversation: status %d, body %s", w.Code, w.Body.String())
	}
	var conv map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&conv); err != nil {
		t.Fatalf("decode conversation: %v", err)
	}
	convID, _ := conv["conversation_id"].(string)
	if convID == "" {
		t.Fatal("response missing conversation_id")
	}

	// List conversations
	req2 := httptest.NewRequest(http.MethodGet, "/api/conversations", nil)
	req2.Header.Set("Authorization", "Bearer "+token)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("list conversations: status %d, body %s", w2.Code, w2.Body.String())
	}
	var listResp struct {
		Conversations []map[string]interface{} `json:"conversations"`
	}
	if err := json.NewDecoder(w2.Body).Decode(&listResp); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if len(listResp.Conversations) < 1 {
		t.Fatalf("expected at least one conversation, got %d", len(listResp.Conversations))
	}
}

func TestMessagesList(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping messaging integration test in short mode")
	}
	router, token := setupMessagingRouter(t)
	bobID := getBobUserID(t, router)
	if bobID == "" {
		t.Skip("could not get bob user_id")
	}

	// Create conversation first
	createBody := map[string]string{"other_user_id": bobID}
	createJSON, _ := json.Marshal(createBody)
	req := httptest.NewRequest(http.MethodPost, "/api/conversations", bytes.NewReader(createJSON))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("create conversation: status %d, body %s", w.Code, w.Body.String())
	}
	var conv map[string]interface{}
	_ = json.NewDecoder(w.Body).Decode(&conv)
	convID, _ := conv["conversation_id"].(string)
	if convID == "" {
		t.Fatal("missing conversation_id")
	}

	// List messages (empty at first)
	req2 := httptest.NewRequest(http.MethodGet, "/api/conversations/"+convID+"/messages", nil)
	req2.Header.Set("Authorization", "Bearer "+token)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("list messages: status %d, body %s", w2.Code, w2.Body.String())
	}
	var msgResp struct {
		Messages []map[string]interface{} `json:"messages"`
	}
	if err := json.NewDecoder(w2.Body).Decode(&msgResp); err != nil {
		t.Fatalf("decode messages: %v", err)
	}
	// May be 0 or more
	_ = msgResp.Messages
}

// TestWebSocketSendMessage connects via WS, sends a message, and verifies it appears in GET /messages.
func TestWebSocketSendMessage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping WebSocket integration test in short mode")
	}
	router, token := setupMessagingRouter(t)
	bobID := getBobUserID(t, router)
	if bobID == "" {
		t.Skip("could not get bob user_id")
	}

	// Create conversation
	createBody := map[string]string{"other_user_id": bobID}
	createJSON, _ := json.Marshal(createBody)
	req := httptest.NewRequest(http.MethodPost, "/api/conversations", bytes.NewReader(createJSON))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("create conversation: %d %s", w.Code, w.Body.String())
	}
	var conv map[string]interface{}
	_ = json.NewDecoder(w.Body).Decode(&conv)
	convID, _ := conv["conversation_id"].(string)
	if convID == "" {
		t.Fatal("missing conversation_id")
	}

	// Start server for WebSocket
	srv := httptest.NewServer(router)
	defer srv.Close()
	wsURL := "ws" + srv.URL[4:] + "/ws?token=" + token
	conn, _, err := gorillawebsocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("ws dial: %v", err)
	}
	defer conn.Close()

	// Send message via WebSocket
	clientMsg := map[string]string{
		"type":            "send_message",
		"conversation_id": convID,
		"content":         "hello from ws",
	}
	payload, _ := json.Marshal(clientMsg)
	if err := conn.WriteMessage(gorillawebsocket.TextMessage, payload); err != nil {
		t.Fatalf("write: %v", err)
	}
	conn.Close() // close so server readPump exits

	// Give server time to persist and broadcast
	time.Sleep(200 * time.Millisecond)

	// Verify message via HTTP GET
	req2 := httptest.NewRequest(http.MethodGet, "/api/conversations/"+convID+"/messages", nil)
	req2.Header.Set("Authorization", "Bearer "+token)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("list messages: %d %s", w2.Code, w2.Body.String())
	}
	var msgResp struct {
		Messages []struct {
			Content string `json:"content"`
		} `json:"messages"`
	}
	if err := json.NewDecoder(w2.Body).Decode(&msgResp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	var found bool
	for _, m := range msgResp.Messages {
		if m.Content == "hello from ws" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected to find message 'hello from ws' in %+v", msgResp.Messages)
	}
}
