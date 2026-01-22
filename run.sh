#!/bin/bash

# Constrictor REST Client - Run script
# Starts both backend and frontend servers

set -e

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${BLUE}🚀 Starting Constrictor REST Client...${NC}\n"

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo -e "${YELLOW}⚠️  Go is not installed. Please install Go 1.21+ to run the backend.${NC}"
    exit 1
fi

# Check if Node.js is installed
if ! command -v node &> /dev/null; then
    echo -e "${YELLOW}⚠️  Node.js is not installed. Please install Node.js to run the frontend.${NC}"
    exit 1
fi

# Check if backend compiles
echo -e "${YELLOW}🔍 Checking backend compilation...${NC}"
if ! go build -o /tmp/constrictor-test ./cmd/constrictor-rest-client/main.go 2>/dev/null; then
    echo -e "${YELLOW}⚠️  Backend has compilation errors. Checking dependencies...${NC}"
    echo -e "${YELLOW}📦 Running go mod tidy...${NC}"
    go mod tidy
    if ! go build -o /tmp/constrictor-test ./cmd/constrictor-rest-client/main.go 2>&1; then
        echo -e "${YELLOW}❌ Backend still has errors. Please fix compilation issues first.${NC}"
        exit 1
    fi
fi
rm -f /tmp/constrictor-test
echo -e "${GREEN}✅ Backend compiles successfully${NC}\n"

# Install dependencies if needed
if [ ! -d "web/node_modules" ]; then
    echo -e "${YELLOW}📦 Installing frontend dependencies...${NC}"
    cd web && npm install && cd ..
fi

# Function to cleanup on exit
cleanup() {
    echo -e "\n${YELLOW}🛑 Shutting down servers...${NC}"
    kill $BACKEND_PID $FRONTEND_PID 2>/dev/null || true
    exit 0
}

# Trap Ctrl+C
trap cleanup INT TERM

# Start backend
echo -e "${GREEN}🔧 Starting backend server on http://localhost:8080...${NC}"
go run cmd/constrictor-rest-client/main.go > /tmp/constrictor-backend.log 2>&1 &
BACKEND_PID=$!

# Wait for backend to be ready
echo -e "${YELLOW}⏳ Waiting for backend to start...${NC}"
BACKEND_READY=false
for i in {1..30}; do
    if curl -s http://localhost:8080/api/health > /dev/null 2>&1; then
        echo -e "${GREEN}✅ Backend is ready!${NC}\n"
        BACKEND_READY=true
        break
    fi
    sleep 1
done

if [ "$BACKEND_READY" = false ]; then
    echo -e "${YELLOW}⚠️  Backend didn't start in time.${NC}"
    echo -e "${YELLOW}📋 Checking backend logs...${NC}"
    if [ -f /tmp/constrictor-backend.log ]; then
        tail -20 /tmp/constrictor-backend.log
    fi
    echo -e "\n${YELLOW}💡 Troubleshooting:${NC}"
    echo -e "${YELLOW}   1. Check if port 8080 is already in use: lsof -i :8080${NC}"
    echo -e "${YELLOW}   2. Check backend logs: tail -f /tmp/constrictor-backend.log${NC}"
    echo -e "${YELLOW}   3. Try running backend manually: go run cmd/constrictor-rest-client/main.go${NC}"
    cleanup
    exit 1
fi

# Start frontend
echo -e "${GREEN}🎨 Starting frontend server on http://localhost:5173...${NC}"
cd web && npm run dev &
FRONTEND_PID=$!
cd ..

echo -e "\n${BLUE}✅ Both servers are running!${NC}"
echo -e "${BLUE}   Backend:  http://localhost:8080${NC}"
echo -e "${BLUE}   Frontend: http://localhost:5173${NC}"
echo -e "\n${YELLOW}Press Ctrl+C to stop both servers${NC}\n"

# Wait for both processes
wait
