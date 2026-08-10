package auth

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
	"github.com/aeroxe/sign-flow/backend/internal/models"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/cryptoser"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/ctxval"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/jwt"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm/schema"
)

var authDBSeq int

func setup(t *testing.T) (*gorm.DB, cache.Cache, *jwt.Manager, *audit.Service) {
	t.Helper()
	schema.RegisterSerializer(cryptoser.Name, cryptoser.Enc{})
	authDBSeq++
	dsn := fmt.Sprintf("file:authmem%d?mode=memory&cache=shared&_loc=UTC", authDBSeq)
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.User{}, &models.Role{}, &audit.LoginLog{}, &audit.Log{}))
	mgr := jwt.New("test-secret-key-32-bytes-long!!", 2*time.Hour)
	return db, cache.NewMemory(), mgr, audit.New(db)
}

func seedUser(t *testing.T, db *gorm.DB, email, password string) models.User {
	t.Helper()
	user := models.User{Name: "Test", Email: email, PasswordHash: hashPassword(t, password), Status: models.UserActive}
	require.NoError(t, db.Create(&user).Error)
	return user
}

func hashPassword(t *testing.T, password string) string {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.NoError(t, err)
	return string(hash)
}

func TestLoginSuccessAndSession(t *testing.T) {
	db, c, mgr, auditSvc := setup(t)
	seedUser(t, db, "user@example.com", "secret123")
	svc := NewService(db, c, mgr, auditSvc, 2*time.Hour)

	ctx := context.Background()
	res, err := svc.Login(ctx, LoginRequest{Email: "user@example.com", Password: "secret123"})
	require.NoError(t, err)
	assert.NotEmpty(t, res.AccessToken)
	assert.Equal(t, "bearer", res.TokenType)

	// token round-trips through the manager
	claims, err := mgr.Parse(res.AccessToken)
	require.NoError(t, err)
	assert.Equal(t, res.User.ID, claims.UserID)

	// session stored
	stored, err := c.Get(ctx, "auth:token:"+res.User.ID)
	require.NoError(t, err)
	assert.Equal(t, md5Hex(res.AccessToken+res.User.ID), stored)
}

func TestLoginWrongPassword(t *testing.T) {
	db, _, mgr, auditSvc := setup(t)
	seedUser(t, db, "user@example.com", "secret123")
	svc := NewService(db, cache.NewMemory(), mgr, auditSvc, 2*time.Hour)

	_, err := svc.Login(context.Background(), LoginRequest{Email: "user@example.com", Password: "wrong"})
	assert.Error(t, err)
}

func TestLoginSuspended(t *testing.T) {
	db, _, mgr, auditSvc := setup(t)
	user := models.User{Name: "X", Email: "suspended@example.com", PasswordHash: hashPassword(t, "secret123"), Status: models.UserSuspended}
	require.NoError(t, db.Create(&user).Error)
	svc := NewService(db, cache.NewMemory(), mgr, auditSvc, 2*time.Hour)

	_, err := svc.Login(context.Background(), LoginRequest{Email: "suspended@example.com", Password: "secret123"})
	assert.Error(t, err)
}

func TestLogoutInvalidatesSession(t *testing.T) {
	db, c, mgr, auditSvc := setup(t)
	user := seedUser(t, db, "logout@example.com", "secret123")
	svc := NewService(db, c, mgr, auditSvc, 2*time.Hour)

	_, err := svc.Login(context.Background(), LoginRequest{Email: "logout@example.com", Password: "secret123"})
	require.NoError(t, err)

	ctx := ctxval.SetUserID(context.Background(), user.ID)
	require.NoError(t, svc.Logout(ctx))
	_, err = c.Get(ctx, "auth:token:"+user.ID)
	assert.Error(t, err)
}

func TestMe(t *testing.T) {
	db, c, mgr, auditSvc := setup(t)
	user := seedUser(t, db, "me@example.com", "secret123")
	svc := NewService(db, c, mgr, auditSvc, 2*time.Hour)

	got, err := svc.Me(ctxval.SetUserID(context.Background(), user.ID))
	require.NoError(t, err)
	assert.Equal(t, user.ID, got.ID)

	_, err = svc.Me(context.Background())
	assert.Error(t, err)
}
