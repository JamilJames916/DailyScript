package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

type Server struct {
	port int
	mux  *http.ServeMux
}

type Response struct {
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	Status    string    `json:"status"`
}

type HealthCheck struct {
	Status    string    `json:"status"`
	Uptime    string    `json:"uptime"`
	Timestamp time.Time `json:"timestamp"`
}

var startTime = time.Now()

func NewServer(port int) *Server {
	return &Server{
		port: port,
		mux:  http.NewServeMux(),
	}
}

func (s *Server) setupRoutes() {
	// Basic routes
	s.mux.HandleFunc("/", s.homeHandler)
	s.mux.HandleFunc("/health", s.healthHandler)
	s.mux.HandleFunc("/api/status", s.statusHandler)
	s.mux.HandleFunc("/api/echo", s.echoHandler)
	s.mux.HandleFunc("/api/time", s.timeHandler)

	// Static file serving
	s.mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))
}

func (s *Server) homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	html := `
<!DOCTYPE html>
<html>
<head>
    <title>Go HTTP Server</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .container { max-width: 600px; margin: 0 auto; }
        .endpoint { background: #f5f5f5; padding: 10px; margin: 10px 0; border-radius: 5px; }
        .method { color: #2196F3; font-weight: bold; }
    </style>
</head>
<body>
    <div class="container">
        <h1>Go HTTP Server</h1>
        <p>Server is running on port ` + strconv.Itoa(s.port) + `</p>
        
        <h2>Available Endpoints:</h2>
        <div class="endpoint">
            <span class="method">GET</span> <code>/health</code> - Health check
        </div>
        <div class="endpoint">
            <span class="method">GET</span> <code>/api/status</code> - Server status
        </div>
        <div class="endpoint">
            <span class="method">POST</span> <code>/api/echo</code> - Echo request body
        </div>
        <div class="endpoint">
            <span class="method">GET</span> <code>/api/time</code> - Current server time
        </div>
        <div class="endpoint">
            <span class="method">GET</span> <code>/static/*</code> - Static files
        </div>
    </div>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	uptime := time.Since(startTime)
	health := HealthCheck{
		Status:    "healthy",
		Uptime:    uptime.String(),
		Timestamp: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}

func (s *Server) statusHandler(w http.ResponseWriter, r *http.Request) {
	response := Response{
		Message:   "Server is running",
		Timestamp: time.Now(),
		Status:    "success",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) echoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var requestData map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"echo":      requestData,
		"timestamp": time.Now(),
		"method":    r.Method,
		"headers":   r.Header,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) timeHandler(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	response := map[string]interface{}{
		"unix":      now.Unix(),
		"rfc3339":   now.Format(time.RFC3339),
		"formatted": now.Format("2006-01-02 15:04:05"),
		"utc":       now.UTC().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) Start() {
	s.setupRoutes()

	fmt.Printf("Starting HTTP server on port %d\n", s.port)
	fmt.Printf("Visit http://localhost:%d for available endpoints\n", s.port)

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", s.port), s.mux))
}

func main() {
	port := 8080

	// Check for port argument
	if len(os.Args) > 1 {
		if p, err := strconv.Atoi(os.Args[1]); err == nil {
			port = p
		} else {
			fmt.Printf("Invalid port number: %s. Using default port 8080\n", os.Args[1])
		}
	}

	server := NewServer(port)
	server.Start()
}
