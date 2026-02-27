package tools

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/muah1987/Aihub/internal/auth"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// ---------- Tool CRUD ----------

func (h *Handler) ListTools(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid project id"})
		return
	}
	tools, err := h.service.List(projectID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list tools"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"tools": tools})
}

func (h *Handler) CreateTool(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid project id"})
		return
	}
	var input CreateToolInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if input.Name == "" || input.ToolType == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name and tool_type are required"})
		return
	}
	tool, err := h.service.Create(projectID, &input)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, map[string]interface{}{"tool": tool})
}

func (h *Handler) UpdateTool(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid project id"})
		return
	}
	toolID, err := uuid.Parse(chi.URLParam(r, "toolId"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid tool id"})
		return
	}
	var input UpdateToolInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	tool, err := h.service.Update(projectID, toolID, &input)
	if err != nil {
		if errors.Is(err, ErrToolNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "tool not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"tool": tool})
}

func (h *Handler) DeleteTool(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid project id"})
		return
	}
	toolID, err := uuid.Parse(chi.URLParam(r, "toolId"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid tool id"})
		return
	}
	if err := h.service.Delete(projectID, toolID); err != nil {
		if errors.Is(err, ErrToolNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "tool not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "tool deleted"})
}

// ---------- Bindings ----------

func (h *Handler) ListBindings(w http.ResponseWriter, r *http.Request) {
	agentID, err := uuid.Parse(chi.URLParam(r, "agentId"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid agent id"})
		return
	}
	bindings, err := h.service.ListBindings(agentID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list bindings"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"bindings": bindings})
}

func (h *Handler) BindTool(w http.ResponseWriter, r *http.Request) {
	agentID, err := uuid.Parse(chi.URLParam(r, "agentId"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid agent id"})
		return
	}
	var body struct {
		ToolID string `json:"tool_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.ToolID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "tool_id is required"})
		return
	}
	toolID, err := uuid.Parse(body.ToolID)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid tool id"})
		return
	}
	if err := h.service.BindTool(agentID, toolID); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"message": "tool bound to agent"})
}

func (h *Handler) UnbindTool(w http.ResponseWriter, r *http.Request) {
	agentID, err := uuid.Parse(chi.URLParam(r, "agentId"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid agent id"})
		return
	}
	toolID, err := uuid.Parse(chi.URLParam(r, "toolId"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid tool id"})
		return
	}
	if err := h.service.UnbindTool(agentID, toolID); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "tool unbound"})
}

// ---------- Execute ----------

func (h *Handler) ExecuteTool(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.GetUserIDFromContext(r.Context())
	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid project id"})
		return
	}
	agentID, err := uuid.Parse(chi.URLParam(r, "agentId"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid agent id"})
		return
	}
	toolID, err := uuid.Parse(chi.URLParam(r, "toolId"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid tool id"})
		return
	}

	var body struct {
		Input json.RawMessage `json:"input"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if body.Input == nil {
		body.Input = json.RawMessage("{}")
	}

	exec, err := h.service.Execute(agentID, toolID, projectID, &userID, body.Input)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, ErrToolNotFound) {
			status = http.StatusNotFound
		} else if errors.Is(err, ErrToolDisabled) {
			status = http.StatusConflict
		}
		writeJSON(w, status, map[string]interface{}{"error": err.Error(), "execution": exec})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"execution": exec})
}

// ---------- Execution log ----------

func (h *Handler) ListExecutions(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid project id"})
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	execs, err := h.service.ListExecutions(projectID, limit)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list executions"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"executions": execs})
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
