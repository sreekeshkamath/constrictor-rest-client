package executor

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/constrictor/constrictor-rest-client/internal/domain"
)

// HTTPExecutor implements Executor using net/http
type HTTPExecutor struct {
	client       *http.Client
	config       Config
	authResolver *AuthResolver
}

// NewHTTPExecutor creates a new HTTP executor
func NewHTTPExecutor(config Config) *HTTPExecutor {
	return &HTTPExecutor{
		client: &http.Client{
			Timeout: config.Timeout,
		},
		config:       config,
		authResolver: NewAuthResolver(),
	}
}

// Execute executes an HTTP request and returns the result
func (e *HTTPExecutor) Execute(ctx context.Context, req *Request) (*ExecutionResult, error) {
	startTime := time.Now()

	// Validate request
	if err := e.validateRequest(req); err != nil {
		return &ExecutionResult{
			Error: &ExecutionError{
				Message: err.Error(),
				Type:    "invalid_request",
			},
		}, nil
	}

	// Create HTTP request
	httpReq, err := e.buildHTTPRequest(ctx, req)
	if err != nil {
		return &ExecutionResult{
			Error: &ExecutionError{
				Message: err.Error(),
				Type:    "invalid_request",
			},
		}, nil
	}

	// Execute request
	resp, err := e.client.Do(httpReq)
	if err != nil {
		errorType := "network"
		if strings.Contains(err.Error(), "timeout") || strings.Contains(err.Error(), "deadline") {
			errorType = "timeout"
		}
		return &ExecutionResult{
			Error: &ExecutionError{
				Message: err.Error(),
				Type:    errorType,
			},
		}, nil
	}
	defer resp.Body.Close()

	// Read response body with size limit
	bodyReader := io.LimitReader(resp.Body, e.config.MaxBodySize)
	bodyBytes, err := io.ReadAll(bodyReader)
	if err != nil {
		return &ExecutionResult{
			Error: &ExecutionError{
				Message: fmt.Sprintf("failed to read response body: %v", err),
				Type:    "network",
			},
		}, nil
	}

	// Check if body was truncated
	body := string(bodyBytes)
	if int64(len(bodyBytes)) >= e.config.MaxBodySize {
		body += "\n\n[Response body truncated - exceeds maximum size]"
	}

	// Collect headers
	headers := make(map[string]string)
	for key, values := range resp.Header {
		headers[key] = strings.Join(values, ", ")
	}

	elapsed := time.Since(startTime)

	return &ExecutionResult{
		Status:     resp.StatusCode,
		StatusText: resp.Status,
		Headers:    headers,
		Body:       body,
		TimeMs:     elapsed.Milliseconds(),
		SizeBytes:  int64(len(bodyBytes)),
	}, nil
}

func (e *HTTPExecutor) validateRequest(req *Request) error {
	if req.Method == "" {
		return fmt.Errorf("method is required")
	}
	if req.URL == "" {
		return fmt.Errorf("URL is required")
	}
	return nil
}

func (e *HTTPExecutor) buildHTTPRequest(ctx context.Context, req *Request) (*http.Request, error) {
	var bodyReader io.Reader

	// Build request body based on body type
	switch req.BodyType {
	case "none":
		bodyReader = nil
	case "json":
		if req.Body != "" {
			bodyReader = strings.NewReader(req.Body)
		}
	case "form-data":
		if len(req.FormData) > 0 {
			formData := url.Values{}
			for key, value := range req.FormData {
				formData.Add(key, value)
			}
			bodyReader = strings.NewReader(formData.Encode())
		}
	case "url-encoded":
		if len(req.FormData) > 0 {
			formData := url.Values{}
			for key, value := range req.FormData {
				formData.Add(key, value)
			}
			bodyReader = strings.NewReader(formData.Encode())
		}
	default:
		return nil, fmt.Errorf("unsupported body type: %s", req.BodyType)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, req.Method, req.URL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Generate auth headers if request item and workspace are provided
	authHeaders := make(map[string]string)
	if req.RequestID != "" {
		// Auth headers will be generated in ExecuteWithAuth
		// For now, we'll handle it in buildHTTPRequest if we have the context
	}

	// Set auth headers first (they take precedence)
	for key, value := range authHeaders {
		httpReq.Header.Set(key, value)
	}

	// Set manual headers (auth headers already set won't be overwritten)
	for key, value := range req.Headers {
		// Only set if not already set by auth
		if httpReq.Header.Get(key) == "" {
			httpReq.Header.Set(key, value)
		}
	}

	// Set Content-Type if not provided and body exists
	if bodyReader != nil {
		if httpReq.Header.Get("Content-Type") == "" {
			switch req.BodyType {
			case "json":
				httpReq.Header.Set("Content-Type", "application/json")
			case "form-data":
				// Note: For true multipart/form-data, we'd need multipart.Writer
				// For now, we'll use application/x-www-form-urlencoded
				httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			case "url-encoded":
				httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			}
		}
	}

	return httpReq, nil
}

// ExecuteWithAuth executes an HTTP request with authentication support
func (e *HTTPExecutor) ExecuteWithAuth(ctx context.Context, req *Request, requestItem *domain.WorkspaceItem, workspace *domain.Workspace) (*ExecutionResult, error) {
	startTime := time.Now()

	// Validate request
	if err := e.validateRequest(req); err != nil {
		return &ExecutionResult{
			Error: &ExecutionError{
				Message: err.Error(),
				Type:    "invalid_request",
			},
		}, nil
	}

	// Resolve auth if request item is provided
	var authHeaders map[string]string
	if requestItem != nil && workspace != nil {
		resolvedAuth, err := e.authResolver.ResolveAuth(requestItem, workspace)
		if err != nil {
			// Log auth resolution error but continue without auth
			fmt.Printf("Warning: Failed to resolve auth for request %s: %v\n", req.RequestID, err)
		} else if resolvedAuth != nil && resolvedAuth.Type != "none" {
			generatedHeaders, err := e.authResolver.GenerateHeaders(resolvedAuth, req.Method, req.URL)
			if err != nil {
				// Log header generation error but continue without auth headers
				fmt.Printf("Warning: Failed to generate auth headers for request %s (type: %s): %v\n", req.RequestID, resolvedAuth.Type, err)
			} else {
				authHeaders = generatedHeaders
				if len(authHeaders) > 0 {
					fmt.Printf("Debug: Generated %d auth headers for request %s (type: %s)\n", len(authHeaders), req.RequestID, resolvedAuth.Type)
				}
			}
		} else {
			fmt.Printf("Debug: No auth config for request %s (resolvedAuth: %v)\n", req.RequestID, resolvedAuth)
		}
	} else {
		if req.RequestID != "" {
			fmt.Printf("Debug: Request %s has no requestItem or workspace for auth resolution\n", req.RequestID)
		}
	}

	// Merge auth headers with manual headers (auth takes precedence)
	// Use canonical header names (Go's http package canonicalizes headers)
	mergedHeaders := make(map[string]string)

	// Add auth headers first
	for key, value := range authHeaders {
		// Canonicalize header key (first letter and letters after hyphens are uppercase)
		canonicalKey := http.CanonicalHeaderKey(key)
		mergedHeaders[canonicalKey] = value
	}

	// Add manual headers, skipping any that conflict with auth headers (case-insensitive check)
	for key, value := range req.Headers {
		canonicalKey := http.CanonicalHeaderKey(key)
		// Only add if not already set by auth (case-insensitive check)
		if _, exists := mergedHeaders[canonicalKey]; !exists {
			mergedHeaders[canonicalKey] = value
		}
	}

	// Create request with merged headers
	reqWithAuth := &Request{
		Method:    req.Method,
		URL:       req.URL,
		Headers:   mergedHeaders,
		BodyType:  req.BodyType,
		Body:      req.Body,
		FormData:  req.FormData,
		RequestID: req.RequestID,
	}

	// Create HTTP request
	httpReq, err := e.buildHTTPRequestWithHeaders(ctx, reqWithAuth, mergedHeaders)
	if err != nil {
		return &ExecutionResult{
			Error: &ExecutionError{
				Message: err.Error(),
				Type:    "invalid_request",
			},
		}, nil
	}

	// Execute request
	resp, err := e.client.Do(httpReq)
	if err != nil {
		errorType := "network"
		if strings.Contains(err.Error(), "timeout") || strings.Contains(err.Error(), "deadline") {
			errorType = "timeout"
		}
		return &ExecutionResult{
			Error: &ExecutionError{
				Message: err.Error(),
				Type:    errorType,
			},
		}, nil
	}
	defer resp.Body.Close()

	// Read response body with size limit
	bodyReader := io.LimitReader(resp.Body, e.config.MaxBodySize)
	bodyBytes, err := io.ReadAll(bodyReader)
	if err != nil {
		return &ExecutionResult{
			Error: &ExecutionError{
				Message: fmt.Sprintf("failed to read response body: %v", err),
				Type:    "network",
			},
		}, nil
	}

	// Check if body was truncated
	body := string(bodyBytes)
	if int64(len(bodyBytes)) >= e.config.MaxBodySize {
		body += "\n\n[Response body truncated - exceeds maximum size]"
	}

	// Collect headers
	headers := make(map[string]string)
	for key, values := range resp.Header {
		headers[key] = strings.Join(values, ", ")
	}

	elapsed := time.Since(startTime)

	return &ExecutionResult{
		Status:     resp.StatusCode,
		StatusText: resp.Status,
		Headers:    headers,
		Body:       body,
		TimeMs:     elapsed.Milliseconds(),
		SizeBytes:  int64(len(bodyBytes)),
	}, nil
}

// buildHTTPRequestWithHeaders builds HTTP request with pre-merged headers
func (e *HTTPExecutor) buildHTTPRequestWithHeaders(ctx context.Context, req *Request, headers map[string]string) (*http.Request, error) {
	var bodyReader io.Reader

	// Build request body based on body type
	switch req.BodyType {
	case "none":
		bodyReader = nil
	case "json":
		if req.Body != "" {
			bodyReader = strings.NewReader(req.Body)
		}
	case "form-data":
		if len(req.FormData) > 0 {
			formData := url.Values{}
			for key, value := range req.FormData {
				formData.Add(key, value)
			}
			bodyReader = strings.NewReader(formData.Encode())
		}
	case "url-encoded":
		if len(req.FormData) > 0 {
			formData := url.Values{}
			for key, value := range req.FormData {
				formData.Add(key, value)
			}
			bodyReader = strings.NewReader(formData.Encode())
		}
	default:
		return nil, fmt.Errorf("unsupported body type: %s", req.BodyType)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, req.Method, req.URL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers (http.Header.Set will canonicalize the keys automatically)
	// Set all headers, including empty values (some headers might intentionally be empty)
	for key, value := range headers {
		httpReq.Header.Set(key, value)
	}

	// Set Content-Type if not provided and body exists
	if bodyReader != nil {
		if httpReq.Header.Get("Content-Type") == "" {
			switch req.BodyType {
			case "json":
				httpReq.Header.Set("Content-Type", "application/json")
			case "form-data":
				// Note: For true multipart/form-data, we'd need multipart.Writer
				// For now, we'll use application/x-www-form-urlencoded
				httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			case "url-encoded":
				httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			}
		}
	}

	return httpReq, nil
}
