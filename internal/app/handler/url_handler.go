package handler

import (
	"io"
	"net/http"
	"strconv"

	"github.com/Roma-F/shortener-url/internal/app/service"
)

func GetMainURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	urlID := r.PathValue("id")

	mainURL, err := service.FetchOriginalURL(urlID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	http.Redirect(w, r, mainURL, http.StatusTemporaryRedirect)
}

func ShortenURL(w http.ResponseWriter, r *http.Request) {
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
	shortURL := service.GenerateShortURL(url)

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Content-Length", strconv.Itoa(len(shortURL)))
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(url))
}
