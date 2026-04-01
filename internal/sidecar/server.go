// Package sidecar provides the HTTP server that runs inside OpenShell sandboxes.
//
// Endpoints:
//   POST /evaluate  — evaluate an agent action against governance rules
//   GET  /healthz   — health check for OpenShell monitoring
//   GET  /metrics   — evaluation count, denial rate, uptime
//   GET  /rules     — list loaded rules (Law 2: Transparency)
package sidecar

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/prahari-ai/grdl/internal/cfais"
)

// Server is the CFAIS sidecar HTTP server.
type Server struct {
	engine    *cfais.Engine
	startTime time.Time
	evalCount atomic.Int64
	denyCount atomic.Int64
}

// NewServer creates a new sidecar server with the given engine.
func NewServer(engine *cfais.Engine) *Server {
	return &Server{
		engine:    engine,
		startTime: time.Now(),
	}
}

// Handler returns an http.Handler with all routes registered.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /evaluate", s.handleEvaluate)
	mux.HandleFunc("GET /healthz", s.handleHealthz)
	mux.HandleFunc("GET /metrics", s.handleMetrics)
	mux.HandleFunc("GET /rules", s.handleRules)
	return mux
}

// ListenAndServe starts the sidecar on the given address.
func (s *Server) ListenAndServe(addr string) error {
	log.Printf("CFAIS sidecar listening on %s (%d rules loaded)", addr, s.engine.RuleCount())
	return http.ListenAndServe(addr, s.Handler())
}

func (s *Server) handleEvaluate(w http.ResponseWriter, r *http.Request) {
	var ctx cfais.EvaluationContext
	if err := json.NewDecoder(r.Body).Decode(&ctx); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err), http.StatusBadRequest)
		return
	}

	result := s.engine.Evaluate(&ctx)
	s.evalCount.Add(1)
	if result.Verdict == cfais.Deny {
		s.denyCount.Add(1)
	}

	status := http.StatusOK
	if result.Verdict == cfais.Deny {
		status = http.StatusForbidden
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(result)
}

func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":        "healthy",
		"engine_loaded": s.engine != nil,
		"rules_loaded":  s.engine.RuleCount(),
		"uptime_s":      int(time.Since(s.startTime).Seconds()),
	})
}

func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	evals := s.evalCount.Load()
	denials := s.denyCount.Load()
	var rate float64
	if evals > 0 {
		rate = float64(denials) / float64(evals)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"evaluations_total": evals,
		"denials_total":     denials,
		"denial_rate":       rate,
		"uptime_s":          int(time.Since(s.startTime).Seconds()),
		"rules_loaded":      s.engine.RuleCount(),
	})
}

func (s *Server) handleRules(w http.ResponseWriter, r *http.Request) {
	// Law 2: Transparency — expose loaded rules
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "rule listing available via compiled cfais-runtime.yaml",
	})
}
