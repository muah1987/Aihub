package compliance

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/muah1987/Aihub/internal/models"
	"gorm.io/gorm"
)

// AuditService handles audit log creation and querying.
type AuditService struct {
	db *gorm.DB
}

func NewAuditService(db *gorm.DB) *AuditService {
	return &AuditService{db: db}
}

type AuditEntry struct {
	Action       string      `json:"action"`
	ResourceType string      `json:"resource_type"`
	ResourceID   *uuid.UUID  `json:"resource_id,omitempty"`
	Details      interface{} `json:"details,omitempty"`
	Severity     string      `json:"severity,omitempty"`
}

// Log creates an audit log entry, extracting IP and user-agent from the request.
func (s *AuditService) Log(r *http.Request, orgID, projectID, userID *uuid.UUID, entry AuditEntry) {
	if entry.Severity == "" {
		entry.Severity = "info"
	}
	detailsJSON, _ := json.Marshal(entry.Details)

	var ip string
	if r != nil {
		ip = r.Header.Get("X-Forwarded-For")
		if ip != "" {
			// Take only the first IP in the chain (client IP)
			if idx := strings.Index(ip, ","); idx != -1 {
				ip = strings.TrimSpace(ip[:idx])
			}
		} else {
			ip = r.RemoteAddr
		}
	}
	var ua string
	if r != nil {
		ua = r.Header.Get("User-Agent")
	}

	log := &models.AuditLog{
		OrganizationID: orgID,
		ProjectID:      projectID,
		UserID:         userID,
		Action:         entry.Action,
		ResourceType:   entry.ResourceType,
		ResourceID:     entry.ResourceID,
		Details:        detailsJSON,
		IPAddress:      ip,
		UserAgent:      ua,
		Severity:       entry.Severity,
	}

	if err := s.db.Create(log).Error; err != nil {
		fmt.Printf("audit: failed to write log: %v\n", err)
	}
}

// LogSimple is a convenience method for logging without an HTTP request context.
func (s *AuditService) LogSimple(orgID, projectID, userID *uuid.UUID, action, resourceType string, resourceID *uuid.UUID, details interface{}) {
	detailsJSON, _ := json.Marshal(details)
	log := &models.AuditLog{
		OrganizationID: orgID,
		ProjectID:      projectID,
		UserID:         userID,
		Action:         action,
		ResourceType:   resourceType,
		ResourceID:     resourceID,
		Details:        detailsJSON,
		Severity:       "info",
	}
	s.db.Create(log)
}

type AuditFilter struct {
	OrganizationID *uuid.UUID
	ProjectID      *uuid.UUID
	UserID         *uuid.UUID
	Action         string
	ResourceType   string
	Severity       string
	Since          *time.Time
	Until          *time.Time
	Limit          int
	Offset         int
}

// Query retrieves audit logs with filtering and pagination.
func (s *AuditService) Query(filter AuditFilter) ([]models.AuditLog, int64, error) {
	q := s.db.Model(&models.AuditLog{})

	if filter.OrganizationID != nil {
		q = q.Where("organization_id = ?", filter.OrganizationID)
	}
	if filter.ProjectID != nil {
		q = q.Where("project_id = ?", filter.ProjectID)
	}
	if filter.UserID != nil {
		q = q.Where("user_id = ?", filter.UserID)
	}
	if filter.Action != "" {
		q = q.Where("action = ?", filter.Action)
	}
	if filter.ResourceType != "" {
		q = q.Where("resource_type = ?", filter.ResourceType)
	}
	if filter.Severity != "" {
		q = q.Where("severity = ?", filter.Severity)
	}
	if filter.Since != nil {
		q = q.Where("created_at >= ?", filter.Since)
	}
	if filter.Until != nil {
		q = q.Where("created_at <= ?", filter.Until)
	}

	var total int64
	q.Count(&total)

	if filter.Limit <= 0 {
		filter.Limit = 50
	}
	if filter.Limit > 500 {
		filter.Limit = 500
	}

	var logs []models.AuditLog
	err := q.Order("created_at DESC").Offset(filter.Offset).Limit(filter.Limit).Find(&logs).Error
	return logs, total, err
}
