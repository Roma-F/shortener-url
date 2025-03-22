package service

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
)

var urlsMAP = make(map[string]string)

func FetchOriginalURL(id string) (string, error) {
	_, ok := urlsMAP[id]
	if ok {
		return urlsMAP[id], nil
	}

	return "", errors.New("this id not found")
}

func GenerateShortURL(originalURL string, host string) string {
	hash := md5.Sum([]byte(originalURL))
	id := hex.EncodeToString(hash[:])[:8]
	hashURL := fmt.Sprintf("http://%s/%s", host, id)

	urlsMAP[id] = originalURL

	return hashURL
}
