package db

import (
	"fmt"
	"net/url"

	"yxbrew/statuspage/internal/utils"
)

// Config stores PostgreSQL connection settings.
type Config struct {
	DatabaseURL string
}

// LoadConfig resolves database configuration from env vars.
func LoadConfig() Config {
	databaseURL := utils.GetEnvOrDefault("DATABASE_URL", "")
	if databaseURL != "" {
		return Config{DatabaseURL: databaseURL}
	}

	host := utils.GetEnvOrDefault("DB_HOST", "localhost")
	port := utils.GetEnvOrDefault("DB_PORT", "5432")
	user := utils.GetEnvOrDefault("DB_USER", "postgres")
	password := utils.GetEnvOrDefault("DB_PASSWORD", "postgres")
	name := utils.GetEnvOrDefault("DB_NAME", "statuspage")
	sslMode := utils.GetEnvOrDefault("DB_SSLMODE", "disable")

	return Config{
		DatabaseURL: fmt.Sprintf(
			"postgres://%s:%s@%s:%s/%s?sslmode=%s",
			url.QueryEscape(user),
			url.QueryEscape(password),
			host,
			port,
			url.QueryEscape(name),
			url.QueryEscape(sslMode),
		),
	}
}
