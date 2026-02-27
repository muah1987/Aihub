package auth

import (
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"net/url"
	"time"

	"crypto/hmac"
	"crypto/sha1"
	"encoding/binary"
	"math"
)

type TOTPService struct{}

func NewTOTPService() *TOTPService {
	return &TOTPService{}
}

func (t *TOTPService) GenerateSecret() (string, error) {
	secret := make([]byte, 20)
	if _, err := rand.Read(secret); err != nil {
		return "", fmt.Errorf("failed to generate secret: %w", err)
	}
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(secret), nil
}

func (t *TOTPService) GenerateQRURL(email, secret string) string {
	return fmt.Sprintf("otpauth://totp/Aihub:%s?secret=%s&issuer=%s&digits=6&period=30",
		url.QueryEscape(email),
		secret,
		url.QueryEscape("Aihub"),
	)
}

func (t *TOTPService) ValidateCode(secret, code string) bool {
	now := time.Now().Unix()

	// Check current and adjacent time steps for clock drift
	for _, offset := range []int64{-1, 0, 1} {
		timeStep := (now / 30) + offset
		expected := generateTOTP(secret, timeStep)
		if expected == code {
			return true
		}
	}
	return false
}

func generateTOTP(secret string, timeStep int64) string {
	key, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(secret)
	if err != nil {
		return ""
	}

	msg := make([]byte, 8)
	binary.BigEndian.PutUint64(msg, uint64(timeStep))

	mac := hmac.New(sha1.New, key)
	mac.Write(msg)
	hash := mac.Sum(nil)

	offset := hash[len(hash)-1] & 0x0f
	truncated := binary.BigEndian.Uint32(hash[offset:offset+4]) & 0x7fffffff

	otp := truncated % uint32(math.Pow10(6))
	return fmt.Sprintf("%06d", otp)
}
