package service

func FetchOriginalURL(id string) (string, error) {
	return "https://practicum.yandex.ru/", nil
}

func GenerateShortURL(url string) string {
	return url + " Test"
}
