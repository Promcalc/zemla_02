// Package parser содержит логику извлечения структурированных данных из текста.
package parser

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/Promcalc/zemla_02/internal/rss"
)

// ExtractedData — результат извлечения
type ExtractedData struct {
	CadastralNumber string
	AuctionDate     time.Time
}

// ExtractFromLot извлекает кадастровый номер и дату аукциона из Lot.
// Следует приоритетам из ТЗ:
// 1. Поля из description, содержащие «кадастровый номер»
// 2. Другие поля item
// 3. Общий текст (title + description)
func ExtractFromLot(lot rss.Lot) ExtractedData {
	var cn string
	var auctionDate time.Time

	// 1. Поиск кадастрового номера в полях из description
	for key, value := range lot.Fields {
		if containsCadastralKey(key) {
			if match := extractCadastralNumber(value); match != "" {
				cn = match
				break
			}
		}
	}

	// 2. Если не нашли — пробежимся по всем полям
	if cn == "" {
		for _, value := range lot.Fields {
			if match := extractCadastralNumber(value); match != "" {
				cn = match
				break
			}
		}
	}

	// 3. Если всё ещё не нашли — ищем в title + description
	if cn == "" {
		text := lot.Title + " " + lot.Description
		cn = extractCadastralNumber(text)
	}

	// Извлечение даты аукциона
	auctionDate = extractAuctionDate(lot)

	return ExtractedData{
		CadastralNumber: cn,
		AuctionDate:     auctionDate,
	}
}

// containsCadastralKey проверяет, содержит ли ключ «кадастровый номер»
func containsCadastralKey(key string) bool {
	keyLower := strings.ToLower(key)
	return strings.Contains(keyLower, "кадастровый") && strings.Contains(keyLower, "номер")
}

// Регулярное выражение для кадастрового номера по ТЗ:
// \b\d{2}:\d{2}:\d{4,19}:\d{1,6}\b
var cadastralRegex = regexp.MustCompile(`\b\d{2}:\d{2}:\d{4,19}:\d{1,6}\b`)

// extractCadastralNumber ищет кадастровый номер в строке
func extractCadastralNumber(text string) string {
	matches := cadastralRegex.FindStringSubmatch(text)
	if len(matches) > 0 {
		return matches[0]
	}
	return ""
}

// extractAuctionDate пытается извлечь дату аукциона из полей
func extractAuctionDate(lot rss.Lot) time.Time {
	// Попробуем найти поле с датой аукциона
	for key, value := range lot.Fields {
		if isAuctionDateField(key) {
			if t, err := parseDateFromString(value); err == nil {
				return t
			}
		}
	}
	// Если не нашли — возвращаем нулевую дату
	return time.Time{}
}

// isAuctionDateField определяет, может ли поле содержать дату аукциона
func isAuctionDateField(key string) bool {
	key = strings.ToLower(key)
	return strings.Contains(key, "аукцион") && strings.Contains(key, "дата") ||
		strings.Contains(key, "дата") && (strings.Contains(key, "торг") || strings.Contains(key, "продаж"))
}

// parseDateFromString парсит дату из строки (поддерживаем несколько форматов)
func parseDateFromString(s string) (time.Time, error) {
	formats := []string{
		"02.01.2006",
		"2006-01-02",
		"02/01/2006",
		"2.1.2006",
		"2006.01.02",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("Unable to parse date: %q", s)
}
