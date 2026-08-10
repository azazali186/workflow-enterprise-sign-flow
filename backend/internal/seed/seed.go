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

// bcryptCost is intentionally above the default (10) for the bootstrap admin.
const bcryptCost = 12

// Run ensures the super admin role and bootstrap user exist.
func Run(db *gorm.DB, opts Options) error {
	email := strings.ToLower(strings.TrimSpace(opts.AdminEmail))
	if email == "" {
		return errors.New("admin email must not be empty")
	}
	if err := validatePassword(opts.AdminPassword); err != nil {
		return err
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
		hash, err := bcrypt.GenerateFromPassword([]byte(opts.AdminPassword), bcryptCost)
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

// validatePassword rejects weak bootstrap admin passwords before seeding.
// 12+ chars, at least one letter, one digit and one symbol.
func validatePassword(pw string) error {
	if pw == "" {
		return errors.New("admin password must not be empty")
	}
	if len(pw) < 12 {
		return errors.New("admin password must be at least 12 characters")
	}
	var letter, digit, symbol bool
	for _, r := range pw {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z':
			letter = true
		case r >= '0' && r <= '9':
			digit = true
		default:
			symbol = true
		}
	}
	if !letter || !digit || !symbol {
		return errors.New("admin password must contain letters, digits and symbols")
	}
	return nil
}
