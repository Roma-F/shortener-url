package models

type ShortenURLResp struct {
	Result string `json:"result"`
}

type ShortenURLReq struct {
	URL string `json:"url"`
}
