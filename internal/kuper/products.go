package kuper

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Product struct {
	Raw map[string]any
}

type productsResp struct {
	Products []map[string]any `json:"products"`
	Items    []map[string]any `json:"items"`
}

func (s *service) ListProducts(ctx context.Context, storeID int, departmentSlug string, page, perPage, offersLimit int) ([]Product, error) {
	url := fmt.Sprintf(
		"%s/api/v3/stores/%d/departments/%s?offers_limit=%d&page=%d&per_page=%d",
		s.baseURL,
		storeID,
		departmentSlug,
		offersLimit,
		page,
		perPage,
	)

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

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"ListProducts: статус=%d url=%s body=%s",
			resp.StatusCode,
			url,
			string(bodyBytes[:min(len(bodyBytes), 4096)]),
		)
	}

	// парсинг структуры json файла с товаром
	var raw map[string]any
	if err := json.Unmarshal(bodyBytes, &raw); err != nil {
		return nil, fmt.Errorf("ListProducts: не удалось распарсить JSON, body=%s", string(bodyBytes[:min(len(bodyBytes), 1024)]))
	}
	if deps, ok := raw["departments"].([]any); ok {
		var all []any
		for _, d := range deps {
			if dep, ok := d.(map[string]any); ok {
				if prods, ok := dep["products"].([]any); ok && len(prods) > 0 {
					all = append(all, prods...)
				}
			}
		}
		if len(all) > 0 {
			return toProducts(all), nil
		}
	}
	// проверки на пустые массивы в полученном json
	if arr, ok := raw["deals"].([]any); ok && len(arr) > 0 {
		return toProducts(arr), nil
	}

	if arr, ok := raw["products"].([]any); ok {
		return toProducts(arr), nil
	}
	if arr, ok := raw["items"].([]any); ok {
		return toProducts(arr), nil
	}

	if data, ok := raw["data"].(map[string]any); ok {
		if arr, ok := data["products"].([]any); ok {
			return toProducts(arr), nil
		}
	}

	if code, ok := raw["code"]; ok {
		msg, _ := raw["message"].(string)
		return nil, fmt.Errorf("ListProducts: api error code=%v message=%s", code, msg)
	}
	return []Product{}, nil

}

func toProducts(arr []any) []Product {
	res := make([]Product, 0, len(arr))
	for _, it := range arr {
		if m, ok := it.(map[string]any); ok {
			res = append(res, Product{Raw: m})
		}
	}
	return res
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
