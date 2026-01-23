package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/constrictor/constrictor-rest-client/internal/domain"
)

// WorkspaceStore defines the interface for workspace persistence
type WorkspaceStore interface {
	Load() (*domain.Workspace, error)
	Save(workspace *domain.Workspace) error
}

// FileStore implements WorkspaceStore using a JSON file
type FileStore struct {
	path string
	mu   sync.RWMutex
}

// NewFileStore creates a new file-backed workspace store
func NewFileStore(workspacePath string) *FileStore {
	return &FileStore{
		path: workspacePath,
	}
}

// Load reads the workspace from the file
func (s *FileStore) Load() (*domain.Workspace, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty workspace if file doesn't exist
			return &domain.Workspace{
				Version: 1,
				Items:   []domain.WorkspaceItem{},
			}, nil
		}
		return nil, fmt.Errorf("failed to read workspace file: %w", err)
	}

	var workspace domain.Workspace
	if err := json.Unmarshal(data, &workspace); err != nil {
		return nil, fmt.Errorf("failed to parse workspace JSON: %w", err)
	}

	// Ensure version is set (for backward compatibility)
	if workspace.Version == 0 {
		workspace.Version = 1
	}

	return &workspace, nil
}

// Save writes the workspace to the file atomically
func (s *FileStore) Save(workspace *domain.Workspace) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Ensure version is set
	if workspace.Version == 0 {
		workspace.Version = 1
	}

	// Marshal to JSON with indentation
	data, err := json.MarshalIndent(workspace, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal workspace: %w", err)
	}

	// Ensure directory exists
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Atomic write: write to temp file, then rename
	tmpPath := s.path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tmpPath, s.path); err != nil {
		os.Remove(tmpPath) // Clean up temp file on error
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	return nil
}
