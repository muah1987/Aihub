package integration

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/muah1987/Aihub/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrConnectionNotFound = errors.New("integration connection not found")
	ErrEventNotFound      = errors.New("outbound event not found")
	ErrRuleNotFound       = errors.New("notification rule not found")
	ErrDigestNotFound     = errors.New("email digest not found")
	ErrUnsafeURL          = errors.New("webhook URL is not allowed: must be HTTPS and not target internal networks")
)

type Service struct {
	db            *gorm.DB
	httpClient    *http.Client
	encryptionKey []byte
}

func NewService(db *gorm.DB, encryptionKey string) *Service {
	return &Service{
		db:            db,
		encryptionKey: []byte(encryptionKey),
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// encrypt encrypts plaintext using AES-GCM with the service encryption key.
func (s *Service) encrypt(plaintext string) (string, error) {
	if len(s.encryptionKey) == 0 {
		return plaintext, nil
	}
	block, err := aes.NewCipher(s.encryptionKey)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decrypt decrypts base64-encoded AES-GCM ciphertext.
func (s *Service) decrypt(encoded string) (string, error) {
	if len(s.encryptionKey) == 0 {
		return encoded, nil
	}
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return encoded, nil // not encrypted (legacy data), return as-is
	}
	block, err := aes.NewCipher(s.encryptionKey)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return encoded, nil // not encrypted, return as-is
	}
	plaintext, err := gcm.Open(nil, data[:nonceSize], data[nonceSize:], nil)
	if err != nil {
		return encoded, nil // decryption failed (legacy data), return as-is
	}
	return string(plaintext), nil
}

// validateWebhookURL ensures the URL is HTTPS and does not target internal networks.
func validateWebhookURL(rawURL string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return ErrUnsafeURL
	}
	if parsed.Scheme != "https" {
		return ErrUnsafeURL
	}
	hostname := parsed.Hostname()
	if hostname == "" {
		return ErrUnsafeURL
	}
	// Block localhost variants
	lower := strings.ToLower(hostname)
	if lower == "localhost" || lower == "127.0.0.1" || lower == "::1" || lower == "0.0.0.0" {
		return ErrUnsafeURL
	}
	// Resolve DNS and block private/reserved IPs
	ips, err := net.LookupIP(hostname)
	if err != nil {
		return ErrUnsafeURL
	}
	for _, ip := range ips {
		if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
			return ErrUnsafeURL
		}
	}
	return nil
}

// ---- Integration Connections CRUD ----

type CreateConnectionInput struct {
	Platform    string `json:"platform" validate:"required"`
	Name        string `json:"name" validate:"required"`
	Config      map[string]interface{} `json:"config"`
	Credentials string `json:"credentials"`
	ChannelID   string `json:"channel_id"`
}

func (s *Service) CreateConnection(projectID uuid.UUID, userID *uuid.UUID, input *CreateConnectionInput) (*models.IntegrationConnection, error) {
	configJSON, err := json.Marshal(input.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	// Validate webhook URLs in config
	for _, key := range []string{"webhook_url", "url"} {
		if u, ok := input.Config[key]; ok {
			if uStr, ok := u.(string); ok && uStr != "" {
				if err := validateWebhookURL(uStr); err != nil {
					return nil, err
				}
			}
		}
	}

	// Encrypt credentials before storage
	encCreds := input.Credentials
	if input.Credentials != "" {
		encCreds, err = s.encrypt(input.Credentials)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt credentials: %w", err)
		}
	}

	conn := &models.IntegrationConnection{
		ProjectID:   projectID,
		CreatedBy:   userID,
		Platform:    input.Platform,
		Name:        input.Name,
		Config:      configJSON,
		Credentials: encCreds,
		ChannelID:   input.ChannelID,
		Status:      "connected",
	}

	if err := s.db.Create(conn).Error; err != nil {
		return nil, fmt.Errorf("failed to create connection: %w", err)
	}
	return conn, nil
}

func (s *Service) GetConnection(projectID, connID uuid.UUID) (*models.IntegrationConnection, error) {
	var conn models.IntegrationConnection
	err := s.db.Where("id = ? AND project_id = ?", connID, projectID).First(&conn).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrConnectionNotFound
		}
		return nil, err
	}
	return &conn, nil
}

func (s *Service) ListConnections(projectID uuid.UUID, platform string) ([]models.IntegrationConnection, error) {
	q := s.db.Where("project_id = ?", projectID)
	if platform != "" {
		q = q.Where("platform = ?", platform)
	}
	var conns []models.IntegrationConnection
	err := q.Order("created_at DESC").Find(&conns).Error
	return conns, err
}

func (s *Service) UpdateConnection(projectID, connID uuid.UUID, updates map[string]interface{}) error {
	result := s.db.Model(&models.IntegrationConnection{}).
		Where("id = ? AND project_id = ?", connID, projectID).
		Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrConnectionNotFound
	}
	return nil
}

func (s *Service) DeleteConnection(projectID, connID uuid.UUID) error {
	result := s.db.Where("id = ? AND project_id = ?", connID, projectID).Delete(&models.IntegrationConnection{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrConnectionNotFound
	}
	return nil
}

func (s *Service) TestConnection(projectID, connID uuid.UUID) (string, error) {
	conn, err := s.GetConnection(projectID, connID)
	if err != nil {
		return "", err
	}

	switch conn.Platform {
	case "slack":
		return s.testSlack(conn)
	case "discord":
		return s.testDiscord(conn)
	case "webhook", "zapier":
		return s.testWebhook(conn)
	default:
		return "", fmt.Errorf("unsupported platform: %s", conn.Platform)
	}
}

// ---- Dispatchers ----

func (s *Service) parseConfig(conn *models.IntegrationConnection) (map[string]string, error) {
	var config map[string]string
	if err := json.Unmarshal(conn.Config, &config); err != nil {
		return nil, fmt.Errorf("failed to parse connection config: %w", err)
	}
	return config, nil
}

func (s *Service) testSlack(conn *models.IntegrationConnection) (string, error) {
	config, err := s.parseConfig(conn)
	if err != nil {
		return "error", err
	}
	webhookURL := config["webhook_url"]
	if webhookURL == "" {
		return "error", fmt.Errorf("webhook_url not configured")
	}

	payload := map[string]string{"text": "Aihub integration test - connection successful!"}
	return s.postJSON(webhookURL, payload)
}

func (s *Service) testDiscord(conn *models.IntegrationConnection) (string, error) {
	config, err := s.parseConfig(conn)
	if err != nil {
		return "error", err
	}
	webhookURL := config["webhook_url"]
	if webhookURL == "" {
		return "error", fmt.Errorf("webhook_url not configured")
	}

	payload := map[string]string{"content": "Aihub integration test - connection successful!"}
	return s.postJSON(webhookURL, payload)
}

func (s *Service) testWebhook(conn *models.IntegrationConnection) (string, error) {
	config, err := s.parseConfig(conn)
	if err != nil {
		return "error", err
	}
	targetURL := config["url"]
	if targetURL == "" {
		return "error", fmt.Errorf("url not configured")
	}

	payload := map[string]interface{}{
		"event":     "test",
		"message":   "Aihub integration test",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}
	return s.postJSON(targetURL, payload)
}

// SendToIntegration dispatches a message to an integration connection.
func (s *Service) SendToIntegration(conn *models.IntegrationConnection, eventType string, payload map[string]interface{}) error {
	config, err := s.parseConfig(conn)
	if err != nil {
		return err
	}

	var url string
	var body interface{}

	switch conn.Platform {
	case "slack":
		url = config["webhook_url"]
		text := fmt.Sprintf("*[%s]* %s", eventType, formatPayloadText(payload))
		body = map[string]string{"text": text}
	case "discord":
		url = config["webhook_url"]
		text := fmt.Sprintf("**[%s]** %s", eventType, formatPayloadText(payload))
		body = map[string]string{"content": text}
	case "webhook", "zapier":
		url = config["url"]
		body = map[string]interface{}{
			"event":     eventType,
			"data":      payload,
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		}
	default:
		return fmt.Errorf("unsupported platform: %s", conn.Platform)
	}

	if url == "" {
		return fmt.Errorf("no URL configured for %s", conn.Platform)
	}

	_, err = s.postJSON(url, body)

	// Update last_used_at
	now := time.Now()
	s.db.Model(conn).Update("last_used_at", &now)

	return err
}

// DispatchEvent creates an outbound event and sends it.
func (s *Service) DispatchEvent(projectID uuid.UUID, eventType string, payload map[string]interface{}) error {
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Find all enabled integrations for this project
	var conns []models.IntegrationConnection
	s.db.Where("project_id = ? AND enabled = true AND status = 'connected'", projectID).Find(&conns)

	var lastErr error
	for i := range conns {
		event := &models.OutboundEvent{
			ProjectID:     projectID,
			IntegrationID: &conns[i].ID,
			EventType:     eventType,
			Payload:       payloadJSON,
			Status:        "pending",
			MaxAttempts:   3,
		}
		if err := s.db.Create(event).Error; err != nil {
			lastErr = err
			continue
		}

		if err := s.SendToIntegration(&conns[i], eventType, payload); err != nil {
			event.Status = "failed"
			event.Attempts = 1
			nextRetry := time.Now().Add(30 * time.Second)
			event.NextRetryAt = &nextRetry
			s.db.Save(event)
			lastErr = err
		} else {
			now := time.Now()
			event.Status = "sent"
			event.Attempts = 1
			event.SentAt = &now
			s.db.Save(event)
		}
	}

	return lastErr
}

// ListOutboundEvents returns recent outbound events for a project.
func (s *Service) ListOutboundEvents(projectID uuid.UUID, limit int) ([]models.OutboundEvent, error) {
	if limit <= 0 {
		limit = 50
	}
	var events []models.OutboundEvent
	err := s.db.Where("project_id = ?", projectID).
		Order("created_at DESC").Limit(limit).Find(&events).Error
	return events, err
}

// RetryFailedEvents retries events that are due for retry.
func (s *Service) RetryFailedEvents() (int, error) {
	var events []models.OutboundEvent
	s.db.Where("status IN ('pending', 'retrying') AND next_retry_at <= ? AND attempts < max_attempts",
		time.Now()).Find(&events)

	retried := 0
	for i := range events {
		var conn models.IntegrationConnection
		if events[i].IntegrationID != nil {
			if err := s.db.First(&conn, "id = ?", events[i].IntegrationID).Error; err != nil {
				continue
			}
		}

		var payload map[string]interface{}
		json.Unmarshal(events[i].Payload, &payload)

		events[i].Attempts++
		if err := s.SendToIntegration(&conn, events[i].EventType, payload); err != nil {
			if events[i].Attempts >= events[i].MaxAttempts {
				events[i].Status = "failed"
			} else {
				events[i].Status = "retrying"
				delay := time.Duration(events[i].Attempts*30) * time.Second
				nextRetry := time.Now().Add(delay)
				events[i].NextRetryAt = &nextRetry
			}
		} else {
			now := time.Now()
			events[i].Status = "sent"
			events[i].SentAt = &now
		}
		s.db.Save(&events[i])
		retried++
	}
	return retried, nil
}

// ---- Notification Rules ----

func (s *Service) CreateRule(userID uuid.UUID, projectID *uuid.UUID, eventType, channel, minSeverity string) (*models.NotificationRule, error) {
	if minSeverity == "" {
		minSeverity = "info"
	}
	rule := &models.NotificationRule{
		UserID:      userID,
		ProjectID:   projectID,
		EventType:   eventType,
		Channel:     channel,
		Enabled:     true,
		MinSeverity: minSeverity,
	}
	result := s.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "project_id"}, {Name: "event_type"}, {Name: "channel"}},
		DoUpdates: clause.AssignmentColumns([]string{"enabled", "min_severity", "updated_at"}),
	}).Create(rule)
	if result.Error != nil {
		return nil, result.Error
	}

	var fetched models.NotificationRule
	s.db.Where("user_id = ? AND event_type = ? AND channel = ?", userID, eventType, channel).First(&fetched)
	return &fetched, nil
}

func (s *Service) ListRules(userID uuid.UUID, projectID *uuid.UUID) ([]models.NotificationRule, error) {
	q := s.db.Where("user_id = ?", userID)
	if projectID != nil {
		q = q.Where("project_id = ? OR project_id IS NULL", projectID)
	}
	var rules []models.NotificationRule
	err := q.Order("event_type, channel").Find(&rules).Error
	return rules, err
}

func (s *Service) UpdateRule(userID, ruleID uuid.UUID, updates map[string]interface{}) error {
	result := s.db.Model(&models.NotificationRule{}).
		Where("id = ? AND user_id = ?", ruleID, userID).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrRuleNotFound
	}
	return nil
}

func (s *Service) DeleteRule(userID, ruleID uuid.UUID) error {
	result := s.db.Where("id = ? AND user_id = ?", ruleID, userID).Delete(&models.NotificationRule{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrRuleNotFound
	}
	return nil
}

// ShouldNotify checks whether a notification should be sent on a given channel.
func (s *Service) ShouldNotify(userID uuid.UUID, projectID *uuid.UUID, eventType, channel string) bool {
	var rule models.NotificationRule
	q := s.db.Where("user_id = ? AND channel = ? AND enabled = true", userID, channel)
	q = q.Where("event_type = ? OR event_type = '*'", eventType)
	if projectID != nil {
		q = q.Where("project_id = ? OR project_id IS NULL", projectID)
	}
	if err := q.First(&rule).Error; err != nil {
		// No rule found = allow by default for in_app, deny for external
		return channel == "in_app"
	}

	// Check quiet hours
	if rule.QuietHoursStart != nil && rule.QuietHoursEnd != nil {
		hour := time.Now().Hour()
		start := *rule.QuietHoursStart
		end := *rule.QuietHoursEnd
		if start < end {
			if hour >= start && hour < end {
				return false
			}
		} else {
			// Wraps midnight
			if hour >= start || hour < end {
				return false
			}
		}
	}

	return rule.Enabled
}

// ---- Email Digests ----

func (s *Service) UpsertDigest(userID uuid.UUID, projectID *uuid.UUID, input *EmailDigestInput) (*models.EmailDigest, error) {
	digest := &models.EmailDigest{
		UserID:               userID,
		ProjectID:            projectID,
		Frequency:            input.Frequency,
		IncludeDeployments:   input.IncludeDeployments,
		IncludeChatSummary:   input.IncludeChatSummary,
		IncludeAgentActivity: input.IncludeAgentActivity,
		IncludeMonitoring:    input.IncludeMonitoring,
		Enabled:              true,
	}

	// Calculate next send time
	now := time.Now()
	switch input.Frequency {
	case "daily":
		next := time.Date(now.Year(), now.Month(), now.Day()+1, 8, 0, 0, 0, now.Location())
		digest.NextSendAt = &next
	case "weekly":
		daysUntilMonday := (8 - int(now.Weekday())) % 7
		if daysUntilMonday == 0 {
			daysUntilMonday = 7
		}
		next := time.Date(now.Year(), now.Month(), now.Day()+daysUntilMonday, 8, 0, 0, 0, now.Location())
		digest.NextSendAt = &next
	}

	result := s.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "user_id"}, {Name: "project_id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"frequency", "include_deployments", "include_chat_summary",
			"include_agent_activity", "include_monitoring", "enabled",
			"next_send_at", "updated_at",
		}),
	}).Create(digest)

	if result.Error != nil {
		return nil, result.Error
	}

	var fetched models.EmailDigest
	q := s.db.Where("user_id = ?", userID)
	if projectID != nil {
		q = q.Where("project_id = ?", projectID)
	} else {
		q = q.Where("project_id IS NULL")
	}
	q.First(&fetched)
	return &fetched, nil
}

type EmailDigestInput struct {
	Frequency            string `json:"frequency"`
	IncludeDeployments   bool   `json:"include_deployments"`
	IncludeChatSummary   bool   `json:"include_chat_summary"`
	IncludeAgentActivity bool   `json:"include_agent_activity"`
	IncludeMonitoring    bool   `json:"include_monitoring"`
}

func (s *Service) GetDigest(userID uuid.UUID, projectID *uuid.UUID) (*models.EmailDigest, error) {
	var digest models.EmailDigest
	q := s.db.Where("user_id = ?", userID)
	if projectID != nil {
		q = q.Where("project_id = ?", projectID)
	} else {
		q = q.Where("project_id IS NULL")
	}
	if err := q.First(&digest).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDigestNotFound
		}
		return nil, err
	}
	return &digest, nil
}

func (s *Service) ListDigests(userID uuid.UUID) ([]models.EmailDigest, error) {
	var digests []models.EmailDigest
	err := s.db.Where("user_id = ?", userID).Find(&digests).Error
	return digests, err
}

func (s *Service) DeleteDigest(userID, digestID uuid.UUID) error {
	result := s.db.Where("id = ? AND user_id = ?", digestID, userID).Delete(&models.EmailDigest{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrDigestNotFound
	}
	return nil
}

// ---- Helpers ----

func (s *Service) postJSON(targetURL string, body interface{}) (string, error) {
	if err := validateWebhookURL(targetURL); err != nil {
		return "error", err
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return "error", err
	}

	resp, err := s.httpClient.Post(targetURL, "application/json", bytes.NewReader(payload))
	if err != nil {
		return "error", fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))

	if resp.StatusCode >= 400 {
		return "error", fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	return "ok", nil
}

func formatPayloadText(payload map[string]interface{}) string {
	if title, ok := payload["title"].(string); ok {
		if msg, ok := payload["message"].(string); ok {
			return fmt.Sprintf("%s: %s", title, msg)
		}
		return title
	}
	if msg, ok := payload["message"].(string); ok {
		return msg
	}
	data, _ := json.Marshal(payload)
	if len(data) > 200 {
		return string(data[:200]) + "..."
	}
	return string(data)
}
