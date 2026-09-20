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
		if route.ProjectSlug == "" || route.Host == "" {
			continue
		}

		// Normalize backend URL
		backendURL := route.Backend
		if !strings.HasPrefix(backendURL, "http://") && !strings.HasPrefix(backendURL, "https://") {
			backendURL = "http://" + backendURL
		}

		name := strings.ReplaceAll(route.Host, ".", "-")

		// Traefik dynamic router
		router := RouterConfig{
			Rule:        fmt.Sprintf("Host(`%s`)", route.Host),
			Service:     name,
			EntryPoints: []string{"web", "websecure"},
		}

		var mws []string
		if route.InternalOnly {
			mws = append(mws, "lan-only@file")
		}
		if route.AuthRequired {
			mws = append(mws, "auth@file")
		}
		if len(mws) > 0 {
			router.Middlewares = mws
		}

		routers[name] = router

		// Traefik dynamic service
		services[name] = ServiceConfig{
			LoadBalancer: LoadBalancerConfig{
				Servers: []ServerConfig{
					{URL: backendURL},
				},
			},
		}
	}

	if len(routes) == 0 {
		return TraefikConfigResponse{}
	}

	return TraefikConfigResponse{
		HTTP: &HTTPConfig{
			Routers:  routers,
			Services: services,
		},
	}
}
