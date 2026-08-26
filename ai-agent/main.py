import os
from dotenv import load_dotenv
from fastapi import FastAPI
from api.routes import router

load_dotenv()

app = FastAPI(title="PayAgent AI Service", version="1.0.0")
app.include_router(router)

@app.get("/health")
def health():
        return {"status": "ok"}