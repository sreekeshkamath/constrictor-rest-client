package main

import (
	"context"
	"time"

	"github.com/constrictor/constrictor-rest-client/internal/config"
	"github.com/constrictor/constrictor-rest-client/internal/domain"
	"github.com/constrictor/constrictor-rest-client/internal/executor"
	"github.com/constrictor/constrictor-rest-client/internal/storage"
)

// App struct
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

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) OnStartup(ctx context.Context) {
	a.ctx = ctx
}

// GetWorkspace returns the current workspace
func (a *App) GetWorkspace() (*domain.Workspace, error) {
	return a.store.Load()
}

// SaveWorkspace saves the workspace
func (a *App) SaveWorkspace(workspace *domain.Workspace) error {
	return a.store.Save(workspace)
}

// ExecuteRequest executes an HTTP request
func (a *App) ExecuteRequest(req *executor.Request) (*executor.ExecutionResult, error) {
	return a.executor.Execute(a.ctx, req)
}
