package storage

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/constrictor/constrictor-rest-client/internal/domain"
)

func TestFileStore_Load_NonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewFileStore(filepath.Join(tmpDir, "workspace.json"))

	workspace, err := store.Load()
	if err != nil {
		t.Fatalf("Expected no error for non-existent file, got: %v", err)
	}

	if workspace.Version != 1 {
		t.Errorf("Expected version 1, got %d", workspace.Version)
	}

	if len(workspace.Items) != 0 {
		t.Errorf("Expected empty items, got %d items", len(workspace.Items))
	}
}

func TestFileStore_SaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewFileStore(filepath.Join(tmpDir, "workspace.json"))

	// Create a test workspace
	workspace := &domain.Workspace{
		Version: 1,
		Items: []domain.WorkspaceItem{
			{
				ID:        "req-1",
				Name:      "Test Request",
				Type:      "request",
				CreatedAt: 1234567890,
				Method:    stringPtr("GET"),
				URL:       stringPtr("https://api.example.com"),
			},
			{
				ID:        "folder-1",
				Name:      "Test Folder",
				Type:      "folder",
				CreatedAt: 1234567891,
			},
		},
	}

	// Save
	if err := store.Save(workspace); err != nil {
		t.Fatalf("Failed to save workspace: %v", err)
	}

	// Load
	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("Failed to load workspace: %v", err)
	}

	// Verify
	if loaded.Version != 1 {
		t.Errorf("Expected version 1, got %d", loaded.Version)
	}

	if len(loaded.Items) != 2 {
		t.Fatalf("Expected 2 items, got %d", len(loaded.Items))
	}

	if loaded.Items[0].ID != "req-1" || loaded.Items[0].Name != "Test Request" {
		t.Errorf("First item mismatch: got %+v", loaded.Items[0])
	}

	if loaded.Items[1].ID != "folder-1" || loaded.Items[1].Name != "Test Folder" {
		t.Errorf("Second item mismatch: got %+v", loaded.Items[1])
	}
}

func TestFileStore_AtomicWrite(t *testing.T) {
	tmpDir := t.TempDir()
	workspacePath := filepath.Join(tmpDir, "workspace.json")
	store := NewFileStore(workspacePath)

	// Save initial workspace
	workspace1 := &domain.Workspace{
		Version: 1,
		Items: []domain.WorkspaceItem{
			{ID: "item-1", Name: "Item 1", Type: "request", CreatedAt: 1000},
		},
	}

	if err := store.Save(workspace1); err != nil {
		t.Fatalf("Failed to save: %v", err)
	}

	// Verify file exists and is valid
	if _, err := os.Stat(workspacePath); err != nil {
		t.Fatalf("Workspace file should exist: %v", err)
	}

	// Save updated workspace
	workspace2 := &domain.Workspace{
		Version: 1,
		Items: []domain.WorkspaceItem{
			{ID: "item-1", Name: "Item 1 Updated", Type: "request", CreatedAt: 1000},
			{ID: "item-2", Name: "Item 2", Type: "folder", CreatedAt: 2000},
		},
	}

	if err := store.Save(workspace2); err != nil {
		t.Fatalf("Failed to save updated workspace: %v", err)
	}

	// Verify temp file was cleaned up
	tmpPath := workspacePath + ".tmp"
	if _, err := os.Stat(tmpPath); err == nil {
		t.Error("Temp file should not exist after successful save")
	}

	// Load and verify
	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("Failed to load: %v", err)
	}

	if len(loaded.Items) != 2 {
		t.Errorf("Expected 2 items after update, got %d", len(loaded.Items))
	}

	if loaded.Items[0].Name != "Item 1 Updated" {
		t.Errorf("Expected updated name, got %s", loaded.Items[0].Name)
	}
}

func TestFileStore_ConcurrentAccess(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewFileStore(filepath.Join(tmpDir, "workspace.json"))

	// Test concurrent reads and writes
	done := make(chan bool, 10)

	for i := 0; i < 5; i++ {
		go func(id int) {
			workspace := &domain.Workspace{
				Version: 1,
				Items: []domain.WorkspaceItem{
					{ID: "item", Name: "Item", Type: "request", CreatedAt: int64(id)},
				},
			}
			store.Save(workspace)
			done <- true
		}(i)
	}

	for i := 0; i < 5; i++ {
		go func() {
			store.Load()
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Final load should succeed
	_, err := store.Load()
	if err != nil {
		t.Fatalf("Failed to load after concurrent access: %v", err)
	}
}

func TestFileStore_BackwardCompatibility(t *testing.T) {
	tmpDir := t.TempDir()
	workspacePath := filepath.Join(tmpDir, "workspace.json")

	// Create a workspace file without version (old format)
	oldWorkspace := `{"items": [{"id": "old-item", "name": "Old", "type": "request", "createdAt": 1000}]}`
	if err := os.WriteFile(workspacePath, []byte(oldWorkspace), 0644); err != nil {
		t.Fatalf("Failed to write old format file: %v", err)
	}

	store := NewFileStore(workspacePath)
	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("Failed to load old format: %v", err)
	}

	// Version should be set to 1
	if loaded.Version != 1 {
		t.Errorf("Expected version 1 for backward compatibility, got %d", loaded.Version)
	}

	if len(loaded.Items) != 1 {
		t.Errorf("Expected 1 item, got %d", len(loaded.Items))
	}
}

func stringPtr(s string) *string {
	return &s
}
