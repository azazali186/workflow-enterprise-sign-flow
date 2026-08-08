package repo

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/aeroxe/sign-flow/backend/internal/pkg/pagination"
)

// sortValue extracts the value of a column from a model for cursor encoding.
// Time columns (value or pointer) are encoded as Unix seconds.
func (r *Repo[T]) sortValue(m T, col string, isTime bool) string {
	if isTime {
		if t := timeOf(fieldOf(m, col)); t != nil {
			return strconv.FormatInt(t.Unix(), 10)
		}
	}
	return pagination.FormatValue(fieldOf(m, col))
}

// timeOf extracts a time.Time from a value or pointer-to-time.
func timeOf(v any) *time.Time {
	switch t := v.(type) {
	case time.Time:
		return &t
	case *time.Time:
		return t
	}
	return nil
}

// fieldByName finds a struct field case-insensitively, so "id" maps to the
// Go field "ID" and "api_key" maps to "ApiKey"/"APIKey". The returned
// index is the full path (including embedded structs).
func fieldByName(t reflect.Type, col string) (reflect.StructField, []int, bool) {
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.Anonymous {
			if ff, idx, ok := fieldByName(f.Type, col); ok {
				return ff, append([]int{i}, idx...), true
			}
			continue
		}
		if strings.EqualFold(f.Name, columnToField(col)) {
			return f, []int{i}, true
		}
	}
	return reflect.StructField{}, nil, false
}

func fieldOf[T any](m T, col string) any {
	v := reflect.ValueOf(m)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	t := v.Type()
	f, idx, ok := fieldByName(t, col)
	if !ok {
		return ""
	}
	if !f.IsExported() {
		return ""
	}
	return v.FieldByIndex(idx).Interface()
}

// idOf extracts the record id.
func (r *Repo[T]) idOf(m T) string {
	return fmt.Sprintf("%v", fieldOf(m, "id"))
}

func columnToField(col string) string {
	parts := strings.Split(col, "_")
	var b strings.Builder
	for _, p := range parts {
		b.WriteString(strings.ToUpper(p[:1]) + p[1:])
	}
	return b.String()
}
