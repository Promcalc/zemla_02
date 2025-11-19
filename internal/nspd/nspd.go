// Package nspd предоставляет клиент для работы с nspd.gov.ru (геопортал).
// Требует инициализации через запрос к карте для получения кук и заголовков.
package nspd

/*
// Создаём клиент один раз при старте
nspdClient := nspd.NewClient(logger)

// Инициализируем сессию (один раз!)
if err := nspdClient.Initialize(ctx); err != nil {
    logger.Error("Не удалось инициализировать nspd", "error", err)
    // Можно пропустить и попробовать позже при повторной обработке
}

// Позже, при обработке лота с кадастровым номером:
if cadastralNumber != "" {
    geoData, err := nspdClient.SearchByCadastralNumber(ctx, cadastralNumber)
    if err != nil {
        logger.Warn("Ошибка запроса к nspd.gov.ru", "cn", cadastralNumber, "error", err)
        // Сохранить err в external_data.nspd_error
    } else if geoData != nil && len(geoData.FeatureCollection.Features) > 0 {
        // Извлечь geometry и centroid
        // Сохранить geoData в external_data.nspd_data
    }
}
*/
import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"

	"github.com/go-resty/resty/v2"
)

// GeoportalResponse — структура ответа от geoportal API
type GeoportalResponse struct {
	FeatureCollection struct {
		Features []struct {
			Properties struct {
				Cn       string `json:"cn"`       // кадастровый номер
				Address  string `json:"adr"`      // адрес
				Area     string `json:"area"`     // площадь
				Category string `json:"category"` // категория
				Use      string `json:"use"`      // вид разрешённого использования
				Status   string `json:"status"`   // статус
			} `json:"properties"`
			Geometry json.RawMessage `json:"geometry"` // геометрия в GeoJSON
		} `json:"features"`
	} `json:"featureCollection"`
	TotalCount int `json:"totalCount"`
}

// Client — клиент для nspd.gov.ru
type Client struct {
	httpClient  *resty.Client
	initialized bool
	logger      *slog.Logger
}

// NewClient создаёт клиент с пустой сессией.
// Перед первым запросом к API необходимо вызвать Initialize().
func NewClient(config NSPDAPIConfig, logger *slog.Logger) *Client {
	// Создаём cookie jar для хранения кук
	jar, _ := cookiejar.New(nil)

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: config.IgnoreSSL, // Теперь настраивается из конфига
		},
	}

	baseClient := &http.Client{
		Timeout:   config.Timeout,
		Transport: transport,
		Jar:       jar, // важно: сохраняем куки между запросами
	}

	rc := resty.NewWithClient(baseClient)
	rc.SetHeader("User-Agent", config.UserAgent)
	rc.SetHeader("Accept", "application/json, text/plain, */*")

	return &Client{
		httpClient: rc,
		logger:     logger.With("service", "nspd"),
	}
}

// Initialize выполняет запрос к карте для получения кук и Referer.
// Должен вызываться один раз перед первым запросом к API.
func (c *Client) Initialize(ctx context.Context) error {
	if c.initialized {
		return nil // уже инициализирован
	}

	c.logger.Info("Инициализация сессии nspd.gov.ru: запрос к карте...")

	// Запрос к карте
	resp, err := c.httpClient.R().
		SetContext(ctx).
		Get("https://nspd.gov.ru/map?thematic=PKK")

	if err != nil {
		return fmt.Errorf("ошибка при инициализации сессии: %w", err)
	}

	if resp.StatusCode() != http.StatusOK && resp.StatusCode() != http.StatusNotFound {
		// Карта может вернуть 404, но куки всё равно установятся
		c.logger.Warn("Неожиданный статус при инициализации", "status", resp.StatusCode())
	}

	// После этого httpClient.Jar содержит куки, а последующие запросы их будут использовать
	c.initialized = true
	c.logger.Info("Сессия nspd.gov.ru успешно инициализирована")
	return nil
}

// SearchByCadastralNumber запрашивает данные по кадастровому номеру.
// Требует предварительного вызова Initialize().
func (c *Client) SearchByCadastralNumber(ctx context.Context, cadastralNumber string) (*GeoportalResponse, error) {
	if !c.initialized {
		return nil, fmt.Errorf("клиент nspd не инициализирован: вызовите Initialize() сначала")
	}

	c.logger.Info("Запрос данных по кадастровому номеру", "cn", cadastralNumber)

	// Требование ТЗ: задержка между запросами
	time.Sleep(500 * time.Millisecond)

	// Формируем URL
	params := url.Values{}
	params.Set("thematicSearchId", "1")
	params.Set("query", cadastralNumber)
	// Используем базовый URL из конфига
	fullURL := "https://nspd.gov.ru/api/geoportal/v2/search/geoportal?" + params.Encode()

	// Выполняем запрос с Referer из карты
	resp, err := c.httpClient.R().
		SetContext(ctx).
		SetHeader("Referer", "https://nspd.gov.ru/map").
		SetResult(&GeoportalResponse{}).
		Get(fullURL)

	if err != nil {
		return nil, fmt.Errorf("ошибка HTTP-запроса к nspd.gov.ru: %w", err)
	}

	if resp.StatusCode() == http.StatusNotFound {
		c.logger.Warn("Кадастровый номер не найден", "cn", cadastralNumber)
		return nil, nil // возвращаем nil без ошибки
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("некорректный статус от nspd.gov.ru: %d", resp.StatusCode())
	}

	result := resp.Result().(*GeoportalResponse)
	c.logger.Debug("Успешно получены данные geoportal", "cn", cadastralNumber, "features", len(result.FeatureCollection.Features))
	return result, nil
}
