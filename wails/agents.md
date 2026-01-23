# wails - Desktop Application

## Purpose

Wails desktop application packaging for Constrictor REST Client. Reuses the same Go backend and React frontend.

## Status

**Implemented** - Wails desktop application is available with bindings.

## Planned Structure

```
wails/
├── app.go              # Wails app entrypoint and bindings (GetWorkspace, SaveWorkspace, ExecuteRequest)
├── main.go             # Wails entrypoint
└── frontend/           # React UI (shared with web/ or symlinked)
```

## How to Run (Future)

```bash
# Development
cd wails
wails dev

# Build
wails build
```

## Integration Plan

1. Reuse Go backend services (executor, storage)
2. Create Wails bindings to expose services to frontend
3. Reuse React UI from `web/` (symlink or shared package)
4. Replace HTTP API calls with Wails bindings in frontend

## Benefits

- Native desktop experience
- No need for HTTP server
- Direct access to file system
- Better performance
- Offline-first architecture
