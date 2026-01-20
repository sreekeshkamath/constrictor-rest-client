# Wails Desktop Application

This directory contains the Wails desktop application for Constrictor REST Client.

## Status

**Note**: This is a basic structure. Full Wails integration requires:
1. Wails CLI installed: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`
2. Frontend build from `web/` directory
3. Proper Wails configuration

## Structure

- `app.go` - Wails app struct with bindings (GetWorkspace, SaveWorkspace, ExecuteRequest)
- `main.go` - Wails application entrypoint
- `frontend/` - Frontend assets (should be built from `web/` directory)

## Setup Instructions

1. **Install Wails CLI**:
   ```bash
   go install github.com/wailsapp/wails/v2/cmd/wails@latest
   ```

2. **Build Frontend**:
   ```bash
   cd ../web
   npm install
   npm run build
   cp -r dist ../wails/frontend/
   ```

3. **Initialize Wails** (if needed):
   ```bash
   cd ../wails
   wails init
   ```

4. **Development**:
   ```bash
   wails dev
   ```

5. **Build**:
   ```bash
   wails build
   ```

## Integration Notes

- The Wails app reuses the same Go backend services (executor, storage)
- Frontend should be adapted to use Wails bindings instead of HTTP API
- The `frontend/` directory should contain the built React app from `web/`

## Future Work

- [ ] Complete Wails configuration
- [ ] Adapt React frontend to use Wails bindings
- [ ] Test desktop app functionality
- [ ] Add build scripts for cross-platform builds
