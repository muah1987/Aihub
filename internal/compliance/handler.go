package compliance

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/muah1987/Aihub/internal/auth"
)

type Handler struct {
	audit     *AuditService
	export    *ExportService
	roles     *RoleService
	retention *RetentionService
}

func NewHandler(audit *AuditService, export *ExportService, roles *RoleService, retention *RetentionService) *Handler {
	return &Handler{audit: audit, export: export, roles: roles, retention: retention}
}

// ---- Audit Logs ----

func (h *Handler) ListAuditLogs(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid project id"})
		return
	}

	filter := AuditFilter{ProjectID: &projectID}
	if action := r.URL.Query().Get("action"); action != "" {
		filter.Action = action
	}
	if rt := r.URL.Query().Get("resource_type"); rt != "" {
		filter.ResourceType = rt
	}
	if sev := r.URL.Query().Get("severity"); sev != "" {
		filter.Severity = sev
	}
	if since := r.URL.Query().Get("since"); since != "" {
		if t, err := time.Parse(time.RFC3339, since); err == nil {
			filter.Since = &t
		}
	}
	if until := r.URL.Query().Get("until"); until != "" {
		if t, err := time.Parse(time.RFC3339, until); err == nil {
			filter.Until = &t
		}
	}
	filter.Limit, _ = strconv.Atoi(r.URL.Query().Get("limit"))
	filter.Offset, _ = strconv.Atoi(r.URL.Query().Get("offset"))

	logs, total, err := h.audit.Query(filter)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to query audit logs"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"audit_logs": logs, "total": total})
}

func (h *Handler) ListOrgAuditLogs(w http.ResponseWriter, r *http.Request) {
	orgID, err := uuid.Parse(chi.URLParam(r, "orgId"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid org id"})
		return
	}

	filter := AuditFilter{OrganizationID: &orgID}
	filter.Limit, _ = strconv.Atoi(r.URL.Query().Get("limit"))
	filter.Offset, _ = strconv.Atoi(r.URL.Query().Get("offset"))
	if action := r.URL.Query().Get("action"); action != "" {
		filter.Action = action
	}
	if sev := r.URL.Query().Get("severity"); sev != "" {
		filter.Severity = sev
	}

	logs, total, err := h.audit.Query(filter)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to query audit logs"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"audit_logs": logs, "total": total})
}

// ---- Data Exports ----

func (h *Handler) ListExports(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid project id"})
		return
	}

	exports, err := h.export.ListExports(projectID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list exports"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"exports": exports})
}

func (h *Handler) CreateExport(w http.ResponseWriter, r *http.Request) {
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

	var input CreateExportInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	export, err := h.export.CreateExport(projectID, &userID, &input)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "export failed"})
		return
	}
	writeJSON(w, http.StatusCreated, map[string]interface{}{"export": export})
}

func (h *Handler) DownloadExport(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid project id"})
		return
	}
	exportID, err := uuid.Parse(chi.URLParam(r, "exportId"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid export id"})
		return
	}

	data, err := h.export.DownloadExport(projectID, exportID)
	if err != nil {
		if errors.Is(err, ErrExportNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "export not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", "attachment; filename=export.json")
	w.Write(data)
}

func (h *Handler) DeleteExport(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid project id"})
		return
	}
	exportID, err := uuid.Parse(chi.URLParam(r, "exportId"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid export id"})
		return
	}

	if err := h.export.DeleteExport(projectID, exportID); err != nil {
		if errors.Is(err, ErrExportNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "export not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to delete"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "export deleted"})
}

// ---- Custom Roles ----

func (h *Handler) ListRoles(w http.ResponseWriter, r *http.Request) {
	orgID, err := uuid.Parse(chi.URLParam(r, "orgId"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid org id"})
		return
	}
	roles, err := h.roles.ListRoles(orgID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list roles"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"roles": roles})
}

func (h *Handler) CreateRole(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	orgID, err := uuid.Parse(chi.URLParam(r, "orgId"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid org id"})
		return
	}

	var input CreateRoleInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil || input.Name == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name is required"})
		return
	}

	role, err := h.roles.CreateRole(orgID, &userID, &input)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create role"})
		return
	}
	writeJSON(w, http.StatusCreated, map[string]interface{}{"role": role})
}

func (h *Handler) UpdateRole(w http.ResponseWriter, r *http.Request) {
	orgID, err := uuid.Parse(chi.URLParam(r, "orgId"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid org id"})
		return
	}
	roleID, err := uuid.Parse(chi.URLParam(r, "roleId"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid role id"})
		return
	}

	var input CreateRoleInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if err := h.roles.UpdateRole(orgID, roleID, &input); err != nil {
		if errors.Is(err, ErrRoleNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "role not found"})
			return
		}
		if errors.Is(err, ErrSystemRole) {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "cannot modify system role"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update role"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "role updated"})
}

func (h *Handler) DeleteRole(w http.ResponseWriter, r *http.Request) {
	orgID, err := uuid.Parse(chi.URLParam(r, "orgId"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid org id"})
		return
	}
	roleID, err := uuid.Parse(chi.URLParam(r, "roleId"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid role id"})
		return
	}

	if err := h.roles.DeleteRole(orgID, roleID); err != nil {
		if errors.Is(err, ErrSystemRole) {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "cannot delete system role"})
			return
		}
		if errors.Is(err, ErrRoleNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "role not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to delete role"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "role deleted"})
}

// ---- Resource Permissions ----

func (h *Handler) ListPermissions(w http.ResponseWriter, r *http.Request) {
	orgID, err := uuid.Parse(chi.URLParam(r, "orgId"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid org id"})
		return
	}

	var userID *uuid.UUID
	if uid := r.URL.Query().Get("user_id"); uid != "" {
		parsed, err := uuid.Parse(uid)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid user_id"})
			return
		}
		userID = &parsed
	}
	resourceType := r.URL.Query().Get("resource_type")

	perms, err := h.roles.ListPermissions(orgID, userID, resourceType)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list permissions"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"permissions": perms})
}

func (h *Handler) GrantPermission(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	orgID, err := uuid.Parse(chi.URLParam(r, "orgId"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid org id"})
		return
	}

	var input GrantPermissionInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	perm, err := h.roles.GrantPermission(orgID, &userID, &input)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to grant permission"})
		return
	}
	writeJSON(w, http.StatusCreated, map[string]interface{}{"permission": perm})
}

func (h *Handler) RevokePermission(w http.ResponseWriter, r *http.Request) {
	orgID, err := uuid.Parse(chi.URLParam(r, "orgId"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid org id"})
		return
	}
	permID, err := uuid.Parse(chi.URLParam(r, "permId"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid permission id"})
		return
	}

	if err := h.roles.RevokePermission(orgID, permID); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to revoke"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "permission revoked"})
}

// ---- Retention Policies ----

func (h *Handler) ListRetentionPolicies(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid project id"})
		return
	}

	policies, err := h.retention.ListPolicies(nil, &projectID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list policies"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"policies": policies})
}

func (h *Handler) UpsertRetentionPolicy(w http.ResponseWriter, r *http.Request) {
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

	var input UpsertPolicyInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil || input.ResourceType == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "resource_type is required"})
		return
	}

	policy, err := h.retention.UpsertPolicy(nil, &projectID, &userID, &input)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to save policy"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"policy": policy})
}

func (h *Handler) DeleteRetentionPolicy(w http.ResponseWriter, r *http.Request) {
	policyID, err := uuid.Parse(chi.URLParam(r, "policyId"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid policy id"})
		return
	}

	if err := h.retention.DeletePolicy(policyID); err != nil {
		if errors.Is(err, ErrPolicyNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "policy not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to delete"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "policy deleted"})
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
