package auth

import (
	"errors"
	"fmt"

	"github.com/muah1987/Aihub/internal/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrEmailTaken      = errors.New("email already registered")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserNotFound    = errors.New("user not found")
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

type Service struct {
	db  *gorm.DB
	jwt *JWTService
}

func NewService(db *gorm.DB, jwt *JWTService) *Service {
	return &Service{db: db, jwt: jwt}
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

	tokens, err := s.jwt.GenerateTokenPair(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, nil, err
	}

	return user, tokens, nil
}

func (s *Service) Login(input *LoginInput) (*models.User, *TokenPair, error) {
	var user models.User
	if err := s.db.Where("email = ?", input.Email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, ErrInvalidCredentials
		}
		return nil, nil, fmt.Errorf("failed to find user: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return nil, nil, ErrInvalidCredentials
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
