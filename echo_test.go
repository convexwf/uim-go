package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// setupRouter sets up a gin engine for testing.
func setupRouter() *gin.Engine {
	router := gin.Default()
	router.GET("/echo", echo)
	return router
}

// TestEcho tests the echo WebSocket server.
func TestEcho(t *testing.T) {
	// Create a test server to serve the echo handler
	router := setupRouter()

	recorder := httptest.NewRecorder()
	request, err := http.NewRequest("GET", "/echo", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	router.ServeHTTP(recorder, request)

	url := "ws" + strings.TrimPrefix(recorder.Result().Header.Get("Location"), "http") + "/echo"

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
