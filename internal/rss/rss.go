// Package rss содержит логику загрузки и парсинга RSS-ленты с torgi.gov.ru.
package rss

/*
// Внутри main или scheduler
lots, err := rssParser.FetchAndParse(ctx)
if err != nil {  обработка  }

dbRepo, err := db.NewRepository(ctx, cfg.Database.URL, logger)
if err != nil {  обработка  }

torgiClient := torgi.NewClient(logger)
nspdClient := nspd.NewClient(logger)
_ = nspdClient.Initialize(ctx)

for _, rssLot := range lots {
	// Извлекаем кадастровый номер и дату
	extracted := parser.ExtractFromLot(rssLot)

	// Запрос к torgi
	lotInfo, torgiErr := torgiClient.GetLotInfo(ctx, rssLot.Link)

	// Запрос к nspd (если есть кадастровый номер)
	var nspdResp *nspd.GeoportalResponse
	var nspdErr error
	if extracted.CadastralNumber != "" {
		nspdResp, nspdErr = nspdClient.SearchByCadastralNumber(ctx, extracted.CadastralNumber)
	}

	// Сохраняем всё атомарно
	_, err := dbRepo.SaveLotWithExternalData(
		ctx, rssLot, extracted,
		lotInfo,
		nspdResp,
		torgiErr, nspdErr,
	)
	if err != nil {
		logger.Error("Ошибка сохранения лота", "guid", rssLot.GUID, "error", err)
	}
}
*/

import (
	"context"
	"crypto/tls"
	"fmt"
	"html"
	"log/slog"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/Promcalc/zemla_02/internal/config"
	"github.com/mmcdole/gofeed"
)

// Lot представляет один лот из RSS-ленты после парсинга.
type Lot struct {
	GUID        string            `json:"guid"`
	Link        string            `json:"link"`
	Title       string            `json:"title"`
	PubDate     time.Time         `json:"pub_date"`
	DCDate      time.Time         `json:"dc_date"`
	Description string            `json:"description"`
	Fields      map[string]string `json:"fields"` // динамические поля из <description>
}

// Parser обёртка вокруг gofeed с поддержкой логирования и конфигурации.
type Parser struct {
	cfg    *config.Config
	logger *slog.Logger
	fp     *gofeed.Parser
}

// NewParser создаёт новый RSS-парсер с заданной конфигурацией и логгером.
func NewParser(cfg *config.Config, logger *slog.Logger) *Parser {
	if cfg == nil || cfg.RSS.Timeout == 0 {
		// fallback
		cfg = &config.Config{
			RSS: config.RSSConfig{
				Timeout: 30 * time.Second,
				URL:     "https://torgi.gov.ru/new/api/public/lotcards/rss?lotStatus=PUBLISHED,APPLICATIONS_SUBMISSION&catCode=2&byFirstVersion=true",
			},
		}
	}

	fp := gofeed.NewParser()

	// Важно: инициализируем Client, если он nil
	if fp.Client == nil {
		fp.Client = &http.Client{
			// Отключаем проверку SSL (по ТЗ: "с verify=False из-за SSL-проблем")
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true, // ← По требованиям ТЗ
				},
			},
		}
	}

	fp.Client.Timeout = cfg.RSS.Timeout

	return &Parser{
		cfg:    cfg,
		logger: logger,
		fp:     fp,
	}
}

// FetchAndParse загружает и парсит RSS-ленту.
// Возвращает список лотов, отсортированных по возрастанию pubDate.
func (p *Parser) FetchAndParse(ctx context.Context) ([]Lot, error) {
	logger := p.logger.With("url", p.cfg.RSS.URL)
	logger.Info("Попытка загрузки RSS-ленты")

	// Загружаем RSS
	feed, err := p.fp.ParseURLWithContext(p.cfg.RSS.URL, ctx)
	if err != nil {
		return nil, fmt.Errorf("ошибка загрузки RSS: %w", err)
	}

	logger.Info("RSS успешно загружен", "items_total", len(feed.Items))

	var lots []Lot
	for i, item := range feed.Items {
		lot, err := p.parseItem(item)
		if err != nil {
			// Не останавливаемся на одной ошибке — пропускаем битый item
			logger.Warn("Ошибка парсинга item", "index", i, "guid", item.GUID, "error", err)
			continue
		}
		lots = append(lots, lot)
	}

	logger.Info("Парсинг RSS завершён", "lots_parsed", len(lots))
	return lots, nil
}

// parseItem преобразует gofeed.Item в структурированный Lot.
// Обрабатывает поля RSS и разбирает description на динамические поля.
func (p *Parser) parseItem(item *gofeed.Item) (Lot, error) {
	if item.GUID == "" {
		return Lot{}, fmt.Errorf("поле GUID отсутствует")
	}

	// Парсим даты
	var pubDate, dcDate time.Time
	// var err error

	if item.PublishedParsed != nil {
		pubDate = *item.PublishedParsed
	}

	// Парсим dc:date
	// var dcDate time.Time
	if dcExt, hasDc := item.Extensions["dc"]; hasDc {
		if dateExts, hasDate := dcExt["date"]; hasDate && len(dateExts) > 0 {
			dcDateStr := dateExts[0].Value
			if dcDateStr != "" {
				if t, err := time.Parse(time.RFC3339, dcDateStr); err == nil {
					dcDate = t
				} else if t, err := time.Parse("2006-01-02T15:04:05-07:00", dcDateStr); err == nil {
					dcDate = t
				} else {
					p.logger.Warn("Не удалось распарсить dc:date", "value", dcDateStr, "error", err)
				}
			}
		}
	}

	// if dcDateStr, ok := item.Extensions["dc"]["date"][0].Value; ok && dcDateStr != "" {
	// 	// Пример: "2025-01-01T10:00:00+03:00"
	// 	dcDate, err = time.Parse("2006-01-02T15:04:05-07:00", dcDateStr)
	// 	if err != nil {
	// 		// Попробуем RFC3339
	// 		dcDate, err = time.Parse(time.RFC3339, dcDateStr)
	// 		if err != nil {
	// 			return Lot{}, fmt.Errorf("не удалось распарсить dc:date '%s': %w", dcDateStr, err)
	// 		}
	// 	}
	// }

	// Обрабатываем description
	fields := p.parseDescription(item.Description)

	lot := Lot{
		GUID:        item.GUID,
		Link:        item.Link,
		Title:       item.Title,
		PubDate:     pubDate,
		DCDate:      dcDate,
		Description: item.Description,
		Fields:      fields,
	}

	return lot, nil
}

// parseDescription разбирает HTML-описание на "Заголовок: Значение".
// Поддерживает теги <br>, <br>, и т.д.
func (p *Parser) parseDescription(desc string) map[string]string {
	if desc == "" {
		return nil
	}

	// Заменяем все виды <br> на \n
	desc = regexp.MustCompile(`(?i)<br\s*/?>`).ReplaceAllString(desc, "\n")
	desc = regexp.MustCompile(`<br\s*/?>`).ReplaceAllString(desc, "\n")

	lines := strings.Split(desc, "\n")
	fields := make(map[string]string)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Удаляем HTML-теги и сущности
		line = regexp.MustCompile(`<[^>]*>`).ReplaceAllString(line, "")
		line = html.UnescapeString(line)

		// Ищем разделитель ":"
		if idx := strings.Index(line, ":"); idx != -1 {
			key := strings.TrimSpace(line[:idx])
			value := strings.TrimSpace(line[idx+1:])

			// Нормализация ключа:
			// - убрать ":" в конце (уже сделано)
			// - заменить ":" на "_" (например, "dc:date" → "dc_date")
			key = strings.ReplaceAll(key, ":", "_")
			// - привести к "ИмяПоля"
			key = p.normalizeFieldName(key)

			if key != "" && value != "" {
				fields[key] = value
			}
		}
	}

	return fields
}

// normalizeFieldName приводит имя поля к формату "ЗаголовокПоля".
// Пример: "Кадастровый номер" → "КадастровыйНомер"
func (p *Parser) normalizeFieldName(name string) string {
	// Разделяем по пробелам, дефисам, подчёркиваниям
	parts := regexp.MustCompile(`[\s\-_]+`).Split(name, -1)
	var normalized []string
	for _, part := range parts {
		if part == "" {
			continue
		}
		// Первая буква — заглавная, остальные — строчные
		normalized = append(normalized, strings.Title(strings.ToLower(part)))
	}
	return strings.Join(normalized, "")
}
