# internal/httpapi

## Purpose

HTTP API handlers and routing. Provides HTTP endpoints for workspace management and request execution.

## Key Files

- `handlers.go` - HTTP handlers (GetWorkspace, PutWorkspace, Execute)
- `router.go` - Route setup function
- `handlers_test.go` - Handler tests with mocks

## Endpoints

- `GET /api/workspace` - Load workspace
- `PUT /api/workspace` - Save workspace
- `POST /api/execute` - Execute HTTP request

## Dependencies

- `internal/storage` - Workspace persistence
- `internal/executor` - Request execution
- `internal/domain` - Domain types

## How to Test

```bash
go test ./internal/httpapi/... -v
```

## Usage

Handlers are created with dependencies and registered via SetupRoutes:

```go
handlers := httpapi.NewHandlers(store, executor)
httpapi.SetupRoutes(router, handlers)
```

## Error Handling

All errors return JSON with error type and message:
```json
{
  "error": "validation_failed",
  "message": "URL is required"
}
```
