package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/lacsar712/snowmelt/internal/model"
)

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", withTimeout(s.handleHealth))
	mux.HandleFunc("GET /snapshot", withTimeout(s.handleSnapshot))
	mux.HandleFunc("GET /telemetry", withTimeout(s.handleTelemetry))
	mux.HandleFunc("GET /health/plant", withTimeout(s.handlePlantHealth))
	mux.HandleFunc("GET /warmup", withTimeout(s.handleWarmup))
	mux.HandleFunc("POST /preheat/start", withTimeout(s.handleStartPreheat))
	mux.HandleFunc("POST /preheat/complete", withTimeout(s.handleCompletePreheat))
	mux.HandleFunc("POST /ignite", withTimeout(s.handleIgnite))
	mux.HandleFunc("POST /trip/reset", withTimeout(s.handleResetTrip))
	mux.HandleFunc("POST /settings", withTimeout(s.handleSettings))
	mux.HandleFunc("GET /glycol/level", withTimeout(s.handleGlycolLevel))
	return cors(logging(mux))
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeOK(w, map[string]string{"status": "ok", "unit": s.app.UnitID()})
}

func (s *Server) handleSnapshot(w http.ResponseWriter, r *http.Request) {
	writeOK(w, s.app.Snapshot())
}

func (s *Server) handleTelemetry(w http.ResponseWriter, r *http.Request) {
	writeOK(w, s.app.Telemetry())
}

func (s *Server) handlePlantHealth(w http.ResponseWriter, r *http.Request) {
	writeOK(w, s.app.PlantHealth())
}

func (s *Server) handleWarmup(w http.ResponseWriter, r *http.Request) {
	ready, detail := s.app.WarmupStatus()
	writeOK(w, map[string]interface{}{
		"ready":  ready,
		"detail": detail,
		"preheat":  s.app.PreheatRemaining(),
		"warmup": s.app.HeatbankWarmupRemaining(),
	})
}

func (s *Server) handleStartPreheat(w http.ResponseWriter, r *http.Request) {
	holder := r.URL.Query().Get("holder")
	if holder == "" {
		holder = "hmi-operator"
	}
	if err := s.app.StartPreheat(r.Context(), holder); err != nil {
		writeErr(w, http.StatusConflict, err)
		return
	}
	writeOK(w, s.app.Snapshot())
}

func (s *Server) handleCompletePreheat(w http.ResponseWriter, r *http.Request) {
	holder := r.URL.Query().Get("holder")
	if holder == "" {
		holder = "hmi-operator"
	}
	if err := s.app.CompletePreheat(r.Context(), holder); err != nil {
		writeErr(w, http.StatusConflict, err)
		return
	}
	writeOK(w, s.app.Snapshot())
}

func (s *Server) handleIgnite(w http.ResponseWriter, r *http.Request) {
	holder := r.URL.Query().Get("holder")
	if holder == "" {
		holder = "hmi-operator"
	}
	if err := s.app.Ignite(r.Context(), holder); err != nil {
		writeErr(w, http.StatusConflict, err)
		return
	}
	writeOK(w, s.app.Snapshot())
}

func (s *Server) handleResetTrip(w http.ResponseWriter, r *http.Request) {
	holder := r.URL.Query().Get("holder")
	if holder == "" {
		holder = "hmi-operator"
	}
	if err := s.app.ResetTrip(r.Context(), holder); err != nil {
		writeErr(w, http.StatusConflict, err)
		return
	}
	writeOK(w, s.app.Snapshot())
}

func (s *Server) handleGlycolLevel(w http.ResponseWriter, r *http.Request) {
	if err := s.app.CheckGlycolLevel(s.app.Snapshot()); err != nil {
		writeErr(w, http.StatusConflict, fmt.Errorf("snowmelt fault: %w", err))
		return
	}
	writeOK(w, map[string]string{"status": "ok"})
}

func (s *Server) handleSettings(w http.ResponseWriter, r *http.Request) {
	var settings model.PlantSettings
	if err := json.NewDecoder(r.Body).Decode(&settings); err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	if err := s.app.UpdateSettings(r.Context(), settings); err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	writeOK(w, s.app.Snapshot())
}
