package registrar

import (
	"encoding/json"
	"testing"
	"time"
)

func TestBuildTraefikConfig_Empty(t *testing.T) {
	cfg := BuildTraefikConfig(nil)

	if cfg.HTTP != nil {
		t.Fatalf("Expected HTTP to be nil for empty routes")
	}

	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	expectedJSON := `{}`
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
			InternalOnly: false,
			UpdatedAt:    time.Now(),
		},
	}

	cfg := BuildTraefikConfig(routes)

	name := "armario-apps-minsau-dev"
	router, exists := cfg.HTTP.Routers[name]
	if !exists {
		t.Fatalf("Expected router for %q not found", name)
	}

	if router.Rule != "Host(`armario.apps.minsau.dev`)" {
		t.Errorf("Unexpected rule: %s", router.Rule)
	}
	if router.Service != name {
		t.Errorf("Unexpected service: %s", router.Service)
	}
	if len(router.EntryPoints) != 2 || router.EntryPoints[0] != "web" || router.EntryPoints[1] != "websecure" {
		t.Errorf("Unexpected entry points: %v", router.EntryPoints)
	}
	if len(router.Middlewares) != 0 {
		t.Errorf("Expected no middlewares, got %v", router.Middlewares)
	}

	svc, exists := cfg.HTTP.Services[name]
	if !exists {
		t.Fatalf("Expected service for %q not found", name)
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
			InternalOnly: false,
			UpdatedAt:    time.Now(),
		},
	}

	cfg := BuildTraefikConfig(routes)

	name := "dashboard-apps-minsau-dev"
	router, exists := cfg.HTTP.Routers[name]
	if !exists {
		t.Fatalf("Expected router for %q not found", name)
	}
	if len(router.Middlewares) != 1 || router.Middlewares[0] != "auth@file" {
		t.Errorf("Expected middleware auth@file, got %v", router.Middlewares)
	}

	svc, exists := cfg.HTTP.Services[name]
	if !exists {
		t.Fatalf("Expected service for %q not found", name)
	}
	// Should not have double http:// prefix
	if svc.LoadBalancer.Servers[0].URL != "http://192.168.0.105:3000" {
		t.Errorf("Unexpected server URL with existing prefix: %s", svc.LoadBalancer.Servers[0].URL)
	}
}

func TestBuildTraefikConfig_InternalOnly(t *testing.T) {
	routes := []Route{
		{
			ProjectSlug:  "secret-app",
			Host:         "secret.apps.minsau.dev",
			Backend:      "192.168.0.105:4000",
			AuthRequired: false,
			InternalOnly: true,
			UpdatedAt:    time.Now(),
		},
	}

	cfg := BuildTraefikConfig(routes)

	name := "secret-apps-minsau-dev"
	router, exists := cfg.HTTP.Routers[name]
	if !exists {
		t.Fatalf("Expected router for %q not found", name)
	}
	if len(router.Middlewares) != 1 || router.Middlewares[0] != "lan-only@file" {
		t.Errorf("Expected middleware [lan-only@file], got %v", router.Middlewares)
	}
}

func TestBuildTraefikConfig_InternalOnlyAndAuthOrdering(t *testing.T) {
	routes := []Route{
		{
			ProjectSlug:  "private-admin",
			Host:         "private.apps.minsau.dev",
			Backend:      "192.168.0.105:5000",
			AuthRequired: true,
			InternalOnly: true,
			UpdatedAt:    time.Now(),
		},
	}

	cfg := BuildTraefikConfig(routes)

	name := "private-apps-minsau-dev"
	router, exists := cfg.HTTP.Routers[name]
	if !exists {
		t.Fatalf("Expected router for %q not found", name)
	}
	if len(router.Middlewares) != 2 || router.Middlewares[0] != "lan-only@file" || router.Middlewares[1] != "auth@file" {
		t.Errorf("Expected middlewares [lan-only@file, auth@file], got %v", router.Middlewares)
	}
}

func TestBuildTraefikConfig_MultiRouteSameProject(t *testing.T) {
	routes := []Route{
		{
			ProjectSlug:  "life-utilities",
			Host:         "lu.apps.minsau.dev",
			Backend:      "http://192.168.0.105:3000",
			AuthRequired: false,
			InternalOnly: false,
			UpdatedAt:    time.Now(),
		},
		{
			ProjectSlug:  "life-utilities",
			Host:         "lu-api.apps.minsau.dev",
			Backend:      "http://192.168.0.105:8080",
			AuthRequired: false,
			InternalOnly: true,
			UpdatedAt:    time.Now(),
		},
	}

	cfg := BuildTraefikConfig(routes)

	webName := "lu-apps-minsau-dev"
	apiName := "lu-api-apps-minsau-dev"

	webRouter, exists := cfg.HTTP.Routers[webName]
	if !exists {
		t.Fatalf("Expected web router %q not found", webName)
	}
	if len(webRouter.Middlewares) != 0 {
		t.Errorf("Expected no middlewares for web, got %v", webRouter.Middlewares)
	}

	apiRouter, exists := cfg.HTTP.Routers[apiName]
	if !exists {
		t.Fatalf("Expected api router %q not found", apiName)
	}
	if len(apiRouter.Middlewares) != 1 || apiRouter.Middlewares[0] != "lan-only@file" {
		t.Errorf("Expected [lan-only@file] for internal api, got %v", apiRouter.Middlewares)
	}

	webSvc, exists := cfg.HTTP.Services[webName]
	if !exists {
		t.Fatalf("Expected web service %q not found", webName)
	}
	if webSvc.LoadBalancer.Servers[0].URL != "http://192.168.0.105:3000" {
		t.Errorf("Unexpected web server URL: %s", webSvc.LoadBalancer.Servers[0].URL)
	}

	apiSvc, exists := cfg.HTTP.Services[apiName]
	if !exists {
		t.Fatalf("Expected api service %q not found", apiName)
	}
	if apiSvc.LoadBalancer.Servers[0].URL != "http://192.168.0.105:8080" {
		t.Errorf("Unexpected api server URL: %s", apiSvc.LoadBalancer.Servers[0].URL)
	}
}
