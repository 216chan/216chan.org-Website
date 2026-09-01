package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port   string
	DBHost string
	DBPort string
	DBUser string
	DBPass string
	DBName string
}

func Load() *Config {
	return &Config{
		Port:   getEnv("PORT", "8080"),
		DBHost: getEnv("POSTGRES_HOST", "localhost"),
		DBPort: getEnv("POSTGRES_PORT", "5432"),
		DBUser: getEnv("POSTGRES_USER", "216chan"),
		DBPass: getEnv("POSTGRES_PASSWORD", "changeme"),
		DBName: getEnv("POSTGRES_DB", "216chan_db"),
	}
}

func (c *Config) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		c.DBHost, c.DBPort, c.DBUser, c.DBPass, c.DBName,
	)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
