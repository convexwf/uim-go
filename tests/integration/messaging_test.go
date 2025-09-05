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

	"github.com/alicebob/miniredis/v2"
	"github.com/convexwf/uim-go/internal/api"
	"github.com/convexwf/uim-go/internal/config"
	"github.com/convexwf/uim-go/internal/middleware"
	"github.com/convexwf/uim-go/internal/pkg/jwt"
	"github.com/convexwf/uim-go/internal/repository"
	"github.com/convexwf/uim-go/internal/service"
	"github.com/convexwf/uim-go/internal/store"
	"github.com/convexwf/uim-go/internal/websocket"
	gorillawebsocket "github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
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
	applyAllMigrationsForTest(t, db)
	jwtMgr := jwt.NewJWTManager(cfg.JWT.Secret, cfg.JWT.AccessExpiry, cfg.JWT.RefreshExpiry)
	userRepo := repository.NewUserRepository(db)
	convRepo := repository.NewConversationRepository(db)
	contactRepo := repository.NewContactRepository(db)
	msgRepo := repository.NewMessageRepository(db)
	authSvc := service.NewAuthService(userRepo, jwtMgr)
	convSvc := service.NewConversationService(convRepo, userRepo, msgRepo)
	contactSvc := service.NewContactService(contactRepo, userRepo, nil)
	hub := websocket.NewHub(convRepo, nil)
	msgSvc := service.NewMessageService(msgRepo, convSvc, hub)
	router := api.SetupRouter(cfg, db, authSvc, jwtMgr, convSvc, contactSvc, msgSvc, hub, nil, nil, nil)
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

// setupMessagingRouterWithRedis is like setupMessagingRouter but with Redis (miniredis) for offline queue and presence.
func setupMessagingRouterWithRedis(t *testing.T) (http.Handler, string, *miniredis.Miniredis) {
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
	applyAllMigrationsForTest(t, db)
	mr, err := miniredis.Run()
	if err != nil {
		t.Skipf("miniredis: %v", err)
	}
	t.Cleanup(func() { mr.Close() })
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	offlineQueue := store.NewRedisOfflineQueue(rdb)
	presenceStore := store.NewRedisPresenceStore(rdb)

	jwtMgr := jwt.NewJWTManager(cfg.JWT.Secret, cfg.JWT.AccessExpiry, cfg.JWT.RefreshExpiry)
	userRepo := repository.NewUserRepository(db)
	convRepo := repository.NewConversationRepository(db)
	contactRepo := repository.NewContactRepository(db)
	msgRepo := repository.NewMessageRepository(db)
	authSvc := service.NewAuthService(userRepo, jwtMgr)
	convSvc := service.NewConversationService(convRepo, userRepo, msgRepo)
	contactSvc := service.NewContactService(contactRepo, userRepo, presenceStore)
	hub := websocket.NewHub(convRepo, offlineQueue)
	msgSvc := service.NewMessageService(msgRepo, convSvc, hub)
	router := api.SetupRouter(cfg, db, authSvc, jwtMgr, convSvc, contactSvc, msgSvc, hub, rdb, offlineQueue, presenceStore)
	router.Use(middleware.CORSMiddleware(cfg))
	router.Use(middleware.LoggerMiddlewareSimple())
	router.Use(middleware.ErrorHandlerMiddleware())

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
	return router, loginResp.AccessToken, mr
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
	// List response includes metadata (other_user, last_message, unread_count)
	first := listResp.Conversations[0]
	if _, ok := first["unread_count"]; !ok {
		t.Error("expected conversation list item to include unread_count")
	}
	if _, ok := first["other_user"]; !ok {
		t.Error("expected 1:1 conversation list item to include other_user")
	}
}

func TestMarkRead(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping messaging integration test in short mode")
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
		t.Fatalf("create conversation: status %d, body %s", w.Code, w.Body.String())
	}
	var conv map[string]interface{}
	_ = json.NewDecoder(w.Body).Decode(&conv)
	convID, _ := conv["conversation_id"].(string)
	if convID == "" {
		t.Fatal("missing conversation_id")
	}
	// Mark read
	readBody := map[string]int64{"last_read_message_id": 0}
	readJSON, _ := json.Marshal(readBody)
	reqRead := httptest.NewRequest(http.MethodPost, "/api/conversations/"+convID+"/read", bytes.NewReader(readJSON))
	reqRead.Header.Set("Content-Type", "application/json")
	reqRead.Header.Set("Authorization", "Bearer "+token)
	wRead := httptest.NewRecorder()
	router.ServeHTTP(wRead, reqRead)
	if wRead.Code != http.StatusNoContent {
		t.Fatalf("mark read: expected 204, got %d, body %s", wRead.Code, wRead.Body.String())
	}
}

func TestConversationDelete(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping messaging integration test in short mode")
	}
	router, token := setupMessagingRouter(t)
	bobID := getBobUserID(t, router)
	if bobID == "" {
		t.Skip("could not get bob user_id")
	}

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

	delReq := httptest.NewRequest(http.MethodDelete, "/api/conversations/"+convID, nil)
	delReq.Header.Set("Authorization", "Bearer "+token)
	delW := httptest.NewRecorder()
	router.ServeHTTP(delW, delReq)
	if delW.Code != http.StatusNoContent {
		t.Fatalf("delete conversation: expected 204, got %d, body %s", delW.Code, delW.Body.String())
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/conversations", nil)
	listReq.Header.Set("Authorization", "Bearer "+token)
	listW := httptest.NewRecorder()
	router.ServeHTTP(listW, listReq)
	if listW.Code != http.StatusOK {
		t.Fatalf("list conversations after delete: status %d, body %s", listW.Code, listW.Body.String())
	}
	var listResp struct {
		Conversations []map[string]interface{} `json:"conversations"`
	}
	if err := json.NewDecoder(listW.Body).Decode(&listResp); err != nil {
		t.Fatalf("decode conversation list: %v", err)
	}
	for _, item := range listResp.Conversations {
		if gotID, _ := item["conversation_id"].(string); gotID == convID {
			t.Fatalf("deleted conversation %s still present in list", convID)
		}
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

// TestOfflineMessageDelivery: Alice sends a message while Bob is offline; Bob connects and receives it from the offline queue.
func TestOfflineMessageDelivery(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping offline message integration test in short mode")
	}
	router, aliceToken, _ := setupMessagingRouterWithRedis(t)
	bobID := getBobUserID(t, router)
	if bobID == "" {
		t.Skip("could not get bob user_id")
	}

	// Create conversation (alice with bob)
	createBody := map[string]string{"other_user_id": bobID}
	createJSON, _ := json.Marshal(createBody)
	req := httptest.NewRequest(http.MethodPost, "/api/conversations", bytes.NewReader(createJSON))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+aliceToken)
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

	// Bob is NOT connected. Alice sends via WebSocket (message goes to offline queue for Bob).
	srv := httptest.NewServer(router)
	defer srv.Close()
	aliceWSURL := "ws" + srv.URL[4:] + "/ws?token=" + aliceToken
	aliceConn, _, err := gorillawebsocket.DefaultDialer.Dial(aliceWSURL, nil)
	if err != nil {
		t.Fatalf("alice ws dial: %v", err)
	}
	clientMsg := map[string]string{
		"type":            "send_message",
		"conversation_id": convID,
		"content":         "offline message for bob",
	}
	payload, _ := json.Marshal(clientMsg)
	if err := aliceConn.WriteMessage(gorillawebsocket.TextMessage, payload); err != nil {
		t.Fatalf("alice write: %v", err)
	}
	time.Sleep(100 * time.Millisecond)
	aliceConn.Close()

	// Bob was offline; message should be in queue. Now Bob logs in and connects WS; he should receive the offline message.
	bobToken := getBobToken(t, router)
	if bobToken == "" {
		t.Skip("could not get bob token")
	}
	bobWSURL := "ws" + srv.URL[4:] + "/ws?token=" + bobToken
	bobConn, _, err := gorillawebsocket.DefaultDialer.Dial(bobWSURL, nil)
	if err != nil {
		t.Fatalf("bob ws dial: %v", err)
	}
	defer bobConn.Close()

	// Bob should receive one new_message (offline delivery)
	_ = bobConn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, raw, err := bobConn.ReadMessage()
	if err != nil {
		t.Fatalf("bob read: %v", err)
	}
	var envelope map[string]interface{}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		t.Fatalf("bob unmarshal: %v", err)
	}
	if envelope["type"] != "new_message" {
		t.Errorf("expected type new_message, got %q", envelope["type"])
	}
	msg, _ := envelope["message"].(map[string]interface{})
	if msg != nil && msg["content"] != "offline message for bob" {
		t.Errorf("expected content 'offline message for bob', got %v", msg)
	}
}

func getBobToken(t *testing.T, router http.Handler) string {
	t.Helper()
	loginBody := map[string]string{"username": "bob", "password": "password123"}
	loginJSON, _ := json.Marshal(loginBody)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(loginJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		return ""
	}
	var loginResp api.AuthResponse
	_ = json.NewDecoder(w.Body).Decode(&loginResp)
	return loginResp.AccessToken
}

// TestPresenceAPI: GET /api/users/:id/presence returns online when user is connected, offline when not.
func TestPresenceAPI(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping presence integration test in short mode")
	}
	router, aliceToken, _ := setupMessagingRouterWithRedis(t)
	aliceID := getAliceUserID(t, router, aliceToken)
	if aliceID == "" {
		t.Skip("could not get alice user_id")
	}

	srv := httptest.NewServer(router)
	defer srv.Close()

	// Before connecting: presence should be offline
	req := httptest.NewRequest(http.MethodGet, "/api/users/"+aliceID+"/presence", nil)
	req.Header.Set("Authorization", "Bearer "+aliceToken)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("GET presence: %d %s", w.Code, w.Body.String())
	}
	var pres api.PresenceResponse
	if err := json.NewDecoder(w.Body).Decode(&pres); err != nil {
		t.Fatalf("decode presence: %v", err)
	}
	if pres.Status != "offline" {
		t.Errorf("before connect: expected offline, got %q", pres.Status)
	}

	// Connect WebSocket -> presence online
	wsURL := "ws" + srv.URL[4:] + "/ws?token=" + aliceToken
	conn, _, err := gorillawebsocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("ws dial: %v", err)
	}
	time.Sleep(100 * time.Millisecond)
	req2 := httptest.NewRequest(http.MethodGet, "/api/users/"+aliceID+"/presence", nil)
	req2.Header.Set("Authorization", "Bearer "+aliceToken)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("GET presence (connected): %d", w2.Code)
	}
	if err := json.NewDecoder(w2.Body).Decode(&pres); err != nil {
		t.Fatalf("decode presence: %v", err)
	}
	if pres.Status != "online" {
		t.Errorf("after connect: expected online, got %q", pres.Status)
	}
	conn.Close()
	time.Sleep(100 * time.Millisecond)

	// After disconnect: presence should be offline
	req3 := httptest.NewRequest(http.MethodGet, "/api/users/"+aliceID+"/presence", nil)
	req3.Header.Set("Authorization", "Bearer "+aliceToken)
	w3 := httptest.NewRecorder()
	router.ServeHTTP(w3, req3)
	if w3.Code != http.StatusOK {
		t.Fatalf("GET presence (disconnected): %d", w3.Code)
	}
	if err := json.NewDecoder(w3.Body).Decode(&pres); err != nil {
		t.Fatalf("decode presence: %v", err)
	}
	if pres.Status != "offline" {
		t.Errorf("after disconnect: expected offline, got %q", pres.Status)
	}
}

func getAliceUserID(t *testing.T, router http.Handler, token string) string {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/api/conversations", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		return ""
	}
	// We need alice's user_id; get it from login response when we have token. For simplicity, decode from token or use a dedicated /me. From setup we have token from alice login - we don't have alice ID stored. So we need to get current user. Check if there's GET /users/me or similar.
	// setupMessagingRouterWithRedis returns token but not user_id. We could return both from setup. Alternatively, call login again for alice and get user_id from response. Let me add a helper that logs in as alice and returns (token, userID).
	loginBody := map[string]string{"username": "alice", "password": "password123"}
	loginJSON, _ := json.Marshal(loginBody)
	reqLogin := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(loginJSON))
	reqLogin.Header.Set("Content-Type", "application/json")
	wLogin := httptest.NewRecorder()
	router.ServeHTTP(wLogin, reqLogin)
	if wLogin.Code != http.StatusOK {
		return ""
	}
	var loginResp api.AuthResponse
	_ = json.NewDecoder(wLogin.Body).Decode(&loginResp)
	userMap, ok := loginResp.User.(map[string]interface{})
	if !ok {
		return ""
	}
	id, _ := userMap["user_id"].(string)
	return id
}
