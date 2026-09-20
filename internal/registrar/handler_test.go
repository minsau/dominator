package registrar

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type mockRepository struct {
	routes []Route
	count  int
	err    error
}

func (m *mockRepository) GetRoutes(ctx context.Context) ([]Route, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.routes, nil
}

func (m *mockRepository) CountRoutes(ctx context.Context) (int, error) {
	if m.err != nil {
		return 0, m.err
	}
	return m.count, nil
}

func (m *mockRepository) Ping(ctx context.Context) error {
	return m.err
}

func TestHandler_TraefikConfig(t *testing.T) {
	repo := &mockRepository{
		routes: []Route{
			{
				ProjectSlug:  "my-app",
				Host:         "my-app.minsau.dev",
				Backend:      "192.168.0.105:8000",
				AuthRequired: true,
				UpdatedAt:    time.Now(),
			},
		},
	}
	controller := NewController(repo)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := NewHandler(controller, logger)

	req := httptest.NewRequest(http.MethodGet, "/traefik-config", nil)
	rec := httptest.NewRecorder()

	handler.TraefikConfig(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", rec.Code)
	}

	var resp TraefikConfigResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	router, exists := resp.HTTP.Routers["my-app"]
	if !exists {
		t.Fatalf("Expected router 'my-app' not found")
	}
	if router.Rule != "Host(`my-app.minsau.dev`)" {
		t.Errorf("Unexpected rule: %s", router.Rule)
	}
	if len(router.Middlewares) != 1 || router.Middlewares[0] != "auth@file" {
		t.Errorf("Expected auth@file middleware")
	}
}

func TestHandler_Health_OK(t *testing.T) {
	repo := &mockRepository{
		count: 3,
	}
	controller := NewController(repo)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := NewHandler(controller, logger)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	handler.Health(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", rec.Code)
	}

	var health HealthResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &health); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if health.Status != "ok" || health.RoutesServed != 3 {
		t.Errorf("Unexpected health response: %+v", health)
	}
}

func TestHandler_Health_Error(t *testing.T) {
	repo := &mockRepository{
		err: errors.New("connection refused"),
	}
	controller := NewController(repo)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := NewHandler(controller, logger)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	handler.Health(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("Expected status 503, got %d", rec.Code)
	}

	var health HealthResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &health); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if health.Status != "error" || health.Error == "" {
		t.Errorf("Unexpected health error response: %+v", health)
	}
}
