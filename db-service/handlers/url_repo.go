package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"go-project/db-service/database"
)

var (
	ErrInvalidURL   = errors.New("invalid longUrl")
	ErrCodeRequired = errors.New("customCode cannot be empty")
	ErrCodeExists   = errors.New("short code already exists")
	ErrNotFound     = errors.New("short code not found")
)

type URLRecord struct {
	Code      string     `json:"code"`
	LongURL   string     `json:"longUrl"`
	CreatedAt time.Time  `json:"createdAt"`
	ExpiresAt *time.Time `json:"expiresAt,omitempty"`
	Visits    int64      `json:"visits"`
}

type CreateURLRequest struct {
	LongURL   string `json:"longUrl"`
	ShortCode string `json:"shortCode"`
	TTLDays   int    `json:"ttlDays"`
}

type Stats struct {
	Code         string `json:"code"`
	TotalVisits  int64  `json:"totalVisits"`
	RecentVisits int64  `json:"recentVisits"`
}

type snapshot struct {
	Records []URLRecord `json:"records"`
}

type Repository struct {
	mu     sync.RWMutex
	byCode map[string]URLRecord
	byLong map[string]string
	store  *database.SnapshotStore
	visits *database.VisitCache
}

func NewRepository(store *database.SnapshotStore, visits *database.VisitCache) (*Repository, error) {
	r := &Repository{
		byCode: make(map[string]URLRecord),
		byLong: make(map[string]string),
		store:  store,
		visits: visits,
	}
	if err := r.load(); err != nil {
		return nil, err
	}
	return r, nil
}

func (r *Repository) Create(req CreateURLRequest) (URLRecord, bool, error) {
	if err := validateLongURL(req.LongURL); err != nil {
		return URLRecord{}, false, err
	}
	if strings.TrimSpace(req.ShortCode) == "" {
		return URLRecord{}, false, ErrCodeRequired
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if existingCode, ok := r.byLong[req.LongURL]; ok {
		rec := r.byCode[existingCode]
		return rec, false, nil
	}
	if _, ok := r.byCode[req.ShortCode]; ok {
		return URLRecord{}, false, ErrCodeExists
	}

	now := time.Now().UTC()
	var expiresAt *time.Time
	if req.TTLDays > 0 {
		exp := now.Add(time.Duration(req.TTLDays) * 24 * time.Hour)
		expiresAt = &exp
	}

	rec := URLRecord{
		Code:      req.ShortCode,
		LongURL:   req.LongURL,
		CreatedAt: now,
		ExpiresAt: expiresAt,
		Visits:    0,
	}
	r.byCode[rec.Code] = rec
	r.byLong[rec.LongURL] = rec.Code

	if err := r.persistLocked(); err != nil {
		return URLRecord{}, false, err
	}
	return rec, true, nil
}

func (r *Repository) GetByCode(code string) (URLRecord, error) {
	r.mu.RLock()
	rec, ok := r.byCode[code]
	r.mu.RUnlock()
	if !ok {
		return URLRecord{}, ErrNotFound
	}
	if rec.ExpiresAt != nil && rec.ExpiresAt.Before(time.Now().UTC()) {
		return URLRecord{}, ErrNotFound
	}
	return rec, nil
}

func (r *Repository) GetByLong(longURL string) (URLRecord, error) {
	r.mu.RLock()
	code, ok := r.byLong[longURL]
	if !ok {
		r.mu.RUnlock()
		return URLRecord{}, ErrNotFound
	}
	rec := r.byCode[code]
	r.mu.RUnlock()
	return rec, nil
}

func (r *Repository) IncrementVisit(code string) (URLRecord, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	rec, ok := r.byCode[code]
	if !ok {
		return URLRecord{}, ErrNotFound
	}
	rec.Visits++
	r.byCode[code] = rec
	r.visits.Increment(code)

	if err := r.persistLocked(); err != nil {
		return URLRecord{}, err
	}
	return rec, nil
}

func (r *Repository) Stats(code string) (Stats, error) {
	r.mu.RLock()
	rec, ok := r.byCode[code]
	r.mu.RUnlock()
	if !ok {
		return Stats{}, ErrNotFound
	}
	return Stats{
		Code:         code,
		TotalVisits:  rec.Visits,
		RecentVisits: r.visits.Get(code),
	}, nil
}

func (r *Repository) load() error {
	data, err := r.store.Load()
	if err != nil || data == nil {
		return err
	}
	var s snapshot
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("decode snapshot: %w", err)
	}
	for _, rec := range s.Records {
		r.byCode[rec.Code] = rec
		r.byLong[rec.LongURL] = rec.Code
	}
	return nil
}

func (r *Repository) persistLocked() error {
	records := make([]URLRecord, 0, len(r.byCode))
	for _, rec := range r.byCode {
		records = append(records, rec)
	}
	payload, err := json.MarshalIndent(snapshot{Records: records}, "", "  ")
	if err != nil {
		return fmt.Errorf("encode snapshot: %w", err)
	}
	if err := r.store.Save(payload); err != nil {
		return fmt.Errorf("save snapshot: %w", err)
	}
	return nil
}

func validateLongURL(raw string) error {
	u, err := url.ParseRequestURI(strings.TrimSpace(raw))
	if err != nil || u.Scheme == "" || u.Host == "" {
		return ErrInvalidURL
	}
	return nil
}
