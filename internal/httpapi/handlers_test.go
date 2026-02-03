package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/constrictor/constrictor-rest-client/internal/domain"
	"github.com/constrictor/constrictor-rest-client/internal/executor"
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

func TestHandleExecute_WithBearerAuth_httptest(t *testing.T) {
	requestID := "req1"
	workspace := &domain.Workspace{
		Version: 1,
		Items: []domain.WorkspaceItem{
			{
				ID:   requestID,
				Type: "request",
				Auth: &domain.AuthConfig{
					Type: "bearer",
					Config: map[string]interface{}{
						"token": "test-bearer-token",
					},
				},
			},
		},
	}
	store := &mockStore{workspace: workspace}
	exec := executor.NewHTTPExecutor(executor.Config{
		Timeout:     30 * time.Second,
		MaxBodySize: 1024 * 1024,
	})
	handlers := NewHandlers(store, exec)

	var receivedAuthHeader string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAuthHeader = r.Header.Get("Authorization")
		t.Log("Server received request, auth header:", receivedAuthHeader)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"auth": "received"}`))
	}))
	defer server.Close()

	reqBody := ExecuteRequest{
		Method:    "GET",
		URL:       server.URL,
		Headers:   []domain.Header{},
		BodyType:  "none",
		RequestID: requestID,
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

	if response.Error != nil {
		t.Fatalf("Execution error: %v", response.Error)
	}

	if response.Status != http.StatusOK {
		t.Errorf("Expected response status 200, got %d", response.Status)
	}

	expectedAuthHeader := "Bearer test-bearer-token"
	if receivedAuthHeader != expectedAuthHeader {
		t.Errorf("Expected Authorization header %q, got %q", expectedAuthHeader, receivedAuthHeader)
	}
}
