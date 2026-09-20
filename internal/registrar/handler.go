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
	mux.HandleFunc("GET /traefik-config", h.TraefikConfig)
	mux.HandleFunc("GET /health", h.Health)
	mux.HandleFunc("POST /api/routes", h.UpsertRoute)
	mux.HandleFunc("GET /api/routes", h.ListRoutes)
	mux.HandleFunc("GET /api/routes/{slug}", h.GetRoutesBySlug)
	mux.HandleFunc("DELETE /api/routes/{slug}", h.DeleteRoutesBySlug)
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

func (h *Handler) writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("failed to encode json response", "error", err)
	}
}

func (h *Handler) UpsertRoute(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var route Route
	if err := json.NewDecoder(r.Body).Decode(&route); err != nil {
		h.logger.Error("failed to decode route request", "error", err)
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	saved, err := h.controller.UpsertRoute(r.Context(), route)
	if err != nil {
		h.logger.Error("failed to upsert route", "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.writeJSON(w, http.StatusCreated, saved)
}

func (h *Handler) ListRoutes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	routes, err := h.controller.ListRoutes(r.Context())
	if err != nil {
		h.logger.Error("failed to list routes", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if routes == nil {
		routes = []Route{}
	}
	h.writeJSON(w, http.StatusOK, routes)
}

func (h *Handler) GetRoutesBySlug(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	slug := r.PathValue("slug")
	if slug == "" {
		slug = r.URL.Query().Get("slug")
	}
	if slug == "" {
		http.Error(w, "Missing slug", http.StatusBadRequest)
		return
	}

	routes, err := h.controller.GetRoutesBySlug(r.Context(), slug)
	if err != nil {
		h.logger.Error("failed to get routes by slug", "slug", slug, "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if routes == nil {
		routes = []Route{}
	}
	h.writeJSON(w, http.StatusOK, routes)
}

func (h *Handler) DeleteRoutesBySlug(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	slug := r.PathValue("slug")
	if slug == "" {
		slug = r.URL.Query().Get("slug")
	}
	if slug == "" {
		http.Error(w, "Missing slug", http.StatusBadRequest)
		return
	}

	if err := h.controller.DeleteRoutesBySlug(r.Context(), slug); err != nil {
		h.logger.Error("failed to delete routes", "slug", slug, "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{"status": "deleted", "slug": slug})
}

