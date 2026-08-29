import os
import httpx
from langchain_core.tools import tool

MERCHANT_BASE_URL = os.getenv("MERCHANT_BASE_URL", "http://localhost:8080")
MERCHANT_API_KEY = os.getenv("MERCHANT_API_KEY", "internal-agent-api-key-change-in-prod")
RAZORPAY_MCP_URL = os.getenv("RAZORPAY_MCP_URL", "http://localhost:3000")

HEADERS = {"X-API-Key": MERCHANT_API_KEY, "Content-Type": "application/json"}


@tool
def catalog_search(session_id: str, query: str = "", category: str = "", max_price_inr: float = 0) -> dict:
    """Search the product catalog. Use this before adding anything to the cart."""
    params = {}
    if query:
        params["q"] = query
    if category:
        params["category"] = category
    if max_price_inr > 0:
        params["max_price_inr"] = max_price_inr

    headers = {**HEADERS, "X-Session-ID": session_id}
    resp = httpx.get(f"{MERCHANT_BASE_URL}/api/v1/catalog/search", params=params, headers=headers, timeout=10)
    return resp.json()


@tool
def cart_add(session_id: str, product_id: str, quantity: int, agent_reasoning: str) -> dict:
    """Add a product to the session cart. Always provide your reasoning for why this product was chosen."""
    payload = {
        "product_id": product_id,
        "quantity": quantity,
        "agent_reasoning": agent_reasoning,
    }
    resp = httpx.post(
        f"{MERCHANT_BASE_URL}/api/v1/cart/{session_id}/add",
        json=payload,
        headers=HEADERS,
        timeout=10,
    )
    return resp.json()


@tool
def cart_get(session_id: str) -> dict:
    """Retrieve the current cart contents and total for a session."""
    resp = httpx.get(f"{MERCHANT_BASE_URL}/api/v1/cart/{session_id}", headers=HEADERS, timeout=10)
    return resp.json()


@tool
def cart_remove(session_id: str, product_id: str) -> dict:
    """Remove a product from the cart, for example when a guardrail is triggered."""
    resp = httpx.delete(
        f"{MERCHANT_BASE_URL}/api/v1/cart/{session_id}/remove/{product_id}",
        headers=HEADERS,
        timeout=10,
    )
    return resp.json()


@tool
def razorpay_create_order(amount_paise: int, currency: str = "INR", notes: dict = {}) -> dict:
    """Create a Razorpay payment order via the MCP server. Called only after explicit user approval."""
    payload = {"amount": amount_paise, "currency": currency, "notes": notes}
    resp = httpx.post(f"{RAZORPAY_MCP_URL}/orders", json=payload, timeout=10)
    return resp.json()


TOOLS = [catalog_search, cart_add, cart_get, cart_remove, razorpay_create_order]