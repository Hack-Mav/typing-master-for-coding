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

	DatabaseURL                    string
	DatabaseMaxOpenConns           int
	DatabaseMaxIdleConns           int
	DatabaseConnMaxLifetimeMinutes int

	RedisURL          string
	RedisPoolSize     int
	RedisMinIdleConns int
}

func Load() *Config {
	cacheMaxSize, _ := strconv.Atoi(getEnv("CACHE_MAX_SIZE", "10000"))
	cacheTTL, _ := strconv.Atoi(getEnv("CACHE_TTL_MINUTES", "30"))

	jwtSecret := getEnv("JWT_SECRET", "")
	if jwtSecret == "" {
		panic("JWT_SECRET environment variable is required and cannot be empty")
	}

	origins := getEnv("ALLOWED_ORIGINS", "")
	if origins == "" {
		panic("ALLOWED_ORIGINS environment variable is required and cannot be empty")
	}
	allowedOrigins := strings.Split(origins, ",")

	dbMaxOpenConns, _ := strconv.Atoi(getEnv("DATABASE_MAX_OPEN_CONNS", "25"))
	dbMaxIdleConns, _ := strconv.Atoi(getEnv("DATABASE_MAX_IDLE_CONNS", "5"))
	dbConnMaxLifetime, _ := strconv.Atoi(getEnv("DATABASE_CONN_MAX_LIFETIME_MINUTES", "30"))

	redisPoolSize, _ := strconv.Atoi(getEnv("REDIS_POOL_SIZE", "10"))
	redisMinIdleConns, _ := strconv.Atoi(getEnv("REDIS_MIN_IDLE_CONNS", "2"))

	return &Config{
		Environment:     getEnv("ENVIRONMENT", "development"),
		ProjectID:       getEnv("GOOGLE_CLOUD_PROJECT", "typing-master-dev"),
		JWTSecret:       jwtSecret,
		Port:            getEnv("PORT", "8080"),
		AllowedOrigins:  allowedOrigins,
		CacheMaxSize:    cacheMaxSize,
		CacheTTLMinutes: cacheTTL,

		DatabaseURL:                    getEnv("DATABASE_URL", ""),
		DatabaseMaxOpenConns:           dbMaxOpenConns,
		DatabaseMaxIdleConns:           dbMaxIdleConns,
		DatabaseConnMaxLifetimeMinutes: dbConnMaxLifetime,

		RedisURL:          getEnv("REDIS_URL", ""),
		RedisPoolSize:     redisPoolSize,
		RedisMinIdleConns: redisMinIdleConns,
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
