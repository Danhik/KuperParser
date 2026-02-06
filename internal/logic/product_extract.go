package logic

import (
	"fmt"
	"strings"

	"kuperparser/internal/kuper"
)

// extractName пытается достать имя товара из разных вариантов поля
func extractName(p kuper.Product) string {
	if v, ok := asString(p.Raw["name"]); ok {
		return v
	}
	if v, ok := asString(p.Raw["title"]); ok {
		return v
	}
	return ""
}

// extractURL возвращает ссылку на страницу товара
func extractURL(baseURL string, p kuper.Product) string {
	if v, ok := asString(p.Raw["canonical_url"]); ok && strings.HasPrefix(v, "http") {
		return v
	}
	if v, ok := asString(p.Raw["url"]); ok && strings.HasPrefix(v, "http") {
		return v
	}

	if v, ok := asString(p.Raw["permalink"]); ok {
		if strings.HasPrefix(v, "http") {
			return v
		}
		if strings.HasPrefix(v, "/") {
			return baseURL + v
		}
		return baseURL + "/" + v
	}

	return ""
}

// extractPrice возвращает цену как строку, если вдруг в каком-то магазине цена будет не в целочисленном формате
func extractPrice(p kuper.Product) string {
	if v, ok := asString(p.Raw["price"]); ok {
		return v
	}
	if v, ok := asNumberString(p.Raw["price"]); ok {
		return v
	}

	if arr, ok := p.Raw["offers"].([]any); ok && len(arr) > 0 {
		if m, ok := arr[0].(map[string]any); ok {
			if v, ok := asNumberString(m["price"]); ok {
				return v
			}
			if pm, ok := m["price"].(map[string]any); ok {
				if v, ok := asNumberString(pm["amount"]); ok {
					return v
				}
				if v, ok := asNumberString(pm["value"]); ok {
					return v
				}
			}
		}
	}

	if v, ok := asNumberString(p.Raw["current_price"]); ok {
		return v
	}
	if v, ok := asNumberString(p.Raw["price_current"]); ok {
		return v
	}

	return ""
}

func asString(v any) (string, bool) {
	s, ok := v.(string)
	if !ok {
		return "", false
	}
	return s, true
}

func asNumberString(v any) (string, bool) {
	switch t := v.(type) {
	case float64:
		if t == float64(int64(t)) {
			return fmt.Sprintf("%d", int64(t)), true
		}
		return fmt.Sprintf("%v", t), true
	case int:
		return fmt.Sprintf("%d", t), true
	case int64:
		return fmt.Sprintf("%d", t), true
	case string:
		if t != "" {
			return t, true
		}
	}
	return "", false
}
