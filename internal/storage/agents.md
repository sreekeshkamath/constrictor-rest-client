# internal/storage

## Purpose

Workspace persistence. Provides an interface for loading and saving workspaces with atomic writes and thread safety.

## Key Files

- `store.go` - WorkspaceStore interface and FileStore implementation
- `store_test.go` - Comprehensive tests

## Interface

```go
type WorkspaceStore interface {
    Load() (*domain.Workspace, error)
    Save(workspace *domain.Workspace) error
}
```

## Features

- Atomic writes (temp file + rename pattern)
- Thread-safe (mutex-protected)
- Versioned JSON format
- Backward compatibility (auto-sets version if missing)
- Returns empty workspace if file doesn't exist

## How to Test

```bash
go test ./internal/storage/... -v
```

## Usage Example

```go
store := storage.NewFileStore("/path/to/workspace.json")

// Load workspace
workspace, err := store.Load()

// Save workspace
err = store.Save(workspace)
```

## File Format

Workspace is stored as JSON:
```json
{
  "version": 1,
  "items": [...]
}
```
