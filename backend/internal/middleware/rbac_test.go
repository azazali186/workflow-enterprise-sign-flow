package middleware

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
	"github.com/aeroxe/sign-flow/backend/internal/models"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/cryptoser"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/metrics"
	"gorm.io/gorm/schema"
)

var rbacDBSeq int

func newRBACDB(t *testing.T) *gorm.DB {
	t.Helper()
	rbacDBSeq++
	dsn := fmt.Sprintf("file:rbacmem%d?mode=memory&cache=shared&_loc=UTC", rbacDBSeq)
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	schema.RegisterSerializer(cryptoser.Name, cryptoser.Enc{})
	require.NoError(t, db.AutoMigrate(&models.User{}, &models.Role{}, &models.Permission{}))
	return db
}

func seedRBAC(t *testing.T, db *gorm.DB, super bool, route string) models.User {
	t.Helper()
	role := models.Role{Name: "Role", Slug: "role", IsSystem: false}
	if super {
		role = models.Role{Name: "Super", Slug: models.RoleSuperAdmin, IsSystem: true}
	}
	require.NoError(t, db.Create(&role).Error)
	if !super {
		perm := models.Permission{Name: "P", Route: route, Path: "/api/v1/x", Method: "POST", Service: "api-gateway"}
		require.NoError(t, db.Create(&perm).Error)
		require.NoError(t, db.Model(&role).Association("Permissions").Append(&perm))
	}
	user := models.User{Name: "T", Email: fmt.Sprintf("u%d@example.com", rbacDBSeq), Status: models.UserActive, Roles: []models.Role{role}}
	require.NoError(t, db.Create(&user).Error)
	return user
}

func TestUserAllows(t *testing.T) {
	assert.True(t, userAllows(`["*"]`, "POST /api/v1/contracts"))
	assert.True(t, userAllows(`["POST /api/v1/contracts"]`, "POST /api/v1/contracts"))
	assert.False(t, userAllows(`["POST /api/v1/contracts"]`, "DELETE /api/v1/contracts"))
	assert.False(t, userAllows(`garbage`, "POST /api/v1/contracts"))
}

func TestCheckSuperAdminAllowsEverything(t *testing.T) {
	db := newRBACDB(t)
	c := cache.NewMemory()
	user := seedRBAC(t, db, true, "")
	rbac := NewRBAC(db, c, map[string]string{}, metrics.New())

	ok, err := rbac.check(context.Background(), user.ID, "DELETE /api/v1/users")
	require.NoError(t, err)
	assert.True(t, ok)

	raw, err := c.Get(context.Background(), RBACPrefix+user.ID)
	require.NoError(t, err)
	assert.Equal(t, `["*"]`, raw)
}

func TestCheckRegularUserHonoursPermissions(t *testing.T) {
	db := newRBACDB(t)
	c := cache.NewMemory()
	user := seedRBAC(t, db, false, "POST /api/v1/contracts")
	rbac := NewRBAC(db, c, map[string]string{}, metrics.New())
	ctx := context.Background()

	ok, err := rbac.check(ctx, user.ID, "POST /api/v1/contracts")
	require.NoError(t, err)
	assert.True(t, ok)

	ok, err = rbac.check(ctx, user.ID, "PATCH /api/v1/contracts")
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestInvalidateUserRBACRefreshesGrants(t *testing.T) {
	db := newRBACDB(t)
	c := cache.NewMemory()
	user := seedRBAC(t, db, false, "POST /api/v1/contracts")
	rbac := NewRBAC(db, c, map[string]string{}, metrics.New())
	ctx := context.Background()

	ok, err := rbac.check(ctx, user.ID, "POST /api/v1/contracts")
	require.NoError(t, err)
	assert.True(t, ok)

	// Revoke the permission in the DB; the stale cache still allows it.
	var role models.Role
	require.NoError(t, db.Where("slug = ?", "role").First(&role).Error)
	require.NoError(t, db.Model(&role).Association("Permissions").Clear())
	ok, _ = rbac.check(ctx, user.ID, "POST /api/v1/contracts")
	assert.True(t, ok, "cached grants must still apply until invalidated")

	// After invalidation the next check reflects the database.
	require.NoError(t, InvalidateUserRBAC(ctx, c, user.ID))
	ok, err = rbac.check(ctx, user.ID, "POST /api/v1/contracts")
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestInvalidateAllRBACClearsEveryUser(t *testing.T) {
	c := cache.NewMemory()
	ctx := context.Background()
	require.NoError(t, c.Set(ctx, RBACPrefix+"u1", "[]", time.Minute))
	require.NoError(t, c.Set(ctx, RBACPrefix+"u2", "[]", time.Minute))
	require.NoError(t, c.Set(ctx, "auth:token:u1", "x", time.Minute))

	require.NoError(t, InvalidateAllRBAC(ctx, c))
	_, err := c.Get(ctx, RBACPrefix+"u1")
	assert.Error(t, err)
	_, err = c.Get(ctx, RBACPrefix+"u2")
	assert.Error(t, err)
	// Unrelated keys survive.
	_, err = c.Get(ctx, "auth:token:u1")
	require.NoError(t, err)
}
