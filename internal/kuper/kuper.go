package kuper

import (
	"context"
	"net/http"

	"kuperparser/internal/client"
)

type KuperService interface {
	ListCategories(ctx context.Context, storeID int) ([]Category, error)

	GetStore(ctx context.Context, storeID int) (StoreInfo, error)

	ListProducts(ctx context.Context, storeID int, departmentSlug string, page, perPage, offersLimit int) ([]Product, error)
}

type service struct {
	transport client.Transport
	baseURL   string
}

func NewKuperService(transport client.Transport) KuperService {
	return &service{
		transport: transport,
		baseURL:   "https://kuper.ru",
	}
}

func (s *service) applyDefaultHeaders(req *http.Request) {
	req.Header.Set(
		"User-Agent",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) "+
			"AppleWebKit/537.36 (KHTML, like Gecko) "+
			"Chrome/144.0.0.0 Safari/537.36",
	)

	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Accept-Language", "ru-RU,ru;q=0.9,en;q=0.8")
	req.Header.Set("Referer", "https://kuper.ru/")
	req.Header.Set("Origin", "https://kuper.ru")

	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Dest", "empty")

}
