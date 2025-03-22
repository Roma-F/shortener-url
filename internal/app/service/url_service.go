package service

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
)

var urlsMAP = make(map[string]string)

func FetchOriginalURL(id string) (string, error) {
	_, ok := urlsMAP[id]
	if ok {
		return urlsMAP[id], nil
	}

	return "", errors.New("this id not found")
}

func GenerateShortURL(originalUrl string, host string) string {
	hash := md5.Sum([]byte(originalUrl))
	id := hex.EncodeToString(hash[:])[:8]
	hashUrl := `http://` + host + "/" + id

	urlsMAP[id] = originalUrl

	return hashUrl
}
