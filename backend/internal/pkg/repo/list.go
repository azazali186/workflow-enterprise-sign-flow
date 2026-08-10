package repo

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/aeroxe/sign-flow/backend/internal/pkg/pagination"
)

// Page is the unified list response: items + pagination summary + db summary.
type Page[T any] struct {
	Items      []T                 `json:"items"`
	Pagination pagination.PageInfo `json:"pagination"`
	Summary    any                 `json:"summary,omitempty"`
}

// List runs a filtered, sorted, cursor-paginated query with optional summary.
func (r *Repo[T]) List(ctx context.Context, q pagination.Query) (*Page[T], error) {
	limit := pagination.NormalizeLimit(q.Limit)

	sortInfo := pagination.ParseSort(q.Sort)
	if sortInfo.Field == "" {
		sortInfo = pagination.ParseSort(r.opts.DefaultSort)
	}
	sortCol, ok := r.sortColumn(sortInfo.Field)
	if !ok {
		sortInfo = pagination.ParseSort(r.opts.DefaultSort)
		sortCol, _ = r.sortColumn(sortInfo.Field)
	}
	if sortCol == "" {
		sortCol, sortInfo = "created_at", pagination.SortInfo{Field: "created_at", Desc: true}
	}
	isTime := r.isTimeColumn(sortCol)
	sortExpr := sortCol
	if isTime {
		sortExpr = r.timeExpr(sortCol)
	}
	order := "asc"
	if sortInfo.Desc {
		order = "desc"
	}

	// total (filters only, no cursor/sort/limit)
	var total int64
	base := r.buildFiltered(ctx, q)
	if err := base.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, err
	}

	// cursor predicate on (sortCol, id); time values are compared as epochs.
	tx := r.buildFiltered(ctx, q)
	if q.Cursor != "" {
		c, err := pagination.DecodeCursor(q.Cursor)
		if err != nil {
			return nil, err
		}
		v := r.cursorValue(sortCol, isTime, c.V)
		if sortInfo.Desc {
			tx = tx.Where("("+sortExpr+" < ?) OR ("+sortExpr+" = ? AND id < ?)", v, v, c.I)
		} else {
			tx = tx.Where("("+sortExpr+" > ?) OR ("+sortExpr+" = ? AND id > ?)", v, v, c.I)
		}
	}
	tx = tx.Order(sortExpr + " " + order).Order("id " + order).Limit(limit + 1)

	var items []T
	if err := tx.Find(&items).Error; err != nil {
		return nil, err
	}

	hasMore := len(items) > limit
	if hasMore {
		items = items[:limit]
	}

	page := &Page[T]{Items: items, Pagination: pagination.PageInfo{Limit: limit, Total: total}}
	if len(items) > 0 {
		last := items[len(items)-1]
		page.Pagination.NextCursor = pagination.EncodeCursor(r.sortValue(last, sortCol, isTime), r.idOf(last))
		page.Pagination.HasMore = hasMore
	}

	if r.opts.Summary != nil {
		summary, err := r.opts.Summary(r.buildFiltered(ctx, q).Session(&gorm.Session{}))
		if err != nil {
			return nil, err
		}
		page.Summary = summary
	}
	return page, nil
}

// buildFiltered applies filters, search and date range to a fresh query.
func (r *Repo[T]) buildFiltered(ctx context.Context, q pagination.Query) *gorm.DB {
	tx := r.db.WithContext(ctx).Model(new(T))
	for _, p := range r.opts.Preloads {
		tx = tx.Preload(p)
	}
	for key, val := range q.Filters {
		col, op, ok := r.filterColumn(key)
		if !ok || val == nil {
			continue
		}
		switch op {
		case "like":
			tx = tx.Where("LOWER("+col+") LIKE LOWER(?)", "%"+fmt.Sprintf("%v", val)+"%")
		case "gte":
			tx = tx.Where(r.compareExpr(col)+" >= ?", val)
		case "lte":
			tx = tx.Where(r.compareExpr(col)+" <= ?", val)
		case "neq":
			tx = tx.Where(col+" != ?", val)
		default:
			tx = tx.Where(col+" = ?", val)
		}
	}
	if q.Search != "" && len(r.opts.Searchable) > 0 {
		var b strings.Builder
		b.WriteString("LOWER(" + r.opts.Searchable[0] + ") LIKE LOWER(?)")
		for _, col := range r.opts.Searchable[1:] {
			b.WriteString(" OR LOWER(" + col + ") LIKE LOWER(?)")
		}
		qargs := make([]any, len(r.opts.Searchable))
		for i := range qargs {
			qargs[i] = "%" + q.Search + "%"
		}
		tx = tx.Where(b.String(), qargs...)
	}
	if q.DateFrom != nil {
		tx = tx.Where(r.timeExpr(dateCol(r.opts.DateFields, q.DateField))+" >= ?", r.timeParam(*q.DateFrom))
	}
	if q.DateTo != nil {
		tx = tx.Where(r.timeExpr(dateCol(r.opts.DateFields, q.DateField))+" <= ?", r.timeParam(*q.DateTo))
	}
	return tx
}

// timeExpr returns a driver-agnostic epoch-seconds expression for a column.
// SQLite's strftime yields TEXT; casting to INTEGER makes both the ORDER BY
// and cursor predicates compare numerically and stay correct across digit
// boundaries (e.g. years beyond 2286).
func (r *Repo[T]) timeExpr(col string) string {
	if r.db.Dialector.Name() == "sqlite" {
		return "CAST(strftime('%s', " + col + ") AS INTEGER)"
	}
	return "EXTRACT(EPOCH FROM " + col + ")"
}

// timeParam renders a time value for comparison against the epoch expression.
// SQLite returns TEXT from strftime and mis-coerces int64 params, so it needs
// the epoch as a string; Postgres compares numerically with a typed int64.
func (r *Repo[T]) timeParam(t time.Time) any {
	if r.db.Dialector.Name() == "sqlite" {
		return strconv.FormatInt(t.Unix(), 10)
	}
	return t.Unix()
}

// compareExpr keeps equality comparisons typed; ranges on non-time columns
// stay plain (values are user-provided and parameterised by GORM).
func (r *Repo[T]) compareExpr(col string) string { return col }

func dateCol(allowed []string, field string) string {
	for _, c := range allowed {
		if c == field {
			return c
		}
	}
	return "created_at"
}

// filterColumn resolves a request filter key to (column, operator).
func (r *Repo[T]) filterColumn(key string) (string, string, bool) {
	op := "eq"
	base := key
	for _, suffix := range []string{"_like", "_gte", "_lte", "_neq"} {
		if strings.HasSuffix(key, suffix) {
			op = strings.TrimPrefix(suffix, "_")
			base = strings.TrimSuffix(key, suffix)
			break
		}
	}
	col, ok := r.opts.Filterable[base]
	if !ok && isColumn(base) {
		col = base
	}
	return col, op, col != ""
}

// sortColumn resolves a request sort field to a db column.
func (r *Repo[T]) sortColumn(field string) (string, bool) {
	if field == "" {
		return "", false
	}
	if col, ok := r.opts.Sortable[field]; ok {
		return col, true
	}
	if isColumn(field) {
		return field, true
	}
	return "", false
}

func isColumn(s string) bool {
	switch s {
	case "id", "created_at", "updated_at":
		return true
	}
	return false
}

// isTimeColumn reports whether a column maps to a time.Time field
// (pointer or value).
func (r *Repo[T]) isTimeColumn(col string) bool {
	return isTimeType(r.fieldType(col))
}

// isTimeType is true for time.Time and *time.Time.
func isTimeType(t reflect.Type) bool {
	for t != nil && t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t == reflect.TypeOf(time.Time{})
}

func (r *Repo[T]) fieldType(col string) reflect.Type {
	t := reflect.TypeOf(new(T)).Elem()
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if f, _, ok := fieldByName(t, col); ok {
		return f.Type
	}
	return nil
}

// cursorValue converts a cursor string into the value used for comparison.
func (r *Repo[T]) cursorValue(col string, isTime bool, v string) any {
	if isTime {
		if r.db.Dialector.Name() == "sqlite" {
			return v // strftime yields TEXT; compare as text
		}
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			return n
		}
		return 0
	}
	return r.typedParam(col, v)
}

// typedParam converts a cursor string value into the Go type of the column.
func (r *Repo[T]) typedParam(col, v string) any {
	t := r.fieldType(col)
	if t == nil {
		return v
	}
	switch t.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			return n
		}
	case reflect.Float32, reflect.Float64:
		if n, err := strconv.ParseFloat(v, 64); err == nil {
			return n
		}
	case reflect.Bool:
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return v
}

// sortValue, fieldOf and friends live in cursor.go to keep this file focused.
