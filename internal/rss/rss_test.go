package rss

import (
	"strings"
	"testing"
	"time"

	"github.com/mmcdole/gofeed"
	"github.com/stretchr/testify/assert"
)

func TestParseDescription(t *testing.T) {
	parser := &Parser{}

	desc := `
		Кадастровый номер: 50:26:0050304:123<br>
		Начальная цена: 1 000 000 руб.<br>
		dc:date: 2025-01-01T10:00:00+03:00
	`

	fields := parser.parseDescription(desc)

	assert.Equal(t, "50:26:0050304:123", fields["КадастровыйНомер"])
	assert.Equal(t, "1 000 000 руб.", fields["НачальнаяЦена"])
	assert.Equal(t, "2025-01-01T10:00:00+03:00", fields["Dc_Date"])
}

func TestParseItem(t *testing.T) {
	item := &gofeed.Item{
		GUID:      "test-guid",
		Link:      "https://example.com/lot/123",
		Title:     "Тестовый лот",
		Published: "Mon, 01 Jan 2025 10:00:00 +0300",
		Description: "Кадастровый номер: 50:26:0050304:123<br>Стартовая цена: 1 млн",
		Extensions: map[string]map[string][]gofeed.Extension{
			"dc": {
				"date": {{Value: "2025-01-01T10:00:00+03:00"}},
			},
		},
	}

	// Вручную распарсим дату публикации для теста
	pubTime, _ := time.Parse(time.RFC1123Z, "Mon, 01 Jan 2025 10:00:00 +0300")
	item.PublishedParsed = &pubTime

	parser := &Parser{}
	lot, err := parser.parseItem(item)

	assert.NoError(t, err)
	assert.Equal(t, "test-guid", lot.GUID)
	assert.Equal(t, "50:26:0050304:123", lot.Fields["КадастровыйНомер"])
	assert.Equal(t, "Тестовый лот", lot.Title)
	assert.Equal(t, "2025-01-01T10:00:00+03:00", lot.DCDate.Format("2006-01-02T15:04:05-07:00"))
}

func TestNormalizeFieldName(t *testing.T) {
	parser := &Parser{}
	cases := []struct {
		input    string
		expected string
	}{
		{"кадастровый номер", "КадастровыйНомер"},
		{"dc:date", "Dc_Date"},
		{"начальная_цена", "Начальная_Цена"},
		{"форма подачи заявок", "ФормаПодачиЗаявок"},
	}

	for _, tc := range cases {
		result := parser.normalizeFieldName(tc.input)
		assert.Equal(t, tc.expected, result, "input: %s", tc.input)
	}
}