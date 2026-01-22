package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/constrictor/constrictor-rest-client/internal/domain"
	"github.com/constrictor/constrictor-rest-client/internal/executor"
	"github.com/constrictor/constrictor-rest-client/internal/gdrive"
	"github.com/constrictor/constrictor-rest-client/internal/storage"
)

// APIError represents an API error response
type APIError struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// Handlers holds dependencies for HTTP handlers
type Handlers struct {
	store         storage.WorkspaceStore
	executor      executor.Executor
	gdriveService gdrive.Service
}

// NewHandlers creates a new Handlers instance
func NewHandlers(store storage.WorkspaceStore, exec executor.Executor) *Handlers {
	return &Handlers{
		store:         store,
		executor:      exec,
		gdriveService: gdrive.NewService(),
	}
}

// HandleGetWorkspace returns the current workspace
func (h *Handlers) HandleGetWorkspace(w http.ResponseWriter, r *http.Request) {
	workspace, err := h.store.Load()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to load workspace", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, workspace)
}

// HandlePutWorkspace saves the workspace
func (h *Handlers) HandlePutWorkspace(w http.ResponseWriter, r *http.Request) {
	var workspace domain.Workspace
	if err := json.NewDecoder(r.Body).Decode(&workspace); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON", err.Error())
		return
	}

	// Validate workspace
	if err := validateWorkspace(&workspace); err != nil {
		respondError(w, http.StatusBadRequest, "validation failed", err.Error())
		return
	}

	if err := h.store.Save(&workspace); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to save workspace", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "saved"})
}

// HandleExecute executes an HTTP request
func (h *Handlers) HandleExecute(w http.ResponseWriter, r *http.Request) {
	var req ExecuteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON", err.Error())
		return
	}

	// Validate request
	if err := validateExecuteRequest(&req); err != nil {
		respondError(w, http.StatusBadRequest, "validation failed", err.Error())
		return
	}

	// Convert to executor.Request
	execReq := &executor.Request{
		Method:   req.Method,
		URL:      req.URL,
		Headers:  convertHeaders(req.Headers),
		BodyType: req.BodyType,
		Body:     req.Body,
		FormData: convertFormData(req.FormData),
	}

	// Execute with request context (timeout handled by executor)
	ctx := r.Context()

	result, err := h.executor.Execute(ctx, execReq)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "execution failed", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, result)
}

// ExecuteRequest represents the request body for /api/execute
type ExecuteRequest struct {
	Method   string                 `json:"method"`
	URL      string                 `json:"url"`
	Headers  []domain.Header        `json:"headers"`
	BodyType string                 `json:"bodyType"`
	Body     string                 `json:"body"`
	FormData []domain.FormDataItem  `json:"formData"`
	Timeout  int                    `json:"timeout"` // seconds, 0 = use default
}

func validateWorkspace(w *domain.Workspace) error {
	if w.Version < 1 {
		return &ValidationError{Message: "version must be >= 1"}
	}
	// Additional validation can be added here
	return nil
}

func validateExecuteRequest(req *ExecuteRequest) error {
	if req.Method == "" {
		return &ValidationError{Message: "method is required"}
	}
	if req.URL == "" {
		return &ValidationError{Message: "URL is required"}
	}
	if req.BodyType == "" {
		req.BodyType = "none"
	}
	validBodyTypes := map[string]bool{
		"none":        true,
		"json":        true,
		"form-data":   true,
		"url-encoded": true,
	}
	if !validBodyTypes[req.BodyType] {
		return &ValidationError{Message: "invalid bodyType"}
	}
	return nil
}

func convertHeaders(headers []domain.Header) map[string]string {
	result := make(map[string]string)
	for _, h := range headers {
		if h.Enabled && h.Key != "" {
			result[h.Key] = h.Value
		}
	}
	return result
}

func convertFormData(formData []domain.FormDataItem) map[string]string {
	result := make(map[string]string)
	for _, item := range formData {
		if item.Enabled && item.Key != "" {
			result[item.Key] = item.Value
		}
	}
	return result
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, errorType, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(APIError{
		Error:   errorType,
		Message: message,
	})
}

// ValidationError represents a validation error
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

// HandleBackupToGDrive backs up the current workspace to Google Drive.
// The workspace is automatically sanitized to remove all sensitive headers and form data.
// Request body: { "accessToken": "...", "filename": "..." }
// Response: { "fileId": "...", "filename": "..." }
func (h *Handlers) HandleBackupToGDrive(w http.ResponseWriter, r *http.Request) {
	var req BackupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON", err.Error())
		return
	}

	if req.AccessToken == "" {
		respondError(w, http.StatusBadRequest, "validation failed", "accessToken is required")
		return
	}

	// Load current workspace
	workspace, err := h.store.Load()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to load workspace", err.Error())
		return
	}

	// Generate filename if not provided
	filename := req.Filename
	if filename == "" {
		filename = "constrictor-workspace-" + time.Now().Format("2006-01-02-150405") + ".json"
	}

	// Backup to Google Drive (workspace is automatically sanitized)
	ctx := r.Context()
	fileID, err := h.gdriveService.BackupWorkspace(ctx, req.AccessToken, workspace, filename)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "backup failed", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"fileId":   fileID,
		"filename": filename,
		"status":   "backed up",
	})
}

// HandleListGDriveBackups lists all workspace backups in Google Drive.
// Query param: accessToken
// Response: { "backups": [...] }
func (h *Handlers) HandleListGDriveBackups(w http.ResponseWriter, r *http.Request) {
	accessToken := r.URL.Query().Get("accessToken")
	if accessToken == "" {
		respondError(w, http.StatusBadRequest, "validation failed", "accessToken query parameter is required")
		return
	}

	ctx := r.Context()
	backups, err := h.gdriveService.ListBackups(ctx, accessToken)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to list backups", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"backups": backups,
	})
}

// HandleRestoreFromGDrive restores a workspace from Google Drive.
// Request body: { "accessToken": "...", "fileId": "..." }
// Response: { "version": 1, "items": [...] }
func (h *Handlers) HandleRestoreFromGDrive(w http.ResponseWriter, r *http.Request) {
	var req RestoreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON", err.Error())
		return
	}

	if req.AccessToken == "" {
		respondError(w, http.StatusBadRequest, "validation failed", "accessToken is required")
		return
	}
	if req.FileID == "" {
		respondError(w, http.StatusBadRequest, "validation failed", "fileId is required")
		return
	}

	ctx := r.Context()
	workspace, err := h.gdriveService.RestoreWorkspace(ctx, req.AccessToken, req.FileID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "restore failed", err.Error())
		return
	}

	// Save restored workspace
	if err := h.store.Save(workspace); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to save restored workspace", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, workspace)
}

// BackupRequest represents the request body for /api/gdrive/backup
type BackupRequest struct {
	AccessToken string `json:"accessToken"`
	Filename    string `json:"filename,omitempty"`
}

// RestoreRequest represents the request body for /api/gdrive/restore
type RestoreRequest struct {
	AccessToken string `json:"accessToken"`
	FileID      string `json:"fileId"`
}
