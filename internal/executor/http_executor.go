package executor

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// HTTPExecutor implements Executor using net/http
type HTTPExecutor struct {
	client  *http.Client
	config  Config
}

// NewHTTPExecutor creates a new HTTP executor
func NewHTTPExecutor(config Config) *HTTPExecutor {
	return &HTTPExecutor{
		client: &http.Client{
			Timeout: config.Timeout,
		},
		config: config,
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

	// Set headers
	for key, value := range req.Headers {
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
