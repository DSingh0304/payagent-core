#!/bin/bash

# Navigate to the project root directory
cd "$(dirname "$0")"

echo "Stopping services by port..."

# Stop Dashboard (3000)
pkill -f "next-server" || true
pkill -f "next dev" || true
echo "Dashboard processes killed."

# Stop AI Agent (8001)
AI_AGENT_PID=$(lsof -t -i:8001)
if [ ! -z "$AI_AGENT_PID" ]; then
    echo "Stopping AI Agent (PID: $AI_AGENT_PID)..."
    kill -9 $AI_AGENT_PID 2>/dev/null
else
    echo "AI Agent is not running on port 8001."
fi

# Stop Merchant Server (8080)
MERCHANT_PID=$(lsof -t -i:8080)
if [ ! -z "$MERCHANT_PID" ]; then
    echo "Stopping Merchant Server (PID: $MERCHANT_PID)..."
    kill -9 $MERCHANT_PID 2>/dev/null
else
    echo "Merchant Server is not running on port 8080."
fi

echo "Stopping Docker containers..."
docker compose down

echo "All services stopped."
