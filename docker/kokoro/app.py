"""Minimal Kokoro TTS HTTP sidecar for Athenaeum narration."""

from __future__ import annotations

import io
import os
from typing import Any

import numpy as np
import soundfile as sf
from fastapi import FastAPI, Header, HTTPException
from fastapi.responses import Response
from pydantic import BaseModel, Field

APP_API_KEY = os.environ.get("KOKORO_API_KEY", "").strip()
DEFAULT_VOICE = os.environ.get("KOKORO_DEFAULT_VOICE", "af_heart")
LANG_CODE = os.environ.get("KOKORO_LANG", "a")

app = FastAPI(title="Athenaeum Kokoro Sidecar", version="1.0.0")

_pipeline = None
_pipeline_error: str | None = None


def get_pipeline():
    global _pipeline, _pipeline_error
    if _pipeline is not None:
        return _pipeline
    if _pipeline_error:
        raise RuntimeError(_pipeline_error)
    try:
        from kokoro import KPipeline

        _pipeline = KPipeline(lang_code=LANG_CODE)
        return _pipeline
    except Exception as exc:  # noqa: BLE001
        _pipeline_error = str(exc)
        raise


KNOWN_VOICES = [
    {"id": "af_heart", "label": "Heart (US female)", "lang": "en-us"},
    {"id": "af_bella", "label": "Bella (US female)", "lang": "en-us"},
    {"id": "af_sarah", "label": "Sarah (US female)", "lang": "en-us"},
    {"id": "am_adam", "label": "Adam (US male)", "lang": "en-us"},
    {"id": "am_michael", "label": "Michael (US male)", "lang": "en-us"},
    {"id": "bf_emma", "label": "Emma (UK female)", "lang": "en-gb"},
    {"id": "bm_george", "label": "George (UK male)", "lang": "en-gb"},
]


class SpeechRequest(BaseModel):
    text: str = Field(min_length=1, max_length=4000)
    voice: str | None = None
    speed: float = Field(default=1.0, ge=0.5, le=2.0)


def check_auth(authorization: str | None) -> None:
    if not APP_API_KEY:
        return
    if not authorization or not authorization.startswith("Bearer "):
        raise HTTPException(status_code=401, detail="missing bearer token")
    token = authorization.removeprefix("Bearer ").strip()
    if token != APP_API_KEY:
        raise HTTPException(status_code=401, detail="invalid api key")


@app.get("/health")
def health() -> dict[str, Any]:
    return {"ok": True, "engine": "kokoro"}


@app.get("/voices")
def voices(authorization: str | None = Header(default=None)) -> dict[str, Any]:
    check_auth(authorization)
    return {"voices": KNOWN_VOICES}


@app.post("/v1/audio/speech")
def speech(body: SpeechRequest, authorization: str | None = Header(default=None)) -> Response:
    check_auth(authorization)
    text = body.text.strip()
    if not text:
        raise HTTPException(status_code=400, detail="text is required")
    voice = (body.voice or DEFAULT_VOICE).strip() or DEFAULT_VOICE
    try:
        pipeline = get_pipeline()
        chunks: list[np.ndarray] = []
        for _gs, _ps, audio in pipeline(text, voice=voice, speed=body.speed):
            chunks.append(np.asarray(audio, dtype=np.float32))
        if not chunks:
            raise HTTPException(status_code=502, detail="no audio generated")
        wav = np.concatenate(chunks)
        buf = io.BytesIO()
        sf.write(buf, wav, 24000, format="WAV")
        return Response(content=buf.getvalue(), media_type="audio/wav")
    except HTTPException:
        raise
    except Exception as exc:  # noqa: BLE001
        raise HTTPException(status_code=502, detail=str(exc)) from exc
