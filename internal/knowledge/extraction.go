package knowledge

import (
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/muah1987/Aihub/internal/models"
	"gorm.io/gorm"
)

var ErrExtractionNotFound = errors.New("extraction not found")

// Extraction types
const (
	ExtractionDecision     = "decision"
	ExtractionRequirement  = "requirement"
	ExtractionArchitecture = "architecture"
	ExtractionBug          = "bug"
	ExtractionActionItem   = "action_item"
)

// ExtractFromMessage analyzes a chat message and creates auto-extractions.
func (s *Service) ExtractFromMessage(projectID uuid.UUID, messageID uuid.UUID, content, senderType string) ([]models.AutoExtraction, error) {
	var extractions []models.AutoExtraction

	// Simple rule-based extraction (can be replaced with AI later)
	patterns := map[string]struct {
		keywords   []string
		confidence float64
	}{
		ExtractionDecision: {
			keywords:   []string{"decided", "we will", "let's go with", "agreed", "decision:", "we chose"},
			confidence: 0.7,
		},
		ExtractionRequirement: {
			keywords:   []string{"must", "should", "requirement:", "need to", "required", "needs to support"},
			confidence: 0.6,
		},
		ExtractionArchitecture: {
			keywords:   []string{"architecture", "design pattern", "database schema", "api design", "system design", "tech stack"},
			confidence: 0.6,
		},
		ExtractionBug: {
			keywords:   []string{"bug:", "issue:", "broken", "doesn't work", "error:", "crash", "fix:"},
			confidence: 0.7,
		},
		ExtractionActionItem: {
			keywords:   []string{"todo:", "action item:", "task:", "TODO", "FIXME", "HACK", "next step:"},
			confidence: 0.8,
		},
	}

	lowerContent := strings.ToLower(content)

	for extractionType, pattern := range patterns {
		for _, keyword := range pattern.keywords {
			if strings.Contains(lowerContent, strings.ToLower(keyword)) {
				// Extract the sentence containing the keyword
				extracted := extractSentence(content, keyword)
				if extracted == "" {
					extracted = content
				}
				// Truncate to reasonable length
				if len(extracted) > 500 {
					extracted = extracted[:500] + "..."
				}

				extraction := models.AutoExtraction{
					ProjectID:        projectID,
					MessageID:        &messageID,
					ExtractedContent: extracted,
					ExtractionType:   extractionType,
					Confidence:       pattern.confidence,
				}
				// Higher confidence for agent messages
				if senderType == "agent" {
					extraction.Confidence += 0.1
					if extraction.Confidence > 1.0 {
						extraction.Confidence = 1.0
					}
				}

				extractions = append(extractions, extraction)
				break // one extraction per type per message
			}
		}
	}

	if len(extractions) > 0 {
		if err := s.db.Create(&extractions).Error; err != nil {
			return nil, fmt.Errorf("failed to save extractions: %w", err)
		}
	}

	return extractions, nil
}

// ListExtractions returns pending/accepted extractions.
func (s *Service) ListExtractions(projectID uuid.UUID, status string, limit int) ([]models.AutoExtraction, error) {
	if limit <= 0 {
		limit = 50
	}
	query := s.db.Where("project_id = ?", projectID)

	switch status {
	case "pending":
		query = query.Where("accepted IS NULL")
	case "accepted":
		query = query.Where("accepted = true")
	case "rejected":
		query = query.Where("accepted = false")
	}

	var extractions []models.AutoExtraction
	err := query.Order("created_at DESC").Limit(limit).Find(&extractions).Error
	return extractions, err
}

// AcceptExtraction accepts an extraction and optionally saves it to memory.
func (s *Service) AcceptExtraction(projectID, extractionID uuid.UUID, saveToMemory bool, userID *uuid.UUID) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var extraction models.AutoExtraction
		if err := tx.Where("id = ? AND project_id = ?", extractionID, projectID).First(&extraction).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrExtractionNotFound
			}
			return err
		}

		accepted := true
		updates := map[string]interface{}{"accepted": accepted}

		if saveToMemory {
			mem := &models.TeamMemory{
				ProjectID:   projectID,
				Category:    extraction.ExtractionType,
				Key:         fmt.Sprintf("auto-%s", extraction.ID.String()[:8]),
				Content:     extraction.ExtractedContent,
				ContentType: "text",
				CreatedBy:   userID,
			}
			if err := tx.Create(mem).Error; err != nil {
				return fmt.Errorf("failed to create memory from extraction: %w", err)
			}
			updates["memory_id"] = mem.ID
		}

		return tx.Model(&models.AutoExtraction{}).Where("id = ?", extractionID).Updates(updates).Error
	})
}

// RejectExtraction marks an extraction as rejected.
func (s *Service) RejectExtraction(projectID, extractionID uuid.UUID) error {
	rejected := false
	result := s.db.Model(&models.AutoExtraction{}).
		Where("id = ? AND project_id = ?", extractionID, projectID).
		Update("accepted", rejected)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrExtractionNotFound
	}
	return nil
}

// extractSentence finds the sentence in text that contains the keyword.
func extractSentence(text, keyword string) string {
	lowerText := strings.ToLower(text)
	lowerKeyword := strings.ToLower(keyword)
	idx := strings.Index(lowerText, lowerKeyword)
	if idx < 0 {
		return ""
	}

	// Find sentence boundaries
	start := idx
	for start > 0 && text[start-1] != '.' && text[start-1] != '\n' && text[start-1] != '!' && text[start-1] != '?' {
		start--
	}

	end := idx + len(keyword)
	for end < len(text) && text[end] != '.' && text[end] != '\n' && text[end] != '!' && text[end] != '?' {
		end++
	}
	if end < len(text) {
		end++ // include the punctuation
	}

	return strings.TrimSpace(text[start:end])
}
