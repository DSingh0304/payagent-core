# PayAgent - Autonomous AI Shopping Assistant

> An AI agent that shops for you. Search → Cart → Pay - fully autonomous with human-in-the-loop safety.

![Go](https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white)
![Python](https://img.shields.io/badge/Python-14354C?style=for-the-badge&logo=python&logoColor=white)
![Next.js](https://img.shields.io/badge/next.js-000000?style=for-the-badge&logo=nextdotjs&logoColor=white)
![LangGraph](https://img.shields.io/badge/LangGraph-FF4F00?style=for-the-badge)
![Razorpay](https://img.shields.io/badge/Razorpay-02042B?style=for-the-badge&logo=razorpay&logoColor=white)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-316192?style=for-the-badge&logo=postgresql&logoColor=white)
![Redis](https://img.shields.io/badge/Redis-DC382D?style=for-the-badge&logo=redis&logoColor=white)

## What It Does
PayAgent is an autonomous shopping assistant that accepts natural language goals (e.g., "Buy running shoes under ₹2000"). It uses LangGraph to independently query a product catalog, build a shopping cart, and present a Razorpay checkout session to the user, strictly adhering to predefined spending guardrails.

## Architecture
```mermaid
graph TD
    User([User]) -->|Natural Language| Dashboard[Next.js Dashboard]
    Dashboard -->|POST /agent/run| Agent[Python LangGraph Agent]
    Agent <-->|Tools via HTTP| Merchant[Go Merchant Server]
    Agent <-->|MCP| Razorpay[Razorpay MCP Server]
    Merchant <--> DB[(PostgreSQL)]
    Merchant <--> Cache[(Redis)]
    Agent -->|SSE Audit Stream| Redis[Redis Pub/Sub]
    Redis -->|SSE| Dashboard
```

## Features
- **Autonomous AI shopping agent** (LangGraph + Groq Llama 3)
- **Spending guardrails** with configurable limits to auto-reject expensive carts
- **Human-in-the-loop** payment approval before final checkout
- **Real-time audit trail** via SSE (Server-Sent Events) and Redis Pub/Sub
- **Razorpay payment integration** via official MCP (Model Context Protocol) Server
- **Voice input support** using Web Speech API
- **Analytics dashboard** for session and order tracking
- **Multi-turn conversations** to tweak cart contents interactively
- **Error recovery** with exponential backoff and retry logic
- **Real-time token cost tracking**

## Quick Start

### The Easy Way
You can run the entire stack using the provided start script.
```bash
./start.sh
```
This will start Docker, the Go server, the Python AI agent, and the Next.js dashboard in the background.
The logs are stored in the `logs/` directory.

To stop everything when you're done, run:
```bash
./stop.sh
```

Navigate to [http://localhost:3000](http://localhost:3000)

### Manual Setup (Alternative)
If you prefer running services manually:

**1. Start Database & Redis**
```bash
docker compose up -d
```

**2. Start Go Merchant Server**
```bash
cd merchant-server
cp .env.example .env
go run cmd/server/main.go
```

**3. Start Python AI Agent**
```bash
cd ai-agent
python3 -m venv venv
source venv/bin/activate
pip install -r requirements.txt
cp .env.example .env # Add GROQ_API_KEY
python3 -m uvicorn main:app --port 8001
```

**4. Start Next.js Dashboard**
```bash
cd dashboard
npm install
npm run dev
```
## Demo Scenarios
Try these queries on the dashboard:
1. `Buy running shoes under ₹2500`
2. `Find wireless earbuds under ₹1500`
3. `I need a laptop bag for college`
4. `Get me a book on system design`
5. `Buy a smartwatch under ₹2000`

## Tech Stack
| Component | Technology | Description |
|-----------|------------|-------------|
| **Frontend** | Next.js 14, React, Recharts | Sleek dark-mode dashboard with SSE, voice input, and charts. |
| **Backend API** | Go (Gin) | Blazing fast merchant API managing catalog, cart, and orders. |
- **Frontend**: Next.js 14 Dashboard
- **Backend (Go)**: Streaming Audit Logs (SSE), Cart State (Redis), Database (Postgres)
- **AI Agent (Python)**: LangGraph state machine, Groq LLM, Razorpay Node MCP

## Running Locally
To run locally, simply use the background scripts:
- `./start.sh` to start all infrastructure and servers
- `./stop.sh` to gracefully stop everything

## Project Structure
```
├── ai-agent/          # Python LangGraph logic
├── dashboard/         # Next.js UI
├── db-init/           # SQL seeds
├── merchant-server/   # Go REST APIs
└── docker-compose.yml
```

## License
MIT
