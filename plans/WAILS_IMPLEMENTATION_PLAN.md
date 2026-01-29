# Wails Native Desktop App - Detailed Implementation Plan

## Overview

Convert Constrictor REST Client from a web server application to a native desktop application using Wails v2. The app will run without a server, using direct Go-to-JavaScript bindings for communication.

## Current State

- **Backend**: Go HTTP server with REST API (`cmd/constrictor-rest-client/main.go`)
- **Frontend**: React app using HTTP fetch calls (`web/src/App.tsx`)
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

### Type Conversion Requirements

The frontend and backend use different data structures that need conversion:

1. **Workspace Format**:
   - Frontend: `SidebarItem[]` (array of items)
   - Backend: `domain.Workspace` (with `Items: []WorkspaceItem`)

2. **Headers/FormData**:
   - Frontend: Arrays `Header[]` and `FormDataItem[]` with `enabled` flags
   - Executor: Maps `map[string]string` (only enabled items)

3. **ExecuteRequest**:
   - Frontend sends: Arrays for headers and formData
   - Executor expects: Maps for headers and formData

## Implementation Steps

### Step 1: Create Type Conversion Layer in Wails App

**File**: `wails/app.go`

**Objective**: Add helper functions to convert between frontend types and backend types.

**Tasks**:
1. Create `SidebarItem` type in Go that matches frontend TypeScript type
2. Add `convertWorkspaceToItems(workspace *domain.Workspace) []SidebarItem` function
3. Add `convertItemsToWorkspace(items []SidebarItem) *domain.Workspace` function
4. Add `convertHeadersArrayToMap(headers []domain.Header) map[string]string` function
5. Add `convertFormDataArrayToMap(formData []domain.FormDataItem) map[string]string` function

**Key Implementation Details**:
- `SidebarItem` should be a struct with JSON tags matching TypeScript interface
- Only include enabled headers/formData items in maps
- Handle nil/empty values gracefully
- Preserve all fields during conversion

**Testing**:
- Create unit tests in `wails/app_test.go` for each conversion function
- Test with various scenarios: empty arrays, disabled items, mixed enabled/disabled
- Verify round-trip conversion (items → workspace → items)

**Manual Verification**:
- Run `go test ./wails/... -v` to verify all tests pass

---

### Step 2: Update GetWorkspace and SaveWorkspace Methods

**File**: `wails/app.go`

**Objective**: Update methods to accept/return frontend-friendly types.

**Tasks**:
1. Change `GetWorkspace()` return type from `*domain.Workspace` to `[]SidebarItem`
2. Implement conversion: Load workspace → convert to items → return items
3. Change `SaveWorkspace()` parameter from `*domain.Workspace` to `[]SidebarItem`
4. Implement conversion: Receive items → convert to workspace → save workspace
5. Add proper error handling and validation

**Key Implementation Details**:
- Use conversion functions from Step 1
- Handle empty workspace (return empty array)
- Validate items before saving
- Return meaningful error messages

**Testing**:
- Add tests for GetWorkspace with empty and populated workspaces
- Add tests for SaveWorkspace with valid and invalid items
- Test error handling (file permissions, invalid data)

**Manual Verification**:
- Run `go test ./wails/... -v`
- Test with `wails dev` (if Wails CLI is installed) to verify methods are callable

---

### Step 3: Update ExecuteRequest Method

**File**: `wails/app.go`

**Objective**: Update method to accept frontend array types and convert to executor map types.

**Tasks**:
1. Create `ExecuteRequestInput` struct matching frontend request format:
   - `Method string`
   - `URL string`
   - `Headers []domain.Header` (array)
   - `BodyType string`
   - `Body string`
   - `FormData []domain.FormDataItem` (array)
2. Change `ExecuteRequest()` signature to accept `*ExecuteRequestInput`
3. Convert arrays to maps using helper functions
4. Create `executor.Request` with converted maps
5. Call executor and return result
6. Handle errors appropriately

**Key Implementation Details**:
- Only include enabled headers/formData in conversion
- Validate method, URL, and bodyType
- Preserve all executor functionality (timeouts, body size limits)

**Testing**:
- Add tests for ExecuteRequest with various scenarios:
  - GET request with headers
  - POST request with JSON body
  - POST request with form-data
  - POST request with url-encoded
  - Request with disabled headers/formData
- Test error cases (invalid URL, timeout, etc.)

**Manual Verification**:
- Run `go test ./wails/... -v`
- Verify executor integration works correctly

---

### Step 4: Create Wails TypeScript Bindings

**File**: `web/src/wails.ts` (NEW FILE)

**Objective**: Create TypeScript definitions and wrapper functions for Wails runtime bindings.

**Tasks**:
1. Define `App` interface matching Go struct methods:
   ```typescript
   interface App {
     GetWorkspace(): Promise<SidebarItem[]>;
     SaveWorkspace(items: SidebarItem[]): Promise<void>;
     ExecuteRequest(req: ExecuteRequestInput): Promise<ExecutionResult>;
   }
   ```
2. Create helper functions that use Wails runtime:
   - `GetWorkspace(): Promise<SidebarItem[]>`
   - `SaveWorkspace(items: SidebarItem[]): Promise<void>`
   - `ExecuteRequest(req: ExecuteRequestInput): Promise<ExecutionResult>`
3. Add runtime detection: `isWailsRuntime(): boolean`
4. Export types: `ExecuteRequestInput`, `ExecutionResult`

**Key Implementation Details**:
- Use `window.go?.main.App.methodName()` for Wails runtime calls
- Handle runtime not available (return null/throw error)
- Match TypeScript types exactly with Go structs
- Add JSDoc comments for IDE support

**Testing**:
- Create unit tests in `web/src/wails.test.ts` (if using Vitest)
- Test runtime detection
- Mock Wails runtime for testing

**Manual Verification**:
- Check TypeScript compilation: `cd web && npm run build`
- Verify no type errors
- Check that functions are exported correctly

---

### Step 5: Update Frontend App.tsx to Use Wails Bindings

**File**: `web/src/App.tsx`

**Objective**: Replace all HTTP fetch calls with Wails bindings.

**Tasks**:
1. Remove `API_BASE = '/api'` constant
2. Import Wails bindings: `import { GetWorkspace, SaveWorkspace, ExecuteRequest } from './wails'`
3. Update `loadWorkspace` function:
   - Replace `fetch('/api/workspace')` with `GetWorkspace()`
   - Handle Wails runtime errors
   - Keep same data normalization logic
4. Update `saveWorkspace` function:
   - Replace `fetch('/api/workspace', { method: 'PUT' })` with `SaveWorkspace(items)`
   - Keep debounce logic
   - Handle errors appropriately
5. Update `handleSendRequest` function:
   - Replace `fetch('/api/execute')` with `ExecuteRequest()`
   - Convert request format to match `ExecuteRequestInput`
   - Handle response format (should match current format)
6. Remove all HTTP-related error handling that's no longer needed
7. Add error handling for Wails runtime not available (show user-friendly message)

**Key Implementation Details**:
- Maintain exact same UI behavior
- Keep all existing state management
- Preserve error handling patterns
- Ensure response format matches current expectations

**Testing**:
- Manual testing: Run `wails dev` and verify:
  - Workspace loads on startup
  - Workspace saves when items change
  - HTTP requests execute correctly
  - All UI features work as before
- Create integration tests if possible

**Manual Verification**:
- Build frontend: `cd web && npm run build`
- Check for TypeScript errors
- Test in Wails dev mode: `wails dev`
- Verify all three main operations: load, save, execute

---

### Step 6: Update Wails Configuration

**File**: `wails/wails.json`

**Objective**: Verify and update Wails configuration for proper build setup.

**Tasks**:
1. Verify `frontend.dir` points to `web/dist` (currently correct: `"dir": "web/dist"`)
2. Add/verify build configuration:
   - Ensure `install`, `build`, `dev` commands are empty (Wails handles this)
3. Verify app metadata:
   - Product name: "Constrictor REST Client"
   - Version: "1.0.0"
   - Author information
4. Add platform-specific build settings if needed

**Key Implementation Details**:
- Frontend directory must point to built React app
- Wails will embed frontend from `web/dist` directory
- Configuration should support cross-platform builds

**Testing**:
- Verify configuration is valid JSON
- Check that paths are correct

**Manual Verification**:
- Run `wails build` (if Wails CLI installed) to verify configuration
- Check for any configuration errors

---

### Step 7: Create Frontend Build Integration

**Objective**: Ensure frontend is built and available for Wails before building native app.

**Tasks**:
1. Verify `web/vite.config.ts` outputs to `web/dist/` (should already be configured)
2. Create Makefile target `wails-build-frontend`:
   - Build frontend: `cd web && npm run build`
   - Verify `web/dist` exists and contains built files
3. Update existing `build` target to also build frontend if needed

**Key Implementation Details**:
- Frontend must be built before Wails can embed it
- Build output should be in `web/dist/`
- Wails reads from `web/dist/` as configured in `wails.json`

**Testing**:
- Run `make wails-build-frontend` (or equivalent)
- Verify `web/dist` contains:
  - `index.html`
  - JavaScript bundles
  - CSS files
  - Assets

**Manual Verification**:
- Run: `cd web && npm run build`
- Check: `ls -la web/dist/` shows built files
- Verify no build errors

---

### Step 8: Update Makefile with Wails Targets

**File**: `Makefile`

**Objective**: Add convenient Makefile targets for Wails development and building.

**Tasks**:
1. Add `wails-dev` target:
   - Build frontend first: `cd web && npm run build`
   - Run `wails dev` for development with hot reload
2. Add `wails-build` target:
   - Build frontend: `cd web && npm run build`
   - Run `wails build` for current platform
3. Add `wails-build-all` target:
   - Build frontend: `cd web && npm run build`
   - Run `wails build` with flags for all platforms (Windows, Linux, macOS)
4. Add platform-specific targets (optional):
   - `wails-build-windows`: Build Windows executable
   - `wails-build-linux`: Build Linux executable
   - `wails-build-darwin`: Build macOS app
5. Add `wails-clean` target to clean Wails build artifacts

**Key Implementation Details**:
- Always build frontend before Wails operations
- Use proper Wails CLI commands
- Handle errors gracefully
- Add helpful echo messages

**Testing**:
- Test each Makefile target:
  - `make wails-build-frontend` (if created separately)
  - `make wails-dev` (requires Wails CLI)
  - `make wails-build` (requires Wails CLI)

**Manual Verification**:
- Run: `make help` to see new targets
- Test: `make wails-build-frontend` (should build frontend)
- Test: `make wails-dev` (if Wails CLI installed)

---

### Step 9: Update Documentation

**File**: `README.md`

**Objective**: Add native app build and development instructions.

**Tasks**:
1. Add "Native Desktop App" section to README
2. Document Wails CLI installation:
   ```bash
   go install github.com/wailsapp/wails/v2/cmd/wails@latest
   ```
3. Document development workflow:
   - `make wails-dev` or `wails dev`
   - Hot reload capabilities
4. Document build process:
   - `make wails-build` for current platform
   - `make wails-build-all` for all platforms
   - Platform-specific requirements (MSVC for Windows, Xcode for macOS, etc.)
5. Add troubleshooting section:
   - Common Wails issues
   - Frontend build requirements
   - Platform-specific notes

**Key Implementation Details**:
- Keep existing web server instructions
- Clearly separate web server vs native app sections
- Include prerequisites for each platform
- Add links to Wails documentation

**Testing**:
- Review documentation for clarity
- Verify all commands work as documented

**Manual Verification**:
- Read through README.md
- Verify instructions are clear and complete
- Test documented commands

---

### Step 10: Final Integration Testing

**Objective**: Comprehensive testing of the complete native desktop app.

**Tasks**:
1. Test workspace operations:
   - Load workspace on startup
   - Create new request
   - Create new folder
   - Rename items
   - Delete items
   - Move items (drag and drop)
   - Save workspace (verify persistence)
2. Test HTTP request execution:
   - GET request
   - POST request with JSON body
   - POST request with form-data
   - POST request with url-encoded
   - Request with custom headers
   - Request with disabled headers/formData
   - Error cases (invalid URL, timeout, etc.)
3. Test workspace persistence:
   - Close and reopen app
   - Verify workspace is restored
   - Verify all items are preserved
4. Test edge cases:
   - Empty workspace
   - Large workspace
   - Special characters in names/URLs
   - Very long response bodies

**Key Implementation Details**:
- Test on target platform (or multiple platforms if possible)
- Verify no HTTP server dependencies remain
- Ensure all UI features work identically to web version

**Testing**:
- Manual testing checklist:
  - [ ] App launches successfully
  - [ ] Workspace loads on startup
  - [ ] Can create/edit/delete requests
  - [ ] Can create/edit/delete folders
  - [ ] HTTP requests execute correctly
  - [ ] Responses display correctly
  - [ ] Workspace persists after app restart
  - [ ] No console errors
  - [ ] No network requests to localhost:8080

**Manual Verification**:
- Run: `wails dev` and test all features
- Run: `wails build` and test built executable
- Verify: No HTTP server needed
- Verify: All functionality works as expected

---

## File Changes Summary

### New Files
- `web/src/wails.ts` - Wails TypeScript bindings
- `wails/app_test.go` - Unit tests for Wails app (optional but recommended)

### Modified Files
- `wails/app.go` - Add type conversion methods, update method signatures
- `web/src/App.tsx` - Replace HTTP fetch with Wails bindings
- `wails/wails.json` - Verify/update configuration (likely minimal changes)
- `Makefile` - Add Wails build targets
- `README.md` - Add native app instructions

### Build Output
- `web/dist/` - Built React app (from `npm run build`)
- Platform-specific executables in `wails/build/bin/` (Wails default)

## Key Technical Decisions

1. **Type Conversion**: Handle frontend array types vs backend map types in Go layer (in `wails/app.go`)
2. **Workspace Format**: Convert between `SidebarItem[]` and `Workspace` structure in Wails app
3. **Error Handling**: Ensure Wails methods return proper error types that JavaScript can handle
4. **Runtime Detection**: Frontend will only use Wails mode (no fallback to HTTP)

## Dependencies

- **Wails CLI**: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`
- **Platform-specific build tools**:
  - Windows: MSVC or MinGW
  - Linux: Standard build tools (gcc, etc.)
  - macOS: Xcode Command Line Tools

## Testing Strategy

1. **Unit Tests**: Test conversion functions and Wails app methods in isolation
2. **Integration Tests**: Test Wails bindings with frontend (manual or automated)
3. **End-to-End Tests**: Test complete app functionality in Wails dev mode
4. **Build Tests**: Verify native app builds successfully on target platforms
5. **Manual Testing**: Comprehensive manual testing checklist (see Step 10)

## Notes

- The Wails app reuses existing Go services (executor, storage) - no changes needed there
- Frontend UI remains the same, only communication layer changes
- Web server mode remains available as alternative build target
- All existing functionality must be preserved
