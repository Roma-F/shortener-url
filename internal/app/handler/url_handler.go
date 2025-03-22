package handler

import (
	"io"
	"net/http"
	"strconv"

	"github.com/Roma-F/shortener-url/internal/app/service"
)

func GetMainUrl(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	urlId := r.PathValue("id")

	mainUrl, err := service.FetchOriginalURL(urlId)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	http.Redirect(w, r, mainUrl, http.StatusTemporaryRedirect)
}

func ShortenUrl(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if r.Header.Get("Content-Type") != "text/plain" {
		http.Error(w, "Content-Type most be text/plain", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	url := string(body)
	var shortUrl string = service.GenerateShortURL(url)

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Content-Length", strconv.Itoa(len(shortUrl)))
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(url))
}
