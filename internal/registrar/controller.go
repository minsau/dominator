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
