import logging
import os
from typing import List

from fastapi import FastAPI, HTTPException
from openai import OpenAI
from pydantic import BaseModel, Field
from dotenv import load_dotenv

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s | %(levelname)s | %(message)s",
    datefmt="%H:%M:%S",
)

# Load env vars from open-router/.env regardless of current working directory.
load_dotenv(dotenv_path=os.path.join(os.path.dirname(__file__), ".env"))

OPENROUTER_BASE_URL = os.getenv("OPENROUTER_BASE_URL", "https://openrouter.ai/api/v1")
OPENROUTER_TIMEOUT = float(os.getenv("OPENROUTER_TIMEOUT", "600"))

app = FastAPI(title="OpenRouter Proxy Service")


class ChatMessage(BaseModel):
    role: str
    content: str


class ChatRequest(BaseModel):
    model: str
    system_prompt: str = ""
    messages: List[ChatMessage] = Field(default_factory=list)
    instruction: str = ""


@app.get("/health")
def health() -> dict:
    return {"status": "ok"}


@app.post("/chat")
def chat(req: ChatRequest) -> dict:
    api_key = os.getenv("OPENROUTER_API_KEY")
    if not api_key:
        raise HTTPException(status_code=500, detail="OPENROUTER_API_KEY is not set")

    client = OpenAI(
        api_key=api_key,
        base_url=OPENROUTER_BASE_URL.rstrip("/"),
        timeout=OPENROUTER_TIMEOUT,
    )

    outgoing_messages = []
    if req.system_prompt.strip():
        outgoing_messages.append({"role": "system", "content": req.system_prompt})

    for msg in req.messages:
        outgoing_messages.append({"role": msg.role, "content": msg.content})

    if req.instruction.strip():
        outgoing_messages.append({"role": "user", "content": req.instruction})

    if not outgoing_messages:
        raise HTTPException(status_code=400, detail="messages are empty")

    extra_headers = {}
    http_referer = os.getenv("OPENROUTER_HTTP_REFERER")
    x_title = os.getenv("OPENROUTER_X_TITLE")
    if http_referer:
        extra_headers["HTTP-Referer"] = http_referer
    if x_title:
        extra_headers["X-Title"] = x_title

    logging.info("OpenRouter request: model=%s messages=%d", req.model, len(outgoing_messages))

    try:
        completion = client.chat.completions.create(
            model=req.model,
            messages=outgoing_messages,
            extra_headers=extra_headers or None,
        )
    except Exception as e:
        raise HTTPException(status_code=502, detail=f"openrouter request failed: {e}")

    text = ""
    if completion.choices and completion.choices[0].message:
        text = completion.choices[0].message.content or ""

    if not text:
        raise HTTPException(status_code=502, detail="empty response from OpenRouter")

    return {
        "status": "ok",
        "model": completion.model or req.model,
        "response": text,
        "id": completion.id,
        "usage": completion.usage.model_dump() if completion.usage else None,
    }
