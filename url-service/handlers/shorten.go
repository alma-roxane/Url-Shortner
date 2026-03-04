package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"go-project/url-service/client"
	"go-project/url-service/models"
	"go-project/url-service/service"
)

type ShortenHandler struct {
	db         *client.DBClient
	generator  *service.Generator
	publicBase string
}

func NewShortenHandler(db *client.DBClient, generator *service.Generator, publicBase string) *ShortenHandler {
	return &ShortenHandler{db: db, generator: generator, publicBase: strings.TrimRight(publicBase, "/")}
}

func (h *ShortenHandler) Handle(w http.ResponseWriter, r *http.Request) {
	var req models.ShortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	if err := validateRequest(req); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}

	if existing, err := h.db.LookupByLongURL(req.LongURL); err == nil && existing != nil {
		writeJSON(w, http.StatusOK, toResponse(*existing, h.publicBase))
		return
	}

	for i := 0; i < 5; i++ {
		code := strings.TrimSpace(req.CustomCode)
		if code == "" {
			generated, err := h.generator.Generate()
			if err != nil {
				writeErr(w, http.StatusInternalServerError, "could not generate short code")
				return
			}
			code = generated
		}

		record, status, err := h.db.CreateShortURL(req.LongURL, code, req.TTLDays)
		if err != nil {
			log.Printf("db create error: %v", err)
			writeErr(w, http.StatusBadGateway, "database service unavailable")
			return
		}
		if record != nil {
			finalStatus := http.StatusCreated
			if status == http.StatusOK {
				finalStatus = http.StatusOK
			}
			writeJSON(w, finalStatus, toResponse(*record, h.publicBase))
			return
		}
		if status == http.StatusConflict && req.CustomCode == "" {
			continue
		}
		if status == http.StatusConflict {
			writeErr(w, http.StatusConflict, "custom code already exists")
			return
		}
		writeErr(w, http.StatusBadRequest, fmt.Sprintf("unable to shorten url (status=%d)", status))
		return
	}

	writeErr(w, http.StatusConflict, "failed to allocate unique short code")
}

func validateRequest(req models.ShortenRequest) error {
	if strings.TrimSpace(req.LongURL) == "" {
		return errors.New("longUrl is required")
	}
	u, err := url.ParseRequestURI(req.LongURL)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return errors.New("longUrl is invalid")
	}
	if req.TTLDays < 0 || req.TTLDays > 3650 {
		return errors.New("ttlDays must be between 0 and 3650")
	}
	if strings.ContainsAny(req.CustomCode, " /?#") {
		return errors.New("customCode contains invalid characters")
	}
	return nil
}

func toResponse(record models.DBURLRecord, publicBase string) models.ShortenResponse {
	resp := models.ShortenResponse{
		Code:      record.Code,
		ShortURL:  publicBase + "/" + record.Code,
		LongURL:   record.LongURL,
		CreatedAt: record.CreatedAt,
	}
	if record.ExpiresAt != nil {
		resp.ExpiresAt = *record.ExpiresAt
	}
	return resp
}

func writeErr(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
