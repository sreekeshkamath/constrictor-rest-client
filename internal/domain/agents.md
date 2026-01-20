# internal/domain

## Purpose

Pure domain models - data structures without business logic. These types represent the core entities of the application.

## Key Files

- `workspace.go` - Workspace, WorkspaceItem, Header, FormDataItem types

## Types

- `Workspace` - Root entity containing all requests and folders
- `WorkspaceItem` - Either a request or a folder
- `Header` - HTTP header with key, value, and enabled flag
- `FormDataItem` - Form data field with key, value, and enabled flag

## Usage

These types are used across the application:
- Storage layer serializes/deserializes Workspace
- HTTP API converts to/from JSON
- Frontend uses equivalent TypeScript types

## Testing

No tests needed - these are pure data structures. Tests are in packages that use these types.
