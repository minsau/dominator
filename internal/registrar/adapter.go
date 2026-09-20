package registrar

import (
	"fmt"
	"strings"
)

// BuildTraefikConfig transforms a slice of Route domain models into Traefik dynamic configuration.
// It is a pure function with no side-effects.
func BuildTraefikConfig(routes []Route) TraefikConfigResponse {
	routers := make(map[string]RouterConfig)
	services := make(map[string]ServiceConfig)

	for _, route := range routes {
		if route.ProjectSlug == "" {
			continue
		}

		// Normalize backend URL
		backendURL := route.Backend
		if !strings.HasPrefix(backendURL, "http://") && !strings.HasPrefix(backendURL, "https://") {
			backendURL = "http://" + backendURL
		}

		// Traefik dynamic router
		router := RouterConfig{
			Rule:        fmt.Sprintf("Host(`%s`)", route.Host),
			Service:     route.ProjectSlug,
			EntryPoints: []string{"web"},
		}

		if route.AuthRequired {
			router.Middlewares = []string{"auth@file"}
		}

		routers[route.ProjectSlug] = router

		// Traefik dynamic service
		services[route.ProjectSlug] = ServiceConfig{
			LoadBalancer: LoadBalancerConfig{
				Servers: []ServerConfig{
					{URL: backendURL},
				},
			},
		}
	}

	return TraefikConfigResponse{
		HTTP: HTTPConfig{
			Routers:  routers,
			Services: services,
		},
	}
}
