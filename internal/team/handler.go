package team

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/muah1987/Aihub/internal/auth"
)

type Handler struct {
	service      *Service
	orchestrator *Orchestrator
}

func NewHandler(service *Service, orchestrator *Orchestrator) *Handler {
	return &Handler{service: service, orchestrator: orchestrator}
}

func (h *Handler) CreateTeam(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid project id"})
		return
	}

	var input CreateTeamInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if input.Name == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name is required"})
		return
	}

	team, err := h.service.CreateTeam(projectID, &input)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create team"})
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{"team": team})
}

func (h *Handler) ListTeams(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid project id"})
		return
	}

	teams, err := h.service.ListTeams(projectID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list teams"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"teams": teams})
}

func (h *Handler) GetTeam(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid project id"})
		return
	}

	teamID, err := uuid.Parse(chi.URLParam(r, "teamId"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid team id"})
		return
	}

	team, err := h.service.GetTeam(projectID, teamID)
	if err != nil {
		if errors.Is(err, ErrTeamNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "team not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get team"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"team": team})
}

func (h *Handler) UpdateTeam(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid project id"})
		return
	}

	teamID, err := uuid.Parse(chi.URLParam(r, "teamId"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid team id"})
		return
	}

	var input UpdateTeamInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	team, err := h.service.UpdateTeam(projectID, teamID, &input)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update team"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"team": team})
}

func (h *Handler) DeleteTeam(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid project id"})
		return
	}

	teamID, err := uuid.Parse(chi.URLParam(r, "teamId"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid team id"})
		return
	}

	if err := h.service.DeleteTeam(projectID, teamID); err != nil {
		if errors.Is(err, ErrTeamNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "team not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to delete team"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "team deleted"})
}

func (h *Handler) AddMember(w http.ResponseWriter, r *http.Request) {
	teamID, err := uuid.Parse(chi.URLParam(r, "teamId"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid team id"})
		return
	}

	var input AddMemberInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	agentID, err := uuid.Parse(input.AgentID)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid agent id"})
		return
	}

	member, err := h.service.AddMember(teamID, agentID, input.RoleInTeam)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to add member"})
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{"member": member})
}

func (h *Handler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	teamID, err := uuid.Parse(chi.URLParam(r, "teamId"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid team id"})
		return
	}

	agentID, err := uuid.Parse(chi.URLParam(r, "agentId"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid agent id"})
		return
	}

	if err := h.service.RemoveMember(teamID, agentID); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to remove member"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "member removed"})
}

func (h *Handler) SetLeader(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid project id"})
		return
	}

	teamID, err := uuid.Parse(chi.URLParam(r, "teamId"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid team id"})
		return
	}

	var input struct {
		AgentID string `json:"agent_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	agentID, err := uuid.Parse(input.AgentID)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid agent id"})
		return
	}

	if err := h.service.SetLeader(projectID, teamID, agentID); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to set leader"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "leader set"})
}

func (h *Handler) ListMembers(w http.ResponseWriter, r *http.Request) {
	teamID, err := uuid.Parse(chi.URLParam(r, "teamId"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid team id"})
		return
	}

	members, err := h.service.ListMembers(teamID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list members"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"members": members})
}

func (h *Handler) InvokeTeam(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid project id"})
		return
	}

	teamID, err := uuid.Parse(chi.URLParam(r, "teamId"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid team id"})
		return
	}

	var input struct {
		Prompt string `json:"prompt"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if input.Prompt == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "prompt is required"})
		return
	}

	result, err := h.orchestrator.InvokeTeam(context.Background(), userID, projectID, teamID, input.Prompt)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message": result.Message,
		"tasks":   result.Tasks,
	})
}

func (h *Handler) ListTasks(w http.ResponseWriter, r *http.Request) {
	teamID, err := uuid.Parse(chi.URLParam(r, "teamId"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid team id"})
		return
	}

	tasks, err := h.service.ListTasks(teamID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list tasks"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"tasks": tasks})
}

func (h *Handler) GetTask(w http.ResponseWriter, r *http.Request) {
	taskID, err := uuid.Parse(chi.URLParam(r, "taskId"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid task id"})
		return
	}

	task, err := h.service.GetTask(taskID)
	if err != nil {
		if errors.Is(err, ErrTaskNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "task not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get task"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"task": task})
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
