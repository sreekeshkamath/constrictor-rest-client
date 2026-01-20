# internal/executor

## Purpose

HTTP request execution. Provides an interface for executing HTTP requests with timeouts, body size limits, and error handling.

## Key Files

- `executor.go` - Executor interface and types (ExecutionResult, ExecutionError, Request)
- `http_executor.go` - HTTPExecutor implementation using net/http
- `http_executor_test.go` - Comprehensive tests

## Interface

```go
type Executor interface {
    Execute(ctx context.Context, req *Request) (*ExecutionResult, error)
}
```

## Features

- Configurable timeout
- Maximum response body size limit
- Support for all HTTP methods
- Body types: none, json, form-data, url-encoded
- Automatic Content-Type header setting
- Comprehensive error handling (timeout, network, invalid_request)

## How to Test

```bash
go test ./internal/executor/... -v
```

## Usage Example

```go
exec := executor.NewHTTPExecutor(executor.Config{
    Timeout: 30 * time.Second,
    MaxBodySize: 10 * 1024 * 1024,
})

result, err := exec.Execute(ctx, &executor.Request{
    Method: "GET",
    URL: "https://api.example.com/endpoint",
    BodyType: "none",
})
```
