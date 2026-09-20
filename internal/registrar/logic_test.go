package registrar

import (
	"encoding/json"
	"testing"
	"time"
)

func TestBuildTraefikConfig_Empty(t *testing.T) {
	cfg := BuildTraefikConfig(nil)

	if cfg.HTTP.Routers == nil || cfg.HTTP.Services == nil {
		t.Fatalf("Routers or Services map is nil")
	}

	if len(cfg.HTTP.Routers) != 0 || len(cfg.HTTP.Services) != 0 {
		t.Fatalf("Expected 0 routers and services, got %d and %d", len(cfg.HTTP.Routers), len(cfg.HTTP.Services))
	}

	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	expectedJSON := `{"http":{"routers":{},"services":{}}}`
	if string(data) != expectedJSON {
		t.Errorf("Expected JSON %s, got %s", expectedJSON, string(data))
	}
}

func TestBuildTraefikConfig_SingleRouteWithoutAuth(t *testing.T) {
	routes := []Route{
		{
			ProjectSlug:  "armario",
			Host:         "armario.apps.minsau.dev",
			Backend:      "192.168.0.105:8000",
			AuthRequired: false,
			UpdatedAt:    time.Now(),
		},
	}

	cfg := BuildTraefikConfig(routes)

	router, exists := cfg.HTTP.Routers["armario"]
	if !exists {
		t.Fatalf("Expected router for 'armario' not found")
	}

	if router.Rule != "Host(`armario.apps.minsau.dev`)" {
		t.Errorf("Unexpected rule: %s", router.Rule)
	}
	if router.Service != "armario" {
		t.Errorf("Unexpected service: %s", router.Service)
	}
	if len(router.EntryPoints) != 1 || router.EntryPoints[0] != "web" {
		t.Errorf("Unexpected entry points: %v", router.EntryPoints)
	}
	if len(router.Middlewares) != 0 {
		t.Errorf("Expected no middlewares, got %v", router.Middlewares)
	}

	svc, exists := cfg.HTTP.Services["armario"]
	if !exists {
		t.Fatalf("Expected service for 'armario' not found")
	}

	if len(svc.LoadBalancer.Servers) != 1 {
		t.Fatalf("Expected 1 server, got %d", len(svc.LoadBalancer.Servers))
	}
	if svc.LoadBalancer.Servers[0].URL != "http://192.168.0.105:8000" {
		t.Errorf("Unexpected server URL: %s", svc.LoadBalancer.Servers[0].URL)
	}
}

func TestBuildTraefikConfig_SingleRouteWithAuth(t *testing.T) {
	routes := []Route{
		{
			ProjectSlug:  "dashboard",
			Host:         "dashboard.apps.minsau.dev",
			Backend:      "http://192.168.0.105:3000",
			AuthRequired: true,
			UpdatedAt:    time.Now(),
		},
	}

	cfg := BuildTraefikConfig(routes)

	router := cfg.HTTP.Routers["dashboard"]
	if len(router.Middlewares) != 1 || router.Middlewares[0] != "auth@file" {
		t.Errorf("Expected middleware auth@file, got %v", router.Middlewares)
	}

	svc := cfg.HTTP.Services["dashboard"]
	// Should not have double http:// prefix
	if svc.LoadBalancer.Servers[0].URL != "http://192.168.0.105:3000" {
		t.Errorf("Unexpected server URL with existing prefix: %s", svc.LoadBalancer.Servers[0].URL)
	}
}
