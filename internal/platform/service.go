package platform

import (
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/muah1987/Aihub/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// escapeLike escapes LIKE special characters in user input.
func escapeLike(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "%", "\\%")
	s = strings.ReplaceAll(s, "_", "\\_")
	return s
}

var (
	ErrPreferencesNotFound = errors.New("preferences not found")
	ErrPinnedNotFound      = errors.New("pinned project not found")
)

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

// ---- User Preferences ----

type PreferencesInput struct {
	Theme              string                 `json:"theme"`
	AccentColor        string                 `json:"accent_color"`
	SidebarCollapsed   *bool                  `json:"sidebar_collapsed"`
	CompactMode        *bool                  `json:"compact_mode"`
	EditorFontSize     int                    `json:"editor_font_size"`
	NotificationsSound *bool                  `json:"notifications_sound"`
	Locale             string                 `json:"locale"`
	Timezone           string                 `json:"timezone"`
	Settings           map[string]interface{} `json:"settings"`
}

func (s *Service) GetPreferences(userID uuid.UUID) (*models.UserPreference, error) {
	var prefs models.UserPreference
	err := s.db.Where("user_id = ?", userID).First(&prefs).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Return defaults
			return &models.UserPreference{
				UserID:             userID,
				Theme:              "system",
				AccentColor:        "blue",
				EditorFontSize:     14,
				NotificationsSound: true,
				Locale:             "en",
				Timezone:           "UTC",
			}, nil
		}
		return nil, err
	}
	return &prefs, nil
}

func (s *Service) UpdatePreferences(userID uuid.UUID, input *PreferencesInput) (*models.UserPreference, error) {
	prefs := &models.UserPreference{UserID: userID}

	// Build updates map
	updates := map[string]interface{}{}
	if input.Theme != "" {
		updates["theme"] = input.Theme
	}
	if input.AccentColor != "" {
		updates["accent_color"] = input.AccentColor
	}
	if input.SidebarCollapsed != nil {
		updates["sidebar_collapsed"] = *input.SidebarCollapsed
	}
	if input.CompactMode != nil {
		updates["compact_mode"] = *input.CompactMode
	}
	if input.EditorFontSize > 0 {
		updates["editor_font_size"] = input.EditorFontSize
	}
	if input.NotificationsSound != nil {
		updates["notifications_sound"] = *input.NotificationsSound
	}
	if input.Locale != "" {
		updates["locale"] = input.Locale
	}
	if input.Timezone != "" {
		updates["timezone"] = input.Timezone
	}

	// Upsert
	result := s.db.Where("user_id = ?", userID).First(prefs)
	if result.Error != nil && errors.Is(result.Error, gorm.ErrRecordNotFound) {
		// Create with defaults + overrides
		prefs.Theme = "system"
		prefs.AccentColor = "blue"
		prefs.EditorFontSize = 14
		prefs.NotificationsSound = true
		prefs.Locale = "en"
		prefs.Timezone = "UTC"
		for k, v := range updates {
			switch k {
			case "theme":
				prefs.Theme = v.(string)
			case "accent_color":
				prefs.AccentColor = v.(string)
			case "sidebar_collapsed":
				prefs.SidebarCollapsed = v.(bool)
			case "compact_mode":
				prefs.CompactMode = v.(bool)
			case "editor_font_size":
				prefs.EditorFontSize = v.(int)
			case "notifications_sound":
				prefs.NotificationsSound = v.(bool)
			case "locale":
				prefs.Locale = v.(string)
			case "timezone":
				prefs.Timezone = v.(string)
			}
		}
		if err := s.db.Create(prefs).Error; err != nil {
			return nil, fmt.Errorf("failed to create preferences: %w", err)
		}
	} else {
		if err := s.db.Model(prefs).Updates(updates).Error; err != nil {
			return nil, fmt.Errorf("failed to update preferences: %w", err)
		}
	}

	return s.GetPreferences(userID)
}

// ---- Pinned Projects ----

func (s *Service) ListPinned(userID uuid.UUID) ([]models.PinnedProject, error) {
	var pins []models.PinnedProject
	err := s.db.Preload("Project").
		Where("user_id = ?", userID).
		Order("pin_order ASC, created_at ASC").
		Find(&pins).Error
	return pins, err
}

func (s *Service) PinProject(userID, projectID uuid.UUID) (*models.PinnedProject, error) {
	// Get next order
	var maxOrder int
	s.db.Model(&models.PinnedProject{}).Where("user_id = ?", userID).
		Select("COALESCE(MAX(pin_order), 0)").Scan(&maxOrder)

	pin := &models.PinnedProject{
		UserID:    userID,
		ProjectID: projectID,
		PinOrder:  maxOrder + 1,
	}

	result := s.db.Clauses(clause.OnConflict{DoNothing: true}).Create(pin)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to pin project: %w", result.Error)
	}

	// Refetch with project
	var fetched models.PinnedProject
	s.db.Preload("Project").Where("user_id = ? AND project_id = ?", userID, projectID).First(&fetched)
	return &fetched, nil
}

func (s *Service) UnpinProject(userID, projectID uuid.UUID) error {
	result := s.db.Where("user_id = ? AND project_id = ?", userID, projectID).Delete(&models.PinnedProject{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrPinnedNotFound
	}
	return nil
}

func (s *Service) ReorderPins(userID uuid.UUID, projectIDs []uuid.UUID) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		for i, pid := range projectIDs {
			if err := tx.Model(&models.PinnedProject{}).
				Where("user_id = ? AND project_id = ?", userID, pid).
				Update("pin_order", i).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// ---- Global Search ----

type SearchResult struct {
	Type      string `json:"type"` // "message", "memory", "agent", "document", "workflow"
	ID        string `json:"id"`
	ProjectID string `json:"project_id"`
	Title     string `json:"title"`
	Snippet   string `json:"snippet"`
	Score     float64 `json:"score"`
}

func (s *Service) GlobalSearch(userID uuid.UUID, query string, limit int) ([]SearchResult, error) {
	if limit <= 0 {
		limit = 20
	}
	var results []SearchResult

	// Search messages
	var messages []models.Message
	s.db.Where("project_id IN (SELECT id FROM projects WHERE user_id = ?) AND to_tsvector('english', content) @@ plainto_tsquery('english', ?)",
		userID, query).Limit(limit / 4).Find(&messages)
	for _, m := range messages {
		snippet := m.Content
		if len(snippet) > 150 {
			snippet = snippet[:150] + "..."
		}
		results = append(results, SearchResult{
			Type: "message", ID: m.ID.String(), ProjectID: m.ProjectID.String(),
			Title: "Chat message", Snippet: snippet,
		})
	}

	// Search memories
	var memories []models.TeamMemory
	s.db.Where("project_id IN (SELECT id FROM projects WHERE user_id = ?) AND to_tsvector('english', content) @@ plainto_tsquery('english', ?)",
		userID, query).Limit(limit / 4).Find(&memories)
	for _, m := range memories {
		snippet := m.Content
		if len(snippet) > 150 {
			snippet = snippet[:150] + "..."
		}
		results = append(results, SearchResult{
			Type: "memory", ID: m.ID.String(), ProjectID: m.ProjectID.String(),
			Title: m.Category + "/" + m.Key, Snippet: snippet,
		})
	}

	// Search agents
	var agents []models.Agent
	escaped := escapeLike(query)
	s.db.Where("project_id IN (SELECT id FROM projects WHERE user_id = ?) AND (name ILIKE ? OR system_prompt ILIKE ?)",
		userID, "%"+escaped+"%", "%"+escaped+"%").Limit(limit / 4).Find(&agents)
	for _, a := range agents {
		snippet := a.SystemPrompt
		if len(snippet) > 150 {
			snippet = snippet[:150] + "..."
		}
		results = append(results, SearchResult{
			Type: "agent", ID: a.ID.String(), ProjectID: a.ProjectID.String(),
			Title: a.Name, Snippet: snippet,
		})
	}

	// Search document chunks
	var chunks []models.DocumentChunk
	s.db.Where("project_id IN (SELECT id FROM projects WHERE user_id = ?) AND to_tsvector('english', content) @@ plainto_tsquery('english', ?)",
		userID, query).Limit(limit / 4).Find(&chunks)
	for _, c := range chunks {
		snippet := c.Content
		if len(snippet) > 150 {
			snippet = snippet[:150] + "..."
		}
		results = append(results, SearchResult{
			Type: "document", ID: c.ID.String(), ProjectID: c.ProjectID.String(),
			Title: fmt.Sprintf("Document chunk #%d", c.ChunkIndex+1), Snippet: snippet,
		})
	}

	return results, nil
}

// ---- Prompt Versions ----

var ErrPromptVersionNotFound = errors.New("prompt version not found")

func (s *Service) ListPromptVersions(agentID uuid.UUID) ([]models.PromptVersion, error) {
	var versions []models.PromptVersion
	err := s.db.Where("agent_id = ?", agentID).Order("created_at DESC").Find(&versions).Error
	return versions, err
}

func (s *Service) CreatePromptVersion(agentID uuid.UUID, userID *uuid.UUID, label, systemPrompt string) (*models.PromptVersion, error) {
	v := &models.PromptVersion{
		AgentID:      agentID,
		VersionLabel: label,
		SystemPrompt: systemPrompt,
		CreatedBy:    userID,
	}
	if err := s.db.Create(v).Error; err != nil {
		return nil, fmt.Errorf("failed to create prompt version: %w", err)
	}
	return v, nil
}

func (s *Service) SetActivePromptVersion(agentID, versionID uuid.UUID) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Deactivate all
		if err := tx.Model(&models.PromptVersion{}).Where("agent_id = ?", agentID).Update("is_active", false).Error; err != nil {
			return err
		}
		// Activate selected
		result := tx.Model(&models.PromptVersion{}).Where("id = ? AND agent_id = ?", versionID, agentID).Update("is_active", true)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return ErrPromptVersionNotFound
		}

		// Also update the agent's system_prompt
		var v models.PromptVersion
		if err := tx.First(&v, "id = ?", versionID).Error; err == nil {
			tx.Model(&models.Agent{}).Where("id = ?", agentID).Update("system_prompt", v.SystemPrompt)
		}

		return nil
	})
}

func (s *Service) DeletePromptVersion(agentID, versionID uuid.UUID) error {
	result := s.db.Where("id = ? AND agent_id = ?", versionID, agentID).Delete(&models.PromptVersion{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrPromptVersionNotFound
	}
	return nil
}
