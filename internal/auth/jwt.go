package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/muah1987/Aihub/internal/config"
)

type TokenType string

const (
	AccessToken      TokenType = "access"
	RefreshToken     TokenType = "refresh"
	PendingTwoFactor TokenType = "2fa_pending"
)

type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	Role   string    `json:"role"`
	Type   TokenType `json:"type"`
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int64  `json:"expires_at"`
}

type JWTService struct {
	secret          []byte
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

func NewJWTService(cfg *config.JWTConfig) *JWTService {
	return &JWTService{
		secret:          []byte(cfg.Secret),
		accessTokenTTL:  cfg.AccessTokenTTL,
		refreshTokenTTL: cfg.RefreshTokenTTL,
	}
}

func (j *JWTService) GenerateTokenPair(userID uuid.UUID, email, role string) (*TokenPair, error) {
	accessToken, expiresAt, err := j.generateToken(userID, email, role, AccessToken, j.accessTokenTTL)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, _, err := j.generateToken(userID, email, role, RefreshToken, j.refreshTokenTTL)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
	}, nil
}

func (j *JWTService) generateToken(userID uuid.UUID, email, role string, tokenType TokenType, ttl time.Duration) (string, int64, error) {
	now := time.Now()
	expiresAt := now.Add(ttl)

	claims := &Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		Type:   tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    "aihub",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(j.secret)
	if err != nil {
		return "", 0, err
	}

	return signed, expiresAt.Unix(), nil
}

func (j *JWTService) GeneratePendingToken(userID uuid.UUID, email, role string) (string, error) {
	token, _, err := j.generateToken(userID, email, role, PendingTwoFactor, 5*time.Minute)
	return token, err
}

func (j *JWTService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}
