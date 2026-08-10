package users

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/aeroxe/sign-flow/backend/internal/audit"
	"github.com/aeroxe/sign-flow/backend/internal/cache"
	"github.com/aeroxe/sign-flow/backend/internal/middleware"
	"github.com/aeroxe/sign-flow/backend/internal/models"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/cryptoser"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/errs"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/pagination"
	"gorm.io/gorm/schema"
)

var usersDBSeq int

func newUsersEnv(t *testing.T) (*gorm.DB, cache.Cache, Service) {
	t.Helper()
	usersDBSeq++
	dsn := fmt.Sprintf("file:usersmem%d?mode=memory&cache=shared&_loc=UTC", usersDBSeq)
	schema.RegisterSerializer(cryptoser.Name, cryptoser.Enc{})
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.User{}, &models.Role{}, &audit.Log{}))
	// A role the tests can assign. The service must share the same cache
	// instance the test inspects for RBAC invalidation.
	role := models.Role{Name: "Manager", Slug: "manager"}
	require.NoError(t, db.Create(&role).Error)
	c := cache.NewMemory()
	return db, c, NewService(db, c, audit.New(db))
}

func TestCreateUserAndDuplicateConflict(t *testing.T) {
	db, _, svc := newUsersEnv(t)
	ctx := context.Background()

	user, err := svc.Create(ctx, CreateRequest{Name: "Alice", Email: "alice@example.com", Password: "secret123"})
	require.NoError(t, err)
	assert.NotEmpty(t, user.ID)
	assert.Equal(t, models.UserActive, user.Status)

	_, err = svc.Create(ctx, CreateRequest{Name: "Alice2", Email: "alice@example.com", Password: "secret123"})
	assert.ErrorIs(t, err, errs.ErrConflict)

	var count int64
	db.Model(&models.User{}).Count(&count)
	assert.Equal(t, int64(1), count)
}

func TestAssignRolesInvalidatesRBACCache(t *testing.T) {
	db, c, svc := newUsersEnv(t)
	ctx := context.Background()

	user, err := svc.Create(ctx, CreateRequest{Name: "Bob", Email: "bob@example.com", Password: "secret123"})
	require.NoError(t, err)

	var role models.Role
	require.NoError(t, db.Where("slug = ?", "manager").First(&role).Error)
	require.NoError(t, c.Set(ctx, middleware.RBACPrefix+user.ID, `["POST /api/v1/contracts"]`, time.Minute))

	_, err = svc.AssignRoles(ctx, AssignRolesRequest{ID: user.ID, RoleIDs: []string{role.ID}})
	require.NoError(t, err)

	_, err = c.Get(ctx, middleware.RBACPrefix+user.ID)
	assert.Error(t, err, "the user's cached grants must be purged when roles change")

	got, err := svc.Detail(ctx, ByIDRequest{ID: user.ID})
	require.NoError(t, err)
	require.Len(t, got.Roles, 1)
	assert.Equal(t, "manager", got.Roles[0].Slug)
}

func TestListPaginatesWithSummary(t *testing.T) {
	_, _, svc := newUsersEnv(t)
	ctx := context.Background()
	for i := 0; i < 3; i++ {
		_, err := svc.Create(ctx, CreateRequest{Name: fmt.Sprintf("U%d", i), Email: fmt.Sprintf("u%d@example.com", i), Password: "secret123"})
		require.NoError(t, err)
	}

	page, err := svc.List(ctx, pagination.Query{Limit: 2, Sort: "-created_at"})
	require.NoError(t, err)
	require.Len(t, page.Items, 2)
	assert.Equal(t, int64(3), page.Pagination.Total)
	assert.True(t, page.Pagination.HasMore)
	assert.NotEmpty(t, page.Pagination.NextCursor)
	assert.NotNil(t, page.Summary, "list responses must include the db summary")
}

func TestDeleteRemovesUser(t *testing.T) {
	_, _, svc := newUsersEnv(t)
	ctx := context.Background()

	user, err := svc.Create(ctx, CreateRequest{Name: "Del", Email: "del@example.com", Password: "secret123"})
	require.NoError(t, err)

	require.NoError(t, svc.Delete(ctx, ByIDRequest{ID: user.ID}))
	_, err = svc.Detail(ctx, ByIDRequest{ID: user.ID})
	assert.ErrorIs(t, err, errs.ErrNotFound)
}
