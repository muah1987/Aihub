package knowledge

import (
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/muah1987/Aihub/internal/models"
	"gorm.io/gorm"
)

var ErrVersionNotFound = errors.New("version not found")

// RecordVersion saves a snapshot of a memory entry's content.
func (s *Service) RecordVersion(memoryID uuid.UUID, content, contentType, changeType string, changedBy *uuid.UUID) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Get the next version number with row-level lock
		var maxVersion int
		if err := tx.Raw(
			"SELECT COALESCE(MAX(version_number), 0) FROM memory_versions WHERE memory_id = ? FOR UPDATE",
			memoryID,
		).Scan(&maxVersion).Error; err != nil {
			return fmt.Errorf("failed to get max version: %w", err)
		}

		version := &models.MemoryVersion{
			MemoryID:      memoryID,
			VersionNumber: maxVersion + 1,
			Content:       content,
			ContentType:   contentType,
			ChangedBy:     changedBy,
			ChangeType:    changeType,
		}

		// Generate diff summary vs previous version
		if maxVersion > 0 {
			var prev models.MemoryVersion
			if err := tx.Where("memory_id = ? AND version_number = ?", memoryID, maxVersion).First(&prev).Error; err == nil {
				version.DiffSummary = generateDiffSummary(prev.Content, content)
			}
		}

		return tx.Create(version).Error
	})
}

// ListVersions returns all versions for a memory entry.
func (s *Service) ListVersions(memoryID uuid.UUID) ([]models.MemoryVersion, error) {
	var versions []models.MemoryVersion
	err := s.db.Where("memory_id = ?", memoryID).
		Order("version_number DESC").
		Find(&versions).Error
	return versions, err
}

// GetVersion returns a specific version.
func (s *Service) GetVersion(memoryID uuid.UUID, versionNumber int) (*models.MemoryVersion, error) {
	var version models.MemoryVersion
	err := s.db.Where("memory_id = ? AND version_number = ?", memoryID, versionNumber).First(&version).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrVersionNotFound
		}
		return nil, err
	}
	return &version, nil
}

// RollbackMemory restores a memory entry to a previous version.
func (s *Service) RollbackMemory(memoryID uuid.UUID, versionNumber int, userID *uuid.UUID) error {
	version, err := s.GetVersion(memoryID, versionNumber)
	if err != nil {
		return err
	}

	// Update the memory entry content
	result := s.db.Model(&models.TeamMemory{}).
		Where("id = ?", memoryID).
		Updates(map[string]interface{}{
			"content":      version.Content,
			"content_type": version.ContentType,
		})
	if result.Error != nil {
		return fmt.Errorf("failed to rollback: %w", result.Error)
	}

	// Record the rollback as a new version
	return s.RecordVersion(memoryID, version.Content, version.ContentType, "rollback", userID)
}

// generateDiffSummary creates a brief description of changes.
func generateDiffSummary(oldContent, newContent string) string {
	oldLines := strings.Split(oldContent, "\n")
	newLines := strings.Split(newContent, "\n")

	added := 0
	removed := 0

	oldSet := make(map[string]bool)
	for _, l := range oldLines {
		oldSet[strings.TrimSpace(l)] = true
	}
	newSet := make(map[string]bool)
	for _, l := range newLines {
		newSet[strings.TrimSpace(l)] = true
	}

	for _, l := range newLines {
		if !oldSet[strings.TrimSpace(l)] {
			added++
		}
	}
	for _, l := range oldLines {
		if !newSet[strings.TrimSpace(l)] {
			removed++
		}
	}

	lenDiff := len(newContent) - len(oldContent)
	sign := "+"
	if lenDiff < 0 {
		sign = ""
	}

	return fmt.Sprintf("%d lines added, %d lines removed (%s%d chars)", added, removed, sign, lenDiff)
}
