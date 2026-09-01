import os
import httpx
from pprint import pprint

key_id = os.getenv("RAZORPAY_KEY_ID")
key_secret = os.getenv("RAZORPAY_KEY_SECRET")
auth = (key_id, key_secret)

payload = {"amount": 1000, "currency": "INR", "notes": {}}
try:
    resp = httpx.post("https://api.razorpay.com/v1/orders", json=payload, auth=auth)
    resp.raise_for_status()
    print("Success:")
    pprint(resp.json())
except Exception as e:
    print("Error:")
    if hasattr(e, 'response') and e.response:
        print(e.response.text)
    else:
        print(e)
