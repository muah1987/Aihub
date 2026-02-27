package organization

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/muah1987/Aihub/internal/auth"
	"github.com/muah1987/Aihub/internal/rbac"
)

type Handler struct {
	service     *Service
	rbacService *rbac.Service
}

func NewHandler(service *Service, rbacService *rbac.Service) *Handler {
	return &Handler{service: service, rbacService: rbacService}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	var input CreateInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if input.Name == "" || input.Slug == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name and slug are required"})
		return
	}

	org, err := h.service.Create(userID, &input)
	if err != nil {
		if errors.Is(err, ErrSlugTaken) {
			writeJSON(w, http.StatusConflict, map[string]string{"error": "slug already taken"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create organization"})
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{"organization": org})
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	orgs, err := h.service.List(userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list organizations"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"organizations": orgs})
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	orgID, err := uuid.Parse(chi.URLParam(r, "orgId"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid organization id"})
		return
	}

	org, err := h.service.Get(orgID)
	if err != nil {
		if errors.Is(err, ErrOrgNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "organization not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get organization"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"organization": org})
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	orgID, err := uuid.Parse(chi.URLParam(r, "orgId"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid organization id"})
		return
	}

	has, _ := h.rbacService.HasPermission(userID, orgID, rbac.PermManageOrg)
	if !has {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "insufficient permissions"})
		return
	}

	var input UpdateInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	org, err := h.service.Update(orgID, &input)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update organization"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"organization": org})
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	orgID, err := uuid.Parse(chi.URLParam(r, "orgId"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid organization id"})
		return
	}

	if err := h.service.Delete(orgID, userID); err != nil {
		if errors.Is(err, ErrNotOwner) {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "only the owner can delete the organization"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to delete organization"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "organization deleted"})
}

func (h *Handler) ListMembers(w http.ResponseWriter, r *http.Request) {
	orgID, err := uuid.Parse(chi.URLParam(r, "orgId"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid organization id"})
		return
	}

	members, err := h.service.ListMembers(orgID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list members"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"members": members})
}

func (h *Handler) InviteMember(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	orgID, err := uuid.Parse(chi.URLParam(r, "orgId"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid organization id"})
		return
	}

	has, _ := h.rbacService.HasPermission(userID, orgID, rbac.PermInviteMembers)
	if !has {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "insufficient permissions"})
		return
	}

	var input InviteInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if input.Email == "" || input.Role == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "email and role are required"})
		return
	}

	invite, err := h.service.InviteMember(orgID, userID, &input)
	if err != nil {
		if errors.Is(err, ErrAlreadyMember) {
			writeJSON(w, http.StatusConflict, map[string]string{"error": "user is already a member"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to send invitation"})
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{"invitation": invite})
}

func (h *Handler) AcceptInvitation(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	token := chi.URLParam(r, "token")

	if err := h.service.AcceptInvitation(token, userID); err != nil {
		if errors.Is(err, ErrInviteNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "invitation not found"})
			return
		}
		if errors.Is(err, ErrInviteExpired) {
			writeJSON(w, http.StatusGone, map[string]string{"error": "invitation expired"})
			return
		}
		if errors.Is(err, ErrAlreadyMember) {
			writeJSON(w, http.StatusConflict, map[string]string{"error": "already a member"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to accept invitation"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "invitation accepted"})
}

func (h *Handler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	orgID, err := uuid.Parse(chi.URLParam(r, "orgId"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid organization id"})
		return
	}

	has, _ := h.rbacService.HasPermission(userID, orgID, rbac.PermInviteMembers)
	if !has {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "insufficient permissions"})
		return
	}

	targetUserID, err := uuid.Parse(chi.URLParam(r, "userId"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid user id"})
		return
	}

	if err := h.service.RemoveMember(orgID, userID, targetUserID); err != nil {
		if errors.Is(err, ErrCannotRemoveSelf) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "cannot remove yourself"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to remove member"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "member removed"})
}

func (h *Handler) UpdateMemberRole(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	orgID, err := uuid.Parse(chi.URLParam(r, "orgId"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid organization id"})
		return
	}

	has, _ := h.rbacService.HasPermission(userID, orgID, rbac.PermInviteMembers)
	if !has {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "insufficient permissions"})
		return
	}

	targetUserID, err := uuid.Parse(chi.URLParam(r, "userId"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid user id"})
		return
	}

	var input struct {
		Role string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if err := h.service.UpdateMemberRole(orgID, targetUserID, input.Role); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update member role"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "role updated"})
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
