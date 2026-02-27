package workflow

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/muah1987/Aihub/internal/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

var (
	ErrWorkflowNotFound = errors.New("workflow not found")
	ErrStepNotFound     = errors.New("workflow step not found")
	ErrRunNotFound      = errors.New("workflow run not found")
)

// StepExecutor is called for each step type during workflow execution.
type StepExecutor interface {
	ExecuteStep(stepType string, config json.RawMessage) (output string, err error)
}

type CreateWorkflowInput struct {
	Name          string          `json:"name"`
	Description   string          `json:"description"`
	TriggerType   string          `json:"trigger_type"`
	TriggerConfig json.RawMessage `json:"trigger_config"`
}

type CreateStepInput struct {
	Name       string          `json:"name"`
	StepOrder  int             `json:"step_order"`
	StepType   string          `json:"step_type"`
	Config     json.RawMessage `json:"config"`
	OnFailure  string          `json:"on_failure"`
	TimeoutSec int             `json:"timeout_sec"`
}

type Service struct {
	db       *gorm.DB
	executor StepExecutor
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

func (s *Service) SetExecutor(e StepExecutor) { s.executor = e }

// ---------- Workflow CRUD ----------

func (s *Service) Create(projectID uuid.UUID, input *CreateWorkflowInput) (*models.Workflow, error) {
	triggerType := input.TriggerType
	if triggerType == "" {
		triggerType = "manual"
	}
	triggerConfig := input.TriggerConfig
	if triggerConfig == nil {
		triggerConfig = json.RawMessage("{}")
	}

	wf := &models.Workflow{
		ProjectID:     projectID,
		Name:          input.Name,
		Description:   input.Description,
		TriggerType:   triggerType,
		TriggerConfig: datatypes.JSON(triggerConfig),
		Enabled:       true,
	}
	if err := s.db.Create(wf).Error; err != nil {
		return nil, fmt.Errorf("failed to create workflow: %w", err)
	}
	return wf, nil
}

func (s *Service) List(projectID uuid.UUID) ([]models.Workflow, error) {
	var wfs []models.Workflow
	err := s.db.Where("project_id = ?", projectID).Order("created_at DESC").Find(&wfs).Error
	return wfs, err
}

func (s *Service) GetByID(projectID, workflowID uuid.UUID) (*models.Workflow, error) {
	var wf models.Workflow
	err := s.db.Preload("Steps", func(db *gorm.DB) *gorm.DB {
		return db.Order("step_order ASC")
	}).Where("id = ? AND project_id = ?", workflowID, projectID).First(&wf).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrWorkflowNotFound
	}
	return &wf, err
}

func (s *Service) Update(projectID, workflowID uuid.UUID, updates map[string]interface{}) (*models.Workflow, error) {
	wf, err := s.GetByID(projectID, workflowID)
	if err != nil {
		return nil, err
	}
	if err := s.db.Model(wf).Updates(updates).Error; err != nil {
		return nil, err
	}
	return s.GetByID(projectID, workflowID)
}

func (s *Service) Delete(projectID, workflowID uuid.UUID) error {
	res := s.db.Where("id = ? AND project_id = ?", workflowID, projectID).Delete(&models.Workflow{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrWorkflowNotFound
	}
	return nil
}

func (s *Service) SetEnabled(projectID, workflowID uuid.UUID, enabled bool) error {
	res := s.db.Model(&models.Workflow{}).Where("id = ? AND project_id = ?", workflowID, projectID).Update("enabled", enabled)
	if res.RowsAffected == 0 {
		return ErrWorkflowNotFound
	}
	return res.Error
}

// ---------- Steps ----------

func (s *Service) AddStep(workflowID uuid.UUID, input *CreateStepInput) (*models.WorkflowStep, error) {
	onFailure := input.OnFailure
	if onFailure == "" {
		onFailure = "abort"
	}
	timeoutSec := input.TimeoutSec
	if timeoutSec <= 0 {
		timeoutSec = 300
	}
	cfg := input.Config
	if cfg == nil {
		cfg = json.RawMessage("{}")
	}

	step := &models.WorkflowStep{
		WorkflowID: workflowID,
		Name:       input.Name,
		StepOrder:  input.StepOrder,
		StepType:   input.StepType,
		Config:     datatypes.JSON(cfg),
		OnFailure:  onFailure,
		TimeoutSec: timeoutSec,
	}
	if err := s.db.Create(step).Error; err != nil {
		return nil, err
	}
	return step, nil
}

func (s *Service) ListSteps(workflowID uuid.UUID) ([]models.WorkflowStep, error) {
	var steps []models.WorkflowStep
	err := s.db.Where("workflow_id = ?", workflowID).Order("step_order ASC").Find(&steps).Error
	return steps, err
}

func (s *Service) UpdateStep(stepID uuid.UUID, updates map[string]interface{}) error {
	return s.db.Model(&models.WorkflowStep{}).Where("id = ?", stepID).Updates(updates).Error
}

func (s *Service) DeleteStep(stepID uuid.UUID) error {
	return s.db.Where("id = ?", stepID).Delete(&models.WorkflowStep{}).Error
}

// ---------- Execution ----------

func (s *Service) Trigger(projectID, workflowID uuid.UUID, userID *uuid.UUID, triggerType string) (*models.WorkflowRun, error) {
	wf, err := s.GetByID(projectID, workflowID)
	if err != nil {
		return nil, err
	}
	if !wf.Enabled {
		return nil, fmt.Errorf("workflow is disabled")
	}

	run := &models.WorkflowRun{
		WorkflowID:  workflowID,
		ProjectID:   projectID,
		TriggeredBy: userID,
		TriggerType: triggerType,
		Status:      "running",
		StartedAt:   time.Now(),
	}
	if err := s.db.Create(run).Error; err != nil {
		return nil, err
	}

	// Create step run entries
	for _, step := range wf.Steps {
		sr := &models.WorkflowStepRun{
			WorkflowRunID: run.ID,
			StepID:        step.ID,
			Status:        "pending",
		}
		s.db.Create(sr)
	}

	// Execute async
	go s.executeRun(run.ID, wf)

	return run, nil
}

func (s *Service) executeRun(runID uuid.UUID, wf *models.Workflow) {
	now := time.Now()
	finalStatus := "success"
	var finalErr string

	for _, step := range wf.Steps {
		// Find the step run record
		var sr models.WorkflowStepRun
		if err := s.db.Where("workflow_run_id = ? AND step_id = ?", runID, step.ID).First(&sr).Error; err != nil {
			continue
		}

		stepStart := time.Now()
		s.db.Model(&sr).Updates(map[string]interface{}{"status": "running", "started_at": stepStart})

		var output string
		var execErr error

		if s.executor != nil {
			output, execErr = s.executor.ExecuteStep(step.StepType, json.RawMessage(step.Config))
		} else {
			output = "no executor configured"
			execErr = fmt.Errorf("no step executor registered")
		}

		stepEnd := time.Now()
		if execErr != nil {
			errMsg := execErr.Error()
			s.db.Model(&sr).Updates(map[string]interface{}{
				"status":        "failed",
				"output":        output,
				"finished_at":   stepEnd,
				"error_message": errMsg,
			})

			if step.OnFailure == "abort" {
				finalStatus = "failed"
				finalErr = fmt.Sprintf("step %q failed: %s", step.Name, errMsg)
				// Mark remaining steps as skipped
				s.db.Model(&models.WorkflowStepRun{}).
					Where("workflow_run_id = ? AND status = ?", runID, "pending").
					Update("status", "skipped")
				break
			}
			// on_failure == "continue": keep going
		} else {
			s.db.Model(&sr).Updates(map[string]interface{}{
				"status":      "success",
				"output":      output,
				"finished_at": stepEnd,
			})
		}
	}

	// Update workflow run
	updates := map[string]interface{}{
		"status":      finalStatus,
		"finished_at": time.Now(),
	}
	if finalErr != "" {
		updates["error_message"] = finalErr
	}
	s.db.Model(&models.WorkflowRun{}).Where("id = ?", runID).Updates(updates)

	// Update workflow last run
	s.db.Model(&models.Workflow{}).Where("id = ?", wf.ID).Updates(map[string]interface{}{
		"last_run_at":     now,
		"last_run_status": finalStatus,
	})
}

// ---------- Run queries ----------

func (s *Service) ListRuns(workflowID uuid.UUID, limit int) ([]models.WorkflowRun, error) {
	if limit <= 0 {
		limit = 20
	}
	var runs []models.WorkflowRun
	err := s.db.Where("workflow_id = ?", workflowID).Order("started_at DESC").Limit(limit).Find(&runs).Error
	return runs, err
}

func (s *Service) GetRun(runID uuid.UUID) (*models.WorkflowRun, error) {
	var run models.WorkflowRun
	err := s.db.Preload("StepRuns").Where("id = ?", runID).First(&run).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrRunNotFound
	}
	return &run, err
}

func (s *Service) ListProjectRuns(projectID uuid.UUID, limit int) ([]models.WorkflowRun, error) {
	if limit <= 0 {
		limit = 20
	}
	var runs []models.WorkflowRun
	err := s.db.Where("project_id = ?", projectID).Order("started_at DESC").Limit(limit).Find(&runs).Error
	return runs, err
}
