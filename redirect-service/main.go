package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"go-project/redirect-service/cache"
	"go-project/redirect-service/handlers"
)

func main() {
	dbURL := getEnv("DB_SERVICE_URL", "http://localhost:8083")
	cacheTTL := 10 * time.Minute

	redirectHandler := handlers.NewRedirectHandler(dbURL, cache.NewStore(cacheTTL))

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "redirect-service"})
	})
	mux.HandleFunc("GET /", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"service": "redirect-service", "message": "use /{code} for redirects"})
	})
	mux.HandleFunc("GET /{code}", redirectHandler.Handle)

	port := getEnv("PORT", "8082")
	log.Printf("redirect-service listening on :%s", port)
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
