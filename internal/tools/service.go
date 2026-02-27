package tools

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/muah1987/Aihub/internal/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrToolNotFound = errors.New("tool not found")
	ErrToolDisabled = errors.New("tool is disabled")
)

// FunctionDef describes the JSON schema a "function" tool expects.
type FunctionDef struct {
	Parameters json.RawMessage `json:"parameters"` // JSON Schema object
}

// HTTPDef describes the config for an "http" tool.
type HTTPDef struct {
	URL     string            `json:"url"`
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers,omitempty"`
	Body    string            `json:"body,omitempty"` // Go template with {{.input}}
}

// ShellDef describes the config for a "shell" tool.
type ShellDef struct {
	Command    string `json:"command"`     // shell command template
	TimeoutSec int    `json:"timeout_sec"` // max seconds, default 30
}

type CreateToolInput struct {
	Name        string          `json:"name" validate:"required"`
	Description string          `json:"description"`
	ToolType    string          `json:"tool_type" validate:"required"`
	Definition  json.RawMessage `json:"definition"`
}

type UpdateToolInput struct {
	Name        *string          `json:"name"`
	Description *string          `json:"description"`
	Enabled     *bool            `json:"enabled"`
	Definition  *json.RawMessage `json:"definition"`
}

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

// ---------- CRUD ----------

func (s *Service) Create(projectID uuid.UUID, input *CreateToolInput) (*models.AgentTool, error) {
	tool := &models.AgentTool{
		ProjectID:   projectID,
		Name:        input.Name,
		Description: input.Description,
		ToolType:    input.ToolType,
		Definition:  datatypes.JSON(input.Definition),
		Enabled:     true,
	}
	if err := s.db.Create(tool).Error; err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return nil, fmt.Errorf("tool name %q already exists in this project", input.Name)
		}
		return nil, fmt.Errorf("failed to create tool: %w", err)
	}
	return tool, nil
}

func (s *Service) List(projectID uuid.UUID) ([]models.AgentTool, error) {
	var tools []models.AgentTool
	err := s.db.Where("project_id = ?", projectID).Order("name ASC").Find(&tools).Error
	return tools, err
}

func (s *Service) GetByID(projectID, toolID uuid.UUID) (*models.AgentTool, error) {
	var tool models.AgentTool
	if err := s.db.Where("id = ? AND project_id = ?", toolID, projectID).First(&tool).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrToolNotFound
		}
		return nil, err
	}
	return &tool, nil
}

func (s *Service) Update(projectID, toolID uuid.UUID, input *UpdateToolInput) (*models.AgentTool, error) {
	tool, err := s.GetByID(projectID, toolID)
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
	if input.Enabled != nil {
		updates["enabled"] = *input.Enabled
	}
	if input.Definition != nil {
		updates["definition"] = datatypes.JSON(*input.Definition)
	}
	if len(updates) > 0 {
		if err := s.db.Model(tool).Updates(updates).Error; err != nil {
			return nil, err
		}
	}
	return s.GetByID(projectID, toolID)
}

func (s *Service) Delete(projectID, toolID uuid.UUID) error {
	res := s.db.Where("id = ? AND project_id = ?", toolID, projectID).Delete(&models.AgentTool{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrToolNotFound
	}
	return nil
}

// ---------- Bindings ----------

func (s *Service) BindTool(agentID, toolID uuid.UUID) error {
	b := &models.AgentToolBinding{AgentID: agentID, ToolID: toolID}
	return s.db.Clauses(clause.OnConflict{DoNothing: true}).Create(b).Error
}

func (s *Service) UnbindTool(agentID, toolID uuid.UUID) error {
	return s.db.Where("agent_id = ? AND tool_id = ?", agentID, toolID).Delete(&models.AgentToolBinding{}).Error
}

func (s *Service) ListBindings(agentID uuid.UUID) ([]models.AgentToolBinding, error) {
	var bindings []models.AgentToolBinding
	err := s.db.Preload("Tool").Where("agent_id = ?", agentID).Find(&bindings).Error
	return bindings, err
}

func (s *Service) GetAgentTools(agentID uuid.UUID) ([]models.AgentTool, error) {
	var tools []models.AgentTool
	err := s.db.
		Joins("JOIN agent_tool_bindings atb ON atb.tool_id = agent_tools.id").
		Where("atb.agent_id = ? AND agent_tools.enabled = true", agentID).
		Find(&tools).Error
	return tools, err
}

// ---------- Execution ----------

// Execute runs a tool and logs the execution.
func (s *Service) Execute(agentID, toolID, projectID uuid.UUID, userID *uuid.UUID, inputData json.RawMessage) (*models.ToolExecution, error) {
	tool, err := s.GetByID(projectID, toolID)
	if err != nil {
		return nil, err
	}
	if !tool.Enabled {
		return nil, ErrToolDisabled
	}

	start := time.Now()
	output, execErr := s.runTool(tool, inputData)
	durationMs := int(time.Since(start).Milliseconds())

	exec := &models.ToolExecution{
		AgentID:    agentID,
		ToolID:     toolID,
		ProjectID:  projectID,
		Input:      datatypes.JSON(inputData),
		Output:     output,
		DurationMs: durationMs,
		Status:     "success",
	}
	if userID != nil {
		exec.TriggeredBy = userID
	}
	if execErr != nil {
		exec.Status = "error"
		errMsg := execErr.Error()
		exec.ErrorMessage = &errMsg
	}

	s.db.Create(exec)
	return exec, execErr
}

func (s *Service) runTool(tool *models.AgentTool, inputData json.RawMessage) (string, error) {
	switch tool.ToolType {
	case "http":
		return s.runHTTPTool(tool, inputData)
	case "shell":
		return s.runShellTool(tool, inputData)
	case "function":
		// Function tools are handled by the AI model natively via tool_use;
		// we just validate and return the input echo for logging.
		return string(inputData), nil
	default:
		return "", fmt.Errorf("unsupported tool type: %s", tool.ToolType)
	}
}

func (s *Service) runHTTPTool(tool *models.AgentTool, inputData json.RawMessage) (string, error) {
	var def HTTPDef
	if err := json.Unmarshal([]byte(tool.Definition), &def); err != nil {
		return "", fmt.Errorf("invalid http tool definition: %w", err)
	}

	method := strings.ToUpper(def.Method)
	if method == "" {
		method = "GET"
	}

	body := def.Body
	if body != "" {
		body = strings.ReplaceAll(body, "{{.input}}", string(inputData))
	}

	var bodyReader io.Reader
	if body != "" {
		bodyReader = strings.NewReader(body)
	}

	req, err := http.NewRequest(method, def.URL, bodyReader)
	if err != nil {
		return "", fmt.Errorf("failed to build request: %w", err)
	}
	for k, v := range def.Headers {
		req.Header.Set(k, v)
	}
	if req.Header.Get("Content-Type") == "" && body != "" {
		req.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()
	respBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20)) // 1MB cap
	if resp.StatusCode >= 400 {
		return string(respBytes), fmt.Errorf("http %d: %s", resp.StatusCode, string(respBytes))
	}
	return string(respBytes), nil
}

func (s *Service) runShellTool(tool *models.AgentTool, inputData json.RawMessage) (string, error) {
	var def ShellDef
	if err := json.Unmarshal([]byte(tool.Definition), &def); err != nil {
		return "", fmt.Errorf("invalid shell tool definition: %w", err)
	}

	timeout := def.TimeoutSec
	if timeout <= 0 {
		timeout = 30
	}

	// Substitute {{.input}} placeholder
	command := strings.ReplaceAll(def.Command, "{{.input}}", string(inputData))

	cmd := exec.Command("bash", "-c", command)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	done := make(chan error, 1)
	go func() { done <- cmd.Run() }()

	select {
	case err := <-done:
		if err != nil {
			return out.String(), fmt.Errorf("shell command failed: %w — output: %s", err, out.String())
		}
		return out.String(), nil
	case <-time.After(time.Duration(timeout) * time.Second):
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
		return out.String(), fmt.Errorf("shell command timed out after %ds", timeout)
	}
}

// ---------- Execution log queries ----------

func (s *Service) ListExecutions(projectID uuid.UUID, limit int) ([]models.ToolExecution, error) {
	if limit <= 0 {
		limit = 50
	}
	var execs []models.ToolExecution
	err := s.db.Where("project_id = ?", projectID).Order("created_at DESC").Limit(limit).Find(&execs).Error
	return execs, err
}

func (s *Service) ListAgentExecutions(agentID uuid.UUID, limit int) ([]models.ToolExecution, error) {
	if limit <= 0 {
		limit = 50
	}
	var execs []models.ToolExecution
	err := s.db.Where("agent_id = ?", agentID).Order("created_at DESC").Limit(limit).Find(&execs).Error
	return execs, err
}
