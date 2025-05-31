package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/bresrch/sawmill"
	"github.com/bresrch/sawmill/plugins"
)

// User represents a user in our system
type User struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Username string `json:"username"`
}

// Server holds our application state
type Server struct {
	logger sawmill.Logger
	users  []User
}

// NewServer creates a new server instance
func NewServer() *Server {
	// Create logger with JSON output
	handler := sawmill.NewJSONHandler(
		sawmill.NewHandlerOptions().
			WithLevel(sawmill.LevelInfo).
			WithStdout(),
	)

	logger := sawmill.New(handler)

	// Sample users data
	users := []User{
		{ID: 1, Name: "Alice Johnson", Email: "alice@example.com", Username: "alice"},
		{ID: 2, Name: "Bob Smith", Email: "bob@example.com", Username: "bob"},
		{ID: 3, Name: "Carol Davis", Email: "carol@example.com", Username: "carol"},
	}

	return &Server{
		logger: logger,
		users:  users,
	}
}

// LoggingMiddleware wraps handlers with automatic request/response logging
func (s *Server) LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Start timing
		start := time.Now()

		// Create response writer wrapper to capture status and size
		wrapper := &ResponseWriterWrapper{
			ResponseWriter: w,
			statusCode:     200, // Default
		}

		// Add request context values (simulating auth middleware)
		ctx := r.Context()
		ctx = context.WithValue(ctx, "request_id", generateRequestID())
		ctx = context.WithValue(ctx, "user_id", 12345) // From auth
		ctx = context.WithValue(ctx, "session_id", "sess-"+generateRequestID())

		// Process request
		next.ServeHTTP(wrapper, r.WithContext(ctx))

		// Calculate duration
		duration := time.Since(start)

		// Log request with plugins
		requestLogger := plugins.WithHTTPResponseWriter(
			plugins.WithHTTPRequest(
				plugins.WithContext(s.logger, ctx), r), wrapper)

		requestLogger.Info("HTTP request completed",
			"duration_ms", duration.Milliseconds())
	})
}

// ResponseWriterWrapper captures response details
type ResponseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
	size       int
}

func (w *ResponseWriterWrapper) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *ResponseWriterWrapper) Write(b []byte) (int, error) {
	n, err := w.ResponseWriter.Write(b)
	w.size += n
	return n, err
}

func (w *ResponseWriterWrapper) Status() int {
	return w.statusCode
}

func (w *ResponseWriterWrapper) Size() int {
	return w.size
}

// GetUsers handles GET /api/users
func (s *Server) GetUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := plugins.WithContext(s.logger, ctx)

	logger.Info("Fetching users list")

	// Parse query parameters
	page := 1
	limit := 10
	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil {
			page = parsed
		}
	}
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}

	logger.Debug("Pagination parameters", "page", page, "limit", limit)

	// Calculate pagination
	start := (page - 1) * limit
	end := start + limit
	if start > len(s.users) {
		start = len(s.users)
	}
	if end > len(s.users) {
		end = len(s.users)
	}

	paginatedUsers := s.users[start:end]

	// Log business logic
	logger.Info("Users retrieved successfully",
		"total_users", len(s.users),
		"returned_users", len(paginatedUsers),
		"page", page,
		"limit", limit)

	// Prepare response
	response := map[string]interface{}{
		"users": paginatedUsers,
		"pagination": map[string]int{
			"page":        page,
			"limit":       limit,
			"total":       len(s.users),
			"total_pages": (len(s.users) + limit - 1) / limit,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Response-ID", "resp-"+generateRequestID())
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error("Failed to encode response", "error", err.Error())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// GetUser handles GET /api/users/{id}
func (s *Server) GetUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := plugins.WithContext(s.logger, ctx)

	// Extract user ID from path
	userIDStr := r.URL.Path[len("/api/users/"):]
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		logger.Warn("Invalid user ID format", "user_id", userIDStr, "error", err.Error())
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	logger.Info("Fetching user by ID", "user_id", userID)

	// Find user
	var user *User
	for _, u := range s.users {
		if u.ID == userID {
			user = &u
			break
		}
	}

	if user == nil {
		logger.Warn("User not found", "user_id", userID)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	logger.Info("User found successfully",
		"user_id", user.ID,
		"username", user.Username)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Response-ID", "resp-"+generateRequestID())
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(user); err != nil {
		logger.Error("Failed to encode user response",
			"user_id", userID,
			"error", err.Error())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// CreateUser handles POST /api/users
func (s *Server) CreateUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := plugins.WithContext(s.logger, ctx)

	logger.Info("Creating new user")

	var newUser User
	if err := json.NewDecoder(r.Body).Decode(&newUser); err != nil {
		logger.Warn("Invalid request body", "error", err.Error())
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if newUser.Name == "" || newUser.Email == "" {
		logger.Warn("Missing required fields",
			"name", newUser.Name,
			"email", newUser.Email)
		http.Error(w, "Name and email are required", http.StatusBadRequest)
		return
	}

	// Assign ID and add to users
	newUser.ID = len(s.users) + 1
	s.users = append(s.users, newUser)

	logger.Info("User created successfully",
		"user_id", newUser.ID,
		"username", newUser.Username,
		"email", newUser.Email)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Response-ID", "resp-"+generateRequestID())
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(newUser); err != nil {
		logger.Error("Failed to encode new user response",
			"user_id", newUser.ID,
			"error", err.Error())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// Health handles GET /health
func (s *Server) Health(w http.ResponseWriter, r *http.Request) {
	logger := plugins.WithRuntime(s.logger)

	logger.Debug("Health check requested")

	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
		"users":     len(s.users),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(health); err != nil {
		logger.Error("Failed to encode health response", "error", err.Error())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	logger.Info("Health check completed")
}

// SetupRoutes configures the HTTP routes
func (s *Server) SetupRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	// Wrap all handlers with logging middleware
	mux.Handle("/api/users", s.LoggingMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			s.GetUsers(w, r)
		case http.MethodPost:
			s.CreateUser(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))

	mux.Handle("/api/users/", s.LoggingMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			s.GetUser(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))

	mux.Handle("/health", s.LoggingMiddleware(http.HandlerFunc(s.Health)))

	return mux
}

// generateRequestID creates a simple request ID
func generateRequestID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano()%100000)
}

func main() {
	fmt.Println("=== Sawmill HTTP Plugin Example ===")

	server := NewServer()
	mux := server.SetupRoutes()

	// Start server
	httpServer := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	go func() {
		server.logger.Info("Starting HTTP server", "port", 8080)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			server.logger.Error("Server failed to start", "error", err.Error())
			os.Exit(1)
		}
	}()

	// Give server time to start
	time.Sleep(500 * time.Millisecond)

	// Make test requests
	fmt.Println("\nMaking test requests...\n")
	makeTestRequests()

	// Shutdown server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	server.logger.Info("Shutting down server")
	if err := httpServer.Shutdown(ctx); err != nil {
		server.logger.Error("Server shutdown failed", "error", err.Error())
	}
}

// makeTestRequests demonstrates the API
func makeTestRequests() {
	client := &http.Client{Timeout: 10 * time.Second}
	baseURL := "http://localhost:8080"

	// Test 1: Get all users
	fmt.Println("1. GET /api/users")
	resp, err := client.Get(baseURL + "/api/users")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Status: %d\n", resp.StatusCode)
		resp.Body.Close()
	}

	time.Sleep(100 * time.Millisecond)

	// Test 2: Get users with pagination
	fmt.Println("\n2. GET /api/users?page=1&limit=2")
	resp, err = client.Get(baseURL + "/api/users?page=1&limit=2")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Status: %d\n", resp.StatusCode)
		resp.Body.Close()
	}

	time.Sleep(100 * time.Millisecond)

	// Test 3: Get specific user
	fmt.Println("\n3. GET /api/users/2")
	resp, err = client.Get(baseURL + "/api/users/2")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Status: %d\n", resp.StatusCode)
		resp.Body.Close()
	}

	time.Sleep(100 * time.Millisecond)

	// Test 4: Get non-existent user
	fmt.Println("\n4. GET /api/users/999 (not found)")
	resp, err = client.Get(baseURL + "/api/users/999")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Status: %d\n", resp.StatusCode)
		resp.Body.Close()
	}

	time.Sleep(100 * time.Millisecond)

	// Test 5: Create new user
	fmt.Println("\n5. POST /api/users (create user)")
	newUser := `{"name": "David Wilson", "email": "david@example.com", "username": "david"}`
	resp, err = client.Post(baseURL+"/api/users", "application/json", strings.NewReader(newUser))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Status: %d\n", resp.StatusCode)
		resp.Body.Close()
	}

	time.Sleep(100 * time.Millisecond)

	// Test 6: Health check
	fmt.Println("\n6. GET /health")
	resp, err = client.Get(baseURL + "/health")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Status: %d\n", resp.StatusCode)
		resp.Body.Close()
	}

	fmt.Println("\nTest requests completed!")
}
