# web - React Frontend

## Purpose

React web application providing the user interface for Constrictor REST Client. Connects to Go backend via HTTP API.

## Key Files

- `src/App.tsx` - Main app component, state management, API integration
- `src/components/` - React components (Sidebar, RequestEditor, ResponseViewer, etc.)
- `src/types.ts` - TypeScript type definitions
- `package.json` - Dependencies and scripts
- `vite.config.ts` - Vite configuration with API proxy

## How to Run

```bash
# Install dependencies
npm install

# Development server
npm run dev

# Build for production
npm run build

# Preview production build
npm run preview
```

## Development Setup

1. Start Go backend: `go run cmd/constrictor-rest-client/main.go`
2. Start React dev server: `npm run dev`
3. Open <http://localhost:5173>

Vite proxy automatically forwards `/api/*` requests to <http://localhost:8080>.

## Architecture

- **State Management**: React hooks (useState, useEffect)
- **Styling**: Tailwind CSS
- **API Integration**: Fetch API with `/api/*` endpoints
- **Workspace Persistence**: Auto-saves to backend with 500ms debounce

## Components

- `Sidebar` - Request/folder list, search, export/import
- `RequestEditor` - Method, URL, headers, body editor
- `ResponseViewer` - Status, headers, body viewer with JSON tree
- `SettingsModal` - Settings dialog (Google Drive sync, appearance)
- `MethodBadge` - HTTP method badge with color coding

## API Integration

- `GET /api/workspace` - Load workspace on mount
- `PUT /api/workspace` - Save workspace on changes (debounced)
- `POST /api/execute` - Execute request on send button click
