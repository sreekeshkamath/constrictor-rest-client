package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/constrictor/constrictor-rest-client/internal/domain"
	"github.com/constrictor/constrictor-rest-client/internal/executor"
	"github.com/constrictor/constrictor-rest-client/internal/storage"
)

type mockStore struct {
	workspace *domain.Workspace
	err       error
}

func (m *mockStore) Load() (*domain.Workspace, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.workspace == nil {
		return &domain.Workspace{Version: 1, Items: []domain.WorkspaceItem{}}, nil
	}
	return m.workspace, nil
}

func (m *mockStore) Save(workspace *domain.Workspace) error {
	if m.err != nil {
		return m.err
	}
	m.workspace = workspace
	return nil
}

type mockExecutor struct {
	result *executor.ExecutionResult
	err    error
}

func (m *mockExecutor) Execute(ctx context.Context, req *executor.Request) (*executor.ExecutionResult, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.result == nil {
		return &executor.ExecutionResult{
			Status:    200,
			StatusText: "OK",
			Headers:    map[string]string{},
			Body:       "{}",
			TimeMs:     100,
			SizeBytes:  2,
		}, nil
	}
	return m.result, nil
}

func TestHandleGetWorkspace(t *testing.T) {
	workspace := &domain.Workspace{
		Version: 1,
		Items: []domain.WorkspaceItem{
			{ID: "item-1", Name: "Test", Type: "request", CreatedAt: 1000},
		},
	}

	store := &mockStore{workspace: workspace}
	exec := &mockExecutor{}
	handlers := NewHandlers(store, exec)

	req := httptest.NewRequest("GET", "/api/workspace", nil)
	w := httptest.NewRecorder()

	handlers.HandleGetWorkspace(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response domain.Workspace
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Version != 1 {
		t.Errorf("Expected version 1, got %d", response.Version)
	}

	if len(response.Items) != 1 {
		t.Errorf("Expected 1 item, got %d", len(response.Items))
	}
}

func TestHandlePutWorkspace(t *testing.T) {
	store := &mockStore{}
	exec := &mockExecutor{}
	handlers := NewHandlers(store, exec)

	workspace := &domain.Workspace{
		Version: 1,
		Items: []domain.WorkspaceItem{
			{ID: "item-1", Name: "Test", Type: "request", CreatedAt: 1000},
		},
	}

	body, _ := json.Marshal(workspace)
	req := httptest.NewRequest("PUT", "/api/workspace", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handlers.HandlePutWorkspace(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Verify workspace was saved
	saved, _ := store.Load()
	if len(saved.Items) != 1 {
		t.Errorf("Expected 1 item to be saved, got %d", len(saved.Items))
	}
}

func TestHandlePutWorkspace_InvalidJSON(t *testing.T) {
	store := &mockStore{}
	exec := &mockExecutor{}
	handlers := NewHandlers(store, exec)

	req := httptest.NewRequest("PUT", "/api/workspace", bytes.NewReader([]byte("invalid json")))
	w := httptest.NewRecorder()

	handlers.HandlePutWorkspace(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	var response APIError
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode error response: %v", err)
	}

	if response.Error != "invalid JSON" {
		t.Errorf("Expected error 'invalid JSON', got %s", response.Error)
	}
}

func TestHandleExecute(t *testing.T) {
	store := &mockStore{}
	exec := &mockExecutor{
		result: &executor.ExecutionResult{
			Status:     200,
			StatusText: "OK",
			Headers:    map[string]string{"Content-Type": "application/json"},
			Body:       `{"success": true}`,
			TimeMs:     150,
			SizeBytes:  17,
		},
	}
	handlers := NewHandlers(store, exec)

	reqBody := ExecuteRequest{
		Method:   "GET",
		URL:      "https://api.example.com/test",
		Headers:  []domain.Header{},
		BodyType: "none",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/execute", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handlers.HandleExecute(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response executor.ExecutionResult
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Status != 200 {
		t.Errorf("Expected status 200, got %d", response.Status)
	}

	if response.Body != `{"success": true}` {
		t.Errorf("Expected body '{\"success\": true}', got %s", response.Body)
	}
}

func TestHandleExecute_WithHeaders(t *testing.T) {
	store := &mockStore{}
	exec := &mockExecutor{}
	handlers := NewHandlers(store, exec)

	reqBody := ExecuteRequest{
		Method: "POST",
		URL:    "https://api.example.com/test",
		Headers: []domain.Header{
			{Key: "Authorization", Value: "Bearer token123", Enabled: true},
			{Key: "X-Custom", Value: "value", Enabled: false}, // Should be ignored
		},
		BodyType: "json",
		Body:     `{"key": "value"}`,
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/execute", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handlers.HandleExecute(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestHandleExecute_InvalidRequest(t *testing.T) {
	store := &mockStore{}
	exec := &mockExecutor{}
	handlers := NewHandlers(store, exec)

	// Missing URL
	reqBody := ExecuteRequest{
		Method:   "GET",
		BodyType: "none",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/execute", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handlers.HandleExecute(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	var response APIError
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode error response: %v", err)
	}

	if response.Error != "validation failed" {
		t.Errorf("Expected error 'validation failed', got %s", response.Error)
	}
}

func TestHandleExecute_InvalidBodyType(t *testing.T) {
	store := &mockStore{}
	exec := &mockExecutor{}
	handlers := NewHandlers(store, exec)

	reqBody := ExecuteRequest{
		Method:   "POST",
		URL:      "https://api.example.com/test",
		BodyType: "invalid-type",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/execute", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handlers.HandleExecute(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestConvertHeaders(t *testing.T) {
	headers := []domain.Header{
		{Key: "Header1", Value: "Value1", Enabled: true},
		{Key: "Header2", Value: "Value2", Enabled: false},
		{Key: "", Value: "Value3", Enabled: true}, // Empty key should be ignored
	}

	result := convertHeaders(headers)

	if len(result) != 1 {
		t.Errorf("Expected 1 header, got %d", len(result))
	}

	if result["Header1"] != "Value1" {
		t.Errorf("Expected Header1=Value1, got %s", result["Header1"])
	}

	if result["Header2"] != "" {
		t.Errorf("Expected Header2 to be empty (disabled), got %s", result["Header2"])
	}
}

func TestConvertFormData(t *testing.T) {
	formData := []domain.FormDataItem{
		{Key: "key1", Value: "value1", Enabled: true},
		{Key: "key2", Value: "value2", Enabled: false},
		{Key: "", Value: "value3", Enabled: true}, // Empty key should be ignored
	}

	result := convertFormData(formData)

	if len(result) != 1 {
		t.Errorf("Expected 1 form data item, got %d", len(result))
	}

	if result["key1"] != "value1" {
		t.Errorf("Expected key1=value1, got %s", result["key1"])
	}
}
