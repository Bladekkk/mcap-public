#!/bin/bash

echo "╔════════════════════════════════════════╗"
echo "║  🛑 Stopping MCAP services...          ║"
echo "╚════════════════════════════════════════╝"
echo ""

echo "Stopping Backend..."
pkill -f "python3 app.py" || echo "  ℹ️  Backend already stopped"

echo "Stopping Frontend..."
pkill -f "next dev" || echo "  ℹ️  Frontend already stopped"

sleep 1

echo ""
echo "✅ All services stopped"
echo ""
