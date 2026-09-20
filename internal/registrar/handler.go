package registrar

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type Handler struct {
	controller *Controller
	logger     *slog.Logger
}

func NewHandler(controller *Controller, logger *slog.Logger) *Handler {
	return &Handler{
		controller: controller,
		logger:     logger,
	}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/traefik-config", h.TraefikConfig)
	mux.HandleFunc("/health", h.Health)
}

func (h *Handler) TraefikConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	config, err := h.controller.GetTraefikConfig(r.Context())
	if err != nil {
		h.logger.Error("failed to generate traefik config", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(config); err != nil {
		h.logger.Error("failed to encode traefik config response", "error", err)
	}
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	health, err := h.controller.GetHealth(r.Context())
	w.Header().Set("Content-Type", "application/json")

	if err != nil {
		h.logger.Error("health check failed", "error", err)
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(health)
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(health)
}
