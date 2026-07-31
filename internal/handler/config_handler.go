package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/vyankatesh-pachpohar/infra-assignments/internal/domain"
	"github.com/vyankatesh-pachpohar/infra-assignments/internal/repository"
	"github.com/vyankatesh-pachpohar/infra-assignments/internal/service"
)

type ConfigHandler struct {
	svc *service.ConfigService
}

func NewConfigHandler(svc *service.ConfigService) *ConfigHandler {
	return &ConfigHandler{svc: svc}
}

// Ping — GET /ping. Returns 200 "pong" if DB is reachable, 503 if not.
// This doubles as both liveness (process is up) and readiness (DB reachable) signal.
func (h *ConfigHandler) Ping(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.HealthCheck(r.Context()); err != nil {
		slog.Error("health check failed", "error", err)
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("db unreachable"))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("pong"))
}

// GetConfig — GET /configs/{id}
func (h *ConfigHandler) GetConfig(w http.ResponseWriter, r *http.Request, id string) {
	cfg, err := h.svc.GetConfig(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			http.Error(w, `{"error":"config not found"}`, http.StatusNotFound)
			return
		}
		slog.Error("get config failed", "id", id, "error", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cfg)
}

// UpsertConfig — POST /configs
func (h *ConfigHandler) UpsertConfig(w http.ResponseWriter, r *http.Request) {
	var cfg domain.Config
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		http.Error(w, `{"error":"invalid json body"}`, http.StatusBadRequest)
		return
	}

	if err := h.svc.SaveConfig(r.Context(), &cfg); err != nil {
		if errors.Is(err, service.ErrInvalidConfig) {
			http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
			return
		}
		slog.Error("upsert config failed", "error", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(cfg)
}

// helper used by main.go's router to parse /configs/{id}
func ParsePathID(path string) (string, bool) {
	const prefix = "/configs/"
	if len(path) <= len(prefix) {
		return "", false
	}
	return path[len(prefix):], true
}

var _ = strconv.Itoa // placeholder import guard if unused later
