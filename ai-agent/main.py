import os
from dotenv import load_dotenv
from fastapi import FastAPI
from api.routes import router

from fastapi.middleware.cors import CORSMiddleware

load_dotenv()

app = FastAPI(title="PayAgent AI Service", version="1.0.0")

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=False,
    allow_methods=["*"],
    allow_headers=["*"],
)

app.include_router(router)

@app.get("/health")
def health():
        return {"status": "ok"}