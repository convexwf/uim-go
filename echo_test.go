package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
)

// TestEcho tests the echo WebSocket server.
func TestEcho(t *testing.T) {
	// Create a test server to serve the echo handler
	server := httptest.NewServer(http.HandlerFunc(echo))
	defer server.Close()

	url := "ws" + strings.TrimPrefix(server.URL, "http") + "/echo"

	// Upgrade the recorded response to a WebSocket connection
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("Dial error: %v", err)
	}
	defer conn.Close()

	// Send a message to the server
	message := []byte("Hello, WebSocket!")
	err = conn.WriteMessage(websocket.TextMessage, message)
	if err != nil {
		t.Fatalf("write error: %v", err)
	}

	// Read the echo message from the server
	_, msg, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read error: %v", err)
	}

	// Verify that the echoed message is the same as the sent message
	if string(msg) != string(message) {
		t.Errorf("unexpected message: got %v want %v", string(msg), string(message))
	}
}
