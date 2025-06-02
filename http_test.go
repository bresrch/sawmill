package sawmill

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHTTPErrorLog(t *testing.T) {
	// Create a buffer to capture log output
	var buf bytes.Buffer
	
	// Create a logger with text handler writing to buffer
	handler := NewTextHandler(WithWriter(&buf))
	logger := New(handler)
	
	// Create HTTP server with sawmill error log
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	
	server := &http.Server{
		Addr:     ":0",
		Handler:  mux,
		ErrorLog: logger.HTTPErrorLog(),
	}
	
	// Test that the HTTPErrorLog method returns a valid *log.Logger
	if server.ErrorLog == nil {
		t.Fatal("HTTPErrorLog() returned nil")
	}
	
	// Simulate an HTTP error by calling the error logger directly
	server.ErrorLog.Println("test http error message")
	
	// Check that the message was logged
	output := buf.String()
	if !strings.Contains(output, "test http error message") {
		t.Errorf("Expected log output to contain 'test http error message', got: %s", output)
	}
	
	// Check that it's logged at ERROR level
	if !strings.Contains(output, "ERROR") {
		t.Errorf("Expected log output to contain 'ERROR' level, got: %s", output)
	}
}

func TestHTTPErrorLogWithRealServer(t *testing.T) {
	// Create a buffer to capture log output
	var buf bytes.Buffer
	
	// Create a logger with text handler writing to buffer
	handler := NewTextHandler(WithWriter(&buf))
	logger := New(handler)
	
	// Create HTTP server with intentionally broken handler to trigger error
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// This should work fine and not generate errors
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	
	server := httptest.NewServer(mux)
	server.Config.ErrorLog = logger.HTTPErrorLog()
	defer server.Close()
	
	// Make a request to ensure server works
	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}