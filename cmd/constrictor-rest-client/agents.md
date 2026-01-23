# cmd/constrictor-rest-client

## Purpose

Main entrypoint for the Constrictor REST Client HTTP server. Initializes dependencies and starts the HTTP server.

## Key Files

- `main.go` - Server entrypoint, route setup, dependency injection
- `main_test.go` - Tests for health endpoint

## How to Run

```bash
# Run server
go run cmd/constrictor-rest-client/main.go

# Run with custom port
PORT=9000 go run cmd/constrictor-rest-client/main.go

# Run tests
go test ./cmd/constrictor-rest-client/...
```

## Dependencies

- `internal/config` - Configuration loading
- `internal/storage` - Workspace persistence
- `internal/executor` - HTTP request execution
- `internal/httpapi` - HTTP handlers and routing

## Routes

- `GET /api/health` - Health check
- `GET /api/workspace` - Load workspace
- `PUT /api/workspace` - Save workspace
- `POST /api/execute` - Execute HTTP request
- `/*` - Serve static files from `web/dist/`
