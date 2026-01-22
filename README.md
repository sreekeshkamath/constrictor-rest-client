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

```bash
# Build frontend for production
cd web
npm run build

# Output in web/dist/
```

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
