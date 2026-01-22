# Constrictor REST Client - Repository Overview

## Purpose

Constrictor REST Client is a REST API testing tool with a Go backend and React frontend. It provides a clean, modern interface for making HTTP requests, organizing requests into folders, and viewing responses.

## Key Entrypoints

- **Web Server**: `cmd/constrictor-rest-client/main.go` - HTTP API server entrypoint
- **Frontend**: `web/` - React application (Vite + TypeScript)
- **Desktop App**: `wails/` - Wails desktop application (future)

## How to Run

### Quick Start (Recommended)

**Option 1: Using npm (requires root package.json setup)**
```bash
# Install root dependencies first
npm install

# Run both backend and frontend
npm run dev
```

**Option 2: Using Make**
```bash
# Install all dependencies first
make install

# Run both servers
make dev
```

**Option 3: Using shell script**
```bash
# Make script executable (first time only)
chmod +x run.sh

# Run both servers
./run.sh
```

### Individual Servers

**Backend (Go)**
```bash
# Run tests
go test ./...

# Run server
go run cmd/constrictor-rest-client/main.go

# Server runs on http://localhost:8080 by default
```

**Frontend (React)**
```bash
cd web
npm install
npm run dev

# Frontend runs on http://localhost:5173 (Vite default)
# API requests are proxied to http://localhost:8080
```

### Full Stack (Manual)

1. Start Go server: `go run cmd/constrictor-rest-client/main.go`
2. Start React dev server: `cd web && npm run dev`
3. Open browser to <http://localhost:5173>

## Architecture

- **Backend**: SOLID principles, clean interfaces, layered architecture
- **Frontend**: React with TypeScript, Tailwind CSS
- **Persistence**: JSON file-based workspace storage
- **API**: RESTful HTTP API with JSON responses

## Key Folders

- `cmd/` - Application entrypoints
- `internal/` - Internal packages (config, domain, executor, storage, httpapi)
- `web/` - React frontend
- `wails/` - Desktop app (future)
- `plans/` - Implementation plans and progress

## Testing

Run all tests:
```bash
go test ./...
```

Run tests for specific package:
```bash
go test ./internal/storage/... -v
```
