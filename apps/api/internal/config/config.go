package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Environment     string
	ProjectID       string
	JWTSecret       string
	Port            string
	AllowedOrigins  []string
	CacheMaxSize    int
	CacheTTLMinutes int
}

func Load() *Config {
	cacheMaxSize, _ := strconv.Atoi(getEnv("CACHE_MAX_SIZE", "10000"))
	cacheTTL, _ := strconv.Atoi(getEnv("CACHE_TTL_MINUTES", "30"))

	origins := getEnv("ALLOWED_ORIGINS", "http://localhost:3000")
	allowedOrigins := strings.Split(origins, ",")

	return &Config{
		Environment:     getEnv("ENVIRONMENT", "development"),
		ProjectID:       getEnv("GOOGLE_CLOUD_PROJECT", "typing-master-dev"),
		JWTSecret:       getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
		Port:            getEnv("PORT", "8080"),
		AllowedOrigins:  allowedOrigins,
		CacheMaxSize:    cacheMaxSize,
		CacheTTLMinutes: cacheTTL,
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
