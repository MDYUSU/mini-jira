package config

import (
	"bufio"
	"os"
	"strings"
	"time"
)

type Config struct {
	AppPort     string
	DatabaseURL string
	JWTSecret   string
	TokenTTL    time.Duration
}

func Load() Config {
	loadEnvFile(".env")

	ttl, err := time.ParseDuration(getEnv("TOKEN_TTL", "24h"))
	if err != nil {
		ttl = 24 * time.Hour
	}

	return Config{
		AppPort:     getEnv("APP_PORT", "8080"),
		DatabaseURL: getEnv("DATABASE_URL", ""),
		JWTSecret:   getEnv("JWT_SECRET", ""),
		TokenTTL:    ttl,
	}
}

func loadEnvFile(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	s := bufio.NewScanner(f)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		k = strings.TrimSpace(k)
		if _, exists := os.LookupEnv(k); !exists {
			os.Setenv(k, strings.TrimSpace(v))
		}
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
