// Package torgi предоставляет клиент для работы с API torgi.gov.ru.
package torgi

/*

// Внутри основного цикла обработки лотов
torgiClient := torgi.NewClient(logger)

lotInfo, err := torgiClient.GetLotInfo(ctx, rssLot.Link)
if err != nil {
    logger.Warn("Ошибка получения данных с torgi.gov.ru",
        "guid", rssLot.GUID,
        "link", rssLot.Link,
        "error", err,
    )
    // Сохранить ошибку в external_data.lot_info_error
} else {
    // Сохранить lotInfo в external_data.lot_info
    logger.Info("Данные с torgi.gov.ru получены", "guid", rssLot.GUID)
}

*/

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
)

// LotInfo представляет структуру ответа от /api/public/lotcards/{lot_id}
// (точную структуру можно уточнить по реальному ответу; здесь — заглушка)
type LotInfo struct {
	ID          string      `json:"id"`
	Title       string      `json:"title"`
	Description string      `json:"description"`
	StartDate   string      `json:"startDate"`
	EndDate     string      `json:"endDate"`
	Price       interface{} `json:"price"` // может быть объектом или числом
	// Другие поля можно добавить по мере необходимости
}

// Client — клиент для запросов к torgi.gov.ru
type Client struct {
	httpClient *resty.Client
	logger     *slog.Logger
}

// NewClient создаёт новый клиент torgi с поддержкой:
// - игнорирования SSL-сертификатов
// - задержки между запросами
// - логирования
func NewClient(logger *slog.Logger) *Client {
	// Создаём HTTP-транспорт с InsecureSkipVerify
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // Требуется по ТЗ из-за SSL-проблем
		},
	}

	// Базовый HTTP-клиент
	baseClient := &http.Client{
		Timeout:   30 * time.Second,
		Transport: transport,
	}

	// Resty-клиент
	rc := resty.NewWithClient(baseClient)
	rc.SetBaseURL("https://torgi.gov.ru")
	rc.SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	rc.SetHeader("Accept", "application/json")

	logger = logger.With("service", "torgi")

	return &Client{
		httpClient: rc,
		logger:     logger,
	}
}

// ExtractLotID извлекает lot_id из RSS link.
// Пример: https://torgi.gov.ru/lot/123456789_0 → "123456789_0"
func ExtractLotID(link string) (string, error) {
	if link == "" {
		return "", fmt.Errorf("empty link")
	}

	// Убираем завершающий слеш
	link = strings.TrimSuffix(link, "/")

	// Берём последнюю часть пути
	parts := strings.Split(link, "/")
	if len(parts) == 0 {
		return "", fmt.Errorf("invalid link format: %s", link)
	}

	lotID := parts[len(parts)-1]

	// Проверяем, что lot_id содержит только цифры, подчёркивания и дефисы
	if !regexp.MustCompile(`^[a-zA-Z0-9_-]+$`).MatchString(lotID) {
		return "", fmt.Errorf("invalid lot_id format: %s", lotID)
	}

	return lotID, nil
}

// GetLotInfo получает детали лота по его RSS link.
// Возвращает JSON-ответ в виде map[string]interface{} или ошибку.
func (c *Client) GetLotInfo(ctx context.Context, rssLink string) (map[string]interface{}, error) {
	lotID, err := ExtractLotID(rssLink)
	if err != nil {
		return nil, fmt.Errorf("ошибка извлечения lot_id из ссылки %q: %w", rssLink, err)
	}

	c.logger.Info("Запрос деталей лота", "lot_id", lotID, "rss_link", rssLink)

	// Добавляем задержку между запросами (требование ТЗ)
	time.Sleep(500 * time.Millisecond)

	// Формируем URL
	url := fmt.Sprintf("/new/api/public/lotcards/%s", lotID)

	// Выполняем запрос
	resp, err := c.httpClient.R().
		SetContext(ctx).
		SetHeader("Referer", rssLink).
		SetResult(map[string]interface{}{}).
		Get(url)

	if err != nil {
		return nil, fmt.Errorf("ошибка HTTP-запроса к torgi.gov.ru: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		// Для 404 можно логировать как "лот не найден", но не паниковать
		if resp.StatusCode() == http.StatusNotFound {
			return nil, fmt.Errorf("лот не найден (404) для lot_id=%s", lotID)
		}
		return nil, fmt.Errorf("некорректный статус ответа от torgi.gov.ru: %d", resp.StatusCode())
	}

	// Проверяем, что тело ответа не пустое
	if resp.Body() == nil || len(resp.Body()) == 0 {
		return nil, fmt.Errorf("получен пустой ответ от torgi.gov.ru для lot_id=%s", lotID)
	}

	// Парсим JSON в map
	var result map[string]interface{}
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("ошибка парсинга JSON от torgi.gov.ru: %w", err)
	}

	c.logger.Debug("Успешно получены детали лота", "lot_id", lotID, "keys", len(result))
	return result, nil
}
