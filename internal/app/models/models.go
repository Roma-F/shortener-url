package models

type ShortenURLResp struct {
	Result string `json:"result"`
}

type ShortenURLReq struct {
	Url string `json:"url"`
}
