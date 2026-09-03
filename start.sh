#!/bin/bash

# Navigate to the project root directory
cd "$(dirname "$0")"

echo "Starting Docker containers..."
docker compose up -d

echo "Starting Merchant Server (Port 8080)..."
cd merchant-server
go run cmd/server/main.go > ../logs/merchant-server.log 2>&1 &
cd ..

echo "Starting AI Agent (Port 8001)..."
cd ai-agent
source venv/bin/activate 2>/dev/null || true
python3 -m uvicorn main:app --port 8001 > ../logs/ai-agent.log 2>&1 &
cd ..

echo "Starting Dashboard (Port 3000)..."
cd dashboard
npm run dev > ../logs/dashboard.log 2>&1 &
cd ..

echo "All services have been started in the background."
echo "Logs are being written to: logs/merchant-server.log, logs/ai-agent.log, logs/dashboard.log"
echo "Run ./stop.sh to close everything."
