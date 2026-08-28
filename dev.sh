#!/bin/bash

# Terminate background processes on Ctrl+C (SIGINT/SIGTERM)
cleanup() {
    echo ""
    echo "=========================================="
    echo " Shutting down StayUp services..."
    echo "=========================================="
    kill $(jobs -p) 2>/dev/null
    exit 0
}

trap cleanup SIGINT SIGTERM

echo "=========================================="
echo " Starting StayUp Development Environment"
echo "=========================================="

# Start Go Backend
echo "[Backend] Starting Go server at http://localhost:8080..."
(cd backend && go run .) &
BACKEND_PID=$!

# Start React Frontend
echo "[Frontend] Starting Vite server at http://localhost:5173..."
(cd frontend && npm run dev) &
FRONTEND_PID=$!

echo ""
echo "=========================================="
echo " StayUp Development Stack is Running!"
echo " - Backend API:  http://localhost:8080"
echo " - Frontend UI:  http://localhost:5173"
echo " Press Ctrl+C to stop both servers."
echo "=========================================="
echo ""

# Wait for background jobs
wait
