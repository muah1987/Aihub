package compliance

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/muah1987/Aihub/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrRoleNotFound      = errors.New("custom role not found")
	ErrPermissionExists  = errors.New("permission already exists")
	ErrSystemRole        = errors.New("cannot modify system role")
)

// RoleService manages custom roles and resource-level permissions.
type RoleService struct {
	db *gorm.DB
}

func NewRoleService(db *gorm.DB) *RoleService {
	return &RoleService{db: db}
}

// ---- Custom Roles ----

type CreateRoleInput struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
}

func (s *RoleService) CreateRole(orgID uuid.UUID, userID *uuid.UUID, input *CreateRoleInput) (*models.CustomRole, error) {
	permsJSON, _ := json.Marshal(input.Permissions)

	role := &models.CustomRole{
		OrganizationID: orgID,
		Name:           input.Name,
		Description:    input.Description,
		Permissions:    permsJSON,
		CreatedBy:      userID,
	}

	result := s.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "organization_id"}, {Name: "name"}},
		DoUpdates: clause.AssignmentColumns([]string{"description", "permissions", "updated_at"}),
	}).Create(role)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to create role: %w", result.Error)
	}

	var fetched models.CustomRole
	s.db.Where("organization_id = ? AND name = ?", orgID, input.Name).First(&fetched)
	return &fetched, nil
}

func (s *RoleService) ListRoles(orgID uuid.UUID) ([]models.CustomRole, error) {
	var roles []models.CustomRole
	err := s.db.Where("organization_id = ?", orgID).Order("is_system DESC, name ASC").Find(&roles).Error
	return roles, err
}

func (s *RoleService) GetRole(orgID, roleID uuid.UUID) (*models.CustomRole, error) {
	var role models.CustomRole
	err := s.db.Where("id = ? AND organization_id = ?", roleID, orgID).First(&role).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRoleNotFound
		}
		return nil, err
	}
	return &role, nil
}

func (s *RoleService) UpdateRole(orgID, roleID uuid.UUID, input *CreateRoleInput) error {
	var role models.CustomRole
	if err := s.db.Where("id = ? AND organization_id = ?", roleID, orgID).First(&role).Error; err != nil {
		return ErrRoleNotFound
	}
	if role.IsSystem {
		return ErrSystemRole
	}

	permsJSON, _ := json.Marshal(input.Permissions)
	return s.db.Model(&role).Updates(map[string]interface{}{
		"name":        input.Name,
		"description": input.Description,
		"permissions": permsJSON,
	}).Error
}

func (s *RoleService) DeleteRole(orgID, roleID uuid.UUID) error {
	var role models.CustomRole
	if err := s.db.Where("id = ? AND organization_id = ?", roleID, orgID).First(&role).Error; err != nil {
		return ErrRoleNotFound
	}
	if role.IsSystem {
		return ErrSystemRole
	}
	return s.db.Delete(&role).Error
}

// ---- Resource Permissions ----

type GrantPermissionInput struct {
	UserID       uuid.UUID `json:"user_id"`
	ResourceType string    `json:"resource_type"`
	ResourceID   uuid.UUID `json:"resource_id"`
	Permission   string    `json:"permission"`
}

func (s *RoleService) GrantPermission(orgID uuid.UUID, grantedBy *uuid.UUID, input *GrantPermissionInput) (*models.ResourcePermission, error) {
	perm := &models.ResourcePermission{
		OrganizationID: orgID,
		UserID:         input.UserID,
		ResourceType:   input.ResourceType,
		ResourceID:     input.ResourceID,
		Permission:     input.Permission,
		GrantedBy:      grantedBy,
	}

	result := s.db.Clauses(clause.OnConflict{DoNothing: true}).Create(perm)
	if result.Error != nil {
		return nil, result.Error
	}
	return perm, nil
}

func (s *RoleService) RevokePermission(orgID, permID uuid.UUID) error {
	result := s.db.Where("id = ? AND organization_id = ?", permID, orgID).Delete(&models.ResourcePermission{})
	if result.RowsAffected == 0 {
		return errors.New("permission not found")
	}
	return result.Error
}

func (s *RoleService) ListPermissions(orgID uuid.UUID, userID *uuid.UUID, resourceType string) ([]models.ResourcePermission, error) {
	q := s.db.Where("organization_id = ?", orgID)
	if userID != nil {
		q = q.Where("user_id = ?", userID)
	}
	if resourceType != "" {
		q = q.Where("resource_type = ?", resourceType)
	}
	var perms []models.ResourcePermission
	err := q.Order("created_at DESC").Find(&perms).Error
	return perms, err
}

// HasPermission checks if a user has a specific permission on a resource.
func (s *RoleService) HasPermission(userID uuid.UUID, resourceType string, resourceID uuid.UUID, permission string) bool {
	var count int64
	s.db.Model(&models.ResourcePermission{}).
		Where("user_id = ? AND resource_type = ? AND resource_id = ? AND permission = ?",
			userID, resourceType, resourceID, permission).
		Count(&count)
	return count > 0
}
