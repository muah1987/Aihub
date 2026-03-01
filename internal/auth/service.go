package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/muah1987/Aihub/internal/email"
	"github.com/muah1987/Aihub/internal/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrEmailTaken         = errors.New("email already registered")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserNotFound       = errors.New("user not found")
	ErrTokenExpired       = errors.New("verification token expired")
	ErrTokenUsed          = errors.New("verification token already used")
	ErrTokenNotFound      = errors.New("verification token not found")
	ErrInvalid2FACode     = errors.New("invalid 2FA code")
)

type RegisterInput struct {
	Email       string `json:"email" validate:"required,email"`
	Password    string `json:"password" validate:"required,min=8"`
	DisplayName string `json:"display_name" validate:"required,min=2,max=100"`
}

type LoginInput struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type LoginResult struct {
	User         *models.User `json:"user,omitempty"`
	Tokens       *TokenPair   `json:"tokens,omitempty"`
	Requires2FA  bool         `json:"requires_2fa,omitempty"`
	PendingToken string       `json:"pending_token,omitempty"`
}

type Service struct {
	db           *gorm.DB
	jwt          *JWTService
	emailService *email.Service
	totp         *TOTPService
}

func NewService(db *gorm.DB, jwt *JWTService, emailService *email.Service) *Service {
	return &Service{
		db:           db,
		jwt:          jwt,
		emailService: emailService,
		totp:         NewTOTPService(),
	}
}

func (s *Service) Register(input *RegisterInput) (*models.User, *TokenPair, error) {
	var existing models.User
	if err := s.db.Where("email = ?", input.Email).First(&existing).Error; err == nil {
		return nil, nil, ErrEmailTaken
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &models.User{
		Email:        input.Email,
		PasswordHash: string(hash),
		DisplayName:  input.DisplayName,
		Role:         "user",
	}

	if err := s.db.Create(user).Error; err != nil {
		return nil, nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Send verification email
	if s.emailService != nil {
		_ = s.SendVerificationEmail(user.ID)
	}

	tokens, err := s.jwt.GenerateTokenPair(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, nil, err
	}

	return user, tokens, nil
}

func (s *Service) Login(input *LoginInput) (*LoginResult, error) {
	var user models.User
	if err := s.db.Where("email = ?", input.Email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	// If 2FA is enabled, return a pending token
	if user.TwoFactorEnabled {
		pendingToken, err := s.jwt.GeneratePendingToken(user.ID, user.Email, user.Role)
		if err != nil {
			return nil, fmt.Errorf("failed to generate pending token: %w", err)
		}
		return &LoginResult{
			Requires2FA:  true,
			PendingToken: pendingToken,
		}, nil
	}

	tokens, err := s.jwt.GenerateTokenPair(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, err
	}

	return &LoginResult{
		User:   &user,
		Tokens: tokens,
	}, nil
}

func (s *Service) LoginVerify2FA(pendingToken, code string) (*models.User, *TokenPair, error) {
	claims, err := s.jwt.ValidateToken(pendingToken)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid pending token: %w", err)
	}
	if claims.Type != PendingTwoFactor {
		return nil, nil, errors.New("invalid token type")
	}

	var user models.User
	if err := s.db.First(&user, "id = ?", claims.UserID).Error; err != nil {
		return nil, nil, ErrUserNotFound
	}

	valid, err := s.Validate2FA(user.ID, user.TwoFactorSecret, code)
	if err != nil {
		return nil, nil, err
	}
	if !valid {
		return nil, nil, ErrInvalid2FACode
	}

	tokens, err := s.jwt.GenerateTokenPair(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, nil, err
	}

	return &user, tokens, nil
}

func (s *Service) RefreshTokens(refreshTokenStr string) (*models.User, *TokenPair, error) {
	claims, err := s.jwt.ValidateToken(refreshTokenStr)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	if claims.Type != RefreshToken {
		return nil, nil, errors.New("token is not a refresh token")
	}

	var user models.User
	if err := s.db.First(&user, "id = ?", claims.UserID).Error; err != nil {
		return nil, nil, ErrUserNotFound
	}

	tokens, err := s.jwt.GenerateTokenPair(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, nil, err
	}

	return &user, tokens, nil
}

func (s *Service) GetUserByID(id string) (*models.User, error) {
	var user models.User
	if err := s.db.First(&user, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	return &user, nil
}

// Email verification

func (s *Service) SendVerificationEmail(userID uuid.UUID) error {
	var user models.User
	if err := s.db.First(&user, "id = ?", userID).Error; err != nil {
		return ErrUserNotFound
	}

	// Invalidate previous tokens
	s.db.Model(&models.EmailVerificationToken{}).
		Where("user_id = ? AND used = false", userID).
		Update("used", true)

	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return fmt.Errorf("failed to generate token: %w", err)
	}
	tokenStr := hex.EncodeToString(tokenBytes)

	token := &models.EmailVerificationToken{
		UserID:    userID,
		Token:     tokenStr,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	if err := s.db.Create(token).Error; err != nil {
		return fmt.Errorf("failed to create verification token: %w", err)
	}

	if s.emailService != nil {
		return s.emailService.SendVerificationEmail(user.Email, user.DisplayName, tokenStr)
	}
	return nil
}

func (s *Service) VerifyEmail(tokenStr string) error {
	var token models.EmailVerificationToken
	if err := s.db.Where("token = ?", tokenStr).First(&token).Error; err != nil {
		return ErrTokenNotFound
	}

	if token.Used {
		return ErrTokenUsed
	}

	if time.Now().After(token.ExpiresAt) {
		return ErrTokenExpired
	}

	tx := s.db.Begin()

	if err := tx.Model(&token).Update("used", true).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update token: %w", err)
	}

	if err := tx.Model(&models.User{}).Where("id = ?", token.UserID).Update("email_verified", true).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to verify email: %w", err)
	}

	return tx.Commit().Error
}

// Two-Factor Authentication

func (s *Service) Setup2FA(userID uuid.UUID) (secret, qrURL string, backupCodes []string, err error) {
	var user models.User
	if err := s.db.First(&user, "id = ?", userID).Error; err != nil {
		return "", "", nil, ErrUserNotFound
	}

	secret, err = s.totp.GenerateSecret()
	if err != nil {
		return "", "", nil, err
	}

	qrURL = s.totp.GenerateQRURL(user.Email, secret)

	// Generate backup codes before opening the transaction so we can return
	// them even if no DB writes have occurred yet.
	backupCodes = make([]string, 10)
	for i := range backupCodes {
		codeBytes := make([]byte, 4)
		if _, err := rand.Read(codeBytes); err != nil {
			return "", "", nil, fmt.Errorf("failed to generate backup code: %w", err)
		}
		backupCodes[i] = hex.EncodeToString(codeBytes)
	}

	// Hash all backup codes before opening the transaction.
	hashes := make([]string, len(backupCodes))
	for i, code := range backupCodes {
		hash, err := bcrypt.GenerateFromPassword([]byte(code), bcrypt.DefaultCost)
		if err != nil {
			return "", "", nil, fmt.Errorf("failed to hash backup code: %w", err)
		}
		hashes[i] = string(hash)
	}

	// Persist secret and backup codes atomically.
	tx := s.db.Begin()
	if tx.Error != nil {
		return "", "", nil, fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}
	defer tx.Rollback() //nolint:errcheck // no-op after Commit; original error takes precedence

	// Store the secret temporarily (not enabled yet until confirmed)
	if err := tx.Model(&user).Update("two_factor_secret", secret).Error; err != nil {
		return "", "", nil, fmt.Errorf("failed to save 2FA secret: %w", err)
	}

	// Delete old backup codes and store new ones
	if err := tx.Where("user_id = ?", userID).Delete(&models.TwoFactorBackupCode{}).Error; err != nil {
		return "", "", nil, fmt.Errorf("failed to delete old backup codes: %w", err)
	}
	for _, h := range hashes {
		if err := tx.Create(&models.TwoFactorBackupCode{
			UserID:   userID,
			CodeHash: h,
		}).Error; err != nil {
			return "", "", nil, fmt.Errorf("failed to save backup code: %w", err)
		}
	}

	if err := tx.Commit().Error; err != nil {
		return "", "", nil, fmt.Errorf("failed to commit 2FA setup: %w", err)
	}

	return secret, qrURL, backupCodes, nil
}

func (s *Service) Confirm2FA(userID uuid.UUID, code string) error {
	var user models.User
	if err := s.db.First(&user, "id = ?", userID).Error; err != nil {
		return ErrUserNotFound
	}

	if user.TwoFactorSecret == "" {
		return errors.New("2FA setup not initiated")
	}

	if !s.totp.ValidateCode(user.TwoFactorSecret, code) {
		return ErrInvalid2FACode
	}

	return s.db.Model(&user).Update("two_factor_enabled", true).Error
}

func (s *Service) Disable2FA(userID uuid.UUID, code string) error {
	var user models.User
	if err := s.db.First(&user, "id = ?", userID).Error; err != nil {
		return ErrUserNotFound
	}

	if !user.TwoFactorEnabled {
		return errors.New("2FA is not enabled")
	}

	valid, err := s.Validate2FA(userID, user.TwoFactorSecret, code)
	if err != nil {
		return err
	}
	if !valid {
		return ErrInvalid2FACode
	}

	s.db.Where("user_id = ?", userID).Delete(&models.TwoFactorBackupCode{})

	return s.db.Model(&user).Updates(map[string]interface{}{
		"two_factor_enabled": false,
		"two_factor_secret":  "",
	}).Error
}

func (s *Service) Validate2FA(userID uuid.UUID, secret, code string) (bool, error) {
	if s.totp.ValidateCode(secret, code) {
		return true, nil
	}

	var backupCodes []models.TwoFactorBackupCode
	s.db.Where("user_id = ? AND used = false", userID).Find(&backupCodes)

	for _, bc := range backupCodes {
		if bcrypt.CompareHashAndPassword([]byte(bc.CodeHash), []byte(code)) == nil {
			s.db.Model(&bc).Update("used", true)
			return true, nil
		}
	}

	return false, nil
}
