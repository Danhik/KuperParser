package logic

import (
	"fmt"
	"regexp"
	"strings"

	"kuperparser/internal/kuper"
)

// ResolveResult результат сопоставления категорий из конфига с категориями магазина
type ResolveResult struct {
	Slugs []string // список уникальных названий категорий для магазина сопоставленных с указанными в конфиге

	NotFoundNames []string
}

// ResolveCategorySlugsByNames сопоставляет список имён категорий из конфига со списком категорий магазина
func ResolveCategorySlugsByNames(requestedNames []string, storeCategories []kuper.Category) ResolveResult {
	index := make(map[string]kuper.Category, len(storeCategories))
	for _, c := range storeCategories {
		n := normalizeCategoryName(c.Name)
		if _, ok := index[n]; !ok {
			index[n] = c
		}
	}

	var res ResolveResult
	res.Slugs = make([]string, 0, len(requestedNames))

	seenSlug := make(map[string]struct{})

	for _, rawName := range requestedNames {
		n := normalizeCategoryName(rawName)
		cat, ok := index[n]
		if !ok {
			res.NotFoundNames = append(res.NotFoundNames, rawName)
			continue
		}

		// дубль проверка категорий конфига
		if _, exists := seenSlug[cat.Slug]; exists {
			continue
		}
		seenSlug[cat.Slug] = struct{}{}

		res.Slugs = append(res.Slugs, cat.Slug)
	}

	return res
}

// BuildAvailableCategoriesHint возвращает строку со списком доступных категорий
func BuildAvailableCategoriesHint(storeCategories []kuper.Category) string {
	lines := make([]string, 0, len(storeCategories))

	for _, c := range storeCategories {
		if strings.ToLower(c.Type) != "department" {
			continue
		}
		lines = append(lines, fmt.Sprintf("- %s", c.Name))
	}

	if len(lines) == 0 {
		return "Доступные категории не найдены (type=department)"
	}

	return "Доступные категории (type=department):\n" + strings.Join(lines, "\n")
}

// normalizeCategoryName приводит название категории к специальному виду для строгого сравнения
func normalizeCategoryName(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, "ё", "е")
	s = collapseSpaces(s)
	return s
}

var spaceRe = regexp.MustCompile(`\s+`)

func collapseSpaces(s string) string {
	return spaceRe.ReplaceAllString(s, " ")
}
