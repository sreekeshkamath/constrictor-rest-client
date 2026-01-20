package executor

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestHTTPExecutor_Execute_GET(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "success"}`))
	}))
	defer server.Close()

	executor := NewHTTPExecutor(Config{
		Timeout:     5 * time.Second,
		MaxBodySize: 10 * 1024 * 1024,
	})

	req := &Request{
		Method:   "GET",
		URL:      server.URL,
		BodyType: "none",
	}

	result, err := executor.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.Error != nil {
		t.Fatalf("Unexpected execution error: %v", result.Error)
	}

	if result.Status != http.StatusOK {
		t.Errorf("Expected status 200, got %d", result.Status)
	}

	if !strings.Contains(result.Body, "success") {
		t.Errorf("Expected body to contain 'success', got: %s", result.Body)
	}

	if result.TimeMs <= 0 {
		t.Errorf("Expected positive time, got %d", result.TimeMs)
	}

	if result.SizeBytes <= 0 {
		t.Errorf("Expected positive size, got %d", result.SizeBytes)
	}
}

func TestHTTPExecutor_Execute_POST_JSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}

		body := make([]byte, 100)
		n, _ := r.Body.Read(body)
		if !strings.Contains(string(body[:n]), "test") {
			t.Errorf("Expected body to contain 'test'")
		}

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id": 123}`))
	}))
	defer server.Close()

	executor := NewHTTPExecutor(Config{
		Timeout:     5 * time.Second,
		MaxBodySize: 10 * 1024 * 1024,
	})

	req := &Request{
		Method:   "POST",
		URL:      server.URL,
		BodyType: "json",
		Body:     `{"name": "test"}`,
		Headers:  map[string]string{},
	}

	result, err := executor.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.Error != nil {
		t.Fatalf("Unexpected execution error: %v", result.Error)
	}

	if result.Status != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", result.Status)
	}
}

func TestHTTPExecutor_Execute_FormData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Errorf("Failed to parse form: %v", err)
		}
		if r.FormValue("key1") != "value1" {
			t.Errorf("Expected form value key1=value1, got %s", r.FormValue("key1"))
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer server.Close()

	executor := NewHTTPExecutor(Config{
		Timeout:     5 * time.Second,
		MaxBodySize: 10 * 1024 * 1024,
	})

	req := &Request{
		Method:   "POST",
		URL:      server.URL,
		BodyType: "form-data",
		FormData: map[string]string{
			"key1": "value1",
			"key2": "value2",
		},
		Headers: map[string]string{},
	}

	result, err := executor.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.Error != nil {
		t.Fatalf("Unexpected execution error: %v", result.Error)
	}

	if result.Status != http.StatusOK {
		t.Errorf("Expected status 200, got %d", result.Status)
	}
}

func TestHTTPExecutor_Execute_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	executor := NewHTTPExecutor(Config{
		Timeout:     100 * time.Millisecond,
		MaxBodySize: 10 * 1024 * 1024,
	})

	req := &Request{
		Method:   "GET",
		URL:      server.URL,
		BodyType: "none",
	}

	result, err := executor.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.Error == nil {
		t.Fatal("Expected timeout error, got none")
	}

	if result.Error.Type != "timeout" {
		t.Errorf("Expected error type 'timeout', got %s", result.Error.Type)
	}
}

func TestHTTPExecutor_Execute_BodySizeLimit(t *testing.T) {
	largeBody := strings.Repeat("x", 2000)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(largeBody))
	}))
	defer server.Close()

	executor := NewHTTPExecutor(Config{
		Timeout:     5 * time.Second,
		MaxBodySize: 1000, // 1KB limit
	})

	req := &Request{
		Method:   "GET",
		URL:      server.URL,
		BodyType: "none",
	}

	result, err := executor.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.Error != nil {
		t.Fatalf("Unexpected execution error: %v", result.Error)
	}

	// Body should be truncated
	if !strings.Contains(result.Body, "truncated") {
		t.Errorf("Expected body to be truncated, got: %s", result.Body[:100])
	}

	if result.SizeBytes > 1000 {
		t.Errorf("Expected size <= 1000, got %d", result.SizeBytes)
	}
}

func TestHTTPExecutor_Execute_InvalidURL(t *testing.T) {
	executor := NewHTTPExecutor(Config{
		Timeout:     5 * time.Second,
		MaxBodySize: 10 * 1024 * 1024,
	})

	req := &Request{
		Method:   "GET",
		URL:      "not-a-valid-url",
		BodyType: "none",
	}

	result, err := executor.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.Error == nil {
		t.Fatal("Expected error for invalid URL, got none")
	}

	if result.Error.Type != "network" {
		t.Errorf("Expected error type 'network', got %s", result.Error.Type)
	}
}

func TestHTTPExecutor_Execute_InvalidRequest(t *testing.T) {
	executor := NewHTTPExecutor(Config{
		Timeout:     5 * time.Second,
		MaxBodySize: 10 * 1024 * 1024,
	})

	// Missing URL
	req := &Request{
		Method:   "GET",
		BodyType: "none",
	}

	result, err := executor.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.Error == nil {
		t.Fatal("Expected validation error, got none")
	}

	if result.Error.Type != "invalid_request" {
		t.Errorf("Expected error type 'invalid_request', got %s", result.Error.Type)
	}
}

func TestHTTPExecutor_Execute_Headers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Custom-Header") != "custom-value" {
			t.Errorf("Expected X-Custom-Header=custom-value, got %s", r.Header.Get("X-Custom-Header"))
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer server.Close()

	executor := NewHTTPExecutor(Config{
		Timeout:     5 * time.Second,
		MaxBodySize: 10 * 1024 * 1024,
	})

	req := &Request{
		Method:   "GET",
		URL:      server.URL,
		BodyType: "none",
		Headers: map[string]string{
			"X-Custom-Header": "custom-value",
		},
	}

	result, err := executor.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.Error != nil {
		t.Fatalf("Unexpected execution error: %v", result.Error)
	}

	if result.Status != http.StatusOK {
		t.Errorf("Expected status 200, got %d", result.Status)
	}
}

func TestHTTPExecutor_Execute_ResponseHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Response-Header", "response-value")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{}"))
	}))
	defer server.Close()

	executor := NewHTTPExecutor(Config{
		Timeout:     5 * time.Second,
		MaxBodySize: 10 * 1024 * 1024,
	})

	req := &Request{
		Method:   "GET",
		URL:      server.URL,
		BodyType: "none",
	}

	result, err := executor.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.Error != nil {
		t.Fatalf("Unexpected execution error: %v", result.Error)
	}

	if result.Headers["X-Response-Header"] != "response-value" {
		t.Errorf("Expected X-Response-Header=response-value, got %s", result.Headers["X-Response-Header"])
	}

	if result.Headers["Content-Type"] != "application/json" {
		t.Errorf("Expected Content-Type=application/json, got %s", result.Headers["Content-Type"])
	}
}
