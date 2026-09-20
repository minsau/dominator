package registrar

// Traefik dynamic HTTP configuration wire models

type ServerConfig struct {
	URL string `json:"url"`
}

type LoadBalancerConfig struct {
	Servers []ServerConfig `json:"servers"`
}

type ServiceConfig struct {
	LoadBalancer LoadBalancerConfig `json:"loadBalancer"`
}

type RouterConfig struct {
	Rule        string   `json:"rule"`
	Service     string   `json:"service"`
	EntryPoints []string `json:"entryPoints"`
	Middlewares []string `json:"middlewares,omitempty"`
}

type HTTPConfig struct {
	Routers  map[string]RouterConfig  `json:"routers,omitempty"`
	Services map[string]ServiceConfig `json:"services,omitempty"`
}

type TraefikConfigResponse struct {
	HTTP *HTTPConfig `json:"http,omitempty"`
}

type HealthResponse struct {
	Status       string `json:"status"`
	RoutesServed int    `json:"routes_served"`
	Error        string `json:"error,omitempty"`
}
