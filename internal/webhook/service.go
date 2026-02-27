package webhook

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/muah1987/Aihub/internal/models"
	"gorm.io/gorm"
)

var (
	ErrWebhookNotFound = errors.New("webhook not found")
	ErrInvalidSignature = errors.New("invalid webhook signature")
)

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

func (s *Service) Create(projectID uuid.UUID, vpsTargetID *uuid.UUID, branch string) (*models.Webhook, error) {
	secret, err := generateSecret()
	if err != nil {
		return nil, fmt.Errorf("generate secret: %w", err)
	}
	if branch == "" {
		branch = "main"
	}
	wh := &models.Webhook{
		ProjectID:   projectID,
		VPSTargetID: vpsTargetID,
		Secret:      secret,
		Branch:      branch,
		Active:      true,
	}
	if err := s.db.Create(wh).Error; err != nil {
		return nil, fmt.Errorf("create webhook: %w", err)
	}
	return wh, nil
}

func (s *Service) List(projectID uuid.UUID) ([]models.Webhook, error) {
	var whs []models.Webhook
	if err := s.db.Where("project_id = ?", projectID).Order("created_at DESC").Find(&whs).Error; err != nil {
		return nil, err
	}
	return whs, nil
}

func (s *Service) GetByID(id uuid.UUID) (*models.Webhook, error) {
	var wh models.Webhook
	if err := s.db.First(&wh, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrWebhookNotFound
		}
		return nil, err
	}
	return &wh, nil
}

func (s *Service) GetSecret(id uuid.UUID) (string, error) {
	wh, err := s.GetByID(id)
	if err != nil {
		return "", err
	}
	return wh.Secret, nil
}

func (s *Service) SetActive(id uuid.UUID, active bool) error {
	return s.db.Model(&models.Webhook{}).Where("id = ?", id).Update("active", active).Error
}

func (s *Service) Delete(id, projectID uuid.UUID) error {
	res := s.db.Where("id = ? AND project_id = ?", id, projectID).Delete(&models.Webhook{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrWebhookNotFound
	}
	return nil
}

// ValidateGitHubSignature checks the X-Hub-Signature-256 header.
func (s *Service) ValidateGitHubSignature(webhookID uuid.UUID, payload []byte, signature string) error {
	wh, err := s.GetByID(webhookID)
	if err != nil {
		return err
	}
	if !wh.Active {
		return errors.New("webhook is inactive")
	}

	expected := computeHMAC([]byte(wh.Secret), payload)
	// signature is "sha256=<hex>"
	if len(signature) < 7 || signature[:7] != "sha256=" {
		return ErrInvalidSignature
	}
	if !hmac.Equal([]byte(signature[7:]), []byte(expected)) {
		return ErrInvalidSignature
	}
	return nil
}

// MatchesBranch returns true if the push ref matches the configured branch.
func MatchesBranch(ref, branch string) bool {
	// ref is like "refs/heads/main"
	return ref == "refs/heads/"+branch || ref == branch
}

func (s *Service) MarkTriggered(id uuid.UUID) {
	now := time.Now()
	s.db.Model(&models.Webhook{}).Where("id = ?", id).Update("last_triggered_at", &now)
}

func generateSecret() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func computeHMAC(secret, payload []byte) string {
	mac := hmac.New(sha256.New, secret)
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}
