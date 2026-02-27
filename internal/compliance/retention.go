package compliance

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/muah1987/Aihub/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrPolicyNotFound        = errors.New("retention policy not found")
	ErrInvalidResourceType   = errors.New("invalid resource_type")
)

// validResourceTypes is the whitelist of allowed retention resource types.
var validResourceTypes = map[string]bool{
	"audit_logs":       true,
	"messages":         true,
	"notifications":    true,
	"outbound_events":  true,
	"server_metrics":   true,
}

// RetentionService manages data retention policies and cleanup.
type RetentionService struct {
	db *gorm.DB
}

func NewRetentionService(db *gorm.DB) *RetentionService {
	return &RetentionService{db: db}
}

type UpsertPolicyInput struct {
	ResourceType  string `json:"resource_type"`
	RetentionDays int    `json:"retention_days"`
	Enabled       bool   `json:"enabled"`
}

func (s *RetentionService) UpsertPolicy(orgID, projectID *uuid.UUID, userID *uuid.UUID, input *UpsertPolicyInput) (*models.RetentionPolicy, error) {
	if !validResourceTypes[input.ResourceType] {
		return nil, ErrInvalidResourceType
	}
	if input.RetentionDays < 1 {
		input.RetentionDays = 90
	}

	policy := &models.RetentionPolicy{
		OrganizationID: orgID,
		ProjectID:      projectID,
		ResourceType:   input.ResourceType,
		RetentionDays:  input.RetentionDays,
		Enabled:        input.Enabled,
		CreatedBy:      userID,
	}

	// Use raw SQL-friendly upsert since the unique constraint uses COALESCE
	var existing models.RetentionPolicy
	q := s.db.Where("resource_type = ?", input.ResourceType)
	if orgID != nil {
		q = q.Where("organization_id = ?", orgID)
	} else {
		q = q.Where("organization_id IS NULL")
	}
	if projectID != nil {
		q = q.Where("project_id = ?", projectID)
	} else {
		q = q.Where("project_id IS NULL")
	}

	if err := q.First(&existing).Error; err == nil {
		// Update existing
		s.db.Model(&existing).Updates(map[string]interface{}{
			"retention_days": input.RetentionDays,
			"enabled":        input.Enabled,
		})
		return &existing, nil
	}

	if err := s.db.Clauses(clause.OnConflict{DoNothing: true}).Create(policy).Error; err != nil {
		return nil, fmt.Errorf("failed to create policy: %w", err)
	}
	return policy, nil
}

func (s *RetentionService) ListPolicies(orgID, projectID *uuid.UUID) ([]models.RetentionPolicy, error) {
	q := s.db.Model(&models.RetentionPolicy{})
	if orgID != nil {
		q = q.Where("organization_id = ? OR organization_id IS NULL", orgID)
	}
	if projectID != nil {
		q = q.Where("project_id = ? OR project_id IS NULL", projectID)
	}
	var policies []models.RetentionPolicy
	err := q.Order("resource_type ASC").Find(&policies).Error
	return policies, err
}

func (s *RetentionService) DeletePolicy(policyID uuid.UUID) error {
	result := s.db.Where("id = ?", policyID).Delete(&models.RetentionPolicy{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrPolicyNotFound
	}
	return nil
}

// tableForResourceType maps resource types to their database tables and timestamp columns.
var tableForResourceType = map[string]struct {
	table     string
	timeCol   string
	scopeCol  string // "project_id" or "organization_id" or ""
}{
	"audit_logs":       {table: "audit_logs", timeCol: "created_at", scopeCol: "organization_id"},
	"messages":         {table: "messages", timeCol: "created_at", scopeCol: "project_id"},
	"notifications":    {table: "notifications", timeCol: "created_at", scopeCol: ""},
	"outbound_events":  {table: "outbound_events", timeCol: "created_at", scopeCol: "project_id"},
	"server_metrics":   {table: "server_metrics", timeCol: "recorded_at", scopeCol: ""},
	"activity_logs":    {table: "activity_logs", timeCol: "created_at", scopeCol: "project_id"},
}

// RunCleanup executes all enabled retention policies and deletes expired data.
func (s *RetentionService) RunCleanup() (int, error) {
	var policies []models.RetentionPolicy
	s.db.Where("enabled = true").Find(&policies)

	totalDeleted := 0

	for i := range policies {
		meta, ok := tableForResourceType[policies[i].ResourceType]
		if !ok {
			continue
		}

		cutoff := time.Now().Add(-time.Duration(policies[i].RetentionDays) * 24 * time.Hour)

		query := fmt.Sprintf("DELETE FROM %s WHERE %s < ?", meta.table, meta.timeCol)
		args := []interface{}{cutoff}

		if meta.scopeCol != "" {
			if policies[i].ProjectID != nil && meta.scopeCol == "project_id" {
				query += fmt.Sprintf(" AND %s = ?", meta.scopeCol)
				args = append(args, policies[i].ProjectID)
			} else if policies[i].OrganizationID != nil && meta.scopeCol == "organization_id" {
				query += fmt.Sprintf(" AND %s = ?", meta.scopeCol)
				args = append(args, policies[i].OrganizationID)
			}
		}

		result := s.db.Exec(query, args...)
		deleted := int(result.RowsAffected)
		totalDeleted += deleted

		now := time.Now()
		s.db.Model(&policies[i]).Updates(map[string]interface{}{
			"last_run_at":     &now,
			"records_deleted": policies[i].RecordsDeleted + deleted,
		})
	}

	return totalDeleted, nil
}
