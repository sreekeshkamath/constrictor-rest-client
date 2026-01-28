# Constrictor REST Client

A modern REST API testing tool with a Go backend and React frontend. Features workspace management, request organization, and Google Drive backup with automatic token sanitization.

## 🚀 Quick Start

### Prerequisites
- **Go** 1.21+ ([install](https://go.dev/doc/install))
- **Node.js** and npm ([install](https://nodejs.org/))
- **curl** (for health checks, usually pre-installed)

### Run Everything with One Command

Choose one of these options:

**Option 1: Using npm (Recommended)**
```bash
# Install root dependencies (first time only)
npm install

# Run both backend and frontend
npm run dev
```

**Option 2: Using Make**
```bash
# Install all dependencies (first time only)
make install

# Run both servers
make dev
```

**Option 3: Using shell script**
```bash
# Make executable (first time only)
chmod +x run.sh

# Run both servers
./run.sh
```

Then open your browser to **http://localhost:5173** 🎉

## 📖 Detailed Setup

### Install Dependencies

**Backend (Go):**
```bash
go mod tidy
```

**Frontend (React):**
```bash
cd web
npm install
cd ..
```

**Root (for npm scripts):**
```bash
npm install
```

### Run Individual Servers

**Backend only:**
```bash
go run cmd/constrictor-rest-client/main.go
# Runs on http://localhost:8080
```

**Frontend only:**
```bash
cd web
npm run dev
# Runs on http://localhost:5173
```

## 🧪 Testing

```bash
# Run all Go tests
go test ./...

# Run specific package tests
go test ./internal/storage/... -v
go test ./internal/domain/... -v
```

## 🏗️ Build

### Web Server Build

```bash
# Build frontend for production
cd web
npm run build

# Output in web/dist/
```

## 🖥️ Native Desktop App

Constrictor REST Client can also be built as a native desktop application using Wails v2. The native app runs without a server, using direct Go-to-JavaScript bindings for communication.

### Prerequisites

**Wails CLI:**
```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

**Platform-specific build tools:**
- **Windows**: MSVC or MinGW
- **Linux**: Standard build tools (gcc, etc.)
- **macOS**: Xcode Command Line Tools

### Development

**Run in development mode:**
```bash
make wails-dev
# or
cd wails && wails dev
```

This will:
1. Build the frontend automatically (one time)
2. Start Wails in development mode
3. **Note:** Frontend changes require manual rebuild (see below)

**Important:** `wails dev` does NOT automatically watch and rebuild frontend changes. When you modify files in `web/src/`, you need to rebuild:

```bash
# Rebuild frontend after making changes
make wails-rebuild-frontend
# or manually:
cd web && npm run build && cd .. && rm -rf wails/frontend && cp -r web/dist wails/frontend
```

Then refresh the Wails app window to see your changes.

**For automatic rebuilding (optional):**
Run the watch script in a separate terminal:
```bash
./wails-dev-watch.sh
```
This will automatically rebuild the frontend when you save changes to `web/src/`.

### Building

**Build for current platform:**
```bash
make wails-build
# or
cd wails && wails build
```

**Build for all platforms:**
```bash
make wails-build-all
```

**Build for specific platform:**
```bash
make wails-build-windows  # Windows
make wails-build-linux    # Linux
make wails-build-darwin   # macOS
```

**Install Linux app (build + install to ~/.local):**
```bash
make wails-install-linux
# or use the helper script:
./make-linux-app.sh
```

This will:
1. Build the Linux binary
2. Install it to `~/.local/bin/constrictor-rest-client`
3. Create a desktop entry in `~/.local/share/applications/`
4. Update the desktop database so it appears in your application menu

After installation, you can:
- Find "Constrictor REST Client" in your application menu
- Run it from terminal: `constrictor-rest-client`

**Create DMG for macOS distribution:**
```bash
make wails-build-dmg
# Creates constrictor-rest-client.dmg in the project root
```

**Build frontend only (for Wails):**
```bash
make wails-build-frontend
```

### Available Wails Commands

- `make wails-build-frontend` - Build frontend for Wails
- `make wails-dev` - Run Wails in development mode
- `make wails-build` - Build for current platform
- `make wails-build-all` - Build for all platforms
- `make wails-build-windows` - Build for Windows
- `make wails-build-linux` - Build for Linux
- `make wails-build-darwin` - Build for macOS
- `make wails-clean` - Clean Wails build artifacts

### Troubleshooting

**Wails CLI not found:**
```bash
# Install Wails CLI
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# Verify installation
wails version
```

**Frontend build errors:**
```bash
# Ensure frontend is built before Wails operations
make wails-build-frontend

# Check that web/dist/ exists
ls -la web/dist/
```

**Platform-specific build issues:**
- **Windows**: Ensure MSVC or MinGW is installed and in PATH
- **macOS**: Install Xcode Command Line Tools: `xcode-select --install`
- **Linux**: Install standard build tools: `sudo apt-get install build-essential` (Ubuntu/Debian)

**Build fails with "frontend not found":**
- Ensure `web/dist/` directory exists (run `make wails-build-frontend` first)
- Verify `wails.json` has `"dir": "web/dist"` in frontend section

**App opens and closes immediately (macOS Gatekeeper):**
- The app is unsigned, so macOS may block it
- **Solution 1 (Recommended)**: Right-click the app → Open → Click "Open" in the dialog
- **Solution 2**: Remove quarantine: `xattr -d com.apple.quarantine wails/build/bin/constrictor-rest-client.app`
- **Solution 3**: System Preferences → Security & Privacy → Allow the app
- **Solution 4**: Run from terminal: `open wails/build/bin/constrictor-rest-client.app`

## 📁 Project Structure

```
constrictor-rest-client/
├── cmd/constrictor-rest-client/  # Backend entrypoint
├── internal/                      # Go packages
│   ├── config/                    # Configuration
│   ├── domain/                    # Domain models + sanitization
│   ├── executor/                  # HTTP request executor
│   ├── gdrive/                    # Google Drive backup service
│   ├── httpapi/                   # HTTP API handlers
│   └── storage/                   # Workspace persistence
├── web/                           # React frontend
├── wails/                         # Desktop app (Wails)
└── data/                          # Workspace storage (gitignored)
```

## 🐛 Troubleshooting

### Backend won't start

**Check if port 8080 is in use:**
```bash
lsof -i :8080
# Kill the process if needed, or change port in config
```

**Check backend logs:**
```bash
# If using run.sh, logs are in /tmp/constrictor-backend.log
tail -f /tmp/constrictor-backend.log
```

**Verify backend compiles:**
```bash
go build ./cmd/constrictor-rest-client/main.go
```

**Install/update dependencies:**
```bash
go mod tidy
```

### Frontend can't connect to backend

**Check if backend is running:**
```bash
curl http://localhost:8080/api/health
# Should return: {"ok":true}
```

**Or use the check script:**
```bash
./check-backend.sh
```

**Manual start (to see errors):**
```bash
# Terminal 1
go run cmd/constrictor-rest-client/main.go

# Terminal 2
cd web && npm run dev
```

## 🔧 Available Commands

### npm scripts (from root)
- `npm run dev` - Run both backend and frontend
- `npm run dev:backend` - Run only backend
- `npm run dev:frontend` - Run only frontend
- `npm run install:all` - Install all dependencies
- `npm run test` - Run Go tests
- `npm run build` - Build frontend

### Make commands
- `make dev` - Run both servers
- `make dev-backend` - Run only backend
- `make dev-frontend` - Run only frontend
- `make install` - Install all dependencies
- `make test` - Run Go tests
- `make build` - Build frontend
- `make clean` - Clean build artifacts
- `make help` - Show all commands

### Wails commands (Native Desktop App)
- `make wails-build-frontend` - Build frontend for Wails
- `make wails-dev` - Run Wails in development mode
- `make wails-build` - Build for current platform
- `make wails-build-all` - Build for all platforms
- `make wails-build-windows` - Build for Windows
- `make wails-build-linux` - Build for Linux
- `make wails-build-darwin` - Build for macOS
- `make wails-build-dmg` - Create DMG file for macOS distribution
- `make wails-clean` - Clean Wails build artifacts

## 🔐 Google Drive Backup

The app includes secure Google Drive backup with automatic token sanitization:

1. **Configure** in Settings → Google Drive Sync
2. **Authenticate** with your Google account
3. **Automatic sync** happens when you save changes
4. **Visual indicator** shows sync status in the sidebar

**Security:** All sensitive headers (Authorization, X-API-Key, etc.) and form data (passwords, tokens) are automatically removed before backup.

## 📚 Documentation

- `DESIGN.md` - Architecture and design documentation
- `plans/IMPLEMENTATION_PLAN.md` - Implementation plan
- `plans/PROGRESS.md` - Progress log
- `agents.md` files in each folder - Developer guides

## 🛠️ Tech Stack

**Backend:**
- Go 1.21+
- gorilla/mux for routing
- Google Drive API for backups

**Frontend:**
- React 19
- TypeScript
- Vite
- Tailwind CSS

**Desktop:**
- Wails v2

## 📝 License

See [LICENSE](LICENSE) file.
