package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

func Connect(databaseURL string) (*sql.DB, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(15 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

func InitSchema(ctx context.Context, db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS routes (
		project_slug   TEXT PRIMARY KEY,
		host           TEXT NOT NULL,
		backend        TEXT NOT NULL,
		auth_required  BOOLEAN NOT NULL DEFAULT false,
		updated_at     TIMESTAMPTZ NOT NULL DEFAULT now()
	);
	`
	_, err := db.ExecContext(ctx, schema)
	if err != nil {
		return fmt.Errorf("failed to initialize routes schema: %w", err)
	}
	return nil
}
