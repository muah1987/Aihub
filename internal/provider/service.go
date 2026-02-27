package provider

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"

	"github.com/google/uuid"
	"github.com/muah1987/Aihub/internal/models"
	"gorm.io/gorm"
)

var (
	ErrProviderNotFound = errors.New("provider connection not found")
	ErrInvalidProvider  = errors.New("invalid provider type")
)

type CreateInput struct {
	ProviderType string `json:"provider_type" validate:"required"`
	ProviderName string `json:"provider_name" validate:"required"`
	AccessToken  string `json:"access_token" validate:"required"`
	Scopes       string `json:"scopes"`
}

type Service struct {
	db            *gorm.DB
	encryptionKey []byte
}

func NewService(db *gorm.DB, encryptionKeyHex string) (*Service, error) {
	key, err := hex.DecodeString(encryptionKeyHex)
	if err != nil || len(key) != 32 {
		// Fallback: use the string directly padded/truncated to 32 bytes
		key = make([]byte, 32)
		copy(key, []byte(encryptionKeyHex))
	}

	return &Service{
		db:            db,
		encryptionKey: key,
	}, nil
}

func (s *Service) Create(userID uuid.UUID, input *CreateInput) (*models.ProviderConnection, error) {
	validTypes := map[string]bool{
		"github": true, "gitlab": true, "bitbucket": true,
		"openai": true, "anthropic": true, "google": true,
		"huggingface": true, "cohere": true, "mistral": true,
	}
	if !validTypes[input.ProviderType] {
		return nil, ErrInvalidProvider
	}

	encrypted, err := s.encrypt(input.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt token: %w", err)
	}

	conn := &models.ProviderConnection{
		UserID:       userID,
		ProviderType: input.ProviderType,
		ProviderName: input.ProviderName,
		AccessToken:  encrypted,
		Scopes:       input.Scopes,
		Status:       "active",
	}

	if err := s.db.Create(conn).Error; err != nil {
		return nil, fmt.Errorf("failed to create provider connection: %w", err)
	}

	return conn, nil
}

func (s *Service) List(userID uuid.UUID) ([]models.ProviderConnection, error) {
	var connections []models.ProviderConnection
	if err := s.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&connections).Error; err != nil {
		return nil, fmt.Errorf("failed to list provider connections: %w", err)
	}
	return connections, nil
}

func (s *Service) GetByID(userID, connectionID uuid.UUID) (*models.ProviderConnection, error) {
	var conn models.ProviderConnection
	if err := s.db.Where("id = ? AND user_id = ?", connectionID, userID).First(&conn).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProviderNotFound
		}
		return nil, fmt.Errorf("failed to get provider connection: %w", err)
	}
	return &conn, nil
}

func (s *Service) Delete(userID, connectionID uuid.UUID) error {
	result := s.db.Where("id = ? AND user_id = ?", connectionID, userID).Delete(&models.ProviderConnection{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete provider connection: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrProviderNotFound
	}
	return nil
}

func (s *Service) GetDecryptedToken(userID, connectionID uuid.UUID) (string, error) {
	conn, err := s.GetByID(userID, connectionID)
	if err != nil {
		return "", err
	}

	token, err := s.decrypt(conn.AccessToken)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt token: %w", err)
	}

	return token, nil
}

func (s *Service) encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(s.encryptionKey)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := aesGCM.Seal(nonce, nonce, []byte(plaintext), nil)
	return hex.EncodeToString(ciphertext), nil
}

func (s *Service) decrypt(ciphertextHex string) (string, error) {
	ciphertext, err := hex.DecodeString(ciphertextHex)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(s.encryptionKey)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := aesGCM.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
