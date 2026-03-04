package routes

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"time"

	"go-project/api-gateway/config"
	"go-project/api-gateway/middleware"
)

func New(cfg config.Config) http.Handler {
	client := &http.Client{
		Timeout: 5 * time.Second,
		// Keep redirect semantics from upstream services; do not follow automatically.
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	mux := http.NewServeMux()
	webAssets := http.FileServer(http.Dir("./web"))

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok","service":"api-gateway"}`))
	})

	writeHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "invalid body", http.StatusBadRequest)
			return
		}

		req, err := http.NewRequest(http.MethodPost, strings.TrimRight(cfg.URLServiceURL, "/")+"/api/v1/urls", bytes.NewReader(body))
		if err != nil {
			http.Error(w, "upstream request error", http.StatusInternalServerError)
			return
		}
		req.Header = r.Header.Clone()
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			http.Error(w, "url-service unavailable", http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		copyResponse(w, resp)
	})
	mux.Handle("POST /api/v1/urls", middleware.APIKey(cfg.APIKey, writeHandler))

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/app", http.StatusTemporaryRedirect)
	})

	mux.Handle("GET /app/", http.StripPrefix("/app/", webAssets))
	mux.HandleFunc("GET /app", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/app/", http.StatusTemporaryRedirect)
	})

	mux.HandleFunc("GET /{code}", func(w http.ResponseWriter, r *http.Request) {
		code := r.PathValue("code")
		if strings.TrimSpace(code) == "" {
			http.NotFound(w, r)
			return
		}
		endpoint := strings.TrimRight(cfg.RedirectSvcURL, "/") + "/" + code
		req, err := http.NewRequest(http.MethodGet, endpoint, nil)
		if err != nil {
			http.Error(w, "upstream request error", http.StatusInternalServerError)
			return
		}
		req.Header = r.Header.Clone()
		resp, err := client.Do(req)
		if err != nil {
			http.Error(w, "redirect-service unavailable", http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()
		copyResponse(w, resp)
	})

	rl := middleware.NewRateLimiter(cfg.RateLimitPerMin)
	return rl.Middleware(mux)
}

func copyResponse(w http.ResponseWriter, resp *http.Response) {
	for k, values := range resp.Header {
		for _, v := range values {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}
