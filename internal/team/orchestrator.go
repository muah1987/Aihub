package team

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/muah1987/Aihub/internal/agent"
	"github.com/muah1987/Aihub/internal/chat"
	"github.com/muah1987/Aihub/internal/memory"
	"github.com/muah1987/Aihub/internal/models"
	"github.com/muah1987/Aihub/internal/provider"
	"gorm.io/gorm"
)

type TaskDecomposition struct {
	Tasks []SubTask `json:"tasks"`
}

type SubTask struct {
	Title        string `json:"title"`
	Description  string `json:"description"`
	AssignedRole string `json:"assigned_role"`
}

type OrchestratorResult struct {
	Message *models.Message  `json:"message"`
	Tasks   []models.AgentTask `json:"tasks"`
}

type Orchestrator struct {
	db              *gorm.DB
	agentService    *agent.Service
	chatService     *chat.Service
	memoryService   *memory.Service
	providerService *provider.Service
}

func NewOrchestrator(
	db *gorm.DB,
	agentService *agent.Service,
	chatService *chat.Service,
	memoryService *memory.Service,
	providerService *provider.Service,
) *Orchestrator {
	return &Orchestrator{
		db:              db,
		agentService:    agentService,
		chatService:     chatService,
		memoryService:   memoryService,
		providerService: providerService,
	}
}

func (o *Orchestrator) InvokeTeam(ctx context.Context, userID, projectID, teamID uuid.UUID, prompt string) (*OrchestratorResult, error) {
	// Load team with members
	var team models.AgentTeam
	if err := o.db.First(&team, "id = ?", teamID).Error; err != nil {
		return nil, fmt.Errorf("team not found: %w", err)
	}

	var members []models.AgentTeamMember
	o.db.Preload("Agent").Where("team_id = ?", teamID).Find(&members)

	if team.LeaderAgentID == nil {
		return nil, fmt.Errorf("team has no leader agent")
	}

	var leader models.Agent
	if err := o.db.First(&leader, "id = ?", *team.LeaderAgentID).Error; err != nil {
		return nil, fmt.Errorf("leader agent not found: %w", err)
	}

	// Mark team as running
	o.db.Model(&team).Update("status", "running")

	// Post the prompt to chat
	o.chatService.SendSystemMessage(projectID, fmt.Sprintf("Team '%s' invoked with: %s", team.Name, prompt), "team_invocation")

	// Get memory context
	memoryContext := ""
	if o.memoryService != nil {
		memories, _ := o.memoryService.GetProjectContext(projectID)
		memoryContext = buildMemoryContext(memories)
	}

	// Step 1: Ask leader to decompose the task
	decomposition, err := o.decomposeTasks(ctx, userID, &leader, members, prompt, memoryContext)
	if err != nil {
		o.db.Model(&team).Update("status", "error")
		return nil, fmt.Errorf("task decomposition failed: %w", err)
	}

	// Step 2: Create task records
	var tasks []models.AgentTask
	for _, st := range decomposition.Tasks {
		assignedAgent := findAgentByRole(members, st.AssignedRole)
		task := models.AgentTask{
			TeamID:      teamID,
			Title:       st.Title,
			Description: st.Description,
			Prompt:      st.Description,
			Status:      "pending",
		}
		if assignedAgent != nil {
			task.AssignedAgentID = &assignedAgent.ID
		}
		o.db.Create(&task)
		tasks = append(tasks, task)
	}

	// Step 3: Execute based on strategy
	switch team.Strategy {
	case "parallel":
		o.executeParallel(ctx, userID, projectID, &leader, tasks, memoryContext)
	default:
		o.executeSequential(ctx, userID, projectID, &leader, tasks, memoryContext)
	}

	// Step 4: Synthesize results
	msg, err := o.synthesizeResults(ctx, userID, projectID, &leader, tasks, memoryContext)
	if err != nil {
		o.db.Model(&team).Update("status", "error")
		return nil, err
	}

	o.db.Model(&team).Update("status", "idle")

	return &OrchestratorResult{Message: msg, Tasks: tasks}, nil
}

func (o *Orchestrator) decomposeTasks(ctx context.Context, userID uuid.UUID, leader *models.Agent, members []models.AgentTeamMember, prompt, memoryContext string) (*TaskDecomposition, error) {
	systemPrompt := buildLeaderSystemPrompt(leader, members, memoryContext)

	decomposePrompt := fmt.Sprintf(`Analyze this task and decompose it into subtasks for your team members.
Output ONLY valid JSON in this format: {"tasks": [{"title": "...", "description": "...", "assigned_role": "..."}]}

Task: %s`, prompt)

	response, err := o.invokeAgent(ctx, userID, leader, systemPrompt, decomposePrompt)
	if err != nil {
		return nil, err
	}

	// Parse JSON response
	var decomp TaskDecomposition
	content := extractJSON(response)
	if err := json.Unmarshal([]byte(content), &decomp); err != nil {
		// Fallback: single task assigned to leader
		decomp = TaskDecomposition{
			Tasks: []SubTask{{
				Title:        "Complete task",
				Description:  prompt,
				AssignedRole: leader.Role,
			}},
		}
	}

	return &decomp, nil
}

func (o *Orchestrator) executeSequential(ctx context.Context, userID, projectID uuid.UUID, leader *models.Agent, tasks []models.AgentTask, memoryContext string) {
	for i := range tasks {
		o.executeTask(ctx, userID, projectID, &tasks[i], memoryContext)
	}
}

func (o *Orchestrator) executeParallel(ctx context.Context, userID, projectID uuid.UUID, leader *models.Agent, tasks []models.AgentTask, memoryContext string) {
	var wg sync.WaitGroup
	for i := range tasks {
		wg.Add(1)
		go func(t *models.AgentTask) {
			defer wg.Done()
			o.executeTask(ctx, userID, projectID, t, memoryContext)
		}(&tasks[i])
	}
	wg.Wait()
}

func (o *Orchestrator) executeTask(ctx context.Context, userID, projectID uuid.UUID, task *models.AgentTask, memoryContext string) {
	now := time.Now()
	o.db.Model(task).Updates(map[string]interface{}{"status": "running", "started_at": now})

	if task.AssignedAgentID == nil {
		o.db.Model(task).Updates(map[string]interface{}{"status": "failed", "result": "no agent assigned"})
		return
	}

	var assignedAgent models.Agent
	if err := o.db.First(&assignedAgent, "id = ?", *task.AssignedAgentID).Error; err != nil {
		o.db.Model(task).Updates(map[string]interface{}{"status": "failed", "result": "agent not found"})
		return
	}

	systemPrompt := assignedAgent.SystemPrompt
	if memoryContext != "" {
		systemPrompt += "\n\n## Project Context (Team Memory)\n" + memoryContext
	}

	response, err := o.invokeAgent(ctx, userID, &assignedAgent, systemPrompt, task.Prompt)
	if err != nil {
		o.db.Model(task).Updates(map[string]interface{}{"status": "failed", "result": err.Error()})
		return
	}

	completedAt := time.Now()
	o.db.Model(task).Updates(map[string]interface{}{
		"status":       "completed",
		"result":       response,
		"completed_at": completedAt,
	})

	o.chatService.SendSystemMessage(projectID,
		fmt.Sprintf("Agent '%s' completed: %s", assignedAgent.Name, task.Title),
		"agent_task_complete")
}

func (o *Orchestrator) synthesizeResults(ctx context.Context, userID, projectID uuid.UUID, leader *models.Agent, tasks []models.AgentTask, memoryContext string) (*models.Message, error) {
	// Reload tasks with results
	for i := range tasks {
		o.db.First(&tasks[i], "id = ?", tasks[i].ID)
	}

	synthesisPrompt := "Here are the results from your team members. Synthesize a coherent final response:\n\n"
	for _, t := range tasks {
		synthesisPrompt += fmt.Sprintf("### %s\nStatus: %s\nResult:\n%s\n\n", t.Title, t.Status, t.Result)
	}

	systemPrompt := leader.SystemPrompt
	if memoryContext != "" {
		systemPrompt += "\n\n## Project Context (Team Memory)\n" + memoryContext
	}

	response, err := o.invokeAgent(ctx, userID, leader, systemPrompt, synthesisPrompt)
	if err != nil {
		return nil, fmt.Errorf("synthesis failed: %w", err)
	}

	msg, err := o.chatService.SendMessage(projectID, leader.ID, leader.Name, "agent", &chat.SendInput{
		Content:     response,
		MessageType: "team_synthesis",
	})
	if err != nil {
		return nil, err
	}

	return msg, nil
}

func (o *Orchestrator) invokeAgent(ctx context.Context, userID uuid.UUID, agentModel *models.Agent, systemPrompt, userPrompt string) (string, error) {
	if agentModel.ProviderConnectionID == nil {
		return "", fmt.Errorf("agent has no provider connection")
	}

	token, err := o.providerService.GetDecryptedToken(userID, *agentModel.ProviderConnectionID)
	if err != nil {
		return "", fmt.Errorf("failed to get provider token: %w", err)
	}

	conn, err := o.providerService.GetByID(userID, *agentModel.ProviderConnectionID)
	if err != nil {
		return "", fmt.Errorf("failed to get provider connection: %w", err)
	}

	aiProvider, err := agent.CreateProvider(conn.ProviderType, token)
	if err != nil {
		return "", fmt.Errorf("failed to create AI provider: %w", err)
	}

	messages := []agent.CompletionMsg{}
	if systemPrompt != "" {
		messages = append(messages, agent.CompletionMsg{Role: "system", Content: systemPrompt})
	}
	messages = append(messages, agent.CompletionMsg{Role: "user", Content: userPrompt})

	completion, err := aiProvider.Complete(&agent.CompletionRequest{
		Model:       agentModel.Model,
		Messages:    messages,
		MaxTokens:   agentModel.MaxTokens,
		Temperature: agentModel.Temperature,
	})
	if err != nil {
		return "", err
	}

	// Track token usage
	totalTokens := int64(completion.InputTokens + completion.OutputTokens)
	o.db.Model(agentModel).Updates(map[string]interface{}{
		"token_usage_total": gorm.Expr("token_usage_total + ?", totalTokens),
		"last_heartbeat":    time.Now(),
	})

	return completion.Content, nil
}

func buildLeaderSystemPrompt(leader *models.Agent, members []models.AgentTeamMember, memoryContext string) string {
	var sb strings.Builder
	sb.WriteString("You are a team leader AI agent. ")
	if leader.SystemPrompt != "" {
		sb.WriteString(leader.SystemPrompt)
		sb.WriteString("\n\n")
	}

	sb.WriteString("Your team members are:\n")
	for _, m := range members {
		sb.WriteString(fmt.Sprintf("- Agent '%s' (role: %s, model: %s)\n", m.Agent.Name, m.Agent.Role, m.Agent.Model))
	}

	sb.WriteString("\nWhen decomposing tasks, assign them to team members by their role.")

	if memoryContext != "" {
		sb.WriteString("\n\n## Project Context (Team Memory)\n")
		sb.WriteString(memoryContext)
	}

	return sb.String()
}

func buildMemoryContext(memories []models.TeamMemory) string {
	if len(memories) == 0 {
		return ""
	}
	var sb strings.Builder
	for _, m := range memories {
		pinned := ""
		if m.Pinned {
			pinned = " [PINNED]"
		}
		sb.WriteString(fmt.Sprintf("- [%s/%s]%s: %s\n", m.Category, m.Key, pinned, m.Content))
	}
	return sb.String()
}

func findAgentByRole(members []models.AgentTeamMember, role string) *models.Agent {
	role = strings.ToLower(role)
	for _, m := range members {
		if strings.ToLower(m.Agent.Role) == role || strings.Contains(strings.ToLower(m.Agent.Role), role) {
			return &m.Agent
		}
	}
	// Fallback: return first member
	if len(members) > 0 {
		return &members[0].Agent
	}
	return nil
}

func extractJSON(s string) string {
	start := strings.Index(s, "{")
	end := strings.LastIndex(s, "}")
	if start >= 0 && end > start {
		return s[start : end+1]
	}
	return s
}
