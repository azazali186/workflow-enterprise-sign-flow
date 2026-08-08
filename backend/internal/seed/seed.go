// Package seed bootstraps the super admin role and admin user on first start.
package seed

import (
	"errors"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/aeroxe/sign-flow/backend/internal/models"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/logger"
	"go.uber.org/zap"
)

// Options configures bootstrap seeding.
type Options struct {
	AdminEmail    string
	AdminPassword string
}

// Run ensures the super admin role and bootstrap user exist.
func Run(db *gorm.DB, opts Options) error {
	email := strings.ToLower(strings.TrimSpace(opts.AdminEmail))
	if email == "" {
		return errors.New("admin email must not be empty")
	}
	var role models.Role
	err := db.Where("slug = ?", models.RoleSuperAdmin).First(&role).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		role = models.Role{
			Name: "Super Admin", Slug: models.RoleSuperAdmin,
			Description: "Full access to every API permission", IsSystem: true,
		}
		if err := db.Create(&role).Error; err != nil {
			return err
		}
		logger.L().Info("super admin role seeded")
	} else if err != nil {
		return err
	}

	var user models.User
	err = db.Where("email = ?", email).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		hash, err := bcrypt.GenerateFromPassword([]byte(opts.AdminPassword), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		user = models.User{
			Name: "Administrator", Email: email, PasswordHash: string(hash),
			Status: models.UserActive, Roles: []models.Role{role},
		}
		if err := db.Create(&user).Error; err != nil {
			return err
		}
		logger.L().Info("bootstrap admin seeded", zap.String("email", email))
	} else if err != nil {
		return err
	}
	return nil
}
