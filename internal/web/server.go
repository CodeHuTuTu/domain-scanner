package web

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"strconv"
	"time"

	"domain_scanner/internal/database"
	"domain_scanner/internal/generator"
	"domain_scanner/internal/types"
	"domain_scanner/internal/worker"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

//go:embed static/*
var staticFiles embed.FS

type Server struct {
	db     *database.Database
	router *mux.Router
}

type ScanRequest struct {
	Length       int    `json:"length"`
	Suffix       string `json:"suffix"`
	Pattern      string `json:"pattern"`
	RegexFilter  string `json:"regex_filter"`
	DictFile     string `json:"dict_file"`
	Delay        int    `json:"delay"`
	Workers      int    `json:"workers"`
}

type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func NewServer(db *database.Database) *Server {
	s := &Server{
		db:     db,
		router: mux.NewRouter(),
	}

	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	// API routes
	api := s.router.PathPrefix("/api").Subrouter()
	api.HandleFunc("/domains", s.handleGetDomains).Methods("GET")
	api.HandleFunc("/domains/search", s.handleSearchDomains).Methods("GET")
	api.HandleFunc("/sessions", s.handleGetSessions).Methods("GET")
	api.HandleFunc("/stats", s.handleGetStats).Methods("GET")
	api.HandleFunc("/scan", s.handleStartScan).Methods("POST")

	// Serve static files
	staticFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		panic(err)
	}
	s.router.PathPrefix("/").Handler(http.FileServer(http.FS(staticFS)))
}

func (s *Server) handleGetDomains(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()

	var available *bool
	if avail := queryParams.Get("available"); avail != "" {
		val := avail == "true"
		available = &val
	}

	limit, _ := strconv.Atoi(queryParams.Get("limit"))
	if limit <= 0 {
		limit = 50
	}

	offset, _ := strconv.Atoi(queryParams.Get("offset"))
	if offset < 0 {
		offset = 0
	}

	domains, err := s.db.GetDomains(available, limit, offset)
	if err != nil {
		s.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.sendSuccess(w, domains)
}

func (s *Server) handleSearchDomains(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	query := queryParams.Get("q")

	if query == "" {
		s.sendError(w, http.StatusBadRequest, "search query is required")
		return
	}

	var available *bool
	if avail := queryParams.Get("available"); avail != "" {
		val := avail == "true"
		available = &val
	}

	limit, _ := strconv.Atoi(queryParams.Get("limit"))
	if limit <= 0 {
		limit = 50
	}

	offset, _ := strconv.Atoi(queryParams.Get("offset"))
	if offset < 0 {
		offset = 0
	}

	domains, err := s.db.SearchDomains(query, available, limit, offset)
	if err != nil {
		s.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.sendSuccess(w, domains)
}

func (s *Server) handleGetSessions(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()

	limit, _ := strconv.Atoi(queryParams.Get("limit"))
	if limit <= 0 {
		limit = 20
	}

	offset, _ := strconv.Atoi(queryParams.Get("offset"))
	if offset < 0 {
		offset = 0
	}

	sessions, err := s.db.GetScanSessions(limit, offset)
	if err != nil {
		s.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.sendSuccess(w, sessions)
}

func (s *Server) handleGetStats(w http.ResponseWriter, r *http.Request) {
	stats, err := s.db.GetStats()
	if err != nil {
		s.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.sendSuccess(w, stats)
}

func (s *Server) handleStartScan(w http.ResponseWriter, r *http.Request) {
	var req ScanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate request
	if req.Length <= 0 {
		req.Length = 3
	}
	if req.Suffix == "" {
		req.Suffix = ".li"
	}
	if req.Pattern == "" {
		req.Pattern = "D"
	}
	if req.Delay <= 0 {
		req.Delay = 1000
	}
	if req.Workers <= 0 {
		req.Workers = 10
	}

	// Start scan in background
	go s.runScan(req)

	s.sendSuccess(w, map[string]string{
		"message": "Scan started successfully",
	})
}

func (s *Server) runScan(req ScanRequest) {
	domainGen := generator.GenerateDomains(req.Length, req.Suffix, req.Pattern, req.RegexFilter, req.DictFile)

	// Create scan session
	sessionID, err := s.db.CreateScanSession(req.Pattern, req.Length, req.Suffix, domainGen.TotalCount)
	if err != nil {
		fmt.Printf("Error creating scan session: %v\n", err)
		return
	}

	fmt.Printf("Started scan session %d: checking %d domains\n", sessionID, domainGen.TotalCount)

	// Create channels
	jobs := make(chan string, 1000)
	results := make(chan types.DomainResult, 1000)

	// Start workers
	for w := 1; w <= req.Workers; w++ {
		go worker.Worker(w, jobs, results, time.Duration(req.Delay)*time.Millisecond)
	}

	// Send jobs
	go func() {
		defer close(jobs)
		for domain := range domainGen.Domains {
			jobs <- domain
		}
	}()

	// Collect results
	processedCount := 0
	for result := range results {
		processedCount++

		if result.Error != nil {
			continue
		}

		err := s.db.SaveDomainResult(sessionID, result.Domain, result.Available, result.Signatures, req.Pattern, req.Length, req.Suffix)
		if err != nil {
			fmt.Printf("Error saving domain result: %v\n", err)
		}

		if processedCount >= domainGen.TotalCount {
			close(results)
			break
		}
	}

	// Complete session
	if err := s.db.CompleteScanSession(sessionID); err != nil {
		fmt.Printf("Error completing scan session: %v\n", err)
	}

	fmt.Printf("Scan session %d completed: processed %d domains\n", sessionID, processedCount)
}

func (s *Server) sendSuccess(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(APIResponse{
		Success: true,
		Data:    data,
	})
}

func (s *Server) sendError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(APIResponse{
		Success: false,
		Error:   message,
	})
}

func (s *Server) Start(addr string) error {
	// Setup CORS
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"*"},
	})

	handler := c.Handler(s.router)

	fmt.Printf("Web server starting on %s\n", addr)
	fmt.Printf("Open http://localhost%s in your browser\n", addr)

	return http.ListenAndServe(addr, handler)
}

