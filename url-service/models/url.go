package models

type ShortenRequest struct {
	LongURL    string `json:"longUrl"`
	CustomCode string `json:"customCode,omitempty"`
	TTLDays    int    `json:"ttlDays,omitempty"`
}

type ShortenResponse struct {
	Code      string `json:"code"`
	ShortURL  string `json:"shortUrl"`
	LongURL   string `json:"longUrl"`
	CreatedAt string `json:"createdAt"`
	ExpiresAt string `json:"expiresAt,omitempty"`
}

type DBURLRecord struct {
	Code      string  `json:"code"`
	LongURL   string  `json:"longUrl"`
	CreatedAt string  `json:"createdAt"`
	ExpiresAt *string `json:"expiresAt,omitempty"`
	Visits    int64   `json:"visits"`
}
