package seed

import (
	"fmt"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/aeroxe/sign-flow/backend/internal/models"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/cryptoser"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm/schema"
)

var seedDBSeq int

func newSeedDB(t *testing.T) *gorm.DB {
	t.Helper()
	seedDBSeq++
	dsn := fmt.Sprintf("file:seedmem%d?mode=memory&cache=shared&_loc=UTC", seedDBSeq)
	schema.RegisterSerializer(cryptoser.Name, cryptoser.Enc{})
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.User{}, &models.Role{}))
	return db
}

func TestRunSeedsAdminAndRole(t *testing.T) {
	db := newSeedDB(t)
	opts := Options{AdminEmail: "admin@signflow.local", AdminPassword: "Str0ngPass!123"}

	require.NoError(t, Run(db, opts))

	var role models.Role
	require.NoError(t, db.Where("slug = ?", models.RoleSuperAdmin).First(&role).Error)
	assert.True(t, role.IsSystem)

	var user models.User
	require.NoError(t, db.Where("email = ?", "admin@signflow.local").First(&user).Error)
	assert.Equal(t, models.UserActive, user.Status)
	require.NoError(t, bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte("Str0ngPass!123")))

	// The many2many link between the bootstrap user and super_admin role
	// must be persisted at create time.
	var joins int64
	db.Table("user_roles").Count(&joins)
	assert.Equal(t, int64(1), joins)
}

func TestRunIsIdempotent(t *testing.T) {
	db := newSeedDB(t)
	opts := Options{AdminEmail: "admin@signflow.local", AdminPassword: "Str0ngPass!123"}

	require.NoError(t, Run(db, opts))
	require.NoError(t, Run(db, opts))
	require.NoError(t, Run(db, opts))

	var users int64
	var roles int64
	db.Model(&models.User{}).Count(&users)
	db.Model(&models.Role{}).Count(&roles)
	assert.Equal(t, int64(1), users)
	assert.Equal(t, int64(1), roles)
}

func TestRunRejectsEmptyEmail(t *testing.T) {
	db := newSeedDB(t)
	err := Run(db, Options{AdminEmail: "  ", AdminPassword: "Str0ngPass!123"})
	require.Error(t, err)
}

func TestRunRejectsWeakPassword(t *testing.T) {
	db := newSeedDB(t)
	opts := Options{AdminEmail: "admin@signflow.local", AdminPassword: "short"}
	err := Run(db, opts)
	require.Error(t, err)

	// No user or role may be created when seeding fails validation.
	var n int64
	db.Model(&models.User{}).Count(&n)
	assert.Zero(t, n)
	db.Model(&models.Role{}).Count(&n)
	assert.Zero(t, n)
}

func TestRunRejectsWeakPasswordMissingSymbol(t *testing.T) {
	db := newSeedDB(t)
	err := Run(db, Options{AdminEmail: "admin@signflow.local", AdminPassword: "abcdefghijklmnop1234"})
	require.Error(t, err)
}
