#!/bin/bash

# Navigate to the project root directory
cd "$(dirname "$0")"

echo "Starting Docker containers..."
docker compose up -d

echo "Starting Merchant Server (Port 8080)..."
cd merchant-server
go run cmd/server/main.go > ../merchant-server.log 2>&1 &
cd ..

echo "Starting AI Agent (Port 8001)..."
cd ai-agent
source venv/bin/activate 2>/dev/null || true
python3 -m uvicorn main:app --port 8001 > ../ai-agent.log 2>&1 &
cd ..

echo "Starting Dashboard (Port 3000)..."
cd dashboard
npm run dev > ../dashboard.log 2>&1 &
cd ..

echo "All services have been started in the background."
echo "Logs are being written to: merchant-server.log, ai-agent.log, dashboard.log"
echo "Run ./stop.sh to close everything."
