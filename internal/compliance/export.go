package compliance

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/muah1987/Aihub/internal/models"
	"gorm.io/gorm"
)

var ErrExportNotFound = errors.New("export not found")

// ExportService handles data export/import for project backup.
type ExportService struct {
	db *gorm.DB
}

func NewExportService(db *gorm.DB) *ExportService {
	return &ExportService{db: db}
}

type CreateExportInput struct {
	ExportType       string `json:"export_type"`
	Format           string `json:"format"`
	IncludeChat      bool   `json:"include_chat"`
	IncludeMemories  bool   `json:"include_memories"`
	IncludeAgents    bool   `json:"include_agents"`
	IncludeWorkflows bool   `json:"include_workflows"`
	IncludeSettings  bool   `json:"include_settings"`
}

func (s *ExportService) CreateExport(projectID uuid.UUID, userID *uuid.UUID, input *CreateExportInput) (*models.DataExport, error) {
	if input.ExportType == "" {
		input.ExportType = "full"
	}
	if input.Format == "" {
		input.Format = "json"
	}

	expires := time.Now().Add(72 * time.Hour) // expires in 3 days

	export := &models.DataExport{
		ProjectID:        projectID,
		RequestedBy:      userID,
		ExportType:       input.ExportType,
		Format:           input.Format,
		Status:           "processing",
		IncludeChat:      input.IncludeChat,
		IncludeMemories:  input.IncludeMemories,
		IncludeAgents:    input.IncludeAgents,
		IncludeWorkflows: input.IncludeWorkflows,
		IncludeSettings:  input.IncludeSettings,
		ExpiresAt:        &expires,
	}

	if err := s.db.Create(export).Error; err != nil {
		return nil, fmt.Errorf("failed to create export: %w", err)
	}

	// Process export synchronously (for MVP; could be async job later)
	data, err := s.gatherData(projectID, input)
	if err != nil {
		s.db.Model(export).Updates(map[string]interface{}{
			"status":        "error",
			"error_message": err.Error(),
		})
		return export, err
	}

	exportJSON, _ := json.MarshalIndent(data, "", "  ")
	now := time.Now()

	export.Status = "ready"
	export.FileSize = int64(len(exportJSON))
	export.FilePath = fmt.Sprintf("exports/%s/%s.json", projectID, export.ID)
	export.CompletedAt = &now

	s.db.Save(export)

	return export, nil
}

func (s *ExportService) GetExport(projectID, exportID uuid.UUID) (*models.DataExport, error) {
	var export models.DataExport
	err := s.db.Where("id = ? AND project_id = ?", exportID, projectID).First(&export).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrExportNotFound
		}
		return nil, err
	}
	return &export, nil
}

func (s *ExportService) ListExports(projectID uuid.UUID) ([]models.DataExport, error) {
	var exports []models.DataExport
	err := s.db.Where("project_id = ?", projectID).Order("created_at DESC").Find(&exports).Error
	return exports, err
}

func (s *ExportService) DeleteExport(projectID, exportID uuid.UUID) error {
	result := s.db.Where("id = ? AND project_id = ?", exportID, projectID).Delete(&models.DataExport{})
	if result.RowsAffected == 0 {
		return ErrExportNotFound
	}
	return result.Error
}

// DownloadExport generates the export data for download.
func (s *ExportService) DownloadExport(projectID, exportID uuid.UUID) ([]byte, error) {
	export, err := s.GetExport(projectID, exportID)
	if err != nil {
		return nil, err
	}
	if export.Status != "ready" {
		return nil, fmt.Errorf("export is not ready (status: %s)", export.Status)
	}

	input := &CreateExportInput{
		IncludeChat:      export.IncludeChat,
		IncludeMemories:  export.IncludeMemories,
		IncludeAgents:    export.IncludeAgents,
		IncludeWorkflows: export.IncludeWorkflows,
		IncludeSettings:  export.IncludeSettings,
	}
	data, err := s.gatherData(projectID, input)
	if err != nil {
		return nil, err
	}
	return json.MarshalIndent(data, "", "  ")
}

func (s *ExportService) gatherData(projectID uuid.UUID, input *CreateExportInput) (map[string]interface{}, error) {
	data := map[string]interface{}{
		"export_version": "1.0",
		"project_id":     projectID.String(),
		"exported_at":    time.Now().UTC().Format(time.RFC3339),
	}

	// Project metadata
	var project models.Project
	if err := s.db.First(&project, "id = ?", projectID).Error; err == nil {
		data["project"] = map[string]interface{}{
			"name":       project.Name,
			"repo_owner": project.RepoOwner,
			"repo_name":  project.RepoName,
		}
	}

	if input.IncludeChat {
		var messages []models.Message
		s.db.Where("project_id = ?", projectID).Order("created_at ASC").Find(&messages)
		data["messages"] = messages
	}

	if input.IncludeMemories {
		var memories []models.TeamMemory
		s.db.Where("project_id = ?", projectID).Find(&memories)
		data["memories"] = memories
	}

	if input.IncludeAgents {
		var agents []models.Agent
		s.db.Where("project_id = ?", projectID).Find(&agents)
		data["agents"] = agents
	}

	if input.IncludeWorkflows {
		var workflows []models.Workflow
		s.db.Where("project_id = ?", projectID).Find(&workflows)
		data["workflows"] = workflows
	}

	if input.IncludeSettings {
		var tools []models.AgentTool
		s.db.Where("project_id = ?", projectID).Find(&tools)
		data["tools"] = tools
	}

	return data, nil
}
