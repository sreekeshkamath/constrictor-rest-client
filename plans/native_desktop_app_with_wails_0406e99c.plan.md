---
name: Native Desktop App with Wails
overview: Convert the Constrictor REST Client from a web server application to a native desktop application using Wails, enabling cross-platform builds for Windows, Linux, and macOS without requiring a server.
todos:
  - id: "1"
    content: "Update wails/app.go: Add type conversion methods and fix ExecuteRequest signature to handle frontend array types"
    status: pending
  - id: "2"
    content: "Create web/src/wails.ts: TypeScript bindings for Wails runtime methods"
    status: pending
  - id: "3"
    content: "Update web/src/App.tsx: Replace HTTP fetch calls with Wails bindings"
    status: pending
  - id: "4"
    content: "Update Makefile: Add Wails development and build targets (dev, build, build-all)"
    status: pending
  - id: "5"
    content: "Create build integration: Script/Makefile target to build frontend and copy to wails/frontend/dist"
    status: pending
  - id: "6"
    content: "Update wails/wails.json: Verify and update configuration for cross-platform builds"
    status: pending
  - id: "7"
    content: "Update README.md: Add native app build and development instructions"
    status: pending
isProject: false
---

# Native Desktop App Implementation Plan

## Overview

Convert the application from a web server model to a native desktop app using Wails v2. The app will run without a server, using direct Go-to-JavaScript bindings for communication.

## Current State

- **Backend**: Go HTTP server with REST API (`cmd/constrictor-rest-client/main.go`)
- **Frontend**: React app using HTTP fetch calls (`/api/workspace`, `/api/execute`)
- **Wails Setup**: Basic structure exists in `wails/` but incomplete
- **Dependencies**: Wails v2.11.0 already in `go.mod`

## Architecture Changes

### Data Flow Transformation

```
Current (Web Server):
Frontend (React) → HTTP Fetch → Go HTTP Server → Services → Storage

New (Native Desktop):
Frontend (React) → Wails Runtime → Go App Methods → Services → Storage
```

## Implementation Steps

### 1. Update Wails App Backend (`wails/app.go`)

**Issues to Fix:**

- `ExecuteRequest` method signature mismatch: frontend sends arrays, executor expects maps
- Need to convert between `SidebarItem[]` (frontend) and `Workspace` (backend)
- Need to handle header/formData array-to-map conversion

**Changes:**

- Add conversion methods for headers/formData arrays → maps
- Add methods to convert between frontend `SidebarItem[]` and backend `Workspace`
- Update `ExecuteRequest` to accept frontend-friendly types and convert internally
- Add proper error handling

### 2. Create Wails TypeScript Bindings (`web/src/wails.ts`)

**New File**: Create TypeScript definitions for Wails runtime bindings

- Define `App` interface matching Go struct methods
- Export typed functions: `GetWorkspace()`, `SaveWorkspace()`, `ExecuteRequest()`
- Use Wails runtime: `window.go?.main.App.methodName()`

### 3. Update Frontend to Use Wails Bindings (`web/src/App.tsx`)

**Replace HTTP calls with Wails bindings:**

- Remove `API_BASE = '/api'` constant
- Replace `fetch('/api/workspace')` → `GetWorkspace()` from Wails
- Replace `fetch('/api/workspace', { method: 'PUT' })` → `SaveWorkspace()` from Wails
- Replace `fetch('/api/execute')` → `ExecuteRequest()` from Wails
- Add detection for Wails runtime vs web mode (optional: support both)

### 4. Update Wails Configuration (`wails/wails.json`)

**Verify/Update:**

- Ensure `frontend.dir` points to `web/dist` (currently correct)
- Add build configuration for cross-platform targets
- Set proper app metadata (name, version, etc.)

### 5. Create Build Scripts

**Update `Makefile` with new targets:**

- `wails-dev`: Run Wails in development mode
- `wails-build`: Build for current platform
- `wails-build-all`: Build for Windows, Linux, macOS
- `wails-build-windows`: Build Windows executable
- `wails-build-linux`: Build Linux executable
- `wails-build-darwin`: Build macOS app

**Create build script** (`build-native.sh` or similar):

- Build frontend (`cd web && npm run build`)
- Copy dist to `wails/frontend/dist`
- Run `wails build` with appropriate flags

### 6. Frontend Build Integration

**Update `web/vite.config.ts`:**

- Ensure build output goes to `dist/` (already configured)
- Add base path configuration if needed for Wails

**Create post-build script** or Makefile target:

- After `npm run build` in `web/`, copy `web/dist` → `wails/frontend/dist`

### 7. Type Conversion Layer

**In `wails/app.go`, add helper functions:**

- `convertWorkspaceToItems(workspace *domain.Workspace) []SidebarItem`
- `convertItemsToWorkspace(items []SidebarItem) *domain.Workspace`
- `convertHeadersArrayToMap(headers []domain.Header) map[string]string`
- `convertFormDataArrayToMap(formData []domain.FormDataItem) map[string]string`

**Note**: May need to create a shared types package or duplicate types in Wails package for `SidebarItem` representation.

### 8. Development Workflow

**Update documentation:**

- Add instructions for Wails CLI installation
- Document development workflow: `wails dev`
- Document build process for each platform

## File Changes Summary

### New Files

- `web/src/wails.ts` - Wails TypeScript bindings
- `build-native.sh` - Cross-platform build script (optional, can use Makefile)

### Modified Files

- `wails/app.go` - Add type conversion methods, fix method signatures
- `web/src/App.tsx` - Replace HTTP fetch with Wails bindings
- `wails/wails.json` - Verify/update configuration
- `Makefile` - Add Wails build targets
- `README.md` - Update with native app instructions

### Build Output

- `wails/frontend/dist/` - Built React app (from `web/dist`)
- Platform-specific executables in `wails/build/bin/` (Wails default)

## Key Technical Decisions

1. **Type Conversion**: Handle frontend array types vs backend map types in Go layer
2. **Workspace Format**: Convert between `SidebarItem[]` and `Workspace` structure
3. **Error Handling**: Ensure Wails methods return proper error types
4. **Development Mode**: Support both `wails dev` (hot reload) and web server mode for flexibility

## Testing Strategy

1. Test Wails bindings in development mode (`wails dev`)
2. Verify workspace save/load functionality
3. Test HTTP request execution
4. Build and test on each target platform
5. Verify no HTTP server dependencies remain

## Dependencies

- Wails CLI: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`
- Platform-specific build tools (see Wails documentation)
  - Windows: MSVC or MinGW
  - Linux: Standard build tools
  - macOS: Xcode Command Line Tools

## Notes

- The Wails app reuses existing Go services (executor, storage) - no changes needed there
- Frontend UI remains the same, only communication layer changes
- Can maintain web server mode as alternative build target if desired
