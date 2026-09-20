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
		project_slug   TEXT NOT NULL,
		host           TEXT NOT NULL,
		backend        TEXT NOT NULL,
		auth_required  BOOLEAN NOT NULL DEFAULT false,
		internal_only  BOOLEAN NOT NULL DEFAULT false,
		updated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
		PRIMARY KEY (project_slug, host)
	);

	ALTER TABLE routes ADD COLUMN IF NOT EXISTS internal_only BOOLEAN NOT NULL DEFAULT false;

	DO $$
	DECLARE
		pk_cols integer;
		c_name text;
	BEGIN
		SELECT tc.constraint_name, count(*) INTO c_name, pk_cols
		FROM information_schema.key_column_usage kcu
		JOIN information_schema.table_constraints tc
		  ON kcu.constraint_name = tc.constraint_name
		 AND kcu.table_schema = tc.table_schema
		WHERE tc.constraint_type = 'PRIMARY KEY'
		  AND tc.table_name = 'routes'
		  AND tc.table_schema = current_schema()
		GROUP BY tc.constraint_name;

		IF pk_cols = 1 THEN
			EXECUTE format('ALTER TABLE routes DROP CONSTRAINT %I', c_name);
			ALTER TABLE routes ADD PRIMARY KEY (project_slug, host);
		END IF;
	END $$;
	`
	_, err := db.ExecContext(ctx, schema)
	if err != nil {
		return fmt.Errorf("failed to initialize routes schema: %w", err)
	}
	return nil
}
