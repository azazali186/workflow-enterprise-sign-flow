package roles

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
	"github.com/aeroxe/sign-flow/backend/internal/pkg/errs"
)

var rolesDBSeq int

func newRolesEnv(t *testing.T) (*gorm.DB, cache.Cache, Service) {
	t.Helper()
	rolesDBSeq++
	dsn := fmt.Sprintf("file:rolesmem%d?mode=memory&cache=shared&_loc=UTC", rolesDBSeq)
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.Role{}, &models.Permission{}, &audit.Log{}))
	c := cache.NewMemory()
	return db, c, NewService(db, c, audit.New(db))
}

func TestCreateRoleAndAssignPermissionsPurgesAllRBAC(t *testing.T) {
	db, c, svc := newRolesEnv(t)
	ctx := context.Background()

	perm := models.Permission{Name: "Create Contract", Route: "POST /api/v1/contracts", Path: "/api/v1/contracts", Method: "POST", Service: "api-gateway"}
	require.NoError(t, db.Create(&perm).Error)

	role, err := svc.Create(ctx, CreateRequest{Name: "Editor", Slug: "editor", Description: "Edits contracts"})
	require.NoError(t, err)
	assert.NotEmpty(t, role.ID)

	// Seed cached grants for two users, then reassign permissions.
	require.NoError(t, c.Set(ctx, middleware.RBACPrefix+"u1", "[]", time.Minute))
	require.NoError(t, c.Set(ctx, middleware.RBACPrefix+"u2", "[]", time.Minute))
	_, err = svc.AssignPermissions(ctx, AssignPermissionsRequest{ID: role.ID, PermissionIDs: []string{perm.ID}})
	require.NoError(t, err)

	_, err = c.Get(ctx, middleware.RBACPrefix+"u1")
	assert.Error(t, err, "all cached grants must be purged after role permission changes")
	_, err = c.Get(ctx, middleware.RBACPrefix+"u2")
	assert.Error(t, err)

	got, err := svc.Detail(ctx, ByIDRequest{ID: role.ID})
	require.NoError(t, err)
	require.Len(t, got.Permissions, 1)
	assert.Equal(t, "POST /api/v1/contracts", got.Permissions[0].Route)
}

func TestSystemRoleIsProtected(t *testing.T) {
	db, _, svc := newRolesEnv(t)
	ctx := context.Background()

	sys := models.Role{Name: "Super Admin", Slug: models.RoleSuperAdmin, IsSystem: true}
	require.NoError(t, db.Create(&sys).Error)

	_, err := svc.Patch(ctx, PatchRequest{ID: sys.ID, Name: ptr("Hacked")})
	assert.Error(t, err)

	err = svc.Delete(ctx, ByIDRequest{ID: sys.ID})
	assert.Error(t, err)

	var count int64
	db.Model(&models.Role{}).Count(&count)
	assert.Equal(t, int64(1), count)
}

func TestDeleteRole(t *testing.T) {
	_, _, svc := newRolesEnv(t)
	ctx := context.Background()

	role, err := svc.Create(ctx, CreateRequest{Name: "Temp", Slug: "temp"})
	require.NoError(t, err)

	require.NoError(t, svc.Delete(ctx, ByIDRequest{ID: role.ID}))
	_, err = svc.Detail(ctx, ByIDRequest{ID: role.ID})
	assert.ErrorIs(t, err, errs.ErrNotFound)
}

func ptr(s string) *string { return &s }
