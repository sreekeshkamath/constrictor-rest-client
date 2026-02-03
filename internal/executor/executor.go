package executor

import (
	"context"
	"time"
)

// ExecutionResult represents the result of executing an HTTP request
type ExecutionResult struct {
	Status     int               `json:"status"`
	StatusText string            `json:"statusText"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`
	TimeMs     int64             `json:"timeMs"`
	SizeBytes  int64             `json:"sizeBytes"`
	Error      *ExecutionError   `json:"error,omitempty"`
}

// ExecutionError represents an error during request execution
type ExecutionError struct {
	Message string `json:"message"`
	Type    string `json:"type"` // "timeout", "network", "invalid_request", "body_too_large", etc.
}

// Request represents an HTTP request to execute
type Request struct {
	Method    string            `json:"method"`
	URL       string            `json:"url"`
	Headers   map[string]string `json:"headers"`
	BodyType  string            `json:"bodyType"` // "none", "json", "form-data", "url-encoded"
	Body      string            `json:"body"`
	FormData  map[string]string `json:"formData"` // For form-data and url-encoded
	RequestID string            `json:"requestId,omitempty"` // ID of the request item for auth resolution
}

// Executor defines the interface for executing HTTP requests
type Executor interface {
	Execute(ctx context.Context, req *Request) (*ExecutionResult, error)
}

// Config holds executor configuration
type Config struct {
	Timeout     time.Duration
	MaxBodySize int64
}
