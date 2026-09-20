package config

import (
	"os"
)

type Config struct {
	DatabaseURL string
	Port        string
}

func Load() *Config {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://dominator_user:JuguitoDeDominator0228.@192.168.0.214:5432/dominator_db?sslmode=disable"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8090"
	}

	return &Config{
		DatabaseURL: dbURL,
		Port:        port,
	}
}
