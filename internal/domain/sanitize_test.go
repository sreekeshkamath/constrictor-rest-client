package domain

import (
	"testing"
)

func TestSanitizeWorkspace(t *testing.T) {
	tests := []struct {
		name     string
		workspace *Workspace
		want     *Workspace
	}{
		{
			name:     "nil workspace",
			workspace: nil,
			want:     nil,
		},
		{
			name: "removes sensitive headers",
			workspace: &Workspace{
				Version: 1,
				Items: []WorkspaceItem{
					{
						ID:   "req1",
						Name: "Test Request",
						Type: "request",
						Headers: []Header{
							{Key: "Content-Type", Value: "application/json", Enabled: true},
							{Key: "Authorization", Value: "Bearer secret-token", Enabled: true},
							{Key: "X-API-Key", Value: "api-key-123", Enabled: true},
							{Key: "User-Agent", Value: "Constrictor", Enabled: true},
						},
					},
				},
			},
			want: &Workspace{
				Version: 1,
				Items: []WorkspaceItem{
					{
						ID:   "req1",
						Name: "Test Request",
						Type: "request",
						Headers: []Header{
							{Key: "Content-Type", Value: "application/json", Enabled: true},
							{Key: "User-Agent", Value: "Constrictor", Enabled: true},
						},
					},
				},
			},
		},
		{
			name: "removes sensitive form data",
			workspace: &Workspace{
				Version: 1,
				Items: []WorkspaceItem{
					{
						ID:   "req1",
						Name: "Test Request",
						Type: "request",
						FormData: []FormDataItem{
							{Key: "username", Value: "user", Enabled: true},
							{Key: "password", Value: "secret123", Enabled: true},
							{Key: "token", Value: "auth-token", Enabled: true},
							{Key: "email", Value: "user@example.com", Enabled: true},
						},
					},
				},
			},
			want: &Workspace{
				Version: 1,
				Items: []WorkspaceItem{
					{
						ID:   "req1",
						Name: "Test Request",
						Type: "request",
						FormData: []FormDataItem{
							{Key: "username", Value: "user", Enabled: true},
							{Key: "email", Value: "user@example.com", Enabled: true},
						},
					},
				},
			},
		},
		{
			name: "preserves body content",
			workspace: &Workspace{
				Version: 1,
				Items: []WorkspaceItem{
					{
						ID:      "req1",
						Name:    "Test Request",
						Type:    "request",
						Body:    stringPtr(`{"data": "test", "token": "should-be-preserved"}`),
						BodyType: stringPtr("json"),
					},
				},
			},
			want: &Workspace{
				Version: 1,
				Items: []WorkspaceItem{
					{
						ID:      "req1",
						Name:    "Test Request",
						Type:    "request",
						Body:    stringPtr(`{"data": "test", "token": "should-be-preserved"}`),
						BodyType: stringPtr("json"),
					},
				},
			},
		},
		{
			name: "case-insensitive header matching",
			workspace: &Workspace{
				Version: 1,
				Items: []WorkspaceItem{
					{
						ID:   "req1",
						Name: "Test Request",
						Type: "request",
						Headers: []Header{
							{Key: "AUTHORIZATION", Value: "Bearer token", Enabled: true},
							{Key: "x-api-key", Value: "key", Enabled: true},
							{Key: "X-Auth-Token", Value: "token", Enabled: true},
						},
					},
				},
			},
			want: &Workspace{
				Version: 1,
				Items: []WorkspaceItem{
					{
						ID:      "req1",
						Name:    "Test Request",
						Type:    "request",
						Headers: []Header{},
					},
				},
			},
		},
		{
			name: "preserves folders",
			workspace: &Workspace{
				Version: 1,
				Items: []WorkspaceItem{
					{
						ID:   "folder1",
						Name: "My Folder",
						Type: "folder",
					},
				},
			},
			want: &Workspace{
				Version: 1,
				Items: []WorkspaceItem{
					{
						ID:   "folder1",
						Name: "My Folder",
						Type: "folder",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizeWorkspace(tt.workspace)
			if !workspaceEqual(got, tt.want) {
				t.Errorf("SanitizeWorkspace() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func stringPtr(s string) *string {
	return &s
}

func workspaceEqual(a, b *Workspace) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	if a.Version != b.Version {
		return false
	}
	if len(a.Items) != len(b.Items) {
		return false
	}
	for i := range a.Items {
		if !workspaceItemEqual(a.Items[i], b.Items[i]) {
			return false
		}
	}
	return true
}

func workspaceItemEqual(a, b WorkspaceItem) bool {
	if a.ID != b.ID || a.Name != b.Name || a.Type != b.Type {
		return false
	}
	if (a.ParentID == nil) != (b.ParentID == nil) {
		return false
	}
	if a.ParentID != nil && b.ParentID != nil && *a.ParentID != *b.ParentID {
		return false
	}
	if len(a.Headers) != len(b.Headers) {
		return false
	}
	for i := range a.Headers {
		if a.Headers[i].Key != b.Headers[i].Key ||
			a.Headers[i].Value != b.Headers[i].Value ||
			a.Headers[i].Enabled != b.Headers[i].Enabled {
			return false
		}
	}
	if len(a.FormData) != len(b.FormData) {
		return false
	}
	for i := range a.FormData {
		if a.FormData[i].Key != b.FormData[i].Key ||
			a.FormData[i].Value != b.FormData[i].Value ||
			a.FormData[i].Enabled != b.FormData[i].Enabled {
			return false
		}
	}
	return true
}
