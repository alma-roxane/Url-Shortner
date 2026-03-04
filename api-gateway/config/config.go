package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port            string
	URLServiceURL   string
	RedirectSvcURL  string
	APIKey          string
	RateLimitPerMin int
}

func Load() Config {
	return Config{
		Port:            getEnv("PORT", "8080"),
		URLServiceURL:   getEnv("URL_SERVICE_URL", "http://localhost:8081"),
		RedirectSvcURL:  getEnv("REDIRECT_SERVICE_URL", "http://localhost:8082"),
		APIKey:          os.Getenv("API_KEY"),
		RateLimitPerMin: getEnvInt("RATE_LIMIT_PER_MIN", 120),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	var n int
	_, err := fmt.Sscanf(v, "%d", &n)
	if err != nil || n <= 0 {
		return fallback
	}
	return n
}
