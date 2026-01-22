# Google Drive Backup Service

## Purpose

This package provides Google Drive backup functionality for Constrictor REST Client workspaces. It handles OAuth authentication, file upload/download, and most importantly, **automatic token sanitization** to ensure no sensitive credentials are backed up to Google Drive.

## Key Files

- `service.go` - Google Drive service implementation with OAuth and file operations
  - `Service` interface - Defines backup/restore operations
  - `GoogleDriveService` - Implementation using Google Drive API v3
  - Automatic workspace sanitization before upload

## Security Features

**CRITICAL**: This service automatically sanitizes workspaces before uploading to Google Drive:

- **Header Sanitization**: Removes sensitive headers like:
  - `Authorization`, `X-API-Key`, `X-Auth-Token`, `Cookie`, etc.
  - Case-insensitive matching (Authorization, authorization, etc.)

- **Form Data Sanitization**: Removes sensitive form fields like:
  - `password`, `token`, `api_key`, `access_token`, etc.

- **OAuth Security**:
  - OAuth handled in frontend (browser)
  - Access tokens passed to backend only for API calls
  - Tokens never stored on server

## How It Works

1. **Authentication**: Frontend handles OAuth flow, gets access token
2. **Backup**: Frontend calls `/api/gdrive/backup` with access token
3. **Sanitization**: Backend automatically sanitizes workspace (removes tokens)
4. **Upload**: Sanitized workspace uploaded to Google Drive
5. **Restore**: Frontend calls `/api/gdrive/restore` to download and restore

## API Endpoints

- `POST /api/gdrive/backup` - Backup workspace (requires accessToken)
- `GET /api/gdrive/backups` - List all backups (requires accessToken)
- `POST /api/gdrive/restore` - Restore workspace (requires accessToken, fileId)

## Dependencies

- `google.golang.org/api/drive/v3` - Google Drive API client
- `golang.org/x/oauth2` - OAuth2 token handling
- `internal/domain` - Workspace types and sanitization utilities

## Testing

```bash
# Run domain sanitization tests
go test ./internal/domain/... -v

# Test backup flow (requires valid OAuth token):
# 1. Authenticate in frontend
# 2. Call POST /api/gdrive/backup
# 3. Verify file in Google Drive (check no Authorization headers)
```

## Usage Example

```go
service := gdrive.NewService()

// Backup workspace (automatically sanitized)
fileID, err := service.BackupWorkspace(ctx, accessToken, workspace, "backup.json")

// List backups
backups, err := service.ListBackups(ctx, accessToken)

// Restore workspace
workspace, err := service.RestoreWorkspace(ctx, accessToken, fileID)
```

## Important Notes

- **Never disable sanitization** - This is a critical security feature
- Access tokens expire - users need to re-authenticate periodically
- OAuth Client ID must be configured in Google Cloud Console
- Requires "drive.file" scope for Google Drive API access
