package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	// Test with default values
	cfg := Load()
	if cfg.Port != "8080" {
		t.Errorf("Expected default port 8080, got %s", cfg.Port)
	}
	if cfg.DataPath != "data" {
		t.Errorf("Expected default data path 'data', got %s", cfg.DataPath)
	}
	if cfg.MaxBodySize != 10*1024*1024 {
		t.Errorf("Expected max body size 10MB, got %d", cfg.MaxBodySize)
	}
	if cfg.Timeout != 30 {
		t.Errorf("Expected timeout 30s, got %d", cfg.Timeout)
	}
}

func TestLoadWithEnvVars(t *testing.T) {
	os.Setenv("PORT", "9000")
	os.Setenv("CONSTRICTOR_DATA_PATH", "/tmp/test-data")
	defer os.Unsetenv("PORT")
	defer os.Unsetenv("CONSTRICTOR_DATA_PATH")

	cfg := Load()
	if cfg.Port != "9000" {
		t.Errorf("Expected port 9000 from env, got %s", cfg.Port)
	}
	if cfg.DataPath != "/tmp/test-data" {
		t.Errorf("Expected data path /tmp/test-data from env, got %s", cfg.DataPath)
	}
}

func TestWorkspacePath(t *testing.T) {
	cfg := &Config{DataPath: "/tmp/test"}
	expected := filepath.Join("/tmp/test", "workspace.json")
	if cfg.WorkspacePath() != expected {
		t.Errorf("Expected workspace path %s, got %s", expected, cfg.WorkspacePath())
	}
}
