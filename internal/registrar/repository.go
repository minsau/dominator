package registrar

import (
	"context"
	"database/sql"
	"fmt"
)

type Repository interface {
	GetRoutes(ctx context.Context) ([]Route, error)
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
		SELECT project_slug, host, backend, auth_required, updated_at
		FROM routes
		ORDER BY project_slug ASC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query routes: %w", err)
	}
	defer rows.Close()

	var routes []Route
	for rows.Next() {
		var route Route
		if err := rows.Scan(
			&route.ProjectSlug,
			&route.Host,
			&route.Backend,
			&route.AuthRequired,
			&route.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan route: %w", err)
		}
		routes = append(routes, route)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error reading route rows: %w", err)
	}

	return routes, nil
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
