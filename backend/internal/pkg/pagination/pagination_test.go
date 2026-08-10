package pagination

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseSort(t *testing.T) {
	tests := []struct {
		in    string
		field string
		desc  bool
	}{
		{"created_at", "created_at", true},
		{"-created_at", "created_at", true},
		{"created_at:asc", "created_at", false},
		{"created_at:desc", "created_at", true},
		{"", "", true},
	}
	for _, tt := range tests {
		got := ParseSort(tt.in)
		assert.Equal(t, tt.field, got.Field, "field for %q", tt.in)
		assert.Equal(t, tt.desc, got.Desc, "desc for %q", tt.in)
	}
}

func TestCursorRoundTrip(t *testing.T) {
	enc := EncodeCursor("2024-01-01T00:00:00Z", "0190abc-123")
	dec, err := DecodeCursor(enc)
	require.NoError(t, err)
	assert.Equal(t, "2024-01-01T00:00:00Z", dec.V)
	assert.Equal(t, "0190abc-123", dec.I)
}

func TestDecodeCursorInvalid(t *testing.T) {
	_, err := DecodeCursor("!!!not-base64!!!")
	assert.Error(t, err)
	_, err = DecodeCursor("e30=") // valid b64 but empty id
	assert.Error(t, err)
}

func TestFormatValue(t *testing.T) {
	assert.Equal(t, "42", FormatValue(42))
	assert.Equal(t, "abc", FormatValue("abc"))
	ts := time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)
	assert.Equal(t, "2024-01-02T03:04:05Z", FormatValue(ts))
	assert.Equal(t, "true", FormatValue(true))
}

func TestNormalizeLimit(t *testing.T) {
	assert.Equal(t, 20, NormalizeLimit(0))
	assert.Equal(t, 5, NormalizeLimit(5))
	assert.Equal(t, 100, NormalizeLimit(500))
}
