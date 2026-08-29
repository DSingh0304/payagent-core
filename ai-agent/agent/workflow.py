import os
import json
from langchain_google_genai import ChatGoogleGenerativeAI
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
2. Always provide clear reasoning when adding items.
3. Before calling razorpay_create_order, you MUST pause and request human approval.
4. Never exceed the budget stated in the goal.
5. If stock is unavailable, find an alternative and explain why.

When you are ready to checkout, state the total and the items clearly, then wait for approval.
"""


def create_workflow():
    llm = ChatGoogleGenerativeAI(model="gemini-3.6-flash", temperature=0)
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

    def after_approval(state: AgentState) -> str:
        if state.get("decision") == "approved":
            return "tools"
        return "end"

    checkpointer = MemorySaver()

    builder = StateGraph(AgentState)
    builder.add_node("agent", agent_node)
    builder.add_node("tools", ToolNode(TOOLS))
    builder.add_node("human_approval", human_approval_node)

    builder.set_entry_point("agent")
    builder.add_conditional_edges("agent", should_interrupt, {
        "tools": "tools",
        "approval": "human_approval",
        "end": END,
    })
    builder.add_edge("tools", "agent")
    builder.add_conditional_edges("human_approval", after_approval, {
        "tools": "tools",
        "end": END,
    })

    return builder.compile(checkpointer=checkpointer, interrupt_before=["human_approval"])