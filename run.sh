#!/bin/bash

# Constrictor REST Client - Run script
# Starts both backend and frontend servers

# Don't exit on error - we handle errors manually
set +e

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
BUILD_OUTPUT=$(go build -o /tmp/constrictor-test ./cmd/constrictor-rest-client/main.go 2>&1)
BUILD_EXIT=$?

if [ $BUILD_EXIT -ne 0 ]; then
    echo -e "${YELLOW}⚠️  Backend has compilation errors. Checking dependencies...${NC}"
    echo -e "${YELLOW}📦 Running go mod tidy...${NC}"
    go mod tidy
    BUILD_OUTPUT=$(go build -o /tmp/constrictor-test ./cmd/constrictor-rest-client/main.go 2>&1)
    BUILD_EXIT=$?
    if [ $BUILD_EXIT -ne 0 ]; then
        echo -e "${YELLOW}❌ Backend compilation failed:${NC}"
        echo "$BUILD_OUTPUT"
        echo -e "\n${YELLOW}💡 Try running: go mod tidy && go mod download${NC}"
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
CLEANUP_EXIT_CODE=0
cleanup() {
    echo -e "\n${YELLOW}🛑 Shutting down servers...${NC}"
    kill $BACKEND_PID 2>/dev/null || true
    kill $FRONTEND_PID 2>/dev/null || true
    rm -f "$LOG_FILE" 2>/dev/null || true
    rm -f /tmp/constrictor-frontend-$$.log 2>/dev/null || true
    exit ${CLEANUP_EXIT_CODE:-0}
}

# Trap Ctrl+C
trap cleanup INT TERM

# Check if curl is available for health checks
HAS_CURL=false
if command -v curl &> /dev/null; then
    HAS_CURL=true
fi

# Start backend
echo -e "${GREEN}🔧 Starting backend server on http://localhost:8080...${NC}"
LOG_FILE="/tmp/constrictor-backend-$$.log"
go run cmd/constrictor-rest-client/main.go > "$LOG_FILE" 2>&1 &
BACKEND_PID=$!

# Wait a moment for backend to start
sleep 2

# Check if backend process is still running
if ! kill -0 $BACKEND_PID 2>/dev/null; then
    echo -e "${YELLOW}❌ Backend process died immediately.${NC}"
    echo -e "${YELLOW}📋 Backend logs:${NC}"
    if [ -f "$LOG_FILE" ]; then
        cat "$LOG_FILE"
    fi
    echo -e "\n${YELLOW}💡 Troubleshooting:${NC}"
    echo -e "${YELLOW}   1. Check if port 8080 is already in use${NC}"
    echo -e "${YELLOW}   2. Try running backend manually: go run cmd/constrictor-rest-client/main.go${NC}"
    exit 1
fi

# Wait for backend to be ready (if curl is available)
if [ "$HAS_CURL" = true ]; then
    echo -e "${YELLOW}⏳ Waiting for backend to start...${NC}"
    BACKEND_READY=false
    for i in {1..30}; do
        if curl -s http://localhost:8080/api/health > /dev/null 2>&1; then
            echo -e "${GREEN}✅ Backend is ready!${NC}\n"
            BACKEND_READY=true
            break
        fi
        # Check if process is still alive
        if ! kill -0 $BACKEND_PID 2>/dev/null; then
            echo -e "${YELLOW}❌ Backend process died.${NC}"
            if [ -f "$LOG_FILE" ]; then
                echo -e "${YELLOW}📋 Backend logs:${NC}"
                cat "$LOG_FILE"
            fi
            exit 1
        fi
        sleep 1
    done

    if [ "$BACKEND_READY" = false ]; then
        echo -e "${YELLOW}⚠️  Backend didn't respond to health check in time.${NC}"
        echo -e "${YELLOW}📋 Checking backend logs...${NC}"
        if [ -f "$LOG_FILE" ]; then
            tail -20 "$LOG_FILE"
        fi
        echo -e "\n${YELLOW}💡 Backend may still be starting. Continuing anyway...${NC}\n"
    fi
else
    echo -e "${YELLOW}⚠️  curl not found, skipping health check.${NC}"
    echo -e "${YELLOW}   Waiting 3 seconds for backend to start...${NC}"
    sleep 3
    if ! kill -0 $BACKEND_PID 2>/dev/null; then
        echo -e "${YELLOW}❌ Backend process died.${NC}"
        if [ -f "$LOG_FILE" ]; then
            echo -e "${YELLOW}📋 Backend logs:${NC}"
            cat "$LOG_FILE"
        fi
        exit 1
    fi
    echo -e "${GREEN}✅ Backend process is running${NC}\n"
fi

# Start frontend
echo -e "${GREEN}🎨 Starting frontend server on http://localhost:5173...${NC}"
cd web
npm run dev > /tmp/constrictor-frontend-$$.log 2>&1 &
FRONTEND_PID=$!
cd ..

# Wait a moment and check if frontend started
sleep 2
if ! kill -0 $FRONTEND_PID 2>/dev/null; then
    echo -e "${YELLOW}❌ Frontend process died immediately.${NC}"
    if [ -f /tmp/constrictor-frontend-$$.log ]; then
        echo -e "${YELLOW}📋 Frontend logs:${NC}"
        cat /tmp/constrictor-frontend-$$.log
    fi
    echo -e "\n${YELLOW}💡 Try: cd web && npm install && npm run dev${NC}"
    cleanup
    exit 1
fi

echo -e "\n${BLUE}✅ Both servers are running!${NC}"
echo -e "${BLUE}   Backend:  http://localhost:8080${NC}"
echo -e "${BLUE}   Frontend: http://localhost:5173${NC}"
echo -e "\n${YELLOW}Press Ctrl+C to stop both servers${NC}\n"

# Function to monitor processes
monitor_processes() {
    while true; do
        if ! kill -0 $BACKEND_PID 2>/dev/null; then
            echo -e "\n${YELLOW}⚠️  Backend process died!${NC}"
            if [ -f "$LOG_FILE" ]; then
                echo -e "${YELLOW}📋 Backend logs:${NC}"
                tail -20 "$LOG_FILE"
            fi
            CLEANUP_EXIT_CODE=1
            cleanup
        fi
        if ! kill -0 $FRONTEND_PID 2>/dev/null; then
            echo -e "\n${YELLOW}⚠️  Frontend process died!${NC}"
            if [ -f /tmp/constrictor-frontend-$$.log ]; then
                echo -e "${YELLOW}📋 Frontend logs:${NC}"
                tail -20 /tmp/constrictor-frontend-$$.log
            fi
            CLEANUP_EXIT_CODE=1
            cleanup
        fi
        sleep 2
    done
}

# Start monitoring in background
monitor_processes &
MONITOR_PID=$!

# Wait for both processes (or until one dies)
while kill -0 $BACKEND_PID 2>/dev/null && kill -0 $FRONTEND_PID 2>/dev/null; do
    sleep 1
done

# Cleanup
kill $MONITOR_PID 2>/dev/null || true
cleanup
