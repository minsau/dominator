package registrar

import (
	"context"
	"fmt"
)

type Controller struct {
	repo Repository
}

func NewController(repo Repository) *Controller {
	return &Controller{repo: repo}
}

func (c *Controller) GetTraefikConfig(ctx context.Context) (TraefikConfigResponse, error) {
	routes, err := c.repo.GetRoutes(ctx)
	if err != nil {
		return TraefikConfigResponse{}, fmt.Errorf("failed to get routes: %w", err)
	}

	return BuildTraefikConfig(routes), nil
}

func (c *Controller) GetHealth(ctx context.Context) (HealthResponse, error) {
	if err := c.repo.Ping(ctx); err != nil {
		return HealthResponse{
			Status: "error",
			Error:  err.Error(),
		}, err
	}

	count, err := c.repo.CountRoutes(ctx)
	if err != nil {
		return HealthResponse{
			Status: "error",
			Error:  err.Error(),
		}, err
	}

	return HealthResponse{
		Status:       "ok",
		RoutesServed: count,
	}, nil
}

func (c *Controller) UpsertRoute(ctx context.Context, route Route) (Route, error) {
	if route.ProjectSlug == "" {
		route.ProjectSlug = route.Slug
	}
	if route.ProjectSlug == "" {
		return Route{}, fmt.Errorf("project_slug is required")
	}
	if route.Host == "" {
		return Route{}, fmt.Errorf("host is required")
	}
	if route.Backend == "" {
		return Route{}, fmt.Errorf("backend is required")
	}
	return c.repo.UpsertRoute(ctx, route)
}

func (c *Controller) DeleteRoutesBySlug(ctx context.Context, slug string) error {
	if slug == "" {
		return fmt.Errorf("slug is required")
	}
	return c.repo.DeleteRoutesBySlug(ctx, slug)
}

func (c *Controller) GetRoutesBySlug(ctx context.Context, slug string) ([]Route, error) {
	if slug == "" {
		return nil, fmt.Errorf("slug is required")
	}
	return c.repo.GetRoutesBySlug(ctx, slug)
}

func (c *Controller) ListRoutes(ctx context.Context) ([]Route, error) {
	return c.repo.GetRoutes(ctx)
}

