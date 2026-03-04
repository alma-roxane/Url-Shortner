package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strings"

	"go-project/db-service/database"
	"go-project/db-service/handlers"
)

func main() {
	store := database.NewSnapshotStore(getEnv("DB_SNAPSHOT_PATH", "db-service/data/urls.json"))
	repo, err := handlers.NewRepository(store, database.NewVisitCache())
	if err != nil {
		log.Fatalf("failed to init repository: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "db-service"})
	})

	mux.HandleFunc("POST /internal/urls", func(w http.ResponseWriter, r *http.Request) {
		var req handlers.CreateURLRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeErr(w, http.StatusBadRequest, "invalid json")
			return
		}
		rec, created, err := repo.Create(req)
		if err != nil {
			switch {
			case errors.Is(err, handlers.ErrInvalidURL), errors.Is(err, handlers.ErrCodeRequired):
				writeErr(w, http.StatusBadRequest, err.Error())
			case errors.Is(err, handlers.ErrCodeExists):
				writeErr(w, http.StatusConflict, err.Error())
			default:
				writeErr(w, http.StatusInternalServerError, "failed to store url")
			}
			return
		}
		status := http.StatusCreated
		if !created {
			status = http.StatusOK
		}
		writeJSON(w, status, rec)
	})

	mux.HandleFunc("GET /internal/urls/by-long", func(w http.ResponseWriter, r *http.Request) {
		longURL := r.URL.Query().Get("longUrl")
		rec, err := repo.GetByLong(longURL)
		if err != nil {
			writeErr(w, http.StatusNotFound, "mapping not found")
			return
		}
		writeJSON(w, http.StatusOK, rec)
	})

	mux.HandleFunc("GET /internal/urls/{code}", func(w http.ResponseWriter, r *http.Request) {
		code := r.PathValue("code")
		rec, err := repo.GetByCode(code)
		if err != nil {
			writeErr(w, http.StatusNotFound, "mapping not found")
			return
		}
		writeJSON(w, http.StatusOK, rec)
	})

	mux.HandleFunc("POST /internal/urls/{code}/visit", func(w http.ResponseWriter, r *http.Request) {
		code := r.PathValue("code")
		rec, err := repo.IncrementVisit(code)
		if err != nil {
			writeErr(w, http.StatusNotFound, "mapping not found")
			return
		}
		writeJSON(w, http.StatusOK, rec)
	})

	mux.HandleFunc("GET /internal/stats/{code}", func(w http.ResponseWriter, r *http.Request) {
		code := strings.TrimSpace(r.PathValue("code"))
		stats, err := repo.Stats(code)
		if err != nil {
			writeErr(w, http.StatusNotFound, "stats not found")
			return
		}
		writeJSON(w, http.StatusOK, stats)
	})

	port := getEnv("PORT", "8083")
	log.Printf("db-service listening on :%s", port)
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

func writeErr(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
