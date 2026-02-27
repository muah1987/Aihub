package activity

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/muah1987/Aihub/internal/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

// Log records an activity event. Safe to call from goroutines.
func (s *Service) Log(projectID uuid.UUID, userID *uuid.UUID, action, resourceType string, resourceID *uuid.UUID, details interface{}) {
	detailsJSON, _ := json.Marshal(details)
	entry := &models.ActivityLog{
		ProjectID:    projectID,
		UserID:       userID,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Details:      datatypes.JSON(detailsJSON),
	}
	s.db.Create(entry)
}

// List returns recent activity for a project.
func (s *Service) List(projectID uuid.UUID, limit, offset int) ([]models.ActivityLog, int64, error) {
	if limit <= 0 {
		limit = 50
	}
	var total int64
	s.db.Model(&models.ActivityLog{}).Where("project_id = ?", projectID).Count(&total)

	var logs []models.ActivityLog
	err := s.db.Where("project_id = ?", projectID).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&logs).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list activity: %w", err)
	}
	return logs, total, nil
}

// ListByAction filters activity by action type.
func (s *Service) ListByAction(projectID uuid.UUID, action string, limit int) ([]models.ActivityLog, error) {
	if limit <= 0 {
		limit = 50
	}
	var logs []models.ActivityLog
	err := s.db.Where("project_id = ? AND action = ?", projectID, action).
		Order("created_at DESC").
		Limit(limit).
		Find(&logs).Error
	return logs, err
}

// ListByResource filters activity by resource type.
func (s *Service) ListByResource(projectID uuid.UUID, resourceType string, limit int) ([]models.ActivityLog, error) {
	if limit <= 0 {
		limit = 50
	}
	var logs []models.ActivityLog
	err := s.db.Where("project_id = ? AND resource_type = ?", projectID, resourceType).
		Order("created_at DESC").
		Limit(limit).
		Find(&logs).Error
	return logs, err
}
