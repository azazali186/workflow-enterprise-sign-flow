// Package pagination implements server-side cursor based pagination with
// filters, date-range filtering and dynamic column sorting.
package pagination

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/aeroxe/sign-flow/backend/internal/pkg/errs"
)

// Query is the common pagination request accepted by every list/report API.
type Query struct {
	Limit     int            `json:"limit"`
	Cursor    string         `json:"cursor"`
	Filters   map[string]any `json:"filters"`
	Search    string         `json:"search"`
	Sort      string         `json:"sort"` // "field" or "field:asc|desc" or "-field"
	DateFrom  *time.Time     `json:"date_from"`
	DateTo    *time.Time     `json:"date_to"`
	DateField string         `json:"date_field"` // whitelisted via Options.DateFields
}

// NormalizeLimit clamps limit to [1, 100], defaulting to 20.
func NormalizeLimit(limit int) int {
	if limit <= 0 {
		return 20
	}
	if limit > 100 {
		return 100
	}
	return limit
}

// Cursor is the opaque, URL-safe cursor payload.
type Cursor struct {
	V string `json:"v"` // sort column value of the last item
	I string `json:"i"` // id of the last item
}

// EncodeCursor serialises a cursor to a base64 string.
func EncodeCursor(v, id string) string {
	b, _ := json.Marshal(Cursor{V: v, I: id})
	return base64.RawURLEncoding.EncodeToString(b)
}

// DecodeCursor parses a cursor string.
func DecodeCursor(s string) (Cursor, error) {
	var c Cursor
	if s == "" {
		return c, nil
	}
	raw, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return c, errs.ErrInvalidCursor
	}
	if err := json.Unmarshal(raw, &c); err != nil {
		return c, errs.ErrInvalidCursor
	}
	if c.I == "" {
		return c, errs.ErrInvalidCursor
	}
	return c, nil
}

// SortInfo is a parsed sort directive.
type SortInfo struct {
	Field string
	Desc  bool
}

// ParseSort parses "field", "field:desc", "field:asc" or "-field".
func ParseSort(sort string) SortInfo {
	s := strings.TrimSpace(sort)
	if s == "" {
		return SortInfo{Field: "", Desc: true}
	}
	if strings.HasPrefix(s, "-") {
		return SortInfo{Field: strings.TrimPrefix(s, "-"), Desc: true}
	}
	parts := strings.Split(s, ":")
	if len(parts) == 2 && strings.EqualFold(parts[1], "asc") {
		return SortInfo{Field: parts[0], Desc: false}
	}
	return SortInfo{Field: parts[0], Desc: true}
}

// FormatValue renders a sort value for cursor encoding.
func FormatValue(v any) string {
	switch t := v.(type) {
	case nil:
		return ""
	case time.Time:
		return t.UTC().Format(time.RFC3339Nano)
	case *time.Time:
		if t == nil {
			return ""
		}
		return t.UTC().Format(time.RFC3339Nano)
	case string:
		return t
	case bool:
		return strconv.FormatBool(t)
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64)
	default:
		return fmt.Sprintf("%v", t)
	}
}

// PageInfo is the pagination summary returned with every list response.
type PageInfo struct {
	NextCursor string `json:"next_cursor"`
	HasMore    bool   `json:"has_more"`
	Limit      int    `json:"limit"`
	Total      int64  `json:"total_count"`
}
