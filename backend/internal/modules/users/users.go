// Package users implements user administration: create, patch, soft delete,
// detail, cursor-paginated list and role assignment.
package users

import (
	"context"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/aeroxe/sign-flow/backend/internal/audit"
	"github.com/aeroxe/sign-flow/backend/internal/cache"
	"github.com/aeroxe/sign-flow/backend/internal/middleware"
	"github.com/aeroxe/sign-flow/backend/internal/models"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/errs"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/pagination"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/repo"
)

// Service is the user use-case boundary.
type Service interface {
	Create(ctx context.Context, req CreateRequest) (*models.User, error)
	Patch(ctx context.Context, req PatchRequest) (*models.User, error)
	Delete(ctx context.Context, req ByIDRequest) error
	Detail(ctx context.Context, req ByIDRequest) (*models.User, error)
	List(ctx context.Context, q pagination.Query) (*repo.Page[models.User], error)
	AssignRoles(ctx context.Context, req AssignRolesRequest) (*models.User, error)
}

type service struct {
	db    *gorm.DB
	cache cache.Cache
	repo  *repo.Repo[models.User]
	audit *audit.Service
}

// NewService wires the user repository with audit logging.
func NewService(db *gorm.DB, c cache.Cache, a *audit.Service) Service {
	return &service{
		db: db,
		repo: repo.New(db, c, repo.Options[models.User]{
			Slug:       "user",
			Searchable: []string{"name", "email"},
			Filterable: map[string]string{"status": "status"},
			Sortable:   map[string]string{"name": "name", "email": "email"},
			Preloads:   []string{"Roles"},
			Summary: func(tx *gorm.DB) (any, error) {
				var rows []struct {
					Status string `gorm:"column:status" json:"status"`
					Count  int64  `gorm:"column:count" json:"count"`
				}
				err := tx.Model(&models.User{}).Select("status, count(*) as count").Group("status").Scan(&rows).Error
				return rows, err
			},
		}),
		audit: a,
		cache: c,
	}
}

// CreateRequest creates a user.
type CreateRequest struct {
	Name     string   `json:"name"`
	Email    string   `json:"email"`
	Password string   `json:"password"`
	Phone    string   `json:"phone"`
	Status   string   `json:"status"`
	RoleIDs  []string `json:"role_ids"`
}

// PatchRequest updates a user (only provided fields change).
type PatchRequest struct {
	ID     string  `json:"id"`
	Name   *string `json:"name"`
	Phone  *string `json:"phone"`
	Status *string `json:"status"`
}

// ByIDRequest targets a single record via the body (no path variables).
type ByIDRequest struct {
	ID string `json:"id"`
}

// AssignRolesRequest replaces a user's roles.
type AssignRolesRequest struct {
	ID      string   `json:"id"`
	RoleIDs []string `json:"role_ids"`
}

func (s *service) Create(ctx context.Context, req CreateRequest) (*models.User, error) {
	if req.Email == "" || req.Password == "" {
		return nil, errs.ErrValidation
	}
	var count int64
	s.db.WithContext(ctx).Model(&models.User{}).Where("email = ?", req.Email).Count(&count)
	if count > 0 {
		return nil, errs.ErrConflict
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	status := req.Status
	if status == "" {
		status = models.UserActive
	}
	var roles []models.Role
	if len(req.RoleIDs) > 0 {
		if err := s.db.WithContext(ctx).Find(&roles, req.RoleIDs).Error; err != nil {
			return nil, err
		}
	}
	user := models.User{
		Name: req.Name, Email: req.Email, Phone: req.Phone,
		PasswordHash: string(hash), Status: status, Roles: roles,
	}
	if err := s.repo.Create(ctx, &user); err != nil {
		return nil, err
	}
	s.audit.Record(ctx, "user.created", "user", user.ID, nil, map[string]any{
		"name": user.Name, "email": user.Email, "status": user.Status,
	})
	return s.Detail(ctx, ByIDRequest{ID: user.ID})
}

func (s *service) Patch(ctx context.Context, req PatchRequest) (*models.User, error) {
	before, err := s.Detail(ctx, ByIDRequest{ID: req.ID})
	if err != nil {
		return nil, err
	}
	updates := map[string]any{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Phone != nil {
		updates["phone"] = *req.Phone
	}
	if req.Status != nil {
		updates["status"] = *req.Status
		if *req.Status != models.UserActive {
			// Suspended/deactivated users lose access immediately.
			_ = middleware.RevokeSession(ctx, s.cache, req.ID)
		}
	}
	after, err := s.repo.Patch(ctx, req.ID, updates)
	if err != nil {
		return nil, err
	}
	s.audit.Record(ctx, "user.patched", "user", req.ID,
		map[string]any{"name": before.Name, "status": before.Status},
		map[string]any{"name": after.Name, "status": after.Status})
	return after, nil
}

func (s *service) Delete(ctx context.Context, req ByIDRequest) error {
	if err := s.repo.Delete(ctx, req.ID); err != nil {
		return err
	}
	// A deleted user must lose access immediately: drop session and grants.
	_ = middleware.RevokeSession(ctx, s.cache, req.ID)
	s.audit.Record(ctx, "user.deleted", "user", req.ID, nil, map[string]any{"deleted": true})
	return nil
}

func (s *service) Detail(ctx context.Context, req ByIDRequest) (*models.User, error) {
	return s.repo.Get(ctx, req.ID)
}

func (s *service) List(ctx context.Context, q pagination.Query) (*repo.Page[models.User], error) {
	return s.repo.List(ctx, q)
}

func (s *service) AssignRoles(ctx context.Context, req AssignRolesRequest) (*models.User, error) {
	user, err := s.Detail(ctx, ByIDRequest{ID: req.ID})
	if err != nil {
		return nil, err
	}
	var roles []models.Role
	if len(req.RoleIDs) > 0 {
		if err := s.db.WithContext(ctx).Find(&roles, req.RoleIDs).Error; err != nil {
			return nil, err
		}
	}
	if err := s.db.WithContext(ctx).Model(user).Association("Roles").Replace(roles); err != nil {
		return nil, err
	}
	// Drop the user's cached grants so the new roles take effect immediately.
	_ = middleware.InvalidateUserRBAC(ctx, s.cache, req.ID)
	s.audit.Record(ctx, "user.roles_assigned", "user", req.ID, nil, map[string]any{"role_ids": req.RoleIDs})
	return s.Detail(ctx, ByIDRequest{ID: req.ID})
}

// Handler exposes HTTP routes. Implemented in handler.go.
