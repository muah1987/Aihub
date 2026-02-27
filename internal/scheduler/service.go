package scheduler

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/muah1987/Aihub/internal/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

var (
	ErrJobNotFound = errors.New("scheduled job not found")
)

type CreateJobInput struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	CronExpr    string          `json:"cron_expr"`
	JobType     string          `json:"job_type"`
	Config      json.RawMessage `json:"config"`
}

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

// ---------- CRUD ----------

func (s *Service) Create(projectID uuid.UUID, input *CreateJobInput) (*models.ScheduledJob, error) {
	if input.CronExpr == "" {
		return nil, fmt.Errorf("cron_expr is required")
	}
	cfg := input.Config
	if cfg == nil {
		cfg = json.RawMessage("{}")
	}

	nextRun := nextCronRun(input.CronExpr)

	job := &models.ScheduledJob{
		ProjectID:   projectID,
		Name:        input.Name,
		Description: input.Description,
		CronExpr:    input.CronExpr,
		JobType:     input.JobType,
		Config:      datatypes.JSON(cfg),
		Enabled:     true,
		NextRunAt:   nextRun,
	}
	if err := s.db.Create(job).Error; err != nil {
		return nil, fmt.Errorf("failed to create job: %w", err)
	}
	return job, nil
}

func (s *Service) List(projectID uuid.UUID) ([]models.ScheduledJob, error) {
	var jobs []models.ScheduledJob
	err := s.db.Where("project_id = ?", projectID).Order("created_at DESC").Find(&jobs).Error
	return jobs, err
}

func (s *Service) GetByID(projectID, jobID uuid.UUID) (*models.ScheduledJob, error) {
	var job models.ScheduledJob
	err := s.db.Where("id = ? AND project_id = ?", jobID, projectID).First(&job).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrJobNotFound
	}
	return &job, err
}

func (s *Service) Update(projectID, jobID uuid.UUID, updates map[string]interface{}) (*models.ScheduledJob, error) {
	job, err := s.GetByID(projectID, jobID)
	if err != nil {
		return nil, err
	}
	if cronExpr, ok := updates["cron_expr"].(string); ok && cronExpr != "" {
		updates["next_run_at"] = nextCronRun(cronExpr)
	}
	if err := s.db.Model(job).Updates(updates).Error; err != nil {
		return nil, err
	}
	return s.GetByID(projectID, jobID)
}

func (s *Service) Delete(projectID, jobID uuid.UUID) error {
	res := s.db.Where("id = ? AND project_id = ?", jobID, projectID).Delete(&models.ScheduledJob{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrJobNotFound
	}
	return nil
}

func (s *Service) SetEnabled(projectID, jobID uuid.UUID, enabled bool) error {
	updates := map[string]interface{}{"enabled": enabled}
	if enabled {
		var job models.ScheduledJob
		if err := s.db.Where("id = ?", jobID).First(&job).Error; err == nil {
			updates["next_run_at"] = nextCronRun(job.CronExpr)
		}
	}
	return s.db.Model(&models.ScheduledJob{}).Where("id = ? AND project_id = ?", jobID, projectID).Updates(updates).Error
}

// ---------- Run history ----------

func (s *Service) ListRuns(jobID uuid.UUID, limit int) ([]models.ScheduledJobRun, error) {
	if limit <= 0 {
		limit = 20
	}
	var runs []models.ScheduledJobRun
	err := s.db.Where("job_id = ?", jobID).Order("started_at DESC").Limit(limit).Find(&runs).Error
	return runs, err
}

func (s *Service) RecordRun(jobID uuid.UUID, status, output string, errMsg *string) (*models.ScheduledJobRun, error) {
	now := time.Now()
	run := &models.ScheduledJobRun{
		JobID:        jobID,
		Status:       status,
		Output:       output,
		StartedAt:    now,
		FinishedAt:   &now,
		ErrorMessage: errMsg,
	}
	if err := s.db.Create(run).Error; err != nil {
		return nil, err
	}

	// Update job last_run info
	s.db.Model(&models.ScheduledJob{}).Where("id = ?", jobID).Updates(map[string]interface{}{
		"last_run_at":     now,
		"last_run_status": status,
	})

	return run, nil
}

// GetDueJobs returns jobs that are due for execution.
func (s *Service) GetDueJobs() ([]models.ScheduledJob, error) {
	var jobs []models.ScheduledJob
	err := s.db.Where("enabled = true AND next_run_at <= ?", time.Now()).Find(&jobs).Error
	return jobs, err
}

// AdvanceNextRun computes and sets the next run time for a job.
func (s *Service) AdvanceNextRun(jobID uuid.UUID, cronExpr string) {
	next := nextCronRun(cronExpr)
	s.db.Model(&models.ScheduledJob{}).Where("id = ?", jobID).Update("next_run_at", next)
}

// nextCronRun is a simplified cron parser for standard 5-field expressions.
// Returns the next run time from now. For production use, this should be
// replaced with a proper cron library.
func nextCronRun(expr string) *time.Time {
	fields := strings.Fields(expr)
	if len(fields) != 5 {
		t := time.Now().Add(1 * time.Hour)
		return &t
	}

	// Parse minute field for simple interval detection
	now := time.Now().UTC()
	minute := now.Minute()
	hour := now.Hour()

	if fields[0] != "*" {
		m, err := strconv.Atoi(fields[0])
		if err == nil {
			minute = m
		}
	}
	if fields[1] != "*" {
		h, err := strconv.Atoi(fields[1])
		if err == nil {
			hour = h
		}
	}

	next := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, time.UTC)
	if next.Before(now) {
		next = next.Add(24 * time.Hour)
	}
	return &next
}
