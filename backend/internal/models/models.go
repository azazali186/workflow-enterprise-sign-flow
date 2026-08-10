// Package models defines the RBAC entities shared across the application.
package models

import (
	"time"

	"github.com/aeroxe/sign-flow/backend/internal/pkg/model"
)

// User statuses.
const (
	UserActive    = "active"
	UserSuspended = "suspended"
)

// Role slugs.
const (
	RoleSuperAdmin = "super_admin"
	RoleViewer     = "viewer"
)

// User is an authenticated actor.
type User struct {
	model.Base
	Name         string     `gorm:"size:120" json:"name"`
	Email        string     `gorm:"size:160;uniqueIndex" json:"email"`
	Phone        string     `gorm:"size:255;serializer:enc" json:"phone"`
	PasswordHash string     `gorm:"size:255" json:"-"`
	Status       string     `gorm:"size:20;index;default:active" json:"status"`
	LastLoginAt  *time.Time `json:"last_login_at"`
	Roles        []Role     `gorm:"many2many:user_roles" json:"roles,omitempty"`
}

// Role groups permissions.
type Role struct {
	model.Base
	Name        string       `gorm:"size:120" json:"name"`
	Slug        string       `gorm:"size:60;uniqueIndex" json:"slug"`
	Description string       `gorm:"size:255" json:"description"`
	IsSystem    bool         `gorm:"default:false" json:"is_system"`
	Permissions []Permission `gorm:"many2many:role_permissions" json:"permissions,omitempty"`
}

// Permission is one API capability, seeded from registered routes.
type Permission struct {
	model.Base
	Name    string `gorm:"size:160" json:"name"`
	Route   string `gorm:"size:200;uniqueIndex" json:"route"` // e.g. "POST /api/v1/contracts"
	Path    string `gorm:"size:200;index" json:"path"`
	Method  string `gorm:"size:10" json:"method"`
	Service string `gorm:"size:60;default:api-gateway" json:"service"`
	Roles   []Role `gorm:"many2many:role_permissions" json:"roles,omitempty"`
}

// UserWithRoles loads a user with roles and their permissions.
func (u *User) PermissionKeys() []string {
	seen := map[string]bool{}
	var keys []string
	for _, role := range u.Roles {
		for _, p := range role.Permissions {
			if !seen[p.Route] {
				seen[p.Route] = true
				keys = append(keys, p.Route)
			}
		}
	}
	return keys
}

// IsSuperAdmin reports whether the user holds the super admin role.
func (u *User) IsSuperAdmin() bool {
	for _, r := range u.Roles {
		if r.Slug == RoleSuperAdmin {
			return true
		}
	}
	return false
}
