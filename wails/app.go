package main

import (
	"context"
	"time"

	"github.com/constrictor/constrictor-rest-client/internal/config"
	"github.com/constrictor/constrictor-rest-client/internal/domain"
	"github.com/constrictor/constrictor-rest-client/internal/executor"
	"github.com/constrictor/constrictor-rest-client/internal/storage"
)

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

// GetWorkspace returns the current workspace
// This method is exposed to the frontend via Wails bindings
func (a *App) GetWorkspace() (*domain.Workspace, error) {
	return a.store.Load()
}

// SaveWorkspace saves the workspace
// This method is exposed to the frontend via Wails bindings
func (a *App) SaveWorkspace(workspace *domain.Workspace) error {
	return a.store.Save(workspace)
}

// ExecuteRequest executes an HTTP request
// This method is exposed to the frontend via Wails bindings
func (a *App) ExecuteRequest(req *executor.Request) (*executor.ExecutionResult, error) {
	return a.executor.Execute(a.ctx, req)
}
