#!/bin/bash

# Check if backend is running
if curl --fail --max-time 10 --connect-timeout 5 -s http://localhost:8080/api/health > /dev/null 2>&1; then
    echo "✅ Backend is running"
    exit 0
else
    echo "❌ Backend is not running on http://localhost:8080"
    echo ""
    echo "To start the backend:"
    echo "  go run cmd/constrictor-rest-client/main.go"
    echo ""
    echo "Or check if port 8080 is already in use:"
    echo "  lsof -i :8080"
    exit 1
fi
