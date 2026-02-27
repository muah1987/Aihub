package webhook

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/muah1987/Aihub/internal/auth"
	"github.com/muah1987/Aihub/internal/deployment"
	"github.com/muah1987/Aihub/internal/notification"
)

type Handler struct {
	service      *Service
	deploySvc    *deployment.Service
	notifySvc    *notification.Service
}

func NewHandler(svc *Service, deploySvc *deployment.Service, notifySvc *notification.Service) *Handler {
	return &Handler{service: svc, deploySvc: deploySvc, notifySvc: notifySvc}
}

// List returns all webhooks for a project (secrets redacted).
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	projectID, ok := parseUUID(w, r, "id")
	if !ok {
		return
	}
	whs, err := h.service.List(projectID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	// Redact secrets, return webhook URL hint
	type safe struct {
		ID              interface{} `json:"id"`
		ProjectID       interface{} `json:"project_id"`
		VPSTargetID     interface{} `json:"vps_target_id,omitempty"`
		Branch          string      `json:"branch"`
		Active          bool        `json:"active"`
		LastTriggeredAt interface{} `json:"last_triggered_at,omitempty"`
		CreatedAt       interface{} `json:"created_at"`
	}
	result := make([]safe, 0, len(whs))
	for _, wh := range whs {
		result = append(result, safe{
			ID: wh.ID, ProjectID: wh.ProjectID, VPSTargetID: wh.VPSTargetID,
			Branch: wh.Branch, Active: wh.Active, LastTriggeredAt: wh.LastTriggeredAt,
			CreatedAt: wh.CreatedAt,
		})
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"webhooks": result})
}

// Create creates a new webhook and returns the secret once.
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	projectID, ok := parseUUID(w, r, "id")
	if !ok {
		return
	}
	var body struct {
		VPSTargetID string `json:"vps_target_id"`
		Branch      string `json:"branch"`
	}
	json.NewDecoder(r.Body).Decode(&body)

	var targetID *uuid.UUID
	if body.VPSTargetID != "" {
		id, err := uuid.Parse(body.VPSTargetID)
		if err == nil {
			targetID = &id
		}
	}

	wh, err := h.service.Create(projectID, targetID, body.Branch)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	// Return secret on creation — won't be shown again
	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"webhook": wh,
		"secret":  wh.Secret,
		"note":    "Save this secret — it will not be shown again",
	})
}

// SetActive enables or disables a webhook.
func (h *Handler) SetActive(w http.ResponseWriter, r *http.Request) {
	webhookID, ok := parseUUID(w, r, "webhookId")
	if !ok {
		return
	}
	var body struct {
		Active bool `json:"active"`
	}
	json.NewDecoder(r.Body).Decode(&body)
	if err := h.service.SetActive(webhookID, body.Active); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"active": body.Active})
}

// Delete removes a webhook.
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	projectID, ok := parseUUID(w, r, "id")
	if !ok {
		return
	}
	webhookID, ok := parseUUID(w, r, "webhookId")
	if !ok {
		return
	}
	if err := h.service.Delete(webhookID, projectID); err != nil {
		if errors.Is(err, ErrWebhookNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "deleted"})
}

// GitHubIncoming handles POST /webhooks/github/{webhookId} (public endpoint).
func (h *Handler) GitHubIncoming(w http.ResponseWriter, r *http.Request) {
	webhookID, ok := parseUUID(w, r, "webhookId")
	if !ok {
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 5*1024*1024))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "cannot read body"})
		return
	}

	// Validate HMAC signature
	sig := r.Header.Get("X-Hub-Signature-256")
	if err := h.service.ValidateGitHubSignature(webhookID, body, sig); err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid signature"})
		return
	}

	// Only react to push events
	event := r.Header.Get("X-GitHub-Event")
	if event != "push" {
		writeJSON(w, http.StatusOK, map[string]string{"message": "ignored"})
		return
	}

	var payload struct {
		Ref        string `json:"ref"`
		Repository struct {
			FullName string `json:"full_name"`
		} `json:"repository"`
		HeadCommit struct {
			ID      string `json:"id"`
			Message string `json:"message"`
		} `json:"head_commit"`
		Pusher struct {
			Name string `json:"name"`
		} `json:"pusher"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid payload"})
		return
	}

	wh, err := h.service.GetByID(webhookID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	if !MatchesBranch(payload.Ref, wh.Branch) {
		writeJSON(w, http.StatusOK, map[string]string{"message": "branch not matched"})
		return
	}

	h.service.MarkTriggered(webhookID)

	// Trigger deployment if target is configured
	if wh.VPSTargetID != nil && h.deploySvc != nil {
		go func() {
			_, _ = h.deploySvc.Deploy(wh.ProjectID, *wh.VPSTargetID, nil)
		}()

		// Send in-app notification to project owner
		if h.notifySvc != nil {
			h.notifySvc.NotifyProjectOwner(wh.ProjectID, "deploy_triggered", "Auto-Deploy Triggered",
				"Push to "+payload.Ref+" by "+payload.Pusher.Name+" triggered a deployment of "+payload.Repository.FullName)
		}
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "webhook processed"})
}

// GetSecret returns the webhook secret (authenticated route, one-time display on create).
func (h *Handler) GetSecret(w http.ResponseWriter, r *http.Request) {
	_, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	webhookID, ok := parseUUID(w, r, "webhookId")
	if !ok {
		return
	}
	secret, err := h.service.GetSecret(webhookID)
	if err != nil {
		if errors.Is(err, ErrWebhookNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"secret": secret})
}

func parseUUID(w http.ResponseWriter, r *http.Request, param string) (uuid.UUID, bool) {
	id, err := uuid.Parse(chi.URLParam(r, param))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid " + param})
		return uuid.Nil, false
	}
	return id, true
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
