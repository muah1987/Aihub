package analytics

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/muah1987/Aihub/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Known model costs in microcents (1/10000 cent) per token.
// These are approximate and can be extended.
var modelCosts = map[string]struct{ Input, Output int64 }{
	// OpenAI
	"gpt-4o":         {25, 100},  // $2.50/1M in, $10/1M out → 0.25 & 1.0 microcent/tok
	"gpt-4o-mini":    {2, 6},     // $0.15/1M in, $0.60/1M out
	"gpt-4-turbo":    {100, 300}, // $10/1M in, $30/1M out
	"gpt-4":          {300, 600},
	"gpt-3.5-turbo":  {5, 15},
	// Anthropic
	"claude-opus-4-20250514":      {150, 750},
	"claude-sonnet-4-20250514":    {30, 150},
	"claude-3-5-sonnet-20241022":  {30, 150},
	"claude-3-5-haiku-20241022":   {10, 50},
	"claude-3-opus-20240229":      {150, 750},
	"claude-3-haiku-20240307":     {2, 12},
}

type DailySummary struct {
	Date         string `json:"date"`
	TotalTokens  int64  `json:"total_tokens"`
	InputTokens  int64  `json:"input_tokens"`
	OutputTokens int64  `json:"output_tokens"`
	Cost         int64  `json:"cost_microcents"`
	Invocations  int64  `json:"invocations"`
}

type ModelBreakdown struct {
	Model        string `json:"model"`
	Provider     string `json:"provider"`
	TotalTokens  int64  `json:"total_tokens"`
	Cost         int64  `json:"cost_microcents"`
	Invocations  int64  `json:"invocations"`
}

type ProjectSummary struct {
	TotalTokens     int64  `json:"total_tokens"`
	TotalCost       int64  `json:"total_cost_microcents"`
	TotalInvocations int64 `json:"total_invocations"`
	ToolCallsTotal  int64  `json:"tool_calls_total"`
}

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

// Record logs a usage record and computes cost.
func (s *Service) Record(userID, projectID uuid.UUID, agentID *uuid.UUID, providerName, model string, inputTokens, outputTokens, toolCalls int) (*models.UsageRecord, error) {
	totalTokens := inputTokens + outputTokens
	costMicrocents := estimateCost(model, inputTokens, outputTokens)

	rec := &models.UsageRecord{
		UserID:         userID,
		ProjectID:      projectID,
		AgentID:        agentID,
		Provider:       providerName,
		Model:          model,
		InputTokens:    inputTokens,
		OutputTokens:   outputTokens,
		TotalTokens:    totalTokens,
		CostMicrocents: costMicrocents,
		ToolCallsCount: toolCalls,
	}

	if err := s.db.Create(rec).Error; err != nil {
		return nil, fmt.Errorf("failed to record usage: %w", err)
	}
	return rec, nil
}

func estimateCost(model string, inTok, outTok int) int64 {
	rates, ok := modelCosts[model]
	if !ok {
		// Fallback: generic cost for unknown models
		return int64(inTok+outTok) * 5 // ~0.5 cent/1K tokens
	}
	return int64(inTok)*rates.Input/10 + int64(outTok)*rates.Output/10
}

// ---------- Queries ----------

func (s *Service) ProjectSummaryForPeriod(projectID uuid.UUID, since time.Time) (*ProjectSummary, error) {
	var result ProjectSummary
	row := s.db.Model(&models.UsageRecord{}).
		Where("project_id = ? AND created_at >= ?", projectID, since).
		Select("COALESCE(SUM(total_tokens),0) as total_tokens, COALESCE(SUM(cost_microcents),0) as total_cost, COUNT(*) as total_invocations, COALESCE(SUM(tool_calls_count),0) as tool_calls_total").
		Row()
	if err := row.Scan(&result.TotalTokens, &result.TotalCost, &result.TotalInvocations, &result.ToolCallsTotal); err != nil {
		return nil, err
	}
	return &result, nil
}

func (s *Service) DailySummary(projectID uuid.UUID, days int) ([]DailySummary, error) {
	if days <= 0 {
		days = 30
	}
	since := time.Now().AddDate(0, 0, -days)
	var results []DailySummary
	err := s.db.Model(&models.UsageRecord{}).
		Where("project_id = ? AND created_at >= ?", projectID, since).
		Select("DATE(created_at) as date, SUM(total_tokens) as total_tokens, SUM(input_tokens) as input_tokens, SUM(output_tokens) as output_tokens, SUM(cost_microcents) as cost, COUNT(*) as invocations").
		Group("DATE(created_at)").
		Order("date ASC").
		Scan(&results).Error
	return results, err
}

func (s *Service) ModelBreakdown(projectID uuid.UUID, since time.Time) ([]ModelBreakdown, error) {
	var results []ModelBreakdown
	err := s.db.Model(&models.UsageRecord{}).
		Where("project_id = ? AND created_at >= ?", projectID, since).
		Select("model, provider, SUM(total_tokens) as total_tokens, SUM(cost_microcents) as cost, COUNT(*) as invocations").
		Group("model, provider").
		Order("cost DESC").
		Scan(&results).Error
	return results, err
}

func (s *Service) RecentRecords(projectID uuid.UUID, limit int) ([]models.UsageRecord, error) {
	if limit <= 0 {
		limit = 50
	}
	var records []models.UsageRecord
	err := s.db.Where("project_id = ?", projectID).Order("created_at DESC").Limit(limit).Find(&records).Error
	return records, err
}

// ---------- Budget ----------

func (s *Service) GetBudget(projectID uuid.UUID) (*models.CostBudget, error) {
	var b models.CostBudget
	if err := s.db.Where("project_id = ?", projectID).First(&b).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &b, nil
}

func (s *Service) SetBudget(projectID uuid.UUID, monthlyLimit int64, alertPct int) (*models.CostBudget, error) {
	b := &models.CostBudget{
		ProjectID:              projectID,
		MonthlyLimitMicrocents: monthlyLimit,
		AlertThresholdPct:      alertPct,
	}
	err := s.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "project_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"monthly_limit_microcents", "alert_threshold_pct", "updated_at"}),
	}).Create(b).Error
	if err != nil {
		return nil, err
	}
	return s.GetBudget(projectID)
}

// CheckBudget returns (currentSpend, limit, overBudget).
func (s *Service) CheckBudget(projectID uuid.UUID) (int64, int64, bool) {
	budget, err := s.GetBudget(projectID)
	if err != nil || budget == nil || budget.MonthlyLimitMicrocents == 0 {
		return 0, 0, false
	}
	// Get current month spend
	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	summary, err := s.ProjectSummaryForPeriod(projectID, monthStart)
	if err != nil {
		return 0, budget.MonthlyLimitMicrocents, false
	}
	return summary.TotalCost, budget.MonthlyLimitMicrocents, summary.TotalCost >= budget.MonthlyLimitMicrocents
}
