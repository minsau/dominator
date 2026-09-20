package registrar

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
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

func (m *mockRepository) GetRoutesBySlug(ctx context.Context, slug string) ([]Route, error) {
	if m.err != nil {
		return nil, m.err
	}
	var res []Route
	for _, r := range m.routes {
		if r.ProjectSlug == slug || r.Slug == slug {
			res = append(res, r)
		}
	}
	return res, nil
}

func (m *mockRepository) UpsertRoute(ctx context.Context, route Route) (Route, error) {
	if m.err != nil {
		return Route{}, m.err
	}
	if route.ProjectSlug == "" {
		route.ProjectSlug = route.Slug
	}
	route.Slug = route.ProjectSlug
	route.UpdatedAt = time.Now()
	// Replace if exists
	found := false
	for i, r := range m.routes {
		if r.ProjectSlug == route.ProjectSlug && r.Host == route.Host {
			m.routes[i] = route
			found = true
			break
		}
	}
	if !found {
		m.routes = append(m.routes, route)
	}
	return route, nil
}

func (m *mockRepository) DeleteRoutesBySlug(ctx context.Context, slug string) error {
	if m.err != nil {
		return m.err
	}
	var filtered []Route
	for _, r := range m.routes {
		if r.ProjectSlug != slug && r.Slug != slug {
			filtered = append(filtered, r)
		}
	}
	m.routes = filtered
	return nil
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

	router, exists := resp.HTTP.Routers["my-app-minsau-dev"]
	if !exists {
		t.Fatalf("Expected router 'my-app-minsau-dev' not found")
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

func TestHandler_RouteEndpoints(t *testing.T) {
	repo := &mockRepository{}
	controller := NewController(repo)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := NewHandler(controller, logger)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	// 1. POST /api/routes
	createBody := `{"project_slug":"test-app","host":"test.internal.minsau.dev","backend":"http://192.168.0.105:3000","internal_only":true,"auth_required":false}`
	req := httptest.NewRequest(http.MethodPost, "/api/routes", strings.NewReader(createBody))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("Expected 201 Created, got %d: %s", rec.Code, rec.Body.String())
	}

	var created Route
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatalf("Failed to decode created route: %v", err)
	}
	if created.ProjectSlug != "test-app" || created.Host != "test.internal.minsau.dev" || !created.InternalOnly {
		t.Errorf("Unexpected created route: %+v", created)
	}

	// 2. GET /api/routes
	req = httptest.NewRequest(http.MethodGet, "/api/routes", nil)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK from GET /api/routes, got %d", rec.Code)
	}
	var routes []Route
	if err := json.Unmarshal(rec.Body.Bytes(), &routes); err != nil {
		t.Fatalf("Failed to decode routes list: %v", err)
	}
	if len(routes) != 1 {
		t.Fatalf("Expected 1 route, got %d", len(routes))
	}

	// 3. GET /api/routes/test-app
	req = httptest.NewRequest(http.MethodGet, "/api/routes/test-app", nil)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK from GET /api/routes/test-app, got %d", rec.Code)
	}
	var appRoutes []Route
	if err := json.Unmarshal(rec.Body.Bytes(), &appRoutes); err != nil {
		t.Fatalf("Failed to decode app routes: %v", err)
	}
	if len(appRoutes) != 1 || appRoutes[0].ProjectSlug != "test-app" {
		t.Fatalf("Unexpected app routes: %+v", appRoutes)
	}

	// 4. DELETE /api/routes/test-app
	req = httptest.NewRequest(http.MethodDelete, "/api/routes/test-app", nil)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK from DELETE /api/routes/test-app, got %d", rec.Code)
	}

	// 5. GET /api/routes/test-app after delete should be empty array
	req = httptest.NewRequest(http.MethodGet, "/api/routes/test-app", nil)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK from GET /api/routes/test-app after delete, got %d", rec.Code)
	}
	var emptyRoutes []Route
	if err := json.Unmarshal(rec.Body.Bytes(), &emptyRoutes); err != nil {
		t.Fatalf("Failed to decode empty routes: %v", err)
	}
	if len(emptyRoutes) != 0 {
		t.Errorf("Expected 0 routes after delete, got %d", len(emptyRoutes))
	}
}

