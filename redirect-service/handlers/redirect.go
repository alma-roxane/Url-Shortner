package handlers

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"go-project/redirect-service/cache"
)

type dbResponse struct {
	Code      string  `json:"code"`
	LongURL   string  `json:"longUrl"`
	ExpiresAt *string `json:"expiresAt,omitempty"`
}

type RedirectHandler struct {
	dbURL string
	http  *http.Client
	cache *cache.Store
}

func NewRedirectHandler(dbURL string, c *cache.Store) *RedirectHandler {
	return &RedirectHandler{
		dbURL: strings.TrimRight(dbURL, "/"),
		http:  &http.Client{Timeout: 4 * time.Second},
		cache: c,
	}
}

func (h *RedirectHandler) Handle(w http.ResponseWriter, r *http.Request) {
	code := strings.TrimSpace(strings.TrimPrefix(r.URL.Path, "/"))
	if code == "" {
		http.NotFound(w, r)
		return
	}

	if longURL, _, ok := h.cache.Get(code); ok {
		go h.trackVisit(code)
		http.Redirect(w, r, longURL, http.StatusTemporaryRedirect)
		return
	}

	lookupResp, err := h.http.Get(h.dbURL + "/internal/urls/" + code)
	if err != nil {
		http.Error(w, "lookup failed", http.StatusBadGateway)
		return
	}
	defer lookupResp.Body.Close()
	if lookupResp.StatusCode == http.StatusNotFound {
		http.NotFound(w, r)
		return
	}
	if lookupResp.StatusCode != http.StatusOK {
		http.Error(w, "lookup failed", http.StatusBadGateway)
		return
	}

	var rec dbResponse
	if err := json.NewDecoder(lookupResp.Body).Decode(&rec); err != nil {
		http.Error(w, "bad response", http.StatusBadGateway)
		return
	}

	var exp *time.Time
	if rec.ExpiresAt != nil {
		parsed, err := time.Parse(time.RFC3339Nano, *rec.ExpiresAt)
		if err == nil {
			exp = &parsed
		}
	}
	h.cache.Set(code, rec.LongURL, exp)
	go h.trackVisit(code)

	http.Redirect(w, r, rec.LongURL, http.StatusTemporaryRedirect)
}

func (h *RedirectHandler) trackVisit(code string) {
	payload := bytes.NewBufferString(`{}`)
	resp, err := h.http.Post(h.dbURL+"/internal/urls/"+code+"/visit", "application/json", payload)
	if err != nil {
		log.Printf("visit tracking failed for %s: %v", code, err)
		return
	}
	_ = resp.Body.Close()
}
