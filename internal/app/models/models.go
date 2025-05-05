package models

type ShortenURLResp struct {
	Result string `json:"result"`
}

type ShortenURLReq struct {
	URL string `json:"url"`
}

type ShortenBatchItem struct {
	CorrelationId string `json:"correlation_id"`
	OriginalUrl   string `json:"original_url"`
}

type ShortenedURLItem struct {
	CorrelationId string `json:"correlation_id"`
	ShortUrl      string `json:"short_url"`
}

type URLPair struct {
	OriginalURL   string
	ShortURL      string
	CorrelationID string
}
