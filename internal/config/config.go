package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
)

type Config struct {
	PostgresUser string
	PostgresPass string
	PostgresDB   string
	PostgresHost string
	PostgresPort string

	RedisHost string
	RedisPort string
	RedisDB   int

	APIHttpPort string

	JWTSecret []byte
	AuthTTL   time.Duration
}

func MustLoad() Config {
	redisDB, err := strconv.Atoi(getEnv("REDIS_DB", "0"))
	if err != nil {
		log.Fatalf("invalid REDIS_DB: %v", err)
	}

	// Преобразуем строку в time.Duration
	authTTL, err := time.ParseDuration(getEnv("AUTH_TTL", "30m"))
	if err != nil {
		log.Fatalf("invalid AUTH_TTL: %v", err)
	}

	return Config{
		PostgresUser: getEnv("POSTGRES_USER", "app"),
		PostgresPass: getEnv("POSTGRES_PASSWORD", "app"),
		PostgresDB:   getEnv("POSTGRES_DB", "microtube"),
		PostgresHost: getEnv("POSTGRES_HOST", "localhost"),
		PostgresPort: getEnv("POSTGRES_PORT", "5432"),

		RedisHost: getEnv("REDIS_HOST", "localhost"),
		RedisPort: getEnv("REDIS_PORT", "6379"),
		RedisDB:   redisDB,

		APIHttpPort: getEnv("API_HTTP_PORT", "8080"),

		JWTSecret: []byte(getEnv("JWT_SECRET", "devsecret")),
		AuthTTL:   authTTL,
	}
}

func (c *Config) PostgresURL() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		c.PostgresUser, c.PostgresPass, c.PostgresHost, c.PostgresPort, c.PostgresDB,
	)
}

func getEnv(key, def string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return def
}
