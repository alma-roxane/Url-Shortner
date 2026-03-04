package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"go-project/url-service/client"
	"go-project/url-service/handlers"
	"go-project/url-service/service"
)

func main() {
	dbURL := getEnv("DB_SERVICE_URL", "http://localhost:8083")
	publicBase := getEnv("PUBLIC_BASE_URL", "http://localhost:8080")

	shorten := handlers.NewShortenHandler(
		client.NewDBClient(dbURL),
		service.NewGenerator(7),
		publicBase,
	)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "url-service"})
	})
	mux.HandleFunc("POST /api/v1/urls", shorten.Handle)

	port := getEnv("PORT", "8081")
	log.Printf("url-service listening on :%s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
