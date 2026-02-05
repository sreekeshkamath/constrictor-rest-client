## Feature: Fix Import Issues and Enhance Response Persistence

### What This PR Does
This PR addresses critical issues with the Insomnia import feature and includes comprehensive fixes for import functionality, response persistence, and desktop app integration. It implements full support for importing Insomnia YAML backup files, fixes multiple bugs related to authentication, encoding, and memory management, and adds response caching with JSON editor enhancements.

### Dependencies

- None

### Key Features

#### 1. Insomnia YAML Import Implementation
- Added complete Insomnia backup file (YAML) import functionality
- Created YAML parser and converter (`internal/importer/insomnia/`) with 91% test coverage
- Supports Insomnia schema version `collection.insomnia.rest/5.0`
- Converts Insomnia folders, requests, authentication, headers, parameters, and body types
- Added POST `/api/import/insomnia` endpoint with multipart file upload support
- 10MB file size limit with proper error handling
- Support for replace and merge modes via query parameter
- Frontend file type detection (JSON vs YAML) with automatic routing
- Desktop app integration via Wails binding (`ImportInsomnia()` method)

#### 2. Response Persistence and JSON Editor
- Implemented response caching to persist responses when switching between requests
- Cache keyed by request ID with automatic cleanup on item deletion
- Added JSON Editor component with syntax highlighting using Prism.js
- JSON prettify functionality with invalid JSON error indicators
- Dark theme styling consistent with app design

#### 3. Critical Bug Fixes
- Fixed double-encoding in query parameters (removed `url.QueryEscape` to let `Encode()` handle properly)
- Fixed OOM risk in file uploads (added `io.LimitReader` with 10MB max, returns 413 error if exceeded)
- Fixed disabled auth handling (check `InsomniaAuth.Disabled` flag, return "none" auth when disabled)
- Fixed Wails binding for Insomnia import (desktop app now uses Wails binding instead of HTTP fetch)
- Fixed stale closure issues in response cache cleanup
- Fixed response cache cleanup for all descendants when deleting items
- Fixed missing `react-simple-code-editor` dependency causing build errors
- Fixed auth inheritance and cycle detection in ResolveAuth
- Fixed RequestEditor to track and clean up stale auth headers
- Fixed AuthConfigModal to disable query option and read from local state

### Testing
- [x] Tested with valid data
- [x] Tested validation errors
- [x] Tested edge cases
- [x] Comprehensive unit tests for Insomnia parser and converter (18 tests, 91% coverage)
- [x] Tested parameter handling, auth conversion, nested folders, and edge cases
- [x] Tested file size limits and error handling
- [x] Tested both web and desktop app import flows

### Environment Variables Required
```
None required
```

### Next Steps
- [ ] Consider adding support for Insomnia environments with variables
- [ ] Consider adding cookie jar support for cookie persistence
- [ ] Consider adding Postman collection format support
- [ ] Consider adding export to Insomnia format functionality

### Notes
- All changes are backward compatible - existing workspaces continue to work without changes
- Insomnia imports can be merged or replace existing workspace via `?mode=merge` query parameter
- Desktop app requires Wails runtime for Insomnia import (falls back gracefully in web mode)
- Response cache automatically cleans up when items are deleted, including all descendants
- JSON Editor uses Prism.js for syntax highlighting with custom dark theme styling
