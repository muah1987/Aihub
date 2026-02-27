package project

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/muah1987/Aihub/internal/models"
	"github.com/muah1987/Aihub/internal/provider"
	"gorm.io/gorm"
)

var (
	ErrProjectNotFound = errors.New("project not found")
	ErrDuplicateRepo   = errors.New("repository already linked to a project")
)

type CreateInput struct {
	Name                 string `json:"name" validate:"required"`
	Description          string `json:"description"`
	RepoOwner            string `json:"repo_owner" validate:"required"`
	RepoName             string `json:"repo_name" validate:"required"`
	ProviderConnectionID string `json:"provider_connection_id" validate:"required"`
}

type UpdateInput struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Status      *string `json:"status"`
}

type Service struct {
	db              *gorm.DB
	providerService *provider.Service
}

func NewService(db *gorm.DB, providerService *provider.Service) *Service {
	return &Service{db: db, providerService: providerService}
}

func (s *Service) Create(userID uuid.UUID, input *CreateInput) (*models.Project, error) {
	connID, err := uuid.Parse(input.ProviderConnectionID)
	if err != nil {
		return nil, fmt.Errorf("invalid provider connection id: %w", err)
	}

	// Get the decrypted token to fetch repo info
	token, err := s.providerService.GetDecryptedToken(userID, connID)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider token: %w", err)
	}

	// Validate repo exists on GitHub
	ghClient := provider.NewGitHubClient(token)
	repo, err := ghClient.GetRepo(input.RepoOwner, input.RepoName)
	if err != nil {
		return nil, fmt.Errorf("failed to access repository: %w", err)
	}

	project := &models.Project{
		UserID:               userID,
		Name:                 input.Name,
		Description:          input.Description,
		RepoProvider:         "github",
		RepoOwner:            input.RepoOwner,
		RepoName:             input.RepoName,
		RepoURL:              repo.HTMLURL,
		RepoDefaultBranch:    repo.DefaultBranch,
		ProviderConnectionID: &connID,
		Status:               "active",
	}

	if err := s.db.Create(project).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, ErrDuplicateRepo
		}
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	return project, nil
}

func (s *Service) List(userID uuid.UUID) ([]models.Project, error) {
	var projects []models.Project
	if err := s.db.Where("user_id = ?", userID).Order("updated_at DESC").Find(&projects).Error; err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}
	return projects, nil
}

func (s *Service) GetByID(userID, projectID uuid.UUID) (*models.Project, error) {
	var project models.Project
	if err := s.db.Where("id = ? AND user_id = ?", projectID, userID).First(&project).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}
	return &project, nil
}

func (s *Service) Update(userID, projectID uuid.UUID, input *UpdateInput) (*models.Project, error) {
	project, err := s.GetByID(userID, projectID)
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
	if input.Status != nil {
		updates["status"] = *input.Status
	}

	if len(updates) > 0 {
		if err := s.db.Model(project).Updates(updates).Error; err != nil {
			return nil, fmt.Errorf("failed to update project: %w", err)
		}
	}

	return s.GetByID(userID, projectID)
}

func (s *Service) Delete(userID, projectID uuid.UUID) error {
	result := s.db.Where("id = ? AND user_id = ?", projectID, userID).Delete(&models.Project{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete project: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrProjectNotFound
	}
	return nil
}

func (s *Service) ListGitHubRepos(userID, connectionID uuid.UUID, page int) ([]provider.GitHubRepo, error) {
	token, err := s.providerService.GetDecryptedToken(userID, connectionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider token: %w", err)
	}

	ghClient := provider.NewGitHubClient(token)
	repos, err := ghClient.ListRepos(page, 30)
	if err != nil {
		return nil, fmt.Errorf("failed to list repos: %w", err)
	}

	return repos, nil
}
