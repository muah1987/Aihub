package monitoring

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/muah1987/Aihub/internal/models"
	"gorm.io/gorm"
)

// SSHRunner is a minimal interface so monitoring can call the deployment SSH logic.
type SSHRunner interface {
	RunSSHCommand(targetID uuid.UUID, cmd string) (string, error)
}

type Service struct {
	db        *gorm.DB
	sshRunner SSHRunner
}

func NewService(db *gorm.DB, runner SSHRunner) *Service {
	return &Service{db: db, sshRunner: runner}
}

// CollectMetrics SSHes into the VPS and records a metric snapshot.
func (s *Service) CollectMetrics(projectID, targetID uuid.UUID) (*models.ServerMetric, error) {
	// Run a single compound command that outputs parseable data
	script := `
cpu=$(top -bn1 | grep "Cpu(s)" | awk '{print $2}' | cut -d'%' -f1 2>/dev/null || echo "0")
mem=$(free | awk '/^Mem/{printf "%.1f", $3/$2*100}' 2>/dev/null || echo "0")
disk=$(df / | awk 'NR==2{printf "%.1f", $5}' | tr -d '%' 2>/dev/null || echo "0")
load=$(cat /proc/loadavg | awk '{print $1,$2,$3}' 2>/dev/null || echo "0 0 0")
uptime=$(awk '{print int($1)}' /proc/uptime 2>/dev/null || echo "0")
echo "cpu=$cpu mem=$mem disk=$disk load=$load uptime=$uptime"
`
	out, err := s.sshRunner.RunSSHCommand(targetID, script)
	if err != nil {
		return nil, fmt.Errorf("collect metrics: %w", err)
	}

	m := &models.ServerMetric{
		VPSTargetID: targetID,
		RecordedAt:  time.Now(),
	}
	for _, pair := range strings.Fields(strings.TrimSpace(out)) {
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) != 2 {
			continue
		}
		switch kv[0] {
		case "cpu":
			if v, err := strconv.ParseFloat(kv[1], 32); err == nil {
				m.CPUPercent = float32(v)
			}
		case "mem":
			if v, err := strconv.ParseFloat(kv[1], 32); err == nil {
				m.MemPercent = float32(v)
			}
		case "disk":
			if v, err := strconv.ParseFloat(kv[1], 32); err == nil {
				m.DiskPercent = float32(v)
			}
		case "uptime":
			if v, err := strconv.ParseInt(kv[1], 10, 64); err == nil {
				m.UptimeSeconds = v
			}
		case "load":
			m.LoadAvg = kv[1]
		}
	}

	if err := s.db.Create(m).Error; err != nil {
		return nil, fmt.Errorf("save metric: %w", err)
	}
	return m, nil
}

// LatestMetric returns the most recent snapshot for a target.
func (s *Service) LatestMetric(targetID uuid.UUID) (*models.ServerMetric, error) {
	var m models.ServerMetric
	if err := s.db.Where("vps_target_id = ?", targetID).
		Order("recorded_at DESC").First(&m).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

// MetricsHistory returns the last N snapshots for graphing.
func (s *Service) MetricsHistory(targetID uuid.UUID, limit int) ([]models.ServerMetric, error) {
	var ms []models.ServerMetric
	if err := s.db.Where("vps_target_id = ?", targetID).
		Order("recorded_at DESC").Limit(limit).Find(&ms).Error; err != nil {
		return nil, err
	}
	// Reverse so oldest is first (better for charts)
	for i, j := 0, len(ms)-1; i < j; i, j = i+1, j-1 {
		ms[i], ms[j] = ms[j], ms[i]
	}
	return ms, nil
}

// Pipeline stage helpers
func (s *Service) ListStages(targetID uuid.UUID) ([]models.PipelineStage, error) {
	var stages []models.PipelineStage
	if err := s.db.Where("vps_target_id = ?", targetID).
		Order("stage_order ASC").Find(&stages).Error; err != nil {
		return nil, err
	}
	return stages, nil
}

func (s *Service) CreateStage(targetID uuid.UUID, name, command, onFailure string, order int) (*models.PipelineStage, error) {
	if onFailure == "" {
		onFailure = "abort"
	}
	st := &models.PipelineStage{
		VPSTargetID: targetID,
		Name:        name,
		Command:     command,
		StageOrder:  order,
		OnFailure:   onFailure,
	}
	if err := s.db.Create(st).Error; err != nil {
		return nil, err
	}
	return st, nil
}

func (s *Service) DeleteStage(id uuid.UUID) error {
	return s.db.Delete(&models.PipelineStage{}, "id = ?", id).Error
}

func (s *Service) UpdateStageOrder(id uuid.UUID, order int) error {
	return s.db.Model(&models.PipelineStage{}).Where("id = ?", id).Update("stage_order", order).Error
}
