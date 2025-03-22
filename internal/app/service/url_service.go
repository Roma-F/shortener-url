package service

import (
	"fmt"
)

func FetchOriginalURL(id string) (string, error) {
	fmt.Println(id, "ID")
	return "https://practicum.yandex.ru/", nil
}

func GenerateShortURL(url string) string {
	return url + " Test"
}
