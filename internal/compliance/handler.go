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
	"github.com/muah1987/Aihub/internal/httputil"
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
		httputil.WriteError(w, http.StatusBadRequest, "invalid project id")
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
		httputil.WriteError(w, http.StatusInternalServerError, "failed to query audit logs")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]interface{}{"audit_logs": logs, "total": total})
}

func (h *Handler) ListOrgAuditLogs(w http.ResponseWriter, r *http.Request) {
	orgID, err := uuid.Parse(chi.URLParam(r, "orgId"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid org id")
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
		httputil.WriteError(w, http.StatusInternalServerError, "failed to query audit logs")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]interface{}{"audit_logs": logs, "total": total})
}

// ---- Data Exports ----

func (h *Handler) ListExports(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid project id")
		return
	}

	exports, err := h.export.ListExports(projectID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to list exports")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]interface{}{"exports": exports})
}

func (h *Handler) CreateExport(w http.ResponseWriter, r *http.Request) {
	httputil.LimitBody(w, r, httputil.MaxBodySize)

	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid project id")
		return
	}

	var input CreateExportInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	export, err := h.export.CreateExport(projectID, &userID, &input)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "export failed")
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, map[string]interface{}{"export": export})
}

func (h *Handler) DownloadExport(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid project id")
		return
	}
	exportID, err := uuid.Parse(chi.URLParam(r, "exportId"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid export id")
		return
	}

	data, err := h.export.DownloadExport(projectID, exportID)
	if err != nil {
		if errors.Is(err, ErrExportNotFound) {
			httputil.WriteError(w, http.StatusNotFound, "export not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", "attachment; filename=export.json")
	w.Write(data)
}

func (h *Handler) DeleteExport(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid project id")
		return
	}
	exportID, err := uuid.Parse(chi.URLParam(r, "exportId"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid export id")
		return
	}

	if err := h.export.DeleteExport(projectID, exportID); err != nil {
		if errors.Is(err, ErrExportNotFound) {
			httputil.WriteError(w, http.StatusNotFound, "export not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "failed to delete")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]string{"message": "export deleted"})
}

// ---- Custom Roles ----

func (h *Handler) ListRoles(w http.ResponseWriter, r *http.Request) {
	orgID, err := uuid.Parse(chi.URLParam(r, "orgId"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid org id")
		return
	}
	roles, err := h.roles.ListRoles(orgID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to list roles")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]interface{}{"roles": roles})
}

func (h *Handler) CreateRole(w http.ResponseWriter, r *http.Request) {
	httputil.LimitBody(w, r, httputil.MaxBodySize)

	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	orgID, err := uuid.Parse(chi.URLParam(r, "orgId"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid org id")
		return
	}

	var input CreateRoleInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil || input.Name == "" {
		httputil.WriteError(w, http.StatusBadRequest, "name is required")
		return
	}

	role, err := h.roles.CreateRole(orgID, &userID, &input)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to create role")
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, map[string]interface{}{"role": role})
}

func (h *Handler) UpdateRole(w http.ResponseWriter, r *http.Request) {
	httputil.LimitBody(w, r, httputil.MaxBodySize)

	orgID, err := uuid.Parse(chi.URLParam(r, "orgId"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid org id")
		return
	}
	roleID, err := uuid.Parse(chi.URLParam(r, "roleId"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid role id")
		return
	}

	var input CreateRoleInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.roles.UpdateRole(orgID, roleID, &input); err != nil {
		if errors.Is(err, ErrRoleNotFound) {
			httputil.WriteError(w, http.StatusNotFound, "role not found")
			return
		}
		if errors.Is(err, ErrSystemRole) {
			httputil.WriteError(w, http.StatusForbidden, "cannot modify system role")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "failed to update role")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]string{"message": "role updated"})
}

func (h *Handler) DeleteRole(w http.ResponseWriter, r *http.Request) {
	orgID, err := uuid.Parse(chi.URLParam(r, "orgId"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid org id")
		return
	}
	roleID, err := uuid.Parse(chi.URLParam(r, "roleId"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid role id")
		return
	}

	if err := h.roles.DeleteRole(orgID, roleID); err != nil {
		if errors.Is(err, ErrSystemRole) {
			httputil.WriteError(w, http.StatusForbidden, "cannot delete system role")
			return
		}
		if errors.Is(err, ErrRoleNotFound) {
			httputil.WriteError(w, http.StatusNotFound, "role not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "failed to delete role")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]string{"message": "role deleted"})
}

// ---- Resource Permissions ----

func (h *Handler) ListPermissions(w http.ResponseWriter, r *http.Request) {
	orgID, err := uuid.Parse(chi.URLParam(r, "orgId"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid org id")
		return
	}

	var userID *uuid.UUID
	if uid := r.URL.Query().Get("user_id"); uid != "" {
		parsed, err := uuid.Parse(uid)
		if err != nil {
			httputil.WriteError(w, http.StatusBadRequest, "invalid user_id")
			return
		}
		userID = &parsed
	}
	resourceType := r.URL.Query().Get("resource_type")

	perms, err := h.roles.ListPermissions(orgID, userID, resourceType)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to list permissions")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]interface{}{"permissions": perms})
}

func (h *Handler) GrantPermission(w http.ResponseWriter, r *http.Request) {
	httputil.LimitBody(w, r, httputil.MaxBodySize)

	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	orgID, err := uuid.Parse(chi.URLParam(r, "orgId"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid org id")
		return
	}

	var input GrantPermissionInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	perm, err := h.roles.GrantPermission(orgID, &userID, &input)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to grant permission")
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, map[string]interface{}{"permission": perm})
}

func (h *Handler) RevokePermission(w http.ResponseWriter, r *http.Request) {
	orgID, err := uuid.Parse(chi.URLParam(r, "orgId"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid org id")
		return
	}
	permID, err := uuid.Parse(chi.URLParam(r, "permId"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid permission id")
		return
	}

	if err := h.roles.RevokePermission(orgID, permID); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to revoke")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]string{"message": "permission revoked"})
}

// ---- Retention Policies ----

func (h *Handler) ListRetentionPolicies(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid project id")
		return
	}

	policies, err := h.retention.ListPolicies(nil, &projectID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to list policies")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]interface{}{"policies": policies})
}

func (h *Handler) UpsertRetentionPolicy(w http.ResponseWriter, r *http.Request) {
	httputil.LimitBody(w, r, httputil.MaxBodySize)

	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid project id")
		return
	}

	var input UpsertPolicyInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil || input.ResourceType == "" {
		httputil.WriteError(w, http.StatusBadRequest, "resource_type is required")
		return
	}

	policy, err := h.retention.UpsertPolicy(nil, &projectID, &userID, &input)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to save policy")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]interface{}{"policy": policy})
}

func (h *Handler) DeleteRetentionPolicy(w http.ResponseWriter, r *http.Request) {
	policyID, err := uuid.Parse(chi.URLParam(r, "policyId"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid policy id")
		return
	}

	if err := h.retention.DeletePolicy(policyID); err != nil {
		if errors.Is(err, ErrPolicyNotFound) {
			httputil.WriteError(w, http.StatusNotFound, "policy not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "failed to delete")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]string{"message": "policy deleted"})
}
