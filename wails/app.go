package main

import (
	"context"
	"time"

	"github.com/constrictor/constrictor-rest-client/internal/config"
	"github.com/constrictor/constrictor-rest-client/internal/domain"
	"github.com/constrictor/constrictor-rest-client/internal/executor"
	"github.com/constrictor/constrictor-rest-client/internal/storage"
)

// SidebarItem represents a sidebar item (request or folder) matching the frontend TypeScript type
type SidebarItem struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Type      string  `json:"type"` // "request" or "folder"
	ParentID  *string `json:"parentId,omitempty"`
	CreatedAt int64   `json:"createdAt"`

	// Request-specific fields (only present when Type == "request")
	Method   *string         `json:"method,omitempty"`
	URL      *string         `json:"url,omitempty"`
	Headers  []domain.Header `json:"headers,omitempty"`
	BodyType *string         `json:"bodyType,omitempty"` // "none", "json", "form-data", "url-encoded"
	Body     *string         `json:"body,omitempty"`
	FormData []domain.FormDataItem `json:"formData,omitempty"`
}

// ExecuteRequestInput represents the input for ExecuteRequest matching frontend format
type ExecuteRequestInput struct {
	Method   string             `json:"method"`
	URL      string             `json:"url"`
	Headers  []domain.Header    `json:"headers"`
	BodyType string             `json:"bodyType"` // "none", "json", "form-data", "url-encoded"
	Body     string             `json:"body"`
	FormData []domain.FormDataItem `json:"formData"`
}

// convertWorkspaceToItems converts a domain.Workspace to an array of SidebarItem
func convertWorkspaceToItems(workspace *domain.Workspace) []SidebarItem {
	if workspace == nil || len(workspace.Items) == 0 {
		return []SidebarItem{}
	}

	items := make([]SidebarItem, 0, len(workspace.Items))
	for _, item := range workspace.Items {
		sidebarItem := SidebarItem{
			ID:        item.ID,
			Name:      item.Name,
			Type:      item.Type,
			CreatedAt: item.CreatedAt,
		}

		if item.ParentID != nil {
			parentID := *item.ParentID
			sidebarItem.ParentID = &parentID
		}

		// Copy request-specific fields if this is a request
		if item.Type == "request" {
			if item.Method != nil {
				method := *item.Method
				sidebarItem.Method = &method
			}
			if item.URL != nil {
				url := *item.URL
				sidebarItem.URL = &url
			}
			if item.BodyType != nil {
				bodyType := *item.BodyType
				sidebarItem.BodyType = &bodyType
			}
			if item.Body != nil {
				body := *item.Body
				sidebarItem.Body = &body
			}
			sidebarItem.Headers = item.Headers
			sidebarItem.FormData = item.FormData
		}

		items = append(items, sidebarItem)
	}

	return items
}

// convertItemsToWorkspace converts an array of SidebarItem to a domain.Workspace
func convertItemsToWorkspace(items []SidebarItem) *domain.Workspace {
	if len(items) == 0 {
		return &domain.Workspace{
			Version: 1,
			Items:   []domain.WorkspaceItem{},
		}
	}

	workspaceItems := make([]domain.WorkspaceItem, 0, len(items))
	for _, item := range items {
		workspaceItem := domain.WorkspaceItem{
			ID:        item.ID,
			Name:      item.Name,
			Type:      item.Type,
			CreatedAt: item.CreatedAt,
		}

		if item.ParentID != nil {
			parentID := *item.ParentID
			workspaceItem.ParentID = &parentID
		}

		// Copy request-specific fields if this is a request
		if item.Type == "request" {
			if item.Method != nil {
				method := *item.Method
				workspaceItem.Method = &method
			}
			if item.URL != nil {
				url := *item.URL
				workspaceItem.URL = &url
			}
			if item.BodyType != nil {
				bodyType := *item.BodyType
				workspaceItem.BodyType = &bodyType
			}
			if item.Body != nil {
				body := *item.Body
				workspaceItem.Body = &body
			}
			workspaceItem.Headers = item.Headers
			workspaceItem.FormData = item.FormData
		}

		workspaceItems = append(workspaceItems, workspaceItem)
	}

	return &domain.Workspace{
		Version: 1,
		Items:   workspaceItems,
	}
}

// convertHeadersArrayToMap converts an array of Header to a map, including only enabled headers
func convertHeadersArrayToMap(headers []domain.Header) map[string]string {
	result := make(map[string]string)
	for _, header := range headers {
		if header.Enabled && header.Key != "" {
			result[header.Key] = header.Value
		}
	}
	return result
}

// convertFormDataArrayToMap converts an array of FormDataItem to a map, including only enabled items
func convertFormDataArrayToMap(formData []domain.FormDataItem) map[string]string {
	result := make(map[string]string)
	for _, item := range formData {
		if item.Enabled && item.Key != "" {
			result[item.Key] = item.Value
		}
	}
	return result
}

// App struct with exported methods for Wails bindings
type App struct {
	ctx      context.Context
	store    storage.WorkspaceStore
	executor executor.Executor
}

// NewApp creates a new App application struct
func NewApp() *App {
	cfg := config.Load()
	store := storage.NewFileStore(cfg.WorkspacePath())
	exec := executor.NewHTTPExecutor(executor.Config{
		Timeout:     time.Duration(cfg.Timeout) * time.Second,
		MaxBodySize: cfg.MaxBodySize,
	})

	return &App{
		store:    store,
		executor: exec,
	}
}

// OnStartup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) OnStartup(ctx context.Context) {
	a.ctx = ctx
}

// GetWorkspace returns the current workspace as an array of SidebarItem
// This method is exposed to the frontend via Wails bindings
func (a *App) GetWorkspace() ([]SidebarItem, error) {
	workspace, err := a.store.Load()
	if err != nil {
		return nil, err
	}
	return convertWorkspaceToItems(workspace), nil
}

// SaveWorkspace saves the workspace from an array of SidebarItem
// This method is exposed to the frontend via Wails bindings
func (a *App) SaveWorkspace(items []SidebarItem) error {
	workspace := convertItemsToWorkspace(items)
	return a.store.Save(workspace)
}

// ExecuteRequest executes an HTTP request
// This method is exposed to the frontend via Wails bindings
func (a *App) ExecuteRequest(req *executor.Request) (*executor.ExecutionResult, error) {
	return a.executor.Execute(a.ctx, req)
}
