#!/bin/bash

# Navigate to the project root directory
cd "$(dirname "$0")"

echo "Stopping services by port..."

# Stop Dashboard (3000)
DASHBOARD_PID=$(lsof -t -i:3000)
if [ ! -z "$DASHBOARD_PID" ]; then
    echo "Stopping Dashboard (PID: $DASHBOARD_PID)..."
    kill -9 $DASHBOARD_PID 2>/dev/null
else
    echo "Dashboard is not running on port 3000."
fi

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
