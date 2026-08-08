package repo

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/aeroxe/sign-flow/backend/internal/cache"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/model"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/pagination"
)

type Widget struct {
	model.Base
	Name   string `gorm:"size:80" json:"name"`
	Status string `gorm:"size:20;index" json:"status"`
}

var dbSeq int

func newTestRepo(t *testing.T) (*Repo[Widget], *gorm.DB) {
	t.Helper()
	dbSeq++
	dsn := fmt.Sprintf("file:memdb%d?mode=memory&cache=shared&_loc=UTC", dbSeq)
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		NowFunc: func() time.Time { return time.Now().UTC() },
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&Widget{}))
	r := New(db, cache.NewMemory(), Options[Widget]{
		Slug:       "widget",
		Searchable: []string{"name"},
		Filterable: map[string]string{"status": "status"},
		Sortable:   map[string]string{"name": "name"},
		Summary: func(tx *gorm.DB) (any, error) {
			var rows []struct {
				Status string `gorm:"column:status"`
				Count  int64  `gorm:"column:count"`
			}
			err := tx.Model(&Widget{}).Select("status, count(*) as count").Group("status").Scan(&rows).Error
			return rows, err
		},
	})
	return r, db
}

func seedWidgets(t *testing.T, db *gorm.DB) []Widget {
	t.Helper()
	names := []string{"alpha", "beta", "gamma", "delta", "epsilon"}
	var out []Widget
	for i, n := range names {
		w := Widget{Name: n, Status: "new"}
		if i%2 == 0 {
			w.Status = "done"
		}
		require.NoError(t, db.Create(&w).Error)
		out = append(out, w)
	}
	return out
}

func TestCrud(t *testing.T) {
	r, _ := newTestRepo(t)
	ctx := context.Background()

	w := Widget{Name: "test", Status: "new"}
	require.NoError(t, r.Create(ctx, &w))
	assert.NotEmpty(t, w.ID)

	got, err := r.Get(ctx, w.ID)
	require.NoError(t, err)
	assert.Equal(t, "test", got.Name)

	updated, err := r.Patch(ctx, w.ID, map[string]any{"status": "done"})
	require.NoError(t, err)
	assert.Equal(t, "done", updated.Status)

	require.NoError(t, r.Delete(ctx, w.ID))
	_, err = r.Get(ctx, w.ID)
	assert.Error(t, err)
}

func TestListCursorPagination(t *testing.T) {
	r, db := newTestRepo(t)
	seedWidgets(t, db)
	ctx := context.Background()

	page1, err := r.List(ctx, pagination.Query{Limit: 2})
	require.NoError(t, err)
	assert.Len(t, page1.Items, 2)
	assert.True(t, page1.Pagination.HasMore)
	assert.NotEmpty(t, page1.Pagination.NextCursor)
	assert.Equal(t, int64(5), page1.Pagination.Total)

	page2, err := r.List(ctx, pagination.Query{Limit: 2, Cursor: page1.Pagination.NextCursor})
	require.NoError(t, err)
	assert.Len(t, page2.Items, 2)
	assert.NotEmpty(t, page2.Pagination.NextCursor)
	// no overlap between pages
	assert.NotEqual(t, page1.Items[0].ID, page2.Items[0].ID)

	page3, err := r.List(ctx, pagination.Query{Limit: 2, Cursor: page2.Pagination.NextCursor})
	require.NoError(t, err)
	assert.Len(t, page3.Items, 1)
	assert.False(t, page3.Pagination.HasMore)
}

func TestListFiltersSearchSort(t *testing.T) {
	r, db := newTestRepo(t)
	seedWidgets(t, db)
	ctx := context.Background()

	// filter
	page, err := r.List(ctx, pagination.Query{Limit: 10, Filters: map[string]any{"status": "done"}})
	require.NoError(t, err)
	assert.Len(t, page.Items, 3)
	assert.Equal(t, int64(3), page.Pagination.Total)

	// search
	page, err = r.List(ctx, pagination.Query{Limit: 10, Search: "beta"})
	require.NoError(t, err)
	assert.Len(t, page.Items, 1)
	assert.Equal(t, "beta", page.Items[0].Name)

	// sort asc by name (alphabetical: alpha, beta, delta, epsilon, gamma)
	page, err = r.List(ctx, pagination.Query{Limit: 10, Sort: "name:asc"})
	require.NoError(t, err)
	assert.Equal(t, "alpha", page.Items[0].Name)
	assert.Equal(t, "gamma", page.Items[len(page.Items)-1].Name)

	// date range
	from := time.Now().Add(-time.Hour)
	to := time.Now().Add(time.Hour)
	page, err = r.List(ctx, pagination.Query{Limit: 10, DateFrom: &from, DateTo: &to})
	require.NoError(t, err)
	assert.Equal(t, int64(5), page.Pagination.Total)
}

func TestListSummary(t *testing.T) {
	r, db := newTestRepo(t)
	seedWidgets(t, db)
	ctx := context.Background()

	page, err := r.List(ctx, pagination.Query{Limit: 2})
	require.NoError(t, err)
	require.NotNil(t, page.Summary)
	rows, ok := page.Summary.([]struct {
		Status string `gorm:"column:status"`
		Count  int64  `gorm:"column:count"`
	})
	assert.True(t, ok)
	assert.Len(t, rows, 2)
}

type PtrTimeWidget struct {
	model.Base
	Name    string     `gorm:"size:80" json:"name"`
	SentAt  *time.Time `json:"sent_at"`
}

func TestListCursorOnPointerTimeColumn(t *testing.T) {
	dbSeq++
	dsn := fmt.Sprintf("file:memdb%d?mode=memory&cache=shared&_loc=UTC", dbSeq)
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		NowFunc: func() time.Time { return time.Now().UTC() },
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&PtrTimeWidget{}))

	r := New(db, cache.NewMemory(), Options[PtrTimeWidget]{
		Slug:       "ptr_time_widget",
		Sortable:   map[string]string{"sent_at": "sent_at"},
		DateFields: []string{"created_at", "sent_at"},
	})
	ctx := context.Background()
	base := time.Now().UTC().Add(-time.Hour)
	for i := 0; i < 5; i++ {
		sent := base.Add(time.Duration(i) * time.Minute)
		require.NoError(t, db.Create(&PtrTimeWidget{Name: fmt.Sprintf("w%d", i), SentAt: &sent}).Error)
	}

	// paginate sorted by the pointer-time column
	page1, err := r.List(ctx, pagination.Query{Limit: 2, Sort: "sent_at:asc"})
	require.NoError(t, err)
	assert.Len(t, page1.Items, 2)
	assert.True(t, page1.Pagination.HasMore)

	page2, err := r.List(ctx, pagination.Query{Limit: 2, Sort: "sent_at:asc", Cursor: page1.Pagination.NextCursor})
	require.NoError(t, err)
	assert.Len(t, page2.Items, 2)
	assert.NotEqual(t, page1.Items[0].ID, page2.Items[0].ID)

	page3, err := r.List(ctx, pagination.Query{Limit: 2, Sort: "sent_at:asc", Cursor: page2.Pagination.NextCursor})
	require.NoError(t, err)
	assert.Len(t, page3.Items, 1)
	assert.False(t, page3.Pagination.HasMore)
}

func TestListInvalidCursor(t *testing.T) {
	r, _ := newTestRepo(t)
	_, err := r.List(context.Background(), pagination.Query{Cursor: "garbage!!"})
	assert.Error(t, err)
}
