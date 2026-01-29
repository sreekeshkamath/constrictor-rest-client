package executor

import (
	"testing"

	"github.com/constrictor/constrictor-rest-client/internal/domain"
)

func TestResolveAuth_None(t *testing.T) {
	resolver := NewAuthResolver()
	item := &domain.WorkspaceItem{
		ID:   "req1",
		Type: "request",
		Auth: &domain.AuthConfig{
			Type:   "none",
			Config: make(map[string]interface{}),
		},
	}
	workspace := &domain.Workspace{
		Version: 1,
		Items:   []domain.WorkspaceItem{*item},
	}

	auth, err := resolver.ResolveAuth(item, workspace)
	if err != nil {
		t.Fatalf("ResolveAuth failed: %v", err)
	}
	if auth.Type != "none" {
		t.Errorf("Expected auth type 'none', got '%s'", auth.Type)
	}
}

func TestResolveAuth_InheritFromParent(t *testing.T) {
	resolver := NewAuthResolver()
	parentID := "folder1"
	parent := &domain.WorkspaceItem{
		ID:   parentID,
		Type: "folder",
		Auth: &domain.AuthConfig{
			Type: "bearer",
			Config: map[string]interface{}{
				"token": "parent-token",
			},
		},
	}
	child := &domain.WorkspaceItem{
		ID:       "req1",
		Type:     "request",
		ParentID: &parentID,
		Auth: &domain.AuthConfig{
			Type:   "inherit",
			Config: make(map[string]interface{}),
		},
	}
	workspace := &domain.Workspace{
		Version: 1,
		Items:   []domain.WorkspaceItem{*parent, *child},
	}

	auth, err := resolver.ResolveAuth(child, workspace)
	if err != nil {
		t.Fatalf("ResolveAuth failed: %v", err)
	}
	if auth.Type != "bearer" {
		t.Errorf("Expected auth type 'bearer', got '%s'", auth.Type)
	}
	if auth.Config["token"] != "parent-token" {
		t.Errorf("Expected token 'parent-token', got '%v'", auth.Config["token"])
	}
}

func TestResolveAuth_InheritFromGrandparent(t *testing.T) {
	resolver := NewAuthResolver()
	grandparentID := "folder1"
	parentID := "folder2"
	grandparent := &domain.WorkspaceItem{
		ID:   grandparentID,
		Type: "folder",
		Auth: &domain.AuthConfig{
			Type: "basic",
			Config: map[string]interface{}{
				"username": "user",
				"password": "pass",
			},
		},
	}
	parent := &domain.WorkspaceItem{
		ID:       parentID,
		Type:     "folder",
		ParentID: &grandparentID,
		Auth: &domain.AuthConfig{
			Type:   "inherit",
			Config: make(map[string]interface{}),
		},
	}
	child := &domain.WorkspaceItem{
		ID:       "req1",
		Type:     "request",
		ParentID: &parentID,
		Auth: &domain.AuthConfig{
			Type:   "inherit",
			Config: make(map[string]interface{}),
		},
	}
	workspace := &domain.Workspace{
		Version: 1,
		Items:   []domain.WorkspaceItem{*grandparent, *parent, *child},
	}

	auth, err := resolver.ResolveAuth(child, workspace)
	if err != nil {
		t.Fatalf("ResolveAuth failed: %v", err)
	}
	if auth.Type != "basic" {
		t.Errorf("Expected auth type 'basic', got '%s'", auth.Type)
	}
}

func TestResolveAuth_NoAuth(t *testing.T) {
	resolver := NewAuthResolver()
	item := &domain.WorkspaceItem{
		ID:   "req1",
		Type: "request",
		Auth: nil,
	}
	workspace := &domain.Workspace{
		Version: 1,
		Items:   []domain.WorkspaceItem{*item},
	}

	auth, err := resolver.ResolveAuth(item, workspace)
	if err != nil {
		t.Fatalf("ResolveAuth failed: %v", err)
	}
	if auth.Type != "none" {
		t.Errorf("Expected auth type 'none', got '%s'", auth.Type)
	}
}

func TestGenerateHeaders_BearerToken(t *testing.T) {
	resolver := NewAuthResolver()
	auth := &domain.AuthConfig{
		Type: "bearer",
		Config: map[string]interface{}{
			"token": "test-token-123",
		},
	}

	headers, err := resolver.GenerateHeaders(auth, "GET", "https://api.example.com")
	if err != nil {
		t.Fatalf("GenerateHeaders failed: %v", err)
	}
	if headers["Authorization"] != "Bearer test-token-123" {
		t.Errorf("Expected 'Bearer test-token-123', got '%s'", headers["Authorization"])
	}
}

func TestGenerateHeaders_BasicAuth(t *testing.T) {
	resolver := NewAuthResolver()
	auth := &domain.AuthConfig{
		Type: "basic",
		Config: map[string]interface{}{
			"username": "user",
			"password": "pass",
		},
	}

	headers, err := resolver.GenerateHeaders(auth, "GET", "https://api.example.com")
	if err != nil {
		t.Fatalf("GenerateHeaders failed: %v", err)
	}
	if headers["Authorization"] == "" {
		t.Error("Expected Authorization header, got empty")
	}
	if headers["Authorization"][:6] != "Basic " {
		t.Errorf("Expected 'Basic ' prefix, got '%s'", headers["Authorization"][:6])
	}
}

func TestGenerateHeaders_APIKey(t *testing.T) {
	resolver := NewAuthResolver()
	auth := &domain.AuthConfig{
		Type: "apikey",
		Config: map[string]interface{}{
			"key":   "X-API-Key",
			"value": "api-key-value",
		},
	}

	headers, err := resolver.GenerateHeaders(auth, "GET", "https://api.example.com")
	if err != nil {
		t.Fatalf("GenerateHeaders failed: %v", err)
	}
	if headers["X-API-Key"] != "api-key-value" {
		t.Errorf("Expected 'api-key-value', got '%s'", headers["X-API-Key"])
	}
}

func TestGenerateHeaders_OAuth2(t *testing.T) {
	resolver := NewAuthResolver()
	auth := &domain.AuthConfig{
		Type: "oauth2",
		Config: map[string]interface{}{
			"accessToken": "oauth-token-123",
			"tokenType":   "Bearer",
		},
	}

	headers, err := resolver.GenerateHeaders(auth, "GET", "https://api.example.com")
	if err != nil {
		t.Fatalf("GenerateHeaders failed: %v", err)
	}
	if headers["Authorization"] != "Bearer oauth-token-123" {
		t.Errorf("Expected 'Bearer oauth-token-123', got '%s'", headers["Authorization"])
	}
}

func TestGenerateHeaders_None(t *testing.T) {
	resolver := NewAuthResolver()
	auth := &domain.AuthConfig{
		Type:   "none",
		Config: make(map[string]interface{}),
	}

	headers, err := resolver.GenerateHeaders(auth, "GET", "https://api.example.com")
	if err != nil {
		t.Fatalf("GenerateHeaders failed: %v", err)
	}
	if len(headers) != 0 {
		t.Errorf("Expected no headers, got %d", len(headers))
	}
}

func TestGenerateHeaders_MissingToken(t *testing.T) {
	resolver := NewAuthResolver()
	auth := &domain.AuthConfig{
		Type:   "bearer",
		Config: make(map[string]interface{}),
	}

	_, err := resolver.GenerateHeaders(auth, "GET", "https://api.example.com")
	if err == nil {
		t.Error("Expected error for missing token, got nil")
	}
}
