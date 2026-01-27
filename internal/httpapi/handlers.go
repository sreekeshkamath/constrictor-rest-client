package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/constrictor/constrictor-rest-client/internal/domain"
	"github.com/constrictor/constrictor-rest-client/internal/executor"
	"github.com/constrictor/constrictor-rest-client/internal/storage"
)

// APIError represents an API error response
type APIError struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// Handlers holds dependencies for HTTP handlers
type Handlers struct {
	store    storage.WorkspaceStore
	executor executor.Executor
}

// NewHandlers creates a new Handlers instance
func NewHandlers(store storage.WorkspaceStore, exec executor.Executor) *Handlers {
	return &Handlers{
		store:    store,
		executor: exec,
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

	// Load workspace for auth resolution
	workspace, err := h.store.Load()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to load workspace", err.Error())
		return
	}

	// Find the request item if requestId is provided
	var requestItem *domain.WorkspaceItem
	if req.RequestID != "" {
		for i := range workspace.Items {
			if workspace.Items[i].ID == req.RequestID && workspace.Items[i].Type == "request" {
				requestItem = &workspace.Items[i]
				break
			}
		}
	}

	// Convert to executor.Request
	execReq := &executor.Request{
		Method:    req.Method,
		URL:       req.URL,
		Headers:   convertHeaders(req.Headers),
		BodyType:  req.BodyType,
		Body:      req.Body,
		FormData:  convertFormData(req.FormData),
		RequestID: req.RequestID,
	}

	// Execute with request context (timeout handled by executor)
	ctx := r.Context()

	// Create executor with workspace and request item for auth resolution
	httpExecutor, ok := h.executor.(*executor.HTTPExecutor)
	if !ok {
		respondError(w, http.StatusInternalServerError, "executor type not supported", "")
		return
	}

	result, err := httpExecutor.ExecuteWithAuth(ctx, execReq, requestItem, workspace)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "execution failed", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, result)
}

// ExecuteRequest represents the request body for /api/execute
type ExecuteRequest struct {
	Method    string                 `json:"method"`
	URL       string                 `json:"url"`
	Headers   []domain.Header        `json:"headers"`
	BodyType  string                 `json:"bodyType"`
	Body      string                 `json:"body"`
	FormData  []domain.FormDataItem  `json:"formData"`
	Timeout   int                    `json:"timeout"` // seconds, 0 = use default
	RequestID string                 `json:"requestId,omitempty"` // ID of the request item for auth resolution
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
