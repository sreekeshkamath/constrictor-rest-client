package main

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/constrictor/constrictor-rest-client/internal/domain"
	"github.com/constrictor/constrictor-rest-client/internal/executor"
	"github.com/constrictor/constrictor-rest-client/internal/storage"
)

func TestConvertWorkspaceToItems(t *testing.T) {
	tests := []struct {
		name     string
		workspace *domain.Workspace
		want     []SidebarItem
	}{
		{
			name:     "nil workspace",
			workspace: nil,
			want:     []SidebarItem{},
		},
		{
			name: "empty workspace",
			workspace: &domain.Workspace{
				Version: 1,
				Items:   []domain.WorkspaceItem{},
			},
			want: []SidebarItem{},
		},
		{
			name: "workspace with folder",
			workspace: &domain.Workspace{
				Version: 1,
				Items: []domain.WorkspaceItem{
					{
						ID:        "folder1",
						Name:      "My Folder",
						Type:      "folder",
						CreatedAt: 1234567890,
					},
				},
			},
			want: []SidebarItem{
				{
					ID:        "folder1",
					Name:      "My Folder",
					Type:      "folder",
					CreatedAt: 1234567890,
				},
			},
		},
		{
			name: "workspace with request",
			workspace: &domain.Workspace{
				Version: 1,
				Items: []domain.WorkspaceItem{
					{
						ID:        "req1",
						Name:      "Get Users",
						Type:      "request",
						Method:    stringPtr("GET"),
						URL:       stringPtr("https://api.example.com/users"),
						BodyType:  stringPtr("none"),
						Body:      stringPtr(""),
						Headers:   []domain.Header{{Key: "Content-Type", Value: "application/json", Enabled: true}},
						FormData:  []domain.FormDataItem{},
						CreatedAt: 1234567890,
					},
				},
			},
			want: []SidebarItem{
				{
					ID:        "req1",
					Name:      "Get Users",
					Type:      "request",
					Method:    stringPtr("GET"),
					URL:       stringPtr("https://api.example.com/users"),
					BodyType:  stringPtr("none"),
					Body:      stringPtr(""),
					Headers:   []domain.Header{{Key: "Content-Type", Value: "application/json", Enabled: true}},
					FormData:  []domain.FormDataItem{},
					CreatedAt: 1234567890,
				},
			},
		},
		{
			name: "workspace with parent ID",
			workspace: &domain.Workspace{
				Version: 1,
				Items: []domain.WorkspaceItem{
					{
						ID:        "req1",
						Name:      "Nested Request",
						Type:      "request",
						ParentID:  stringPtr("folder1"),
						CreatedAt: 1234567890,
					},
				},
			},
			want: []SidebarItem{
				{
					ID:        "req1",
					Name:      "Nested Request",
					Type:      "request",
					ParentID:  stringPtr("folder1"),
					CreatedAt: 1234567890,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertWorkspaceToItems(tt.workspace)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("convertWorkspaceToItems() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConvertItemsToWorkspace(t *testing.T) {
	tests := []struct {
		name  string
		items []SidebarItem
		want  *domain.Workspace
	}{
		{
			name:  "empty items",
			items: []SidebarItem{},
			want: &domain.Workspace{
				Version: 1,
				Items:   []domain.WorkspaceItem{},
			},
		},
		{
			name: "items with folder",
			items: []SidebarItem{
				{
					ID:        "folder1",
					Name:      "My Folder",
					Type:      "folder",
					CreatedAt: 1234567890,
				},
			},
			want: &domain.Workspace{
				Version: 1,
				Items: []domain.WorkspaceItem{
					{
						ID:        "folder1",
						Name:      "My Folder",
						Type:      "folder",
						CreatedAt: 1234567890,
					},
				},
			},
		},
		{
			name: "items with request",
			items: []SidebarItem{
				{
					ID:        "req1",
					Name:      "Get Users",
					Type:      "request",
					Method:    stringPtr("GET"),
					URL:       stringPtr("https://api.example.com/users"),
					BodyType:  stringPtr("none"),
					Body:      stringPtr(""),
					Headers:   []domain.Header{{Key: "Content-Type", Value: "application/json", Enabled: true}},
					FormData:  []domain.FormDataItem{},
					CreatedAt: 1234567890,
				},
			},
			want: &domain.Workspace{
				Version: 1,
				Items: []domain.WorkspaceItem{
					{
						ID:        "req1",
						Name:      "Get Users",
						Type:      "request",
						Method:    stringPtr("GET"),
						URL:       stringPtr("https://api.example.com/users"),
						BodyType:  stringPtr("none"),
						Body:      stringPtr(""),
						Headers:   []domain.Header{{Key: "Content-Type", Value: "application/json", Enabled: true}},
						FormData:  []domain.FormDataItem{},
						CreatedAt: 1234567890,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertItemsToWorkspace(tt.items)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("convertItemsToWorkspace() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConvertWorkspaceRoundTrip(t *testing.T) {
	original := &domain.Workspace{
		Version: 1,
		Items: []domain.WorkspaceItem{
			{
				ID:        "folder1",
				Name:      "My Folder",
				Type:      "folder",
				CreatedAt: 1234567890,
			},
			{
				ID:        "req1",
				Name:      "Get Users",
				Type:      "request",
				Method:    stringPtr("GET"),
				URL:       stringPtr("https://api.example.com/users"),
				BodyType:  stringPtr("json"),
				Body:      stringPtr(`{"key": "value"}`),
				Headers: []domain.Header{
					{Key: "Content-Type", Value: "application/json", Enabled: true},
					{Key: "Authorization", Value: "Bearer token", Enabled: false},
				},
				FormData: []domain.FormDataItem{
					{Key: "field1", Value: "value1", Enabled: true},
					{Key: "field2", Value: "value2", Enabled: false},
				},
				ParentID:  stringPtr("folder1"),
				CreatedAt: 1234567891,
			},
		},
	}

	// Convert workspace to items
	items := convertWorkspaceToItems(original)

	// Convert items back to workspace
	result := convertItemsToWorkspace(items)

	// Compare
	if !reflect.DeepEqual(result, original) {
		t.Errorf("Round-trip conversion failed:\nOriginal: %+v\nResult: %+v", original, result)
	}
}

func TestConvertHeadersArrayToMap(t *testing.T) {
	tests := []struct {
		name    string
		headers []domain.Header
		want    map[string]string
	}{
		{
			name:    "empty headers",
			headers: []domain.Header{},
			want:    map[string]string{},
		},
		{
			name: "enabled headers only",
			headers: []domain.Header{
				{Key: "Content-Type", Value: "application/json", Enabled: true},
				{Key: "Authorization", Value: "Bearer token", Enabled: false},
				{Key: "X-Custom", Value: "value", Enabled: true},
			},
			want: map[string]string{
				"Content-Type": "application/json",
				"X-Custom":     "value",
			},
		},
		{
			name: "empty key ignored",
			headers: []domain.Header{
				{Key: "", Value: "value", Enabled: true},
				{Key: "Valid-Key", Value: "value", Enabled: true},
			},
			want: map[string]string{
				"Valid-Key": "value",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertHeadersArrayToMap(tt.headers)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("convertHeadersArrayToMap() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConvertFormDataArrayToMap(t *testing.T) {
	tests := []struct {
		name     string
		formData []domain.FormDataItem
		want     map[string]string
	}{
		{
			name:     "empty form data",
			formData: []domain.FormDataItem{},
			want:     map[string]string{},
		},
		{
			name: "enabled items only",
			formData: []domain.FormDataItem{
				{Key: "field1", Value: "value1", Enabled: true},
				{Key: "field2", Value: "value2", Enabled: false},
				{Key: "field3", Value: "value3", Enabled: true},
			},
			want: map[string]string{
				"field1": "value1",
				"field3": "value3",
			},
		},
		{
			name: "empty key ignored",
			formData: []domain.FormDataItem{
				{Key: "", Value: "value", Enabled: true},
				{Key: "valid-key", Value: "value", Enabled: true},
			},
			want: map[string]string{
				"valid-key": "value",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertFormDataArrayToMap(tt.formData)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("convertFormDataArrayToMap() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}

// mockStore is a mock implementation of WorkspaceStore for testing
type mockStore struct {
	workspace *domain.Workspace
	loadError error
	saveError error
}

func (m *mockStore) Load() (*domain.Workspace, error) {
	if m.loadError != nil {
		return nil, m.loadError
	}
	if m.workspace == nil {
		return &domain.Workspace{
			Version: 1,
			Items:   []domain.WorkspaceItem{},
		}, nil
	}
	return m.workspace, nil
}

func (m *mockStore) Save(workspace *domain.Workspace) error {
	if m.saveError != nil {
		return m.saveError
	}
	m.workspace = workspace
	return nil
}

func TestApp_GetWorkspace(t *testing.T) {
	tests := []struct {
		name      string
		store     *mockStore
		want      []SidebarItem
		wantError bool
	}{
		{
			name: "empty workspace",
			store: &mockStore{
				workspace: &domain.Workspace{
					Version: 1,
					Items:   []domain.WorkspaceItem{},
				},
			},
			want:      []SidebarItem{},
			wantError: false,
		},
		{
			name: "workspace with items",
			store: &mockStore{
				workspace: &domain.Workspace{
					Version: 1,
					Items: []domain.WorkspaceItem{
						{
							ID:        "folder1",
							Name:      "My Folder",
							Type:      "folder",
							CreatedAt: 1234567890,
						},
						{
							ID:        "req1",
							Name:      "Get Users",
							Type:      "request",
							Method:    stringPtr("GET"),
							URL:       stringPtr("https://api.example.com/users"),
							BodyType:  stringPtr("none"),
							Body:      stringPtr(""),
							Headers:   []domain.Header{},
							FormData:  []domain.FormDataItem{},
							CreatedAt: 1234567891,
						},
					},
				},
			},
			want: []SidebarItem{
				{
					ID:        "folder1",
					Name:      "My Folder",
					Type:      "folder",
					CreatedAt: 1234567890,
				},
				{
					ID:        "req1",
					Name:      "Get Users",
					Type:      "request",
					Method:    stringPtr("GET"),
					URL:       stringPtr("https://api.example.com/users"),
					BodyType:  stringPtr("none"),
					Body:      stringPtr(""),
					Headers:   []domain.Header{},
					FormData:  []domain.FormDataItem{},
					CreatedAt: 1234567891,
				},
			},
			wantError: false,
		},
		{
			name: "load error",
			store: &mockStore{
				loadError: errors.New("failed to load"),
			},
			want:      nil,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := &App{
				store: tt.store,
			}
			got, err := app.GetWorkspace()
			if (err != nil) != tt.wantError {
				t.Errorf("App.GetWorkspace() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("App.GetWorkspace() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestApp_SaveWorkspace(t *testing.T) {
	tests := []struct {
		name      string
		items     []SidebarItem
		store     *mockStore
		wantError bool
		wantSaved *domain.Workspace
	}{
		{
			name:  "save empty items",
			items: []SidebarItem{},
			store: &mockStore{},
			wantSaved: &domain.Workspace{
				Version: 1,
				Items:   []domain.WorkspaceItem{},
			},
			wantError: false,
		},
		{
			name: "save items with folder and request",
			items: []SidebarItem{
				{
					ID:        "folder1",
					Name:      "My Folder",
					Type:      "folder",
					CreatedAt: 1234567890,
				},
				{
					ID:        "req1",
					Name:      "Get Users",
					Type:      "request",
					Method:    stringPtr("GET"),
					URL:       stringPtr("https://api.example.com/users"),
					BodyType:  stringPtr("none"),
					Body:      stringPtr(""),
					Headers:   []domain.Header{},
					FormData:  []domain.FormDataItem{},
					CreatedAt: 1234567891,
				},
			},
			store: &mockStore{},
			wantSaved: &domain.Workspace{
				Version: 1,
				Items: []domain.WorkspaceItem{
					{
						ID:        "folder1",
						Name:      "My Folder",
						Type:      "folder",
						CreatedAt: 1234567890,
					},
					{
						ID:        "req1",
						Name:      "Get Users",
						Type:      "request",
						Method:    stringPtr("GET"),
						URL:       stringPtr("https://api.example.com/users"),
						BodyType:  stringPtr("none"),
						Body:      stringPtr(""),
						Headers:   []domain.Header{},
						FormData:  []domain.FormDataItem{},
						CreatedAt: 1234567891,
					},
				},
			},
			wantError: false,
		},
		{
			name: "save error",
			items: []SidebarItem{
				{
					ID:        "req1",
					Name:      "Test Request",
					Type:      "request",
					CreatedAt: 1234567890,
				},
			},
			store: &mockStore{
				saveError: errors.New("failed to save"),
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := &App{
				store: tt.store,
			}
			err := app.SaveWorkspace(tt.items)
			if (err != nil) != tt.wantError {
				t.Errorf("App.SaveWorkspace() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError && tt.wantSaved != nil {
				if !reflect.DeepEqual(tt.store.workspace, tt.wantSaved) {
					t.Errorf("App.SaveWorkspace() saved workspace = %v, want %v", tt.store.workspace, tt.wantSaved)
				}
			}
		})
	}
}

// mockExecutor is a mock implementation of Executor for testing
type mockExecutor struct {
	result *executor.ExecutionResult
	err    error
}

func (m *mockExecutor) Execute(ctx context.Context, req *executor.Request) (*executor.ExecutionResult, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.result, nil
}

func TestApp_ExecuteRequest(t *testing.T) {
	tests := []struct {
		name      string
		input     *ExecuteRequestInput
		executor  *mockExecutor
		want      *executor.ExecutionResult
		wantError bool
		errorMsg  string
	}{
		{
			name: "GET request with headers",
			input: &ExecuteRequestInput{
				Method:   "GET",
				URL:      "https://api.example.com/users",
				Headers:  []domain.Header{{Key: "Authorization", Value: "Bearer token", Enabled: true}},
				BodyType: "none",
				Body:     "",
				FormData: []domain.FormDataItem{},
			},
			executor: &mockExecutor{
				result: &executor.ExecutionResult{
					Status:     200,
					StatusText: "OK",
					Headers:    map[string]string{"Content-Type": "application/json"},
					Body:       `[{"id": 1, "name": "John"}]`,
					TimeMs:     150,
					SizeBytes:  25,
				},
			},
			want: &executor.ExecutionResult{
				Status:     200,
				StatusText: "OK",
				Headers:    map[string]string{"Content-Type": "application/json"},
				Body:       `[{"id": 1, "name": "John"}]`,
				TimeMs:     150,
				SizeBytes:  25,
			},
			wantError: false,
		},
		{
			name: "POST request with JSON body",
			input: &ExecuteRequestInput{
				Method:   "POST",
				URL:      "https://api.example.com/users",
				Headers:  []domain.Header{{Key: "Content-Type", Value: "application/json", Enabled: true}},
				BodyType: "json",
				Body:     `{"name": "John", "email": "john@example.com"}`,
				FormData: []domain.FormDataItem{},
			},
			executor: &mockExecutor{
				result: &executor.ExecutionResult{
					Status:     201,
					StatusText: "Created",
					Headers:    map[string]string{"Content-Type": "application/json"},
					Body:       `{"id": 1, "name": "John"}`,
					TimeMs:     200,
					SizeBytes:  20,
				},
			},
			want: &executor.ExecutionResult{
				Status:     201,
				StatusText: "Created",
				Headers:    map[string]string{"Content-Type": "application/json"},
				Body:       `{"id": 1, "name": "John"}`,
				TimeMs:     200,
				SizeBytes:  20,
			},
			wantError: false,
		},
		{
			name: "POST request with form-data",
			input: &ExecuteRequestInput{
				Method:   "POST",
				URL:      "https://api.example.com/upload",
				Headers:  []domain.Header{{Key: "Content-Type", Value: "multipart/form-data", Enabled: true}},
				BodyType: "form-data",
				Body:     "",
				FormData: []domain.FormDataItem{
					{Key: "field1", Value: "value1", Enabled: true},
					{Key: "field2", Value: "value2", Enabled: false}, // Should be filtered out
					{Key: "field3", Value: "value3", Enabled: true},
				},
			},
			executor: &mockExecutor{
				result: &executor.ExecutionResult{
					Status:     200,
					StatusText: "OK",
					Headers:    map[string]string{"Content-Type": "application/json"},
					Body:       `{"success": true}`,
					TimeMs:     100,
					SizeBytes:  15,
				},
			},
			want: &executor.ExecutionResult{
				Status:     200,
				StatusText: "OK",
				Headers:    map[string]string{"Content-Type": "application/json"},
				Body:       `{"success": true}`,
				TimeMs:     100,
				SizeBytes:  15,
			},
			wantError: false,
		},
		{
			name: "POST request with url-encoded",
			input: &ExecuteRequestInput{
				Method:   "POST",
				URL:      "https://api.example.com/login",
				Headers:  []domain.Header{{Key: "Content-Type", Value: "application/x-www-form-urlencoded", Enabled: true}},
				BodyType: "url-encoded",
				Body:     "",
				FormData: []domain.FormDataItem{
					{Key: "username", Value: "user", Enabled: true},
					{Key: "password", Value: "pass", Enabled: true},
				},
			},
			executor: &mockExecutor{
				result: &executor.ExecutionResult{
					Status:     200,
					StatusText: "OK",
					Headers:    map[string]string{"Content-Type": "application/json"},
					Body:       `{"token": "abc123"}`,
					TimeMs:     120,
					SizeBytes:  18,
				},
			},
			want: &executor.ExecutionResult{
				Status:     200,
				StatusText: "OK",
				Headers:    map[string]string{"Content-Type": "application/json"},
				Body:       `{"token": "abc123"}`,
				TimeMs:     120,
				SizeBytes:  18,
			},
			wantError: false,
		},
		{
			name: "nil input",
			input: nil,
			executor: &mockExecutor{},
			want:      nil,
			wantError: true,
			errorMsg:  "request input cannot be nil",
		},
		{
			name: "empty method",
			input: &ExecuteRequestInput{
				Method:   "",
				URL:      "https://api.example.com/users",
				Headers:  []domain.Header{},
				BodyType: "none",
				Body:     "",
				FormData: []domain.FormDataItem{},
			},
			executor: &mockExecutor{},
			want:      nil,
			wantError: true,
			errorMsg:  "method is required",
		},
		{
			name: "empty URL",
			input: &ExecuteRequestInput{
				Method:   "GET",
				URL:      "",
				Headers:  []domain.Header{},
				BodyType: "none",
				Body:     "",
				FormData: []domain.FormDataItem{},
			},
			executor: &mockExecutor{},
			want:      nil,
			wantError: true,
			errorMsg:  "url is required",
		},
		{
			name: "executor error",
			input: &ExecuteRequestInput{
				Method:   "GET",
				URL:      "https://api.example.com/users",
				Headers:  []domain.Header{},
				BodyType: "none",
				Body:     "",
				FormData: []domain.FormDataItem{},
			},
			executor: &mockExecutor{
				err: errors.New("network error"),
			},
			want:      nil,
			wantError: true,
			errorMsg:  "network error",
		},
		{
			name: "request with disabled headers/formData",
			input: &ExecuteRequestInput{
				Method:   "GET",
				URL:      "https://api.example.com/users",
				Headers: []domain.Header{
					{Key: "Authorization", Value: "Bearer token", Enabled: true},
					{Key: "X-Custom", Value: "value", Enabled: false}, // Should be filtered out
				},
				BodyType: "none",
				Body:     "",
				FormData: []domain.FormDataItem{
					{Key: "field1", Value: "value1", Enabled: false}, // Should be filtered out
				},
			},
			executor: &mockExecutor{
				result: &executor.ExecutionResult{
					Status:     200,
					StatusText: "OK",
					Headers:    map[string]string{"Content-Type": "application/json"},
					Body:       `[]`,
					TimeMs:     100,
					SizeBytes:  2,
				},
			},
			want: &executor.ExecutionResult{
				Status:     200,
				StatusText: "OK",
				Headers:    map[string]string{"Content-Type": "application/json"},
				Body:       `[]`,
				TimeMs:     100,
				SizeBytes:  2,
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := &App{
				ctx:      context.Background(),
				executor: tt.executor,
			}
			got, err := app.ExecuteRequest(tt.input)
			if (err != nil) != tt.wantError {
				t.Errorf("App.ExecuteRequest() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if tt.wantError {
				if err != nil && tt.errorMsg != "" && err.Error() != tt.errorMsg {
					t.Errorf("App.ExecuteRequest() error = %v, want error message %v", err, tt.errorMsg)
				}
			} else {
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("App.ExecuteRequest() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}
