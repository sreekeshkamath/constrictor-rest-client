# Constrictor REST Client - Design Document

## Architecture Overview

Constrictor REST Client is a REST API testing tool built with Go backend and React frontend, designed with SOLID principles and clean architecture.

### High-Level Architecture

```
┌─────────────────┐         ┌─────────────────┐
│   React Web UI  │────────▶│   Go HTTP API   │
│   (Vite + TS)   │         │   (gorilla/mux) │
└─────────────────┘         └─────────────────┘
                                      │
                    ┌─────────────────┼─────────────────┐
                    │                 │                 │
            ┌───────▼──────┐  ┌───────▼──────┐  ┌──────▼──────┐
            │   Executor   │  │   Storage    │  │   Config    │
            │  (HTTP reqs) │  │ (Workspace)  │  │  (Settings) │
            └──────────────┘  └──────────────┘  └─────────────┘
```

## Package Structure

### Backend (Go)

```
constrictor-rest-client/
├── cmd/
│   └── constrictor-rest-client/    # Main entrypoint (HTTP server)
├── internal/
│   ├── config/                      # Configuration management
│   ├── domain/                      # Domain models (pure types)
│   ├── executor/                    # HTTP request execution
│   ├── storage/                     # Workspace persistence
│   └── httpapi/                     # HTTP handlers and routing
├── web/                             # React frontend (Vite + TS)
├── wails/                           # Wails desktop app (future)
└── plans/                           # Implementation plans
```

### Frontend (React)

```
web/
├── src/
│   ├── components/                  # React components
│   │   ├── Sidebar.tsx
│   │   ├── RequestEditor.tsx
│   │   ├── ResponseViewer.tsx
│   │   ├── SettingsModal.tsx
│   │   └── MethodBadge.tsx
│   ├── App.tsx                      # Main app component
│   ├── types.ts                     # TypeScript types
│   ├── index.tsx                    # Entry point
│   └── index.css                    # Styles (Tailwind)
├── package.json
├── vite.config.ts
└── tsconfig.json
```

## Domain Model

### Workspace

The workspace is the root entity containing all requests and folders:

```go
type Workspace struct {
    Version int            `json:"version"`
    Items   []WorkspaceItem `json:"items"`
}
```

### WorkspaceItem

Represents either a request or a folder:

```go
type WorkspaceItem struct {
    ID        string  `json:"id"`
    Name      string  `json:"name"`
    Type      string  `json:"type"` // "request" or "folder"
    ParentID  *string `json:"parentId,omitempty"`
    CreatedAt int64   `json:"createdAt"`

    // Request-specific fields
    Method   *string     `json:"method,omitempty"`
    URL      *string     `json:"url,omitempty"`
    Headers  []Header    `json:"headers,omitempty"`
    BodyType *string     `json:"bodyType,omitempty"`
    Body     *string     `json:"body,omitempty"`
    FormData []FormDataItem `json:"formData,omitempty"`
}
```

## API Contract

### Endpoints

#### `GET /api/health`
Health check endpoint.

**Response:**
```json
{
  "ok": true
}
```

#### `GET /api/workspace`
Loads the current workspace.

**Response:**
```json
{
  "version": 1,
  "items": [...]
}
```

#### `PUT /api/workspace`
Saves the workspace.

**Request:**
```json
{
  "version": 1,
  "items": [...]
}
```

**Response:**
```json
{
  "status": "saved"
}
```

#### `POST /api/execute`
Executes an HTTP request.

**Request:**
```json
{
  "method": "GET",
  "url": "https://api.example.com/endpoint",
  "headers": [
    {"key": "Authorization", "value": "Bearer token", "enabled": true}
  ],
  "bodyType": "json",
  "body": "{\"key\": \"value\"}",
  "formData": []
}
```

**Response:**
```json
{
  "status": 200,
  "statusText": "200 OK",
  "headers": {"Content-Type": "application/json"},
  "body": "{\"result\": \"success\"}",
  "timeMs": 150,
  "sizeBytes": 25
}
```

**Error Response:**
```json
{
  "status": 0,
  "error": {
    "message": "Request timeout",
    "type": "timeout"
  }
}
```

## Persistence

### Workspace Storage

- **Location**: `data/workspace.json` (configurable via `CONSTRICTOR_DATA_PATH`)
- **Format**: Versioned JSON
- **Atomic Writes**: Uses temp file + rename pattern
- **Thread Safety**: Mutex-protected read/write operations

### Storage Interface

```go
type WorkspaceStore interface {
    Load() (*domain.Workspace, error)
    Save(workspace *domain.Workspace) error
}
```

## Request Execution

### Executor Interface

```go
type Executor interface {
    Execute(ctx context.Context, req *Request) (*ExecutionResult, error)
}
```

### Features

- **Timeouts**: Configurable request timeout (default: 30s)
- **Body Size Limit**: Maximum response body size (default: 10MB)
- **Body Types**: Supports `none`, `json`, `form-data`, `url-encoded`
- **Automatic Headers**: Content-Type set automatically based on body type
- **Error Handling**: Categorizes errors (timeout, network, invalid_request)

## Configuration

### Environment Variables

- `PORT`: Server port (default: 8080)
- `CONSTRICTOR_DATA_PATH`: Data directory path (default: `data`)
- `MAX_BODY_SIZE`: Maximum response body size in bytes (default: 10MB)
- `TIMEOUT`: Request timeout in seconds (default: 30)

### Config Structure

```go
type Config struct {
    Port        string
    DataPath    string
    MaxBodySize int64
    Timeout     int
}
```

## Frontend Architecture

### Component Hierarchy

```
App
├── Sidebar
│   ├── Search
│   ├── ItemList (folders/requests)
│   └── Actions (export/import/settings)
├── RequestEditor
│   ├── URL Bar (method + URL + send button)
│   ├── Headers Tab
│   └── Body Tab (json/form-data/url-encoded)
└── ResponseViewer
    ├── Status/Time/Size
    ├── Body Tab (JSON tree or raw text)
    └── Headers Tab
```

### State Management

- **Workspace State**: Managed in App component, synced to backend
- **Active Request**: Selected request ID tracked in App state
- **Response State**: Execution result stored in App state
- **Settings**: Local state in App component (not persisted to backend yet)

### API Integration

- **Workspace Loading**: `GET /api/workspace` on mount
- **Workspace Saving**: `PUT /api/workspace` with debounce (500ms)
- **Request Execution**: `POST /api/execute` on send button click

## Wails Integration (Future)

The Wails desktop app will:

1. Reuse the same React UI from `web/`
2. Use Wails bindings instead of HTTP API
3. Share the same Go backend services (executor, storage)
4. Provide native desktop experience

### Wails Structure

```
wails/
├── app.go              # Wails app entrypoint
├── bindings.go         # Wails bindings (expose Go services)
└── frontend/           # Shared React UI (symlink or copy from web/)
```

## Testing Strategy

### Backend Tests

- **Unit Tests**: Each package has comprehensive unit tests
- **Integration Tests**: HTTP API handlers tested with mock dependencies
- **Test Coverage**: Aim for >80% coverage

### Frontend Tests

- **Component Tests**: Test individual components (future)
- **Integration Tests**: Test API integration (future)

## Error Handling

### Backend Error Model

```go
type APIError struct {
    Error   string `json:"error"`
    Message string `json:"message,omitempty"`
}
```

### Execution Errors

```go
type ExecutionError struct {
    Message string `json:"message"`
    Type    string `json:"type"` // "timeout", "network", "invalid_request", etc.
}
```

## Security Considerations

1. **CORS**: Not applicable for desktop app, but should be configured for web deployment
2. **Input Validation**: All API inputs validated before processing
3. **Body Size Limits**: Prevent memory exhaustion from large responses
4. **Timeout Limits**: Prevent hanging requests
5. **File System**: Atomic writes prevent corruption

## Deployment

### Web Deployment

1. Build React app: `cd web && npm run build`
2. Serve static files from `web/dist/`
3. Run Go server: `go run cmd/constrictor-rest-client/main.go`

### Desktop Deployment (Wails)

1. Build Wails app: `cd wails && wails build`
2. Distribute native binaries

## Future Enhancements

1. **Authentication**: User authentication for multi-user scenarios
2. **Cloud Sync**: Google Drive integration (UI already prepared)
3. **Request History**: Save execution history
4. **Environment Variables**: Support for environment-specific configs
5. **Collections**: Organize requests into collections
6. **Scripts**: Pre/post request scripts
7. **Tests**: Assertions and test suites
