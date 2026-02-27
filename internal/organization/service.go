package organization

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/muah1987/Aihub/internal/email"
	"github.com/muah1987/Aihub/internal/models"
	"gorm.io/gorm"
)

var (
	ErrOrgNotFound      = errors.New("organization not found")
	ErrSlugTaken        = errors.New("slug already taken")
	ErrNotOwner         = errors.New("only the owner can perform this action")
	ErrAlreadyMember    = errors.New("user is already a member")
	ErrMemberNotFound   = errors.New("member not found")
	ErrInviteNotFound   = errors.New("invitation not found")
	ErrInviteExpired    = errors.New("invitation expired")
	ErrCannotRemoveSelf = errors.New("cannot remove yourself")
)

type CreateInput struct {
	Name        string `json:"name" validate:"required"`
	Slug        string `json:"slug" validate:"required"`
	Description string `json:"description"`
}

type UpdateInput struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

type InviteInput struct {
	Email string `json:"email" validate:"required"`
	Role  string `json:"role" validate:"required"`
}

type Service struct {
	db           *gorm.DB
	emailService *email.Service
}

func NewService(db *gorm.DB, emailService *email.Service) *Service {
	return &Service{db: db, emailService: emailService}
}

func (s *Service) Create(ownerID uuid.UUID, input *CreateInput) (*models.Organization, error) {
	slug := slugify(input.Slug)

	var existing models.Organization
	if s.db.Where("slug = ?", slug).First(&existing).Error == nil {
		return nil, ErrSlugTaken
	}

	org := &models.Organization{
		Name:        input.Name,
		Slug:        slug,
		Description: input.Description,
		OwnerID:     ownerID,
	}

	tx := s.db.Begin()

	if err := tx.Create(org).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create organization: %w", err)
	}

	member := &models.OrganizationMember{
		OrganizationID: org.ID,
		UserID:         ownerID,
		Role:           "owner",
	}

	if err := tx.Create(member).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to add owner as member: %w", err)
	}

	return org, tx.Commit().Error
}

func (s *Service) Get(orgID uuid.UUID) (*models.Organization, error) {
	var org models.Organization
	if err := s.db.First(&org, "id = ?", orgID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrgNotFound
		}
		return nil, err
	}
	return &org, nil
}

func (s *Service) List(userID uuid.UUID) ([]models.Organization, error) {
	var orgs []models.Organization
	err := s.db.Joins("JOIN organization_members ON organization_members.organization_id = organizations.id").
		Where("organization_members.user_id = ?", userID).
		Find(&orgs).Error
	return orgs, err
}

func (s *Service) Update(orgID uuid.UUID, input *UpdateInput) (*models.Organization, error) {
	org, err := s.Get(orgID)
	if err != nil {
		return nil, err
	}

	updates := map[string]interface{}{}
	if input.Name != nil {
		updates["name"] = *input.Name
	}
	if input.Description != nil {
		updates["description"] = *input.Description
	}

	if len(updates) > 0 {
		if err := s.db.Model(org).Updates(updates).Error; err != nil {
			return nil, err
		}
	}

	return s.Get(orgID)
}

func (s *Service) Delete(orgID, userID uuid.UUID) error {
	var org models.Organization
	if err := s.db.First(&org, "id = ?", orgID).Error; err != nil {
		return ErrOrgNotFound
	}
	if org.OwnerID != userID {
		return ErrNotOwner
	}
	return s.db.Delete(&org).Error
}

func (s *Service) ListMembers(orgID uuid.UUID) ([]models.OrganizationMember, error) {
	var members []models.OrganizationMember
	err := s.db.Preload("User").Where("organization_id = ?", orgID).Find(&members).Error
	return members, err
}

func (s *Service) InviteMember(orgID, inviterID uuid.UUID, input *InviteInput) (*models.OrganizationInvitation, error) {
	validRoles := map[string]bool{"admin": true, "member": true, "viewer": true}
	if !validRoles[input.Role] {
		return nil, fmt.Errorf("invalid role: %s", input.Role)
	}

	// Check if already a member
	var existing models.OrganizationMember
	var user models.User
	if s.db.Where("email = ?", input.Email).First(&user).Error == nil {
		if s.db.Where("organization_id = ? AND user_id = ?", orgID, user.ID).First(&existing).Error == nil {
			return nil, ErrAlreadyMember
		}
	}

	tokenBytes := make([]byte, 32)
	rand.Read(tokenBytes)
	tokenStr := hex.EncodeToString(tokenBytes)

	invite := &models.OrganizationInvitation{
		OrganizationID: orgID,
		Email:          input.Email,
		Role:           input.Role,
		Token:          tokenStr,
		InvitedBy:      inviterID,
		ExpiresAt:      time.Now().Add(7 * 24 * time.Hour),
	}

	if err := s.db.Create(invite).Error; err != nil {
		return nil, fmt.Errorf("failed to create invitation: %w", err)
	}

	// Send invitation email
	if s.emailService != nil {
		var org models.Organization
		s.db.First(&org, "id = ?", orgID)
		var inviter models.User
		s.db.First(&inviter, "id = ?", inviterID)
		_ = s.emailService.SendInvitationEmail(input.Email, org.Name, inviter.DisplayName, input.Role, tokenStr)
	}

	return invite, nil
}

func (s *Service) AcceptInvitation(token string, userID uuid.UUID) error {
	var invite models.OrganizationInvitation
	if err := s.db.Where("token = ? AND accepted = false", token).First(&invite).Error; err != nil {
		return ErrInviteNotFound
	}

	if time.Now().After(invite.ExpiresAt) {
		return ErrInviteExpired
	}

	tx := s.db.Begin()

	if err := tx.Model(&invite).Update("accepted", true).Error; err != nil {
		tx.Rollback()
		return err
	}

	member := &models.OrganizationMember{
		OrganizationID: invite.OrganizationID,
		UserID:         userID,
		Role:           invite.Role,
		InvitedBy:      &invite.InvitedBy,
	}

	if err := tx.Create(member).Error; err != nil {
		tx.Rollback()
		if strings.Contains(err.Error(), "duplicate") {
			return ErrAlreadyMember
		}
		return err
	}

	return tx.Commit().Error
}

func (s *Service) RemoveMember(orgID, requesterID, targetUserID uuid.UUID) error {
	if requesterID == targetUserID {
		return ErrCannotRemoveSelf
	}

	result := s.db.Where("organization_id = ? AND user_id = ?", orgID, targetUserID).Delete(&models.OrganizationMember{})
	if result.RowsAffected == 0 {
		return ErrMemberNotFound
	}
	return result.Error
}

func (s *Service) UpdateMemberRole(orgID, targetUserID uuid.UUID, newRole string) error {
	validRoles := map[string]bool{"admin": true, "member": true, "viewer": true}
	if !validRoles[newRole] {
		return fmt.Errorf("invalid role: %s", newRole)
	}

	result := s.db.Model(&models.OrganizationMember{}).
		Where("organization_id = ? AND user_id = ?", orgID, targetUserID).
		Update("role", newRole)

	if result.RowsAffected == 0 {
		return ErrMemberNotFound
	}
	return result.Error
}

var slugRegex = regexp.MustCompile(`[^a-z0-9-]`)

func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = slugRegex.ReplaceAllString(s, "-")
	s = regexp.MustCompile(`-+`).ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}
