import os
import json
from langchain_groq import ChatGroq
from langchain_core.messages import SystemMessage, HumanMessage
from langgraph.graph import StateGraph, END
from langgraph.checkpoint.memory import MemorySaver
from langgraph.prebuilt import ToolNode
from langgraph.types import interrupt

from agent.state import AgentState
from agent.tools import TOOLS

REDIS_URL = os.getenv("REDIS_URL", "redis://redis:6379")

SYSTEM_PROMPT = """You are PayAgent, an autonomous AI shopping assistant.
Your job is to fulfill a buyer's shopping goal by searching the catalog, building a cart, and preparing a Razorpay order.

Rules you must follow:
1. Always search the catalog before adding anything to the cart.
2. Only add items to the cart that the user explicitly wants or conceptually matched.
3. If multiple items match, autonomously pick the best or cheapest one and ADD IT TO THE CART IMMEDIATELY. Do not ask the user for clarification. You are a fully autonomous agent.
4. Once the cart is ready, execute razorpay_create_order tool to request user payment.
5. You MUST stop and wait after calling razorpay_create_order. Do not proceed until approved.
6. Do not invent products. Only use what is returned by the catalog search.
"""


def create_workflow():
    llm = ChatGroq(model="openai/gpt-oss-120b", temperature=0)
    llm_with_tools = llm.bind_tools(TOOLS)

    def agent_node(state: AgentState) -> dict:
        messages = [SystemMessage(content=SYSTEM_PROMPT)] + state["messages"]
        response = llm_with_tools.invoke(messages)
        return {"messages": [response]}

    def human_approval_node(state: AgentState) -> dict:
        # Pause execution here and surface the pending action to the dashboard.
        # The workflow will remain suspended in Redis until /resume is called.
        decision = interrupt({
            "type": "APPROVAL_REQUIRED",
            "cart": state.get("cart", {}),
            "message": "Agent wants to create a Razorpay payment order. Please approve or reject.",
        })
        return {"decision": decision}

    import httpx
    from langchain_core.messages import AIMessage
    
    MERCHANT_BASE_URL = os.getenv("MERCHANT_BASE_URL", "http://localhost:8080")
    MERCHANT_API_KEY = os.getenv("MERCHANT_API_KEY", "internal-agent-api-key-change-in-prod")
    HEADERS = {"X-API-Key": MERCHANT_API_KEY, "Content-Type": "application/json"}
    MAX_ORDER_INR = float(os.getenv("MAX_ORDER_AMOUNT_INR", "5000"))

    def guardrail_node(state: AgentState, config: dict) -> dict:
        """Check the cart total and reject if it exceeds the spending limit."""
        messages = state["messages"]
        last = messages[-1] if messages else None
        
        if last and hasattr(last, "tool_calls"):
            for tc in last.tool_calls:
                if tc["name"] == "razorpay_create_order":
                    amount_paise = tc["args"].get("amount_paise", 0)
                    amount_inr = amount_paise / 100
                    if amount_inr > MAX_ORDER_INR:
                        rejection = AIMessage(content=f"⚠️ GUARDRAIL: Order of ₹{amount_inr:.2f} exceeds the spending limit of ₹{MAX_ORDER_INR:.2f}. Order automatically rejected for safety.")
                        
                        try:
                            session_id = config.get("configurable", {}).get("thread_id")
                            if session_id:
                                httpx.post(f"{MERCHANT_BASE_URL}/api/v1/audit", json={
                                    "session_id": session_id,
                                    "event_type": "GUARDRAIL_TRIGGERED",
                                    "actor": "system",
                                    "reasoning": f"Order of ₹{amount_inr:.2f} exceeds limit of ₹{MAX_ORDER_INR:.2f}",
                                    "outcome": "blocked"
                                }, headers=HEADERS, timeout=5)
                        except Exception as e:
                            print(f"Failed to post guardrail audit event: {e}")
                            
                        return {"messages": [rejection], "completed": True, "guardrail": "blocked"}
        
        return {"guardrail": "passed"}

    def should_interrupt(state: AgentState) -> str:
        messages = state["messages"]
        last = messages[-1] if messages else None
        if not last or not hasattr(last, "tool_calls"):
            return "end"

        tool_names = [tc["name"] for tc in last.tool_calls]

        if "razorpay_create_order" in tool_names:
            return "approval"

        if last.tool_calls:
            return "tools"

        return "end"
        
    def check_guardrail_result(state: AgentState) -> str:
        if state.get("guardrail") == "blocked":
            return "end"
        return "passed"

    def after_approval(state: AgentState) -> str:
        if state.get("decision") == "approved":
            return "tools"
        return "end"

    checkpointer = MemorySaver()

    builder = StateGraph(AgentState)
    builder.add_node("agent", agent_node)
    builder.add_node("tools", ToolNode(TOOLS))
    builder.add_node("guardrail", guardrail_node)
    builder.add_node("human_approval", human_approval_node)

    builder.set_entry_point("agent")
    builder.add_conditional_edges("agent", should_interrupt, {
        "tools": "tools",
        "approval": "guardrail",
        "end": END,
    })
    builder.add_edge("tools", "agent")
    builder.add_conditional_edges("guardrail", check_guardrail_result, {
        "blocked": END,
        "passed": "human_approval",
    })
    builder.add_conditional_edges("human_approval", after_approval, {
        "tools": "tools",
        "end": END,
    })

    return builder.compile(checkpointer=checkpointer, interrupt_before=["human_approval"])