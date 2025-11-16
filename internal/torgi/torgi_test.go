package torgi

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractLotID(t *testing.T) {
	cases := []struct {
		link     string
		expected string
		hasError bool
	}{
		{"https://torgi.gov.ru/lot/123456789_0", "123456789_0", false},
		{"https://torgi.gov.ru/lot/ABC-123_def", "ABC-123_def", false},
		{"https://torgi.gov.ru/lot/", "", true},
		{"", "", true},
		{"https://example.com/lot/123<script>", "", true}, // защита от инъекций
	}

	for _, tc := range cases {
		result, err := ExtractLotID(tc.link)
		if tc.hasError {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, result)
		}
	}
}