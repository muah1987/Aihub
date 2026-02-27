package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server     ServerConfig
	Database   DatabaseConfig
	JWT        JWTConfig
	Encryption EncryptionConfig
	CORS       CORSConfig
	Docker     DockerConfig
	SMTP       SMTPConfig
}

type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	BaseURL  string
}

type ServerConfig struct {
	Port        string
	Host        string
	Environment string
}

type DatabaseConfig struct {
	URL string
}

type JWTConfig struct {
	Secret          string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

type EncryptionConfig struct {
	Key string
}

type CORSConfig struct {
	AllowedOrigins string
}

type DockerConfig struct {
	Host        string
	SandboxImage string
	MemoryLimit string
	CPULimit    string
}

func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port:        getEnv("SERVER_PORT", "8080"),
			Host:        getEnv("SERVER_HOST", "0.0.0.0"),
			Environment: getEnv("ENVIRONMENT", "development"),
		},
		Database: DatabaseConfig{
			URL: getEnv("DATABASE_URL", "postgres://aihub:aihub_secret@localhost:5432/aihub?sslmode=disable"),
		},
		JWT: JWTConfig{
			Secret:          getEnv("JWT_SECRET", "dev-secret-change-me"),
			AccessTokenTTL:  parseDuration(getEnv("JWT_ACCESS_TOKEN_TTL", "15m")),
			RefreshTokenTTL: parseDuration(getEnv("JWT_REFRESH_TOKEN_TTL", "168h")),
		},
		Encryption: EncryptionConfig{
			Key: getEnv("ENCRYPTION_KEY", "0123456789abcdef0123456789abcdef"),
		},
		CORS: CORSConfig{
			AllowedOrigins: getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:5173"),
		},
		Docker: DockerConfig{
			Host:        getEnv("DOCKER_HOST", "unix:///var/run/docker.sock"),
			SandboxImage: getEnv("SANDBOX_IMAGE", "aihub-sandbox:latest"),
			MemoryLimit: getEnv("SANDBOX_MEMORY_LIMIT", "256m"),
			CPULimit:    getEnv("SANDBOX_CPU_LIMIT", "0.5"),
		},
		SMTP: SMTPConfig{
			Host:     getEnv("SMTP_HOST", ""),
			Port:     parseIntEnv("SMTP_PORT", 587),
			Username: getEnv("SMTP_USERNAME", ""),
			Password: getEnv("SMTP_PASSWORD", ""),
			From:     getEnv("SMTP_FROM", "noreply@aihub.dev"),
			BaseURL:  getEnv("APP_BASE_URL", "http://localhost:5173"),
		},
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func parseIntEnv(key string, fallback int) int {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	i, err := strconv.Atoi(val)
	if err != nil {
		return fallback
	}
	return i
}

func parseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		return 15 * time.Minute
	}
	return d
}
