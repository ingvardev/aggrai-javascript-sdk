package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/ingvar/aiaggregator/apps/api/internal/middleware"
	"github.com/ingvar/aiaggregator/packages/domain"
	"github.com/ingvar/aiaggregator/packages/usecases"
)

// AdminHandler handles admin API endpoints for managing API users and keys.
type AdminHandler struct {
	authService    *usecases.AuthService
	webAuthService *usecases.WebAuthService
}

// NewAdminHandler creates a new admin handler.
func NewAdminHandler(authService *usecases.AuthService, webAuthService *usecases.WebAuthService) *AdminHandler {
	return &AdminHandler{
		authService:    authService,
		webAuthService: webAuthService,
	}
}

// getAuthContext tries to get auth context from API key or session token.
// Returns AuthContext if authenticated, or nil if not.
func (h *AdminHandler) getAuthContext(r *http.Request) *domain.AuthContext {
	// First try API key auth
	authCtx := middleware.AuthContextFromContext(r.Context())
	if authCtx != nil {
		return authCtx
	}

	// Then try session auth (web dashboard) - owners have admin scope
	webCtx := middleware.WebAuthContextFromContext(r.Context())
	if webCtx != nil {
		// Convert WebAuthContext to AuthContext with admin scope
		return &domain.AuthContext{
			TenantID: webCtx.TenantID,
			Scopes:   []string{string(domain.ScopeAdmin), string(domain.ScopeRead), string(domain.ScopeWrite)},
		}
	}

	return nil
}

// --- Request/Response types ---

// CreateUserRequest is the request body for creating an API user.
type CreateUserRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// UserResponse is the response for API user operations.
type UserResponse struct {
	ID          string `json:"id"`
	TenantID    string `json:"tenant_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Active      bool   `json:"active"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// CreateKeyRequest is the request body for creating an API key.
type CreateKeyRequest struct {
	UserID string   `json:"user_id"`
	Name   string   `json:"name"`
	Scopes []string `json:"scopes,omitempty"`
}

// KeyResponse is the response for API key operations.
type KeyResponse struct {
	ID         string   `json:"id"`
	UserID     string   `json:"user_id"`
	KeyPrefix  string   `json:"key_prefix"`
	Name       string   `json:"name"`
	Scopes     []string `json:"scopes"`
	Active     bool     `json:"active"`
	ExpiresAt  *string  `json:"expires_at,omitempty"`
	LastUsedAt *string  `json:"last_used_at,omitempty"`
	UsageCount int64    `json:"usage_count"`
	CreatedAt  string   `json:"created_at"`
	RevokedAt  *string  `json:"revoked_at,omitempty"`
}

// CreateKeyResponse includes the raw key (shown only once).
type CreateKeyResponse struct {
	KeyResponse
	Key string `json:"key"` // Raw key - ONLY SHOWN ONCE
}

// ActivityEntry represents an audit log entry in API response.
type ActivityEntry struct {
	ID        string                 `json:"id"`
	Action    string                 `json:"action"`
	Details   map[string]interface{} `json:"details,omitempty"`
	IPAddress string                 `json:"ip_address,omitempty"`
	UserAgent string                 `json:"user_agent,omitempty"`
	CreatedAt string                 `json:"created_at"`
}

// ErrorResponse is the standard error response.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// --- Handlers ---

// CreateUser handles POST /api/admin/users
func (h *AdminHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	authCtx := h.getAuthContext(r)
	if authCtx == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "")
		return
	}

	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Invalid JSON body")
		return
	}

	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "validation_error", "name is required")
		return
	}

	user, err := h.authService.CreateAPIUser(r.Context(), authCtx, req.Name, req.Description)
	if err != nil {
		if err == domain.ErrInsufficientScope {
			writeError(w, http.StatusForbidden, "forbidden", "Admin scope required")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to create user")
		return
	}

	resp := toUserResponse(user)
	writeJSON(w, http.StatusCreated, resp)
}

// ListUsers handles GET /api/admin/users
func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	authCtx := h.getAuthContext(r)
	if authCtx == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "")
		return
	}

	users, err := h.authService.ListAPIUsers(r.Context(), authCtx)
	if err != nil {
		if err == domain.ErrInsufficientScope {
			writeError(w, http.StatusForbidden, "forbidden", "Admin scope required")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to list users")
		return
	}

	resp := make([]UserResponse, len(users))
	for i, u := range users {
		resp[i] = toUserResponse(u)
	}

	writeJSON(w, http.StatusOK, resp)
}

// CreateKey handles POST /api/admin/api-keys
func (h *AdminHandler) CreateKey(w http.ResponseWriter, r *http.Request) {
	authCtx := h.getAuthContext(r)
	if authCtx == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "")
		return
	}

	var req CreateKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Invalid JSON body")
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "validation_error", "Invalid user_id format")
		return
	}

	keyName := req.Name
	if keyName == "" {
		keyName = "Default"
	}

	keyWithRaw, err := h.authService.CreateAPIKey(r.Context(), authCtx, userID, keyName, req.Scopes)
	if err != nil {
		if err == domain.ErrInsufficientScope {
			writeError(w, http.StatusForbidden, "forbidden", "Admin scope required")
			return
		}
		if err == domain.ErrAPIUserNotFound {
			writeError(w, http.StatusNotFound, "not_found", "User not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to create API key")
		return
	}

	resp := CreateKeyResponse{
		KeyResponse: toKeyResponse(&keyWithRaw.APIKey),
		Key:         keyWithRaw.RawKey, // Only time this is shown!
	}

	writeJSON(w, http.StatusCreated, resp)
}

// ListKeys handles GET /api/admin/users/{user_id}/api-keys
func (h *AdminHandler) ListKeys(w http.ResponseWriter, r *http.Request) {
	authCtx := h.getAuthContext(r)
	if authCtx == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "")
		return
	}

	userIDStr := chi.URLParam(r, "user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "validation_error", "Invalid user_id format")
		return
	}

	keys, err := h.authService.ListAPIKeys(r.Context(), authCtx, userID)
	if err != nil {
		if err == domain.ErrInsufficientScope {
			writeError(w, http.StatusForbidden, "forbidden", "Admin scope required")
			return
		}
		if err == domain.ErrAPIUserNotFound {
			writeError(w, http.StatusNotFound, "not_found", "User not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to list keys")
		return
	}

	resp := make([]KeyResponse, len(keys))
	for i, k := range keys {
		resp[i] = toKeyResponse(k)
	}

	writeJSON(w, http.StatusOK, resp)
}

// RevokeKey handles DELETE /api/admin/api-keys/{id}
func (h *AdminHandler) RevokeKey(w http.ResponseWriter, r *http.Request) {
	authCtx := h.getAuthContext(r)
	if authCtx == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "")
		return
	}

	keyIDStr := chi.URLParam(r, "id")
	keyID, err := uuid.Parse(keyIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "validation_error", "Invalid key id format")
		return
	}

	if err := h.authService.RevokeAPIKey(r.Context(), authCtx, keyID); err != nil {
		if err == domain.ErrInsufficientScope {
			writeError(w, http.StatusForbidden, "forbidden", "Admin scope required")
			return
		}
		if err == domain.ErrAPIKeyNotFound {
			writeError(w, http.StatusNotFound, "not_found", "API key not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to revoke key")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetUserActivity handles GET /api/admin/users/{user_id}/activity
func (h *AdminHandler) GetUserActivity(w http.ResponseWriter, r *http.Request) {
	authCtx := h.getAuthContext(r)
	if authCtx == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "")
		return
	}

	userIDStr := chi.URLParam(r, "user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "validation_error", "Invalid user_id format")
		return
	}

	// Default limit and offset
	limit := 50
	offset := 0

	entries, err := h.authService.GetUserActivity(r.Context(), authCtx, userID, limit, offset)
	if err != nil {
		if err == domain.ErrInsufficientScope {
			writeError(w, http.StatusForbidden, "forbidden", "Admin scope required")
			return
		}
		if err == domain.ErrAPIUserNotFound {
			writeError(w, http.StatusNotFound, "not_found", "User not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to get activity")
		return
	}

	resp := make([]ActivityEntry, len(entries))
	for i, e := range entries {
		resp[i] = toActivityEntry(e)
	}

	writeJSON(w, http.StatusOK, resp)
}

// --- Helper functions ---

func toUserResponse(u *domain.APIUser) UserResponse {
	return UserResponse{
		ID:          u.ID.String(),
		TenantID:    u.TenantID.String(),
		Name:        u.Name,
		Description: u.Description,
		Active:      u.Active,
		CreatedAt:   u.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   u.UpdatedAt.Format(time.RFC3339),
	}
}

func toKeyResponse(k *domain.APIKey) KeyResponse {
	resp := KeyResponse{
		ID:         k.ID.String(),
		UserID:     k.APIUserID.String(),
		KeyPrefix:  k.KeyPrefix,
		Name:       k.Name,
		Scopes:     k.Scopes,
		Active:     k.Active,
		UsageCount: k.UsageCount,
		CreatedAt:  k.CreatedAt.Format(time.RFC3339),
	}

	if k.ExpiresAt != nil {
		s := k.ExpiresAt.Format(time.RFC3339)
		resp.ExpiresAt = &s
	}
	if k.LastUsedAt != nil {
		s := k.LastUsedAt.Format(time.RFC3339)
		resp.LastUsedAt = &s
	}
	if k.RevokedAt != nil {
		s := k.RevokedAt.Format(time.RFC3339)
		resp.RevokedAt = &s
	}

	return resp
}

func toActivityEntry(e *domain.AuditLogEntry) ActivityEntry {
	entry := ActivityEntry{
		ID:        e.ID.String(),
		Action:    string(e.Action),
		Details:   e.Details,
		UserAgent: e.UserAgent,
		CreatedAt: e.CreatedAt.Format(time.RFC3339),
	}
	if e.IPAddress != nil {
		entry.IPAddress = e.IPAddress.String()
	}
	return entry
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, errCode, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(ErrorResponse{
		Error:   errCode,
		Message: message,
	})
}
