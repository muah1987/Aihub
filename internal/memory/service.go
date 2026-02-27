package memory

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/muah1987/Aihub/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrMemoryNotFound = errors.New("memory entry not found")
)

type SetInput struct {
	Category    string `json:"category"`
	Key         string `json:"key" validate:"required"`
	Content     string `json:"content" validate:"required"`
	ContentType string `json:"content_type"`
	Pinned      *bool  `json:"pinned"`
}

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

func (s *Service) Set(projectID uuid.UUID, userID *uuid.UUID, input *SetInput) (*models.TeamMemory, error) {
	category := input.Category
	if category == "" {
		category = "general"
	}
	contentType := input.ContentType
	if contentType == "" {
		contentType = "text"
	}

	mem := &models.TeamMemory{
		ProjectID:   projectID,
		Category:    category,
		Key:         input.Key,
		Content:     input.Content,
		ContentType: contentType,
		CreatedBy:   userID,
	}
	if input.Pinned != nil {
		mem.Pinned = *input.Pinned
	}

	// Upsert on (project_id, category, key)
	result := s.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "project_id"}, {Name: "category"}, {Name: "key"}},
		DoUpdates: clause.AssignmentColumns([]string{"content", "content_type", "pinned", "updated_at"}),
	}).Create(mem)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to set memory: %w", result.Error)
	}

	// Refetch to get the full record
	var fetched models.TeamMemory
	s.db.Where("project_id = ? AND category = ? AND key = ?", projectID, category, input.Key).First(&fetched)
	return &fetched, nil
}

func (s *Service) Get(projectID uuid.UUID, category, key string) (*models.TeamMemory, error) {
	var mem models.TeamMemory
	err := s.db.Where("project_id = ? AND category = ? AND key = ?", projectID, category, key).First(&mem).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrMemoryNotFound
		}
		return nil, err
	}
	return &mem, nil
}

func (s *Service) List(projectID uuid.UUID, category string) ([]models.TeamMemory, error) {
	var memories []models.TeamMemory
	query := s.db.Where("project_id = ?", projectID)
	if category != "" {
		query = query.Where("category = ?", category)
	}
	err := query.Order("pinned DESC, updated_at DESC").Find(&memories).Error
	return memories, err
}

func (s *Service) Delete(projectID, memoryID uuid.UUID) error {
	result := s.db.Where("id = ? AND project_id = ?", memoryID, projectID).Delete(&models.TeamMemory{})
	if result.RowsAffected == 0 {
		return ErrMemoryNotFound
	}
	return result.Error
}

func (s *Service) Search(projectID uuid.UUID, query string) ([]models.TeamMemory, error) {
	var memories []models.TeamMemory
	err := s.db.Where("project_id = ? AND to_tsvector('english', content) @@ plainto_tsquery('english', ?)", projectID, query).
		Order("pinned DESC, updated_at DESC").
		Find(&memories).Error
	return memories, err
}

func (s *Service) GetProjectContext(projectID uuid.UUID) ([]models.TeamMemory, error) {
	var memories []models.TeamMemory
	err := s.db.Where("project_id = ? AND (pinned = true OR updated_at > NOW() - INTERVAL '7 days')", projectID).
		Order("pinned DESC, updated_at DESC").
		Limit(50).
		Find(&memories).Error
	return memories, err
}
