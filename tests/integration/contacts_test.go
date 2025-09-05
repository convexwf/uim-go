// Copyright 2026 convexwf
//
// Project: uim-go
// File: contacts_test.go
// Description: Integration tests for contacts and exact username search endpoints.
package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/convexwf/uim-go/internal/api"
)

func registerContactTestUser(t *testing.T, router http.Handler, username string) api.AuthResponse {
	t.Helper()
	body := map[string]string{
		"username": username,
		"email":    fmt.Sprintf("%s@example.com", username),
		"password": "password123",
	}
	raw, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("register %s: status %d body %s", username, w.Code, w.Body.String())
	}
	var resp api.AuthResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode register response: %v", err)
	}
	return resp
}

func authUserID(t *testing.T, resp api.AuthResponse) string {
	t.Helper()
	userMap, ok := resp.User.(map[string]interface{})
	if !ok {
		t.Fatalf("user is not map: %T", resp.User)
	}
	userID, _ := userMap["user_id"].(string)
	if userID == "" {
		t.Fatal("missing user_id")
	}
	return userID
}

func TestContactsSearchAddAndList(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping contacts integration test in short mode")
	}
	router, _ := setupMessagingRouter(t)
	suffix := time.Now().UnixNano()
	alice := registerContactTestUser(t, router, fmt.Sprintf("contacts_a_%d", suffix))
	bob := registerContactTestUser(t, router, fmt.Sprintf("contacts_b_%d", suffix))
	aliceToken := alice.AccessToken
	bobToken := bob.AccessToken
	bobID := authUserID(t, bob)
	bobUsername := fmt.Sprintf("contacts_b_%d", suffix)

	// Exact username search before adding.
	searchReq := httptest.NewRequest(http.MethodGet, "/api/users/search?q="+bobUsername, nil)
	searchReq.Header.Set("Authorization", "Bearer "+aliceToken)
	searchW := httptest.NewRecorder()
	router.ServeHTTP(searchW, searchReq)
	if searchW.Code != http.StatusOK {
		t.Fatalf("search users: status %d body %s", searchW.Code, searchW.Body.String())
	}
	var searchResp struct {
		Users []map[string]interface{} `json:"users"`
	}
	if err := json.NewDecoder(searchW.Body).Decode(&searchResp); err != nil {
		t.Fatalf("decode search response: %v", err)
	}
	if len(searchResp.Users) != 1 {
		t.Fatalf("expected exactly one search result, got %d", len(searchResp.Users))
	}
	if added, _ := searchResp.Users[0]["already_added"].(bool); added {
		t.Fatal("expected already_added=false before adding contact")
	}

	// Add bob as alice's contact.
	addBody := map[string]string{"contact_user_id": bobID}
	addJSON, _ := json.Marshal(addBody)
	addReq := httptest.NewRequest(http.MethodPost, "/api/contacts", bytes.NewReader(addJSON))
	addReq.Header.Set("Content-Type", "application/json")
	addReq.Header.Set("Authorization", "Bearer "+aliceToken)
	addW := httptest.NewRecorder()
	router.ServeHTTP(addW, addReq)
	if addW.Code != http.StatusCreated {
		t.Fatalf("add contact: status %d body %s", addW.Code, addW.Body.String())
	}

	// Search again should mark already_added=true.
	searchReq2 := httptest.NewRequest(http.MethodGet, "/api/users/search?q="+bobUsername, nil)
	searchReq2.Header.Set("Authorization", "Bearer "+aliceToken)
	searchW2 := httptest.NewRecorder()
	router.ServeHTTP(searchW2, searchReq2)
	if searchW2.Code != http.StatusOK {
		t.Fatalf("search users after add: status %d body %s", searchW2.Code, searchW2.Body.String())
	}
	var searchResp2 struct {
		Users []map[string]interface{} `json:"users"`
	}
	if err := json.NewDecoder(searchW2.Body).Decode(&searchResp2); err != nil {
		t.Fatalf("decode second search response: %v", err)
	}
	if len(searchResp2.Users) != 1 {
		t.Fatalf("expected one search result after add, got %d", len(searchResp2.Users))
	}
	if added, _ := searchResp2.Users[0]["already_added"].(bool); !added {
		t.Fatal("expected already_added=true after adding contact")
	}

	// Alice should list bob.
	listReq := httptest.NewRequest(http.MethodGet, "/api/contacts", nil)
	listReq.Header.Set("Authorization", "Bearer "+aliceToken)
	listW := httptest.NewRecorder()
	router.ServeHTTP(listW, listReq)
	if listW.Code != http.StatusOK {
		t.Fatalf("list contacts: status %d body %s", listW.Code, listW.Body.String())
	}
	var listResp struct {
		Contacts []map[string]interface{} `json:"contacts"`
	}
	if err := json.NewDecoder(listW.Body).Decode(&listResp); err != nil {
		t.Fatalf("decode contacts response: %v", err)
	}
	if len(listResp.Contacts) != 1 {
		t.Fatalf("expected one contact, got %d", len(listResp.Contacts))
	}
	if gotID, _ := listResp.Contacts[0]["user_id"].(string); gotID != bobID {
		t.Fatalf("expected bob in alice contacts, got %q", gotID)
	}
	if _, ok := listResp.Contacts[0]["presence"]; !ok {
		t.Fatal("expected list item to include presence")
	}

	// Bob should not see alice automatically (single-direction contacts).
	listReqBob := httptest.NewRequest(http.MethodGet, "/api/contacts", nil)
	listReqBob.Header.Set("Authorization", "Bearer "+bobToken)
	listWBob := httptest.NewRecorder()
	router.ServeHTTP(listWBob, listReqBob)
	if listWBob.Code != http.StatusOK {
		t.Fatalf("bob list contacts: status %d body %s", listWBob.Code, listWBob.Body.String())
	}
	var bobListResp struct {
		Contacts []map[string]interface{} `json:"contacts"`
	}
	if err := json.NewDecoder(listWBob.Body).Decode(&bobListResp); err != nil {
		t.Fatalf("decode bob contacts response: %v", err)
	}
	if len(bobListResp.Contacts) != 0 {
		t.Fatalf("expected bob to have zero contacts, got %d", len(bobListResp.Contacts))
	}

	// Delete contact and verify it disappears.
	delReq := httptest.NewRequest(http.MethodDelete, "/api/contacts/"+bobID, nil)
	delReq.Header.Set("Authorization", "Bearer "+aliceToken)
	delW := httptest.NewRecorder()
	router.ServeHTTP(delW, delReq)
	if delW.Code != http.StatusNoContent {
		t.Fatalf("delete contact: status %d body %s", delW.Code, delW.Body.String())
	}

	listReqAfterDelete := httptest.NewRequest(http.MethodGet, "/api/contacts", nil)
	listReqAfterDelete.Header.Set("Authorization", "Bearer "+aliceToken)
	listWAfterDelete := httptest.NewRecorder()
	router.ServeHTTP(listWAfterDelete, listReqAfterDelete)
	if listWAfterDelete.Code != http.StatusOK {
		t.Fatalf("list contacts after delete: status %d body %s", listWAfterDelete.Code, listWAfterDelete.Body.String())
	}
	var listAfterDelete struct {
		Contacts []map[string]interface{} `json:"contacts"`
	}
	if err := json.NewDecoder(listWAfterDelete.Body).Decode(&listAfterDelete); err != nil {
		t.Fatalf("decode contacts after delete: %v", err)
	}
	if len(listAfterDelete.Contacts) != 0 {
		t.Fatalf("expected zero contacts after delete, got %d", len(listAfterDelete.Contacts))
	}
}
