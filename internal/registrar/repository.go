package registrar

import (
	"context"
	"database/sql"
	"fmt"
)

type Repository interface {
	GetRoutes(ctx context.Context) ([]Route, error)
	GetRoutesBySlug(ctx context.Context, slug string) ([]Route, error)
	UpsertRoute(ctx context.Context, route Route) (Route, error)
	DeleteRoutesBySlug(ctx context.Context, slug string) error
	CountRoutes(ctx context.Context) (int, error)
	Ping(ctx context.Context) error
}

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) GetRoutes(ctx context.Context) ([]Route, error) {
	query := `
		SELECT project_slug, host, backend, auth_required, internal_only, updated_at
		FROM routes
		ORDER BY project_slug ASC, host ASC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query routes: %w", err)
	}
	defer rows.Close()

	routes := make([]Route, 0)
	for rows.Next() {
		var route Route
		if err := rows.Scan(
			&route.ProjectSlug,
			&route.Host,
			&route.Backend,
			&route.AuthRequired,
			&route.InternalOnly,
			&route.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan route: %w", err)
		}
		route.Slug = route.ProjectSlug
		routes = append(routes, route)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error reading route rows: %w", err)
	}

	return routes, nil
}

func (r *PostgresRepository) GetRoutesBySlug(ctx context.Context, slug string) ([]Route, error) {
	query := `
		SELECT project_slug, host, backend, auth_required, internal_only, updated_at
		FROM routes
		WHERE project_slug = $1
		ORDER BY host ASC
	`
	rows, err := r.db.QueryContext(ctx, query, slug)
	if err != nil {
		return nil, fmt.Errorf("failed to query routes by slug: %w", err)
	}
	defer rows.Close()

	routes := make([]Route, 0)
	for rows.Next() {
		var route Route
		if err := rows.Scan(
			&route.ProjectSlug,
			&route.Host,
			&route.Backend,
			&route.AuthRequired,
			&route.InternalOnly,
			&route.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan route: %w", err)
		}
		route.Slug = route.ProjectSlug
		routes = append(routes, route)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error reading route rows: %w", err)
	}

	return routes, nil
}

func (r *PostgresRepository) UpsertRoute(ctx context.Context, route Route) (Route, error) {
	query := `
		INSERT INTO routes (project_slug, host, backend, auth_required, internal_only, updated_at)
		VALUES ($1, $2, $3, $4, $5, now())
		ON CONFLICT (project_slug, host) DO UPDATE
		SET backend = EXCLUDED.backend,
		    auth_required = EXCLUDED.auth_required,
		    internal_only = EXCLUDED.internal_only,
		    updated_at = now()
		RETURNING project_slug, host, backend, auth_required, internal_only, updated_at
	`
	var saved Route
	err := r.db.QueryRowContext(ctx, query,
		route.ProjectSlug,
		route.Host,
		route.Backend,
		route.AuthRequired,
		route.InternalOnly,
	).Scan(
		&saved.ProjectSlug,
		&saved.Host,
		&saved.Backend,
		&saved.AuthRequired,
		&saved.InternalOnly,
		&saved.UpdatedAt,
	)
	if err != nil {
		return Route{}, fmt.Errorf("failed to upsert route: %w", err)
	}
	saved.Slug = saved.ProjectSlug
	return saved, nil
}

func (r *PostgresRepository) DeleteRoutesBySlug(ctx context.Context, slug string) error {
	query := `DELETE FROM routes WHERE project_slug = $1`
	_, err := r.db.ExecContext(ctx, query, slug)
	if err != nil {
		return fmt.Errorf("failed to delete routes by slug: %w", err)
	}
	return nil
}

func (r *PostgresRepository) CountRoutes(ctx context.Context) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM routes`
	if err := r.db.QueryRowContext(ctx, query).Scan(&count); err != nil {
		return 0, fmt.Errorf("failed to count routes: %w", err)
	}
	return count, nil
}

func (r *PostgresRepository) Ping(ctx context.Context) error {
	if err := r.db.PingContext(ctx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}
	return nil
}
