package config

import (
	"os"
	"path/filepath"
)

type Config struct {
	Port        string
	DataPath    string
	MaxBodySize int64
	Timeout     int // seconds
}

func Load() *Config {
	dataPath := os.Getenv("CONSTRICTOR_DATA_PATH")
	if dataPath == "" {
		dataPath = "data"
	}

	// Ensure data directory exists
	if err := os.MkdirAll(dataPath, 0755); err != nil {
		panic(err)
	}

	return &Config{
		Port:        getEnv("PORT", "8080"),
		DataPath:    dataPath,
		MaxBodySize: 10 * 1024 * 1024, // 10MB
		Timeout:     30,               // 30 seconds
	}
}

func (c *Config) WorkspacePath() string {
	return filepath.Join(c.DataPath, "workspace.json")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
