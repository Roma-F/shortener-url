package models

type ShortenURLResp struct {
	Result string `json:"result"`
}

type ShortenURLReq struct {
	URL string `json:"url"`
}

type ShortenBatchItem struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type ShortenedURLItem struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

type URLPair struct {
	OriginalURL   string
	ShortURL      string
	CorrelationID string
	UserID        string
}

type UserURL struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}
