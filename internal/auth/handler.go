package auth

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var input RegisterInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if input.Email == "" || input.Password == "" || input.DisplayName == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "email, password, and display_name are required"})
		return
	}

	if len(input.Password) < 8 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "password must be at least 8 characters"})
		return
	}

	user, tokens, err := h.service.Register(&input)
	if err != nil {
		if errors.Is(err, ErrEmailTaken) {
			writeJSON(w, http.StatusConflict, map[string]string{"error": "email already registered"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to register"})
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"user":   user.ToResponse(),
		"tokens": tokens,
	})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var input LoginInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if input.Email == "" || input.Password == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "email and password are required"})
		return
	}

	result, err := h.service.Login(&input)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid email or password"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to login"})
		return
	}

	if result.Requires2FA {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"requires_2fa":  true,
			"pending_token": result.PendingToken,
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"user":   result.User.ToResponse(),
		"tokens": result.Tokens,
	})
}

func (h *Handler) LoginVerify2FA(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Token string `json:"token"`
		Code  string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if input.Token == "" || input.Code == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "token and code are required"})
		return
	}

	user, tokens, err := h.service.LoginVerify2FA(input.Token, input.Code)
	if err != nil {
		if errors.Is(err, ErrInvalid2FACode) {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid 2FA code"})
			return
		}
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid token"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"user":   user.ToResponse(),
		"tokens": tokens,
	})
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	var input struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if input.RefreshToken == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "refresh_token is required"})
		return
	}

	user, tokens, err := h.service.RefreshTokens(input.RefreshToken)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid refresh token"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"user":   user.ToResponse(),
		"tokens": tokens,
	})
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(ContextKeyUserID).(uuid.UUID)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	user, err := h.service.GetUserByID(userID.String())
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "user not found"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"user": user.ToResponse(),
	})
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"message": "logged out successfully"})
}

// Email verification

func (h *Handler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if input.Token == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "token is required"})
		return
	}

	if err := h.service.VerifyEmail(input.Token); err != nil {
		if errors.Is(err, ErrTokenNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "invalid verification token"})
			return
		}
		if errors.Is(err, ErrTokenExpired) {
			writeJSON(w, http.StatusGone, map[string]string{"error": "verification token expired"})
			return
		}
		if errors.Is(err, ErrTokenUsed) {
			writeJSON(w, http.StatusConflict, map[string]string{"error": "token already used"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "verification failed"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "email verified successfully"})
}

func (h *Handler) ResendVerification(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(ContextKeyUserID).(uuid.UUID)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	if err := h.service.SendVerificationEmail(userID); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to send verification email"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "verification email sent"})
}

// Two-Factor Authentication

func (h *Handler) Setup2FA(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(ContextKeyUserID).(uuid.UUID)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	secret, qrURL, backupCodes, err := h.service.Setup2FA(userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to setup 2FA"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"secret":       secret,
		"qr_url":       qrURL,
		"backup_codes": backupCodes,
	})
}

func (h *Handler) Confirm2FA(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(ContextKeyUserID).(uuid.UUID)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	var input struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if err := h.service.Confirm2FA(userID, input.Code); err != nil {
		if errors.Is(err, ErrInvalid2FACode) {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid code"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "2FA enabled successfully"})
}

func (h *Handler) Disable2FA(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(ContextKeyUserID).(uuid.UUID)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	var input struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if err := h.service.Disable2FA(userID, input.Code); err != nil {
		if errors.Is(err, ErrInvalid2FACode) {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid code"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "2FA disabled successfully"})
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
