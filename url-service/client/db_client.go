package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"go-project/url-service/models"
)

type DBClient struct {
	baseURL string
	http    *http.Client
}

type createReq struct {
	LongURL   string `json:"longUrl"`
	ShortCode string `json:"shortCode"`
	TTLDays   int    `json:"ttlDays"`
}

func NewDBClient(baseURL string) *DBClient {
	return &DBClient{baseURL: baseURL, http: &http.Client{}}
}

func (d *DBClient) LookupByLongURL(longURL string) (*models.DBURLRecord, error) {
	endpoint := fmt.Sprintf("%s/internal/urls/by-long?longUrl=%s", d.baseURL, url.QueryEscape(longURL))
	resp, err := d.http.Get(endpoint)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("lookup failed: status=%d", resp.StatusCode)
	}
	var record models.DBURLRecord
	if err := json.NewDecoder(resp.Body).Decode(&record); err != nil {
		return nil, err
	}
	return &record, nil
}

func (d *DBClient) CreateShortURL(longURL, code string, ttlDays int) (*models.DBURLRecord, int, error) {
	payload, err := json.Marshal(createReq{LongURL: longURL, ShortCode: code, TTLDays: ttlDays})
	if err != nil {
		return nil, 0, err
	}
	resp, err := d.http.Post(d.baseURL+"/internal/urls", "application/json", bytes.NewReader(payload))
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return nil, resp.StatusCode, nil
	}
	var record models.DBURLRecord
	if err := json.NewDecoder(resp.Body).Decode(&record); err != nil {
		return nil, 0, err
	}
	return &record, resp.StatusCode, nil
}
