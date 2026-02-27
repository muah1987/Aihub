package agent

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/muah1987/Aihub/internal/chat"
	"github.com/muah1987/Aihub/internal/memory"
	"github.com/muah1987/Aihub/internal/models"
	"github.com/muah1987/Aihub/internal/provider"
	"gorm.io/gorm"
)

// UsageRecorder is satisfied by analytics.Service.
type UsageRecorder interface {
	Record(userID, projectID uuid.UUID, agentID *uuid.UUID, providerName, model string, inputTokens, outputTokens, toolCalls int) (*models.UsageRecord, error)
}

// ToolProvider is satisfied by tools.Service.
type ToolProvider interface {
	GetAgentTools(agentID uuid.UUID) ([]models.AgentTool, error)
	Execute(agentID, toolID, projectID uuid.UUID, userID *uuid.UUID, inputData json.RawMessage) (*models.ToolExecution, error)
}

var (
	ErrAgentNotFound = errors.New("agent not found")
)

type CreateInput struct {
	Name         string  `json:"name" validate:"required"`
	Role         string  `json:"role" validate:"required"`
	Model        string  `json:"model" validate:"required"`
	SystemPrompt string  `json:"system_prompt"`
	Temperature  float64 `json:"temperature"`
	MaxTokens    int     `json:"max_tokens"`
	ProviderConnectionID string `json:"provider_connection_id" validate:"required"`
}

type UpdateInput struct {
	Name         *string  `json:"name"`
	Role         *string  `json:"role"`
	Model        *string  `json:"model"`
	SystemPrompt *string  `json:"system_prompt"`
	Temperature  *float64 `json:"temperature"`
	MaxTokens    *int     `json:"max_tokens"`
}

type InvokeInput struct {
	Prompt string `json:"prompt" validate:"required"`
}

type Service struct {
	db              *gorm.DB
	providerService *provider.Service
	chatService     *chat.Service
	memoryService   *memory.Service
	usageRecorder   UsageRecorder
	toolProvider    ToolProvider
}

func NewService(db *gorm.DB, providerService *provider.Service, chatService *chat.Service, memoryService *memory.Service) *Service {
	return &Service{
		db:              db,
		providerService: providerService,
		chatService:     chatService,
		memoryService:   memoryService,
	}
}

// SetUsageRecorder wires the analytics service (avoids import cycle).
func (s *Service) SetUsageRecorder(ur UsageRecorder) { s.usageRecorder = ur }

// SetToolProvider wires the tools service (avoids import cycle).
func (s *Service) SetToolProvider(tp ToolProvider) { s.toolProvider = tp }

func (s *Service) Create(projectID uuid.UUID, input *CreateInput) (*models.Agent, error) {
	connID, err := uuid.Parse(input.ProviderConnectionID)
	if err != nil {
		return nil, fmt.Errorf("invalid provider connection id: %w", err)
	}

	temp := input.Temperature
	if temp <= 0 {
		temp = 0.7
	}
	maxTokens := input.MaxTokens
	if maxTokens <= 0 {
		maxTokens = 4096
	}

	agent := &models.Agent{
		ProjectID:            projectID,
		Name:                 input.Name,
		Role:                 input.Role,
		Model:                input.Model,
		ProviderConnectionID: &connID,
		SystemPrompt:         input.SystemPrompt,
		Temperature:          temp,
		MaxTokens:            maxTokens,
		Status:               "idle",
	}

	if err := s.db.Create(agent).Error; err != nil {
		return nil, fmt.Errorf("failed to create agent: %w", err)
	}

	return agent, nil
}

func (s *Service) List(projectID uuid.UUID) ([]models.Agent, error) {
	var agents []models.Agent
	if err := s.db.Where("project_id = ?", projectID).Order("created_at ASC").Find(&agents).Error; err != nil {
		return nil, fmt.Errorf("failed to list agents: %w", err)
	}
	return agents, nil
}

func (s *Service) GetByID(projectID, agentID uuid.UUID) (*models.Agent, error) {
	var agent models.Agent
	if err := s.db.Where("id = ? AND project_id = ?", agentID, projectID).First(&agent).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAgentNotFound
		}
		return nil, fmt.Errorf("failed to get agent: %w", err)
	}
	return &agent, nil
}

func (s *Service) Update(projectID, agentID uuid.UUID, input *UpdateInput) (*models.Agent, error) {
	agent, err := s.GetByID(projectID, agentID)
	if err != nil {
		return nil, err
	}

	updates := map[string]interface{}{}
	if input.Name != nil {
		updates["name"] = *input.Name
	}
	if input.Role != nil {
		updates["role"] = *input.Role
	}
	if input.Model != nil {
		updates["model"] = *input.Model
	}
	if input.SystemPrompt != nil {
		updates["system_prompt"] = *input.SystemPrompt
	}
	if input.Temperature != nil {
		updates["temperature"] = *input.Temperature
	}
	if input.MaxTokens != nil {
		updates["max_tokens"] = *input.MaxTokens
	}

	if len(updates) > 0 {
		if err := s.db.Model(agent).Updates(updates).Error; err != nil {
			return nil, fmt.Errorf("failed to update agent: %w", err)
		}
	}

	return s.GetByID(projectID, agentID)
}

func (s *Service) Delete(projectID, agentID uuid.UUID) error {
	result := s.db.Where("id = ? AND project_id = ?", agentID, projectID).Delete(&models.Agent{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete agent: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrAgentNotFound
	}
	return nil
}

func (s *Service) Invoke(userID, projectID, agentID uuid.UUID, input *InvokeInput) (*models.Message, error) {
	agent, err := s.GetByID(projectID, agentID)
	if err != nil {
		return nil, err
	}

	if agent.ProviderConnectionID == nil {
		return nil, fmt.Errorf("agent has no provider connection")
	}

	// Mark agent as running
	s.db.Model(agent).Updates(map[string]interface{}{
		"status":         "running",
		"last_heartbeat": time.Now(),
	})

	// Get the AI provider token
	token, err := s.providerService.GetDecryptedToken(userID, *agent.ProviderConnectionID)
	if err != nil {
		s.db.Model(agent).Update("status", "error")
		return nil, fmt.Errorf("failed to get provider token: %w", err)
	}

	// Determine provider type
	conn, err := s.providerService.GetByID(userID, *agent.ProviderConnectionID)
	if err != nil {
		s.db.Model(agent).Update("status", "error")
		return nil, fmt.Errorf("failed to get provider connection: %w", err)
	}

	aiProvider, err := CreateProvider(conn.ProviderType, token)
	if err != nil {
		s.db.Model(agent).Update("status", "error")
		return nil, fmt.Errorf("failed to create AI provider: %w", err)
	}

	// Build messages with memory context
	systemContent := agent.SystemPrompt
	if s.memoryService != nil {
		memories, _ := s.memoryService.GetProjectContext(projectID)
		if len(memories) > 0 {
			var sb strings.Builder
			for _, m := range memories {
				pinned := ""
				if m.Pinned {
					pinned = " [PINNED]"
				}
				sb.WriteString(fmt.Sprintf("- [%s/%s]%s: %s\n", m.Category, m.Key, pinned, m.Content))
			}
			systemContent += "\n\n## Project Context (Team Memory)\n" + sb.String()
		}
	}

	// Inject tool descriptions into the system prompt so the model knows what's available
	var toolCount int
	if s.toolProvider != nil {
		agentTools, _ := s.toolProvider.GetAgentTools(agentID)
		if len(agentTools) > 0 {
			toolCount = len(agentTools)
			var tb strings.Builder
			tb.WriteString("\n\n## Available Tools\nYou have the following tools available. To use a tool, include a JSON block in your response with the format: ```tool\n{\"tool\": \"<name>\", \"input\": {<params>}}\n```\n\n")
			for _, t := range agentTools {
				tb.WriteString(fmt.Sprintf("### %s (%s)\n%s\nDefinition: %s\n\n", t.Name, t.ToolType, t.Description, string(t.Definition)))
			}
			systemContent += tb.String()
		}
	}

	messages := []CompletionMsg{}
	if systemContent != "" {
		messages = append(messages, CompletionMsg{
			Role:    "system",
			Content: systemContent,
		})
	}
	messages = append(messages, CompletionMsg{
		Role:    "user",
		Content: input.Prompt,
	})

	// Call AI provider
	completion, err := aiProvider.Complete(&CompletionRequest{
		Model:       agent.Model,
		Messages:    messages,
		MaxTokens:   agent.MaxTokens,
		Temperature: agent.Temperature,
	})
	if err != nil {
		s.db.Model(agent).Update("status", "error")
		return nil, fmt.Errorf("AI completion failed: %w", err)
	}

	// Update token usage
	totalTokens := int64(completion.InputTokens + completion.OutputTokens)
	s.db.Model(agent).Updates(map[string]interface{}{
		"status":            "idle",
		"token_usage_total": gorm.Expr("token_usage_total + ?", totalTokens),
		"last_heartbeat":    time.Now(),
	})

	// Record analytics
	if s.usageRecorder != nil {
		s.usageRecorder.Record(userID, projectID, &agentID, conn.ProviderType, agent.Model, completion.InputTokens, completion.OutputTokens, toolCount)
	}

	// Post agent response to project chat
	msg, err := s.chatService.SendMessage(
		projectID,
		agent.ID,
		agent.Name,
		"agent",
		&chat.SendInput{
			Content:     completion.Content,
			MessageType: "agent_output",
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to post agent response to chat: %w", err)
	}

	return msg, nil
}
