#!/bin/bash

# E2E Test Runner Script
# This script starts the backend with DEV_MODE=true and runs E2E tests

set -e

echo "🎯 Starting E2E Test Environment"

# Function to cleanup on exit
cleanup() {
    echo "🧹 Cleaning up..."
    if [[ -n $BACKEND_PID ]]; then
        kill $BACKEND_PID 2>/dev/null || true
    fi
    if [[ -n $FRONTEND_PID ]]; then
        kill $FRONTEND_PID 2>/dev/null || true  
    fi
}

trap cleanup EXIT INT TERM

# Check if ports are already in use
if lsof -Pi :8080 -sTCP:LISTEN -t >/dev/null ; then
    echo "⚠️  Port 8080 already in use. Please stop the running backend or use a different port."
    exit 1
fi

if lsof -Pi :3000 -sTCP:LISTEN -t >/dev/null ; then
    echo "⚠️  Port 3000 already in use. Please stop the running frontend or use a different port."
    exit 1
fi

# Start backend with DEV_MODE
echo "🚀 Starting backend with DEV_MODE=true on port 8080..."
cd backend
DEV_MODE=true go run ./cmd/server &
BACKEND_PID=$!
cd ..

# Start frontend
echo "🚀 Starting frontend on port 3000..."
cd frontend
npm run dev > /dev/null 2>&1 &
FRONTEND_PID=$!
cd ..

# Wait for services to start
echo "⏳ Waiting for services to start..."
sleep 8

# Check if services are responding
echo "🔍 Checking backend..."
if ! curl -s http://localhost:8080/health > /dev/null; then
    echo "❌ Backend not responding on port 8080"
    exit 1
fi

echo "🔍 Checking frontend..."
if ! curl -s http://localhost:3000 > /dev/null; then
    echo "❌ Frontend not responding on port 3000"  
    exit 1
fi

echo "✅ Services ready!"

# Run E2E tests
echo "🧪 Running Playwright E2E tests..."
cd frontend

if [[ "$1" == "--headed" ]]; then
    npm test -- --headed
elif [[ -n "$1" ]]; then
    # Pass any arguments to playwright (e.g., --grep "Development Cards")
    npm test -- "$@"
else
    npm test
fi

echo "🎉 E2E tests completed!"