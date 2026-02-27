package rbac

import (
	"errors"

	"github.com/google/uuid"
	"github.com/muah1987/Aihub/internal/models"
	"gorm.io/gorm"
)

type Permission string

const (
	PermManageOrg      Permission = "manage_org"
	PermInviteMembers  Permission = "invite_members"
	PermManageProjects Permission = "manage_projects"
	PermManageAgents   Permission = "manage_agents"
	PermSendMessages   Permission = "send_messages"
	PermViewProject    Permission = "view_project"
	PermManageTerminal Permission = "manage_terminal"
	PermDeleteOrg      Permission = "delete_org"
)

var rolePermissions = map[string][]Permission{
	"owner":  {PermManageOrg, PermInviteMembers, PermManageProjects, PermManageAgents, PermSendMessages, PermViewProject, PermManageTerminal, PermDeleteOrg},
	"admin":  {PermManageOrg, PermInviteMembers, PermManageProjects, PermManageAgents, PermSendMessages, PermViewProject, PermManageTerminal},
	"member": {PermManageProjects, PermManageAgents, PermSendMessages, PermViewProject, PermManageTerminal},
	"viewer": {PermViewProject},
}

var (
	ErrAccessDenied = errors.New("access denied")
)

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

func (s *Service) HasPermission(userID, orgID uuid.UUID, perm Permission) (bool, error) {
	role, err := s.GetUserRole(userID, orgID)
	if err != nil {
		return false, err
	}

	perms, ok := rolePermissions[role]
	if !ok {
		return false, nil
	}

	for _, p := range perms {
		if p == perm {
			return true, nil
		}
	}
	return false, nil
}

func (s *Service) GetUserRole(userID, orgID uuid.UUID) (string, error) {
	var member models.OrganizationMember
	err := s.db.Where("user_id = ? AND organization_id = ?", userID, orgID).First(&member).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", ErrAccessDenied
		}
		return "", err
	}
	return member.Role, nil
}

func (s *Service) GetUserProjectAccess(userID, projectID uuid.UUID) (string, error) {
	var project models.Project
	if err := s.db.First(&project, "id = ?", projectID).Error; err != nil {
		return "", err
	}

	// Personal project: owner check
	if project.OrganizationID == nil {
		if project.UserID == userID {
			return "owner", nil
		}
		return "", ErrAccessDenied
	}

	// Org project: check membership
	return s.GetUserRole(userID, *project.OrganizationID)
}

func HasRole(role string, perm Permission) bool {
	perms, ok := rolePermissions[role]
	if !ok {
		return false
	}
	for _, p := range perms {
		if p == perm {
			return true
		}
	}
	return false
}
