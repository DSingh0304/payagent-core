import uuid
from fastapi import APIRouter, HTTPException
from pydantic import BaseModel
from typing import Optional
from langchain_core.messages import HumanMessage
import os
import json
from datetime import datetime
import redis as redis_lib

from agent.workflow import create_workflow
from agent.state import AgentState

REDIS_URL = os.getenv("REDIS_URL", "redis://localhost:6379")
redis_client = redis_lib.from_url(REDIS_URL)

router = APIRouter()
graph = create_workflow()

def publish_events(session_id: str, messages: list):
    for msg in messages:
        if hasattr(msg, "tool_calls") and msg.tool_calls:
            for tc in msg.tool_calls:
                event = {
                    "session_id": session_id,
                    "event_type": f"TOOL_CALL_{tc['name'].upper()}",
                    "actor": "agent",
                    "reasoning": f"Called {tc['name']} with {json.dumps(tc['args'])}",
                    "outcome": "invoked",
                    "created_at": datetime.utcnow().isoformat(),
                }
                redis_client.publish(f"audit:{session_id}", json.dumps(event))
    
    last_msg = messages[-1] if messages else None
    if last_msg and hasattr(last_msg, "content") and last_msg.content:
        event = {
            "session_id": session_id,
            "event_type": "AGENT_RESPONSE",
            "actor": "agent",
            "reasoning": last_msg.content,
            "outcome": "completed",
            "created_at": datetime.utcnow().isoformat(),
        }
        redis_client.publish(f"audit:{session_id}", json.dumps(event))


class RunRequest(BaseModel):
    goal: str
    session_id: str | None = None


class ResumeRequest(BaseModel):
    decision: str


@router.post("/agent/run")
async def run_agent(req: RunRequest):
    session_id = req.session_id or str(uuid.uuid4())
    config = {"configurable": {"thread_id": session_id}}

    event = {
        "session_id": session_id,
        "event_type": "USER_GOAL",
        "actor": "user",
        "reasoning": req.goal,
        "outcome": "started",
        "created_at": datetime.utcnow().isoformat(),
    }
    redis_client.publish(f"audit:{session_id}", json.dumps(event))

    initial_state: AgentState = {
        "session_id": session_id,
        "goal": req.goal,
        "messages": [HumanMessage(content=f"Shopping goal: {req.goal}. My session ID is {session_id}.")],
        "cart": {},
        "wishlist": [],
        "pending_action": None,
        "decision": None,
        "completed": False,
        "guardrail": None,
    }

    result = await graph.ainvoke(initial_state, config=config)

    state_snapshot = await graph.aget_state(config)
    is_interrupted = bool(state_snapshot.tasks)

    messages = result.get("messages", [])
    publish_events(session_id, messages)
    
    if is_interrupted:
        event = {
            "session_id": session_id,
            "event_type": "APPROVAL_REQUIRED",
            "actor": "system",
            "reasoning": "Agent paused for human approval to create order.",
            "outcome": "pending",
            "created_at": datetime.utcnow().isoformat(),
        }
        redis_client.publish(f"audit:{session_id}", json.dumps(event))

    total_input_tokens = 0
    total_output_tokens = 0
    for msg in result.get("messages", []):
        usage = getattr(msg, "usage_metadata", None)
        if usage:
            total_input_tokens += usage.get("input_tokens", 0)
            total_output_tokens += usage.get("output_tokens", 0)

    return {
        "session_id": session_id,
        "status": "awaiting_approval" if is_interrupted else "completed",
        "messages": [
            m.content if not getattr(m, "tool_calls", None) else f"TOOL_CALL: {m.tool_calls}"
            for m in result.get("messages", []) if hasattr(m, "content")
        ],
        "token_usage": {
            "input_tokens": total_input_tokens,
            "output_tokens": total_output_tokens,
            "total_tokens": total_input_tokens + total_output_tokens,
            "estimated_cost_usd": round((total_input_tokens * 0.0000005 + total_output_tokens * 0.000001), 6),
        }
    }


@router.post("/agent/{session_id}/resume")
async def resume_agent(session_id: str, req: ResumeRequest):
    config = {"configurable": {"thread_id": session_id}}

    state_snapshot = await graph.aget_state(config)
    if not state_snapshot.tasks:
        raise HTTPException(status_code=404, detail="No interrupted agent found for this session")

    result = await graph.ainvoke({"decision": req.decision}, config=config)

    messages = result.get("messages", [])
    publish_events(session_id, messages)

    total_input_tokens = 0
    total_output_tokens = 0
    for msg in result.get("messages", []):
        usage = getattr(msg, "usage_metadata", None)
        if usage:
            total_input_tokens += usage.get("input_tokens", 0)
            total_output_tokens += usage.get("output_tokens", 0)

    return {
        "session_id": session_id,
        "status": "completed",
        "decision": req.decision,
        "messages": [m.content for m in result.get("messages", []) if hasattr(m, "content")],
        "token_usage": {
            "input_tokens": total_input_tokens,
            "output_tokens": total_output_tokens,
            "total_tokens": total_input_tokens + total_output_tokens,
            "estimated_cost_usd": round((total_input_tokens * 0.0000005 + total_output_tokens * 0.000001), 6),
        }
    }


class MessageRequest(BaseModel):
    message: str

@router.post("/agent/{session_id}/message")
async def send_message(session_id: str, req: MessageRequest):
    config = {"configurable": {"thread_id": session_id}}
    
    state_snapshot = await graph.aget_state(config)
    if not state_snapshot:
        raise HTTPException(status_code=404, detail="Session not found")
        
    result = await graph.ainvoke({"messages": [HumanMessage(content=req.message)]}, config=config)
    
    new_snapshot = await graph.aget_state(config)
    is_interrupted = bool(new_snapshot.tasks)

    messages = result.get("messages", [])
    publish_events(session_id, messages)
    
    total_input_tokens = 0
    total_output_tokens = 0
    for msg in result.get("messages", []):
        usage = getattr(msg, "usage_metadata", None)
        if usage:
            total_input_tokens += usage.get("input_tokens", 0)
            total_output_tokens += usage.get("output_tokens", 0)

    return {
        "session_id": session_id,
        "status": "awaiting_approval" if is_interrupted else "completed",
        "messages": [m.content for m in result.get("messages", []) if hasattr(m, "content")],
        "token_usage": {
            "input_tokens": total_input_tokens,
            "output_tokens": total_output_tokens,
            "total_tokens": total_input_tokens + total_output_tokens,
            "estimated_cost_usd": round((total_input_tokens * 0.0000005 + total_output_tokens * 0.000001), 6),
        }
    }

@router.get("/agent/{session_id}/state")
async def get_state(session_id: str):
    config = {"configurable": {"thread_id": session_id}}
    state_snapshot = await graph.aget_state(config)
    if not state_snapshot:
        raise HTTPException(status_code=404, detail="Session not found")
        
    messages_formatted = []
    total_input_tokens = 0
    total_output_tokens = 0

    for msg in state_snapshot.values.get("messages", []):
        if msg.type == "human":
            content = msg.content
            if content.startswith("Shopping goal: "):
                messages_formatted.append(f"USER: {content.split('. My session ID')[0].replace('Shopping goal: ', '')}")
            else:
                messages_formatted.append(f"USER: {content}")
        elif msg.type == "ai" and getattr(msg, "content", None):
            messages_formatted.append(msg.content)
            
        usage = getattr(msg, "usage_metadata", None)
        if usage:
            total_input_tokens += usage.get("input_tokens", 0)
            total_output_tokens += usage.get("output_tokens", 0)

    return {
        "session_id": session_id,
        "goal": state_snapshot.values.get("goal", ""),
        "status": "awaiting_approval" if bool(state_snapshot.tasks) else "completed",
        "messages": messages_formatted,
        "token_usage": {
            "input_tokens": total_input_tokens,
            "output_tokens": total_output_tokens,
            "total_tokens": total_input_tokens + total_output_tokens,
            "estimated_cost_usd": round((total_input_tokens * 0.0000005 + total_output_tokens * 0.000001), 6),
        }
    }
