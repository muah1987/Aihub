package workflow

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/muah1987/Aihub/internal/auth"
	"gorm.io/datatypes"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// ---------- Workflow CRUD ----------

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid project id"})
		return
	}
	wfs, err := h.service.List(projectID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list workflows"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"workflows": wfs})
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	projectID, _ := uuid.Parse(chi.URLParam(r, "id"))
	workflowID, err := uuid.Parse(chi.URLParam(r, "workflowId"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid workflow id"})
		return
	}
	wf, err := h.service.GetByID(projectID, workflowID)
	if err != nil {
		if errors.Is(err, ErrWorkflowNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "workflow not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"workflow": wf})
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid project id"})
		return
	}
	var input CreateWorkflowInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if input.Name == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name is required"})
		return
	}
	wf, err := h.service.Create(projectID, &input)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, map[string]interface{}{"workflow": wf})
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	projectID, _ := uuid.Parse(chi.URLParam(r, "id"))
	workflowID, err := uuid.Parse(chi.URLParam(r, "workflowId"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid workflow id"})
		return
	}
	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	// Convert trigger_config to datatypes.JSON if present
	if tc, ok := body["trigger_config"]; ok {
		tcBytes, _ := json.Marshal(tc)
		body["trigger_config"] = datatypes.JSON(tcBytes)
	}
	wf, err := h.service.Update(projectID, workflowID, body)
	if err != nil {
		if errors.Is(err, ErrWorkflowNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "workflow not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"workflow": wf})
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	projectID, _ := uuid.Parse(chi.URLParam(r, "id"))
	workflowID, err := uuid.Parse(chi.URLParam(r, "workflowId"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid workflow id"})
		return
	}
	if err := h.service.Delete(projectID, workflowID); err != nil {
		if errors.Is(err, ErrWorkflowNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "workflow not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "workflow deleted"})
}

func (h *Handler) SetEnabled(w http.ResponseWriter, r *http.Request) {
	projectID, _ := uuid.Parse(chi.URLParam(r, "id"))
	workflowID, err := uuid.Parse(chi.URLParam(r, "workflowId"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid workflow id"})
		return
	}
	var body struct {
		Enabled bool `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if err := h.service.SetEnabled(projectID, workflowID, body.Enabled); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "updated"})
}

// ---------- Steps ----------

func (h *Handler) AddStep(w http.ResponseWriter, r *http.Request) {
	workflowID, err := uuid.Parse(chi.URLParam(r, "workflowId"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid workflow id"})
		return
	}
	var input CreateStepInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if input.Name == "" || input.StepType == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name and step_type are required"})
		return
	}
	step, err := h.service.AddStep(workflowID, &input)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, map[string]interface{}{"step": step})
}

func (h *Handler) DeleteStep(w http.ResponseWriter, r *http.Request) {
	stepID, err := uuid.Parse(chi.URLParam(r, "stepId"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid step id"})
		return
	}
	if err := h.service.DeleteStep(stepID); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "step deleted"})
}

// ---------- Trigger & Runs ----------

func (h *Handler) Trigger(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.GetUserIDFromContext(r.Context())
	projectID, _ := uuid.Parse(chi.URLParam(r, "id"))
	workflowID, err := uuid.Parse(chi.URLParam(r, "workflowId"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid workflow id"})
		return
	}
	run, err := h.service.Trigger(projectID, workflowID, &userID, "manual")
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]interface{}{"run": run})
}

func (h *Handler) ListRuns(w http.ResponseWriter, r *http.Request) {
	workflowID, err := uuid.Parse(chi.URLParam(r, "workflowId"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid workflow id"})
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	runs, err := h.service.ListRuns(workflowID, limit)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list runs"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"runs": runs})
}

func (h *Handler) GetRun(w http.ResponseWriter, r *http.Request) {
	runID, err := uuid.Parse(chi.URLParam(r, "runId"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid run id"})
		return
	}
	run, err := h.service.GetRun(runID)
	if err != nil {
		if errors.Is(err, ErrRunNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "run not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"run": run})
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
