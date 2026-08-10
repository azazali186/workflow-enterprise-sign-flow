package registry

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/aeroxe/sign-flow/backend/internal/cache"
	"github.com/aeroxe/sign-flow/backend/internal/middleware"
	"github.com/aeroxe/sign-flow/backend/internal/models"
)

var regDBSeq int

func newRegDB(t *testing.T) *gorm.DB {
	t.Helper()
	regDBSeq++
	dsn := fmt.Sprintf("file:regmem%d?mode=memory&cache=shared&_loc=UTC", regDBSeq)
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.Permission{}))
	return db
}

// chdirToTemp isolates the routes.json side-effect of SeedPermissions.
// NOTE: this package must never use t.Parallel() — os.Chdir is process-wide.
func chdirToTemp(t *testing.T) {
	t.Helper()
	old, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(t.TempDir()))
	t.Cleanup(func() { _ = os.Chdir(old) })
}

func TestSeedPermissionsSkipsPublicAndSeedsAPI(t *testing.T) {
	db := newRegDB(t)
	chdirToTemp(t)
	reg := New()
	reg.Register("POST", "/api/v1/contracts", "Create Contract", middleware.GuardAPI)
	reg.Register("POST", "/api/v1/auth/login", "Login", middleware.GuardPublic)
	reg.Register("GET", "/metrics", "Metrics", middleware.GuardPublic)

	require.NoError(t, reg.SeedPermissions(db, cache.NewMemory()))

	var perms []models.Permission
	require.NoError(t, db.Find(&perms).Error)
	require.Len(t, perms, 1)
	assert.Equal(t, "POST /api/v1/contracts", perms[0].Route)
	assert.Equal(t, "api-gateway", perms[0].Service)
	assert.Equal(t, "POST", perms[0].Method)
}

func TestSeedPermissionsIsIdempotentAndUpserts(t *testing.T) {
	db := newRegDB(t)
	chdirToTemp(t)
	reg := New()
	reg.Register("POST", "/api/v1/contracts", "Create Contract", middleware.GuardAPI)

	require.NoError(t, reg.SeedPermissions(db, cache.NewMemory()))
	require.NoError(t, reg.SeedPermissions(db, cache.NewMemory()))

	var count int64
	db.Model(&models.Permission{}).Count(&count)
	assert.Equal(t, int64(1), count)

	// A renamed route updates the existing permission instead of duplicating.
	reg2 := New()
	reg2.Register("POST", "/api/v1/contracts", "Create New Contract", middleware.GuardAPI)
	require.NoError(t, reg2.SeedPermissions(db, cache.NewMemory()))

	var perm models.Permission
	require.NoError(t, db.First(&perm).Error)
	assert.Equal(t, "Create New Contract", perm.Name)
}

func TestGuardsTable(t *testing.T) {
	reg := New()
	reg.Register("POST", "/api/v1/x", "X", middleware.GuardAPI)
	reg.Register("GET", "/metrics", "Metrics", middleware.GuardPublic)

	guards := reg.Guards()
	require.Len(t, guards, 2)
	assert.Equal(t, "POST /api/v1/x", guards[0].Key)
	assert.Equal(t, middleware.GuardPublic, guards[1].Guard)
}

func TestFormatRouteName(t *testing.T) {
	assert.Equal(t, "Contracts List", FormatRouteName("/api/v1/contracts/list"))
	assert.Equal(t, "Health", FormatRouteName("/api/v1/health"))
	assert.Equal(t, "", FormatRouteName("/api/v1"))
}

func TestSeedPermissionsStoresRouteTableInCache(t *testing.T) {
	db := newRegDB(t)
	chdirToTemp(t)
	c := cache.NewMemory()
	reg := New()
	reg.Register("POST", "/api/v1/contracts", "Create Contract", middleware.GuardAPI)

	require.NoError(t, reg.SeedPermissions(db, c))
	raw, err := c.Get(context.Background(), redisRouteKey)
	require.NoError(t, err)
	assert.Contains(t, string(raw), "/api/v1/contracts")
}
