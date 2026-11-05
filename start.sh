#!/bin/bash

set -e

PROJECT_DIR="/home/bladekkk/MCAP-project"
BACKEND_DIR="$PROJECT_DIR/backend"
FRONTEND_DIR="$PROJECT_DIR/frontend"

echo "╔════════════════════════════════════════╗"
echo "║  🎮 MCAP - Minecraft Admin Panel      ║"
echo "║  Starting all services...              ║"
echo "╚════════════════════════════════════════╝"
echo ""

# Kill any existing processes
echo "🛑 Cleaning up old processes..."
pkill -f "python3 app.py" || true
pkill -f "next dev" || true
sleep 1

# Create data directories
mkdir -p "$PROJECT_DIR/data"
mkdir -p "$PROJECT_DIR/data/cache"
mkdir -p "$PROJECT_DIR/data/jdk"
mkdir -p "$PROJECT_DIR/data/servers"

echo "📁 Directories created/verified"
echo ""

# Start backend
echo "▶️  Starting Backend (Flask + SQLite)..."
cd "$BACKEND_DIR"
source venv/bin/activate

python3 app.py > "$PROJECT_DIR/backend.log" 2>&1 &
BACKEND_PID=$!
echo "   ✅ Backend started (PID: $BACKEND_PID)"

# Wait for backend to be ready
sleep 2
if ! curl -s http://localhost:8080/api/v1/servers > /dev/null 2>&1; then
    echo "   ❌ Backend failed to start!"
    cat "$PROJECT_DIR/backend.log"
    exit 1
fi

echo ""

# Start frontend
echo "▶️  Starting Frontend (Next.js)..."
cd "$FRONTEND_DIR"
npm run dev > "$PROJECT_DIR/frontend.log" 2>&1 &
FRONTEND_PID=$!
echo "   ✅ Frontend started (PID: $FRONTEND_PID)"

# Wait for frontend to be ready
sleep 3
if ! curl -s http://localhost:3000/ > /dev/null 2>&1; then
    echo "   ⚠️  Frontend still loading..."
fi

echo ""
echo "╔════════════════════════════════════════╗"
echo "║  ✅ All services started successfully! ║"
echo "╚════════════════════════════════════════╝"
echo ""
echo "🌐 Dashboard:     http://localhost:3000"
echo "📡 Backend API:   http://localhost:8080/api/v1"
echo ""
echo "📊 Logs:"
echo "   Backend:  $PROJECT_DIR/backend.log"
echo "   Frontend: $PROJECT_DIR/frontend.log"
echo ""
echo "🛑 To stop all services, run: pkill -f 'python3 app.py'; pkill -f 'next dev'"
echo ""

# Keep script running
wait
