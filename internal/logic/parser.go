package logic

import (
	"context"
	"fmt"
	"kuperparser/internal/client"
	"kuperparser/internal/config"
	"kuperparser/internal/kuper"
	"kuperparser/storage"
	"log"
	"os"
	"time"
)

func Run(ctx context.Context, cfg *config.Config) error {
	// Настройка http клиента
	timeout := time.Duration(cfg.HTTP.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	httpClient := client.NewHTTPClient(timeout)

	tcfg := client.TransportConfig{
		Timeout: timeout,
		Retries: cfg.HTTP.Retries,
		Workers: cfg.Concurrency.Workers,
	}

	// Выбор режима прокси из конфига
	switch client.ProxyMode(cfg.Proxy.Mode) {
	case client.ProxyDisabled, "":
		tcfg.ProxyMode = client.ProxyDisabled
	case client.ProxyList:
		tcfg.ProxyMode = client.ProxyList
		tcfg.ProxyList = cfg.Proxy.List
	case client.ProxyRotation:
		tcfg.ProxyMode = client.ProxyRotation
		tcfg.RotationURL = cfg.Proxy.RotationURL
	default:
		return fmt.Errorf("неизвестный proxy.mode=%q (ожидается disabled|list|rotation)", cfg.Proxy.Mode)
	}

	transport, err := client.Build(httpClient, tcfg)
	if err != nil {
		return fmt.Errorf("ошибка сборки transport слоя: %w", err)
	}
	// Создание клиента kuper
	kuperSvc := kuper.NewKuperService(transport)
	// Загрузка списка категорий магазина и получение slug при сравнении с выбранной категорией из конфига
	storeID := cfg.Kuper.StoreID
	log.Printf("Получаем категории магазина store_id=%d ...", storeID)

	categories, err := kuperSvc.ListCategories(ctx, storeID)
	if err != nil {
		return fmt.Errorf("не удалось получить категории: %w", err)
	}

	log.Printf("Категорий получено: %d", len(categories))

	log.Println(BuildAvailableCategoriesHint(categories))

	res := ResolveCategorySlugsByNames(cfg.Departments.Names, categories)

	for _, name := range res.NotFoundNames {
		log.Printf("WARN: Категория %q не найдена в магазине store_id=%d — пропускаю", name, storeID)
	}

	if len(res.Slugs) == 0 {
		return fmt.Errorf(
			"ни одна категория из конфига не найдена для store_id=%d\n%s",
			storeID,
			BuildAvailableCategoriesHint(categories),
		)
	}

	log.Println("Найденные slug'и категорий:")
	for _, slug := range res.Slugs {
		log.Printf("- %s", slug)
	}

	// Подготовка и сборка выходного файла
	storeInfo, err := kuperSvc.GetStore(ctx, storeID)
	if err != nil {
		return fmt.Errorf("не удалось получить информацию о магазине: %w", err)
	}

	if err := ensureDir(cfg.Output.Directory); err != nil {
		return fmt.Errorf("не удалось создать output директорию: %w", err)
	}

	for _, slug := range res.Slugs {
		fileName := fmt.Sprintf(
			"%s_%s_%s.csv",
			sanitizeFilePart(storeInfo.RetailerName),
			sanitizeFilePart(storeInfo.StoreAddress),
			sanitizeFilePart(slug),
		)

		fullPath := cfg.Output.Directory + "/" + fileName
		log.Printf("Пишем файл: %s", fullPath)

		w, err := storage.NewCSVWriter(fullPath)
		if err != nil {
			return fmt.Errorf("ошибка создания csv: %w", err)
		}

		page := 1
		perPage := cfg.Pagination.PerPage
		if perPage <= 0 {
			perPage = 5
		}
		if perPage > 5 {
			log.Printf("WARN: per_page=%d больше 5, для этого API максимум 5. Ставлю 5.", perPage)
			perPage = 5
		}

		offersLimit := cfg.Pagination.OffersLimit
		if offersLimit <= 0 {
			offersLimit = 10
		}

		total := 0
		for {
			prods, err := kuperSvc.ListProducts(ctx, storeID, slug, page, perPage, offersLimit)
			if err != nil {
				_ = w.Close()
				return fmt.Errorf("ошибка получения товаров (slug=%s page=%d): %w", slug, page, err)
			}

			if len(prods) == 0 {
				break
			}

			for _, p := range prods {
				name := extractName(p)
				price := extractPrice(p)
				url := extractURL(cfg.Kuper.BaseURL, p)

				if err := w.WriteRow(name, price, url); err != nil {
					_ = w.Close()
					return fmt.Errorf("ошибка записи csv: %w", err)
				}
				total++
			}

			page++
			if page > 500 {
				log.Printf("WARN: достигнут лимит страниц для slug=%s, останавливаемся", slug)
				break
			}
		}

		_ = w.Close()
		log.Printf("Готово: slug=%s, строк=%d", slug, total)
	}

	return nil

}

func ensureDir(path string) error {
	return os.MkdirAll(path, 0o755)
}
