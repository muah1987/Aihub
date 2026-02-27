package knowledge

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/muah1987/Aihub/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrSharedMemoryNotFound = errors.New("shared memory not found")

// ShareMemory shares a memory entry with an organization.
func (s *Service) ShareMemory(orgID, projectID, memoryID uuid.UUID, userID *uuid.UUID, accessLevel string) (*models.SharedMemory, error) {
	if accessLevel == "" {
		accessLevel = "read"
	}

	shared := &models.SharedMemory{
		OrganizationID:  orgID,
		SourceProjectID: projectID,
		SourceMemoryID:  memoryID,
		SharedBy:        userID,
		AccessLevel:     accessLevel,
	}

	result := s.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "organization_id"}, {Name: "source_memory_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"access_level"}),
	}).Create(shared)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to share memory: %w", result.Error)
	}

	// Refetch with relations
	var fetched models.SharedMemory
	s.db.Preload("Memory").Preload("Project").
		Where("organization_id = ? AND source_memory_id = ?", orgID, memoryID).
		First(&fetched)

	return &fetched, nil
}

// UnshareMemory removes a shared memory entry.
func (s *Service) UnshareMemory(orgID, sharedID uuid.UUID) error {
	result := s.db.Where("id = ? AND organization_id = ?", sharedID, orgID).Delete(&models.SharedMemory{})
	if result.RowsAffected == 0 {
		return ErrSharedMemoryNotFound
	}
	return result.Error
}

// ListSharedMemories returns all memories shared with an organization.
func (s *Service) ListSharedMemories(orgID uuid.UUID) ([]models.SharedMemory, error) {
	var shared []models.SharedMemory
	err := s.db.Preload("Memory").Preload("Project").
		Where("organization_id = ?", orgID).
		Order("created_at DESC").
		Find(&shared).Error
	return shared, err
}

// GetSharedMemoriesForProject returns shared memories accessible to a project through its org.
func (s *Service) GetSharedMemoriesForProject(projectID uuid.UUID) ([]models.SharedMemory, error) {
	// Find orgs this project's owner belongs to, then find shared memories
	var shared []models.SharedMemory
	err := s.db.Preload("Memory").
		Joins("JOIN organization_members om ON om.organization_id = shared_memories.organization_id").
		Joins("JOIN projects p ON p.user_id = om.user_id AND p.id = ?", projectID).
		Where("shared_memories.source_project_id != ?", projectID).
		Find(&shared).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return shared, nil
}
