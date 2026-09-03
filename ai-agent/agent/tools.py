import os
import httpx
import time
from langchain_core.tools import tool

MERCHANT_BASE_URL = os.getenv("MERCHANT_BASE_URL", "http://localhost:8080")
MERCHANT_API_KEY = os.getenv("MERCHANT_API_KEY", "internal-agent-api-key-change-in-prod")

HEADERS = {"X-API-Key": MERCHANT_API_KEY, "Content-Type": "application/json"}


def _request_with_retry(method, url, max_retries=2, **kwargs):
    """Execute an HTTP request with exponential backoff retry."""
    kwargs.setdefault("timeout", 10)
    for attempt in range(max_retries + 1):
        try:
            if method == "GET":
                resp = httpx.get(url, **kwargs)
            elif method == "POST":
                resp = httpx.post(url, **kwargs)
            elif method == "DELETE":
                resp = httpx.delete(url, **kwargs)
            elif method == "PATCH":
                resp = httpx.patch(url, **kwargs)
            resp.raise_for_status()
            return resp
        except (httpx.ConnectError, httpx.TimeoutException, httpx.HTTPStatusError) as e:
            if attempt == max_retries:
                return {"error": str(e), "status": "TOOL_ERROR"}
            time.sleep(0.5 * (2 ** attempt))


@tool
def catalog_search(session_id: str, query: str = "", category: str = "", max_price_inr: float = 0, min_price_inr: float = 0) -> dict:
    """Search the product catalog by keyword, category, or price range.
    Always call this before suggesting, comparing, or adding products to cart.
    Returns a list of matching products with their IDs, names, prices, descriptions, stock, and tags.
    """
    params = {}
    if query:
        params["q"] = query
    if category:
        params["category"] = category
    if max_price_inr > 0:
        params["max_price_inr"] = max_price_inr
    if min_price_inr > 0:
        params["min_price_inr"] = min_price_inr

    headers = {**HEADERS, "X-Session-ID": session_id}
    resp = _request_with_retry("GET", f"{MERCHANT_BASE_URL}/api/v1/catalog/search", params=params, headers=headers)
    if isinstance(resp, dict) and "error" in resp:
        return resp
    return resp.json()


@tool
def catalog_get_product(product_id: str) -> dict:
    """Fetch full details of a single product by its ID.
    Use this when the user asks 'tell me more about X', 'what are the full specs of Y?',
    or when you need to verify stock or description before recommending.
    """
    resp = _request_with_retry("GET", f"{MERCHANT_BASE_URL}/api/v1/catalog/{product_id}", headers=HEADERS)
    if isinstance(resp, dict) and "error" in resp:
        return resp
    return resp.json()


@tool
def catalog_compare(session_id: str, product_ids: list) -> dict:
    """Compare multiple products side-by-side given a list of product IDs.
    Use this when the user says 'compare these', 'which is better?', 'difference between X and Y'.
    Returns a structured comparison of all products with price, stock, features, and value rating.
    """
    products = []
    for pid in product_ids:
        resp = _request_with_retry("GET", f"{MERCHANT_BASE_URL}/api/v1/catalog/{pid}", headers=HEADERS)
        if isinstance(resp, dict) and "error" in resp:
            products.append({"id": pid, "error": resp["error"]})
        else:
            products.append(resp.json())

    if not products:
        return {"error": "No products found to compare"}

    # Build a structured comparison
    comparison = {
        "products": products,
        "summary": {
            "cheapest": min(products, key=lambda p: p.get("price_paise", float("inf"))).get("name", "N/A"),
            "most_expensive": max(products, key=lambda p: p.get("price_paise", 0)).get("name", "N/A"),
            "best_stocked": max(products, key=lambda p: p.get("stock", 0)).get("name", "N/A"),
        }
    }
    return comparison


@tool
def cart_add(session_id: str, product_id: str, quantity: int, agent_reasoning: str) -> dict:
    """Add a product to the session cart.
    Always provide reasoning explaining why this specific product was chosen.
    Use catalog_search or catalog_get_product first to confirm the product exists.
    """
    payload = {
        "product_id": product_id,
        "quantity": quantity,
        "agent_reasoning": agent_reasoning,
    }
    resp = _request_with_retry(
        "POST",
        f"{MERCHANT_BASE_URL}/api/v1/cart/{session_id}/add",
        json=payload,
        headers=HEADERS,
    )
    if isinstance(resp, dict) and "error" in resp:
        return resp
    return resp.json()


@tool
def cart_get(session_id: str) -> dict:
    """Retrieve the current cart contents, item details, and total price for a session.
    Use this to show the user what's in their cart or before proceeding to checkout.
    """
    resp = _request_with_retry("GET", f"{MERCHANT_BASE_URL}/api/v1/cart/{session_id}", headers=HEADERS)
    if isinstance(resp, dict) and "error" in resp:
        return resp
    return resp.json()


@tool
def cart_remove(session_id: str, product_id: str, quantity: int = 0) -> dict:
    """Remove a specific product from the cart or decrease its quantity.
    Use when user says 'remove X', 'delete Y from cart', 'I don't want X anymore' (set quantity=0).
    Use when user says 'remove 1 of X', 'decrease quantity' (set quantity to the amount to decrease).
    """
    url = f"{MERCHANT_BASE_URL}/api/v1/cart/{session_id}/remove/{product_id}"
    if quantity > 0:
        url += f"?qty={quantity}"
    resp = _request_with_retry(
        "DELETE",
        url,
        headers=HEADERS,
    )
    if isinstance(resp, dict) and "error" in resp:
        return resp
    return resp.json()


@tool
def cart_clear(session_id: str) -> dict:
    """Remove ALL items from the cart at once.
    Use when user says 'clear my cart', 'start over', 'empty the cart'.
    """
    # Get current cart first, then remove each item
    cart_resp = _request_with_retry("GET", f"{MERCHANT_BASE_URL}/api/v1/cart/{session_id}", headers=HEADERS)
    if isinstance(cart_resp, dict) and "error" in cart_resp:
        return cart_resp
    cart = cart_resp.json()
    items = cart.get("items", [])
    removed = []
    for item in items:
        pid = item.get("product_id")
        if pid:
            _request_with_retry("DELETE", f"{MERCHANT_BASE_URL}/api/v1/cart/{session_id}/remove/{pid}", headers=HEADERS)
            removed.append(pid)
    return {"cleared": True, "items_removed": len(removed)}


@tool
def razorpay_create_order(session_id: str, amount_paise: int, currency: str = "INR", notes: dict = {}) -> dict:
    """Create a Razorpay payment order. ONLY call this after the user explicitly says they want to checkout/buy/pay.
    This will pause the agent and require human approval before executing.
    amount_paise is the total in paise (INR × 100). E.g., ₹1499 = 149900 paise.
    """
    payload = {"amount": amount_paise, "currency": currency, "notes": notes}

    auth = None
    key_id = os.getenv("RAZORPAY_KEY_ID")
    key_secret = os.getenv("RAZORPAY_KEY_SECRET")
    if key_id and key_secret:
        auth = (key_id, key_secret)

    resp = _request_with_retry("POST", "https://api.razorpay.com/v1/orders", json=payload, auth=auth)
    if isinstance(resp, dict) and "error" in resp:
        return resp
    rzp_data = resp.json()

    if "id" in rzp_data:
        go_payload = {
            "session_id": session_id,
            "razorpay_order_id": rzp_data["id"],
            "amount_paise": amount_paise,
            "items": {}
        }
        go_resp = httpx.post(f"{MERCHANT_BASE_URL}/api/v1/orders", json=go_payload, headers=HEADERS, timeout=10)
        rzp_data["merchant_order"] = go_resp.json()

    return rzp_data


TOOLS = [
    catalog_search,
    catalog_get_product,
    catalog_compare,
    cart_add,
    cart_get,
    cart_remove,
    cart_clear,
    razorpay_create_order,
]