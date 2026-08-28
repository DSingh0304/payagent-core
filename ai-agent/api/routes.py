import uuid
from fastapi import APIRouter, HTTPException
from pydantic import BaseModel
from langchain_core.messages import HumanMessage

from agent.workflow import create_workflow
from agent.state import AgentState

router = APIRouter()
graph = create_workflow()


class RunRequest(BaseModel):
    goal: str
    session_id: str | None = None


class ResumeRequest(BaseModel):
    decision: str


@router.post("/agent/run")
async def run_agent(req: RunRequest):
    session_id = req.session_id or str(uuid.uuid4())
    config = {"configurable": {"thread_id": session_id}}

    initial_state: AgentState = {
        "session_id": session_id,
        "goal": req.goal,
        "messages": [HumanMessage(content=f"Shopping goal: {req.goal}. My session ID is {session_id}.")],
        "cart": {},
        "pending_action": None,
        "decision": None,
        "completed": False,
    }

    result = await graph.ainvoke(initial_state, config=config)

    state_snapshot = await graph.aget_state(config)
    is_interrupted = bool(state_snapshot.tasks)

    return {
        "session_id": session_id,
        "status": "awaiting_approval" if is_interrupted else "completed",
        "messages": [m.content for m in result.get("messages", []) if hasattr(m, "content")],
    }


@router.post("/agent/{session_id}/resume")
async def resume_agent(session_id: str, req: ResumeRequest):
    config = {"configurable": {"thread_id": session_id}}

    state_snapshot = await graph.aget_state(config)
    if not state_snapshot.tasks:
        raise HTTPException(status_code=404, detail="No interrupted agent found for this session")

    result = await graph.ainvoke({"decision": req.decision}, config=config)

    return {
        "session_id": session_id,
        "status": "completed",
        "decision": req.decision,
        "messages": [m.content for m in result.get("messages", []) if hasattr(m, "content")],
    }


@router.get("/agent/{session_id}/state")
async def get_state(session_id: str):
    config = {"configurable": {"thread_id": session_id}}
    state_snapshot = await graph.aget_state(config)
    if not state_snapshot:
        raise HTTPException(status_code=404, detail="Session not found")
    return {
        "session_id": session_id,
        "is_interrupted": bool(state_snapshot.tasks),
        "values": state_snapshot.values,
    }
