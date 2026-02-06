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
	//req.AddCookie(&http.Cookie{Name: "spid", Value: "1770393537224_be3d811503b542395747123a0c0cd01c_2c05mwtp59om0avi"})
	//req.AddCookie(&http.Cookie{Name: "spsc", Value: "1770393840000_9db4fb8bf8bee3313518af9f7611d73b_cYnyno5ttEieJDY1kL15kCpcFaEC04iMYaBsLMRQDhwZ"})

}
