package kuper

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Category struct {
	ID            int    `json:"id"`
	ParentID      int    `json:"parent_id"`
	Type          string `json:"type"`
	Name          string `json:"name"`
	Slug          string `json:"slug"`
	ProductsCount int    `json:"products_count"`
	CategoryType  string `json:"category_type"`
	HasChildren   bool   `json:"has_children"`
}

type categoriesResp struct {
	Categories []Category `json:"categories"`
}

// ListCategories возвращает список доступных категорий по id магазина
func (s *service) ListCategories(ctx context.Context, storeID int) ([]Category, error) {
	url := fmt.Sprintf("%s/api/v3/stores/%d/categories", s.baseURL, storeID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	s.applyDefaultHeaders(req)

	resp, err := s.transport.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// доп проверка для отладки если вернуло не 200 резульатат
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("ListCategories: статус=%d body=%s", resp.StatusCode, string(b))
	}

	var out categoriesResp
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}

	return out.Categories, nil
}
