package team

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/muah1987/Aihub/internal/models"
	"gorm.io/gorm"
)

var (
	ErrTeamNotFound   = errors.New("team not found")
	ErrMemberNotFound = errors.New("team member not found")
	ErrTaskNotFound   = errors.New("task not found")
)

type CreateTeamInput struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
	Strategy    string `json:"strategy"`
}

type UpdateTeamInput struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Strategy    *string `json:"strategy"`
}

type AddMemberInput struct {
	AgentID    string `json:"agent_id" validate:"required"`
	RoleInTeam string `json:"role_in_team"`
}

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

func (s *Service) CreateTeam(projectID uuid.UUID, input *CreateTeamInput) (*models.AgentTeam, error) {
	strategy := input.Strategy
	if strategy == "" {
		strategy = "sequential"
	}

	team := &models.AgentTeam{
		ProjectID:   projectID,
		Name:        input.Name,
		Description: input.Description,
		Strategy:    strategy,
		Status:      "idle",
	}

	if err := s.db.Create(team).Error; err != nil {
		return nil, fmt.Errorf("failed to create team: %w", err)
	}

	return team, nil
}

func (s *Service) GetTeam(projectID, teamID uuid.UUID) (*models.AgentTeam, error) {
	var team models.AgentTeam
	err := s.db.Where("id = ? AND project_id = ?", teamID, projectID).First(&team).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTeamNotFound
		}
		return nil, err
	}
	return &team, nil
}

func (s *Service) ListTeams(projectID uuid.UUID) ([]models.AgentTeam, error) {
	var teams []models.AgentTeam
	err := s.db.Where("project_id = ?", projectID).Order("created_at ASC").Find(&teams).Error
	return teams, err
}

func (s *Service) UpdateTeam(projectID, teamID uuid.UUID, input *UpdateTeamInput) (*models.AgentTeam, error) {
	team, err := s.GetTeam(projectID, teamID)
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
	if input.Strategy != nil {
		updates["strategy"] = *input.Strategy
	}

	if len(updates) > 0 {
		if err := s.db.Model(team).Updates(updates).Error; err != nil {
			return nil, err
		}
	}

	return s.GetTeam(projectID, teamID)
}

func (s *Service) DeleteTeam(projectID, teamID uuid.UUID) error {
	result := s.db.Where("id = ? AND project_id = ?", teamID, projectID).Delete(&models.AgentTeam{})
	if result.RowsAffected == 0 {
		return ErrTeamNotFound
	}
	return result.Error
}

func (s *Service) AddMember(teamID, agentID uuid.UUID, roleInTeam string) (*models.AgentTeamMember, error) {
	if roleInTeam == "" {
		roleInTeam = "member"
	}

	member := &models.AgentTeamMember{
		TeamID:     teamID,
		AgentID:    agentID,
		RoleInTeam: roleInTeam,
	}

	if err := s.db.Create(member).Error; err != nil {
		return nil, fmt.Errorf("failed to add member: %w", err)
	}

	return member, nil
}

func (s *Service) RemoveMember(teamID, agentID uuid.UUID) error {
	result := s.db.Where("team_id = ? AND agent_id = ?", teamID, agentID).Delete(&models.AgentTeamMember{})
	if result.RowsAffected == 0 {
		return ErrMemberNotFound
	}
	return result.Error
}

func (s *Service) SetLeader(projectID, teamID, agentID uuid.UUID) error {
	return s.db.Model(&models.AgentTeam{}).
		Where("id = ? AND project_id = ?", teamID, projectID).
		Update("leader_agent_id", agentID).Error
}

func (s *Service) ListMembers(teamID uuid.UUID) ([]models.AgentTeamMember, error) {
	var members []models.AgentTeamMember
	err := s.db.Preload("Agent").Where("team_id = ?", teamID).Order("priority DESC").Find(&members).Error
	return members, err
}

func (s *Service) ListTasks(teamID uuid.UUID) ([]models.AgentTask, error) {
	var tasks []models.AgentTask
	err := s.db.Preload("AssignedAgent").Where("team_id = ?", teamID).Order("created_at ASC").Find(&tasks).Error
	return tasks, err
}

func (s *Service) GetTask(taskID uuid.UUID) (*models.AgentTask, error) {
	var task models.AgentTask
	err := s.db.Preload("AssignedAgent").First(&task, "id = ?", taskID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTaskNotFound
		}
		return nil, err
	}
	return &task, nil
}
