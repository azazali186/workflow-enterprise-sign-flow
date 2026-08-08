// Package roles implements role administration with permission assignment.
// System roles are protected from mutation and deletion.
package roles

import (
	"context"

	"gorm.io/gorm"

	"github.com/aeroxe/sign-flow/backend/internal/audit"
	"github.com/aeroxe/sign-flow/backend/internal/cache"
	"github.com/aeroxe/sign-flow/backend/internal/models"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/errs"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/pagination"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/repo"
)

// Service is the role use-case boundary.
type Service interface {
	Create(ctx context.Context, req CreateRequest) (*models.Role, error)
	Patch(ctx context.Context, req PatchRequest) (*models.Role, error)
	Delete(ctx context.Context, req ByIDRequest) error
	Detail(ctx context.Context, req ByIDRequest) (*models.Role, error)
	List(ctx context.Context, q pagination.Query) (*repo.Page[models.Role], error)
	AssignPermissions(ctx context.Context, req AssignPermissionsRequest) (*models.Role, error)
}

type service struct {
	db    *gorm.DB
	repo  *repo.Repo[models.Role]
	audit *audit.Service
}

// NewService wires the role repository.
func NewService(db *gorm.DB, c cache.Cache, a *audit.Service) Service {
	return &service{
		db: db,
		repo: repo.New(db, c, repo.Options[models.Role]{
			Slug:       "role",
			Searchable: []string{"name", "slug"},
			Filterable: map[string]string{"slug": "slug"},
			Sortable:   map[string]string{"name": "name"},
			Preloads:   []string{"Permissions"},
			Summary: func(tx *gorm.DB) (any, error) {
				var out struct {
					Total int64 `json:"total_roles"`
					System int64 `json:"system_roles"`
				}
				tx.Model(&models.Role{}).Count(&out.Total)
				tx.Model(&models.Role{}).Where("is_system = ?", true).Count(&out.System)
				return out, nil
			},
		}),
		audit: a,
	}
}

// CreateRequest creates a role.
type CreateRequest struct {
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
}

// PatchRequest updates a role.
type PatchRequest struct {
	ID          string  `json:"id"`
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

// ByIDRequest targets a record via the body.
type ByIDRequest struct {
	ID string `json:"id"`
}

// AssignPermissionsRequest replaces a role's permissions.
type AssignPermissionsRequest struct {
	ID             string   `json:"id"`
	PermissionIDs  []string `json:"permission_ids"`
}

func (s *service) Create(ctx context.Context, req CreateRequest) (*models.Role, error) {
	if req.Name == "" || req.Slug == "" {
		return nil, errs.ErrValidation
	}
	role := models.Role{Name: req.Name, Slug: req.Slug, Description: req.Description}
	if err := s.repo.Create(ctx, &role); err != nil {
		return nil, err
	}
	s.audit.Record(ctx, "role.created", "role", role.ID, nil, map[string]any{"name": role.Name, "slug": role.Slug})
	return s.Detail(ctx, ByIDRequest{ID: role.ID})
}

func (s *service) Patch(ctx context.Context, req PatchRequest) (*models.Role, error) {
	before, err := s.Detail(ctx, ByIDRequest{ID: req.ID})
	if err != nil {
		return nil, err
	}
	if before.IsSystem {
		return nil, errs.New(400, 40010, "system roles cannot be modified")
	}
	updates := map[string]any{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	after, err := s.repo.Patch(ctx, req.ID, updates)
	if err != nil {
		return nil, err
	}
	s.audit.Record(ctx, "role.patched", "role", req.ID,
		map[string]any{"name": before.Name},
		map[string]any{"name": after.Name})
	return after, nil
}

func (s *service) Delete(ctx context.Context, req ByIDRequest) error {
	role, err := s.Detail(ctx, req)
	if err != nil {
		return err
	}
	if role.IsSystem {
		return errs.New(400, 40010, "system roles cannot be deleted")
	}
	if err := s.repo.Delete(ctx, req.ID); err != nil {
		return err
	}
	s.audit.Record(ctx, "role.deleted", "role", req.ID, nil, map[string]any{"deleted": true})
	return nil
}

func (s *service) Detail(ctx context.Context, req ByIDRequest) (*models.Role, error) {
	return s.repo.Get(ctx, req.ID)
}

func (s *service) List(ctx context.Context, q pagination.Query) (*repo.Page[models.Role], error) {
	return s.repo.List(ctx, q)
}

func (s *service) AssignPermissions(ctx context.Context, req AssignPermissionsRequest) (*models.Role, error) {
	role, err := s.Detail(ctx, ByIDRequest{ID: req.ID})
	if err != nil {
		return nil, err
	}
	if role.IsSystem {
		return nil, errs.New(400, 40010, "system roles cannot be modified")
	}
	var perms []models.Permission
	if len(req.PermissionIDs) > 0 {
		if err := s.db.WithContext(ctx).Find(&perms, req.PermissionIDs).Error; err != nil {
			return nil, err
		}
	}
	if err := s.db.WithContext(ctx).Model(role).Association("Permissions").Replace(perms); err != nil {
		return nil, err
	}
	s.audit.Record(ctx, "role.permissions_assigned", "role", req.ID, nil, map[string]any{"permission_ids": req.PermissionIDs})
	return s.Detail(ctx, ByIDRequest{ID: req.ID})
}

// Handler exposes HTTP routes. Implemented in handler.go.
