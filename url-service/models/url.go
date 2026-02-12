package models

type URLRequest struct {
	LongURL string `json:"long_url"`

}

type URLResponse struct {
	ShortCode string `json:"short_code"`
	ShortURL string `json:"short_url"`
}