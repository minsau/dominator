package config

import (
	"errors"
	"os"
)

// ErrMissingDatabaseURL is returned when DATABASE_URL is not set. There is no
// default: the DSN carries a credential and must come from the environment
// (Ansible injects it from the vault), never from source.
var ErrMissingDatabaseURL = errors.New("DATABASE_URL is required")

type Config struct {
	DatabaseURL string
	Port        string
}

func Load() (*Config, error) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, ErrMissingDatabaseURL
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8090"
	}

	return &Config{
		DatabaseURL: dbURL,
		Port:        port,
	}, nil
}
