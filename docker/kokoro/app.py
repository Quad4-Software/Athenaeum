"""Kokoro TTS sidecar for Athenaeum narration and audiobook jobs.

Speaks the OpenAI /v1/audio/speech shape, so the bundled image can be
swapped for Kokoro-FastAPI, Speaches, or a hosted endpoint without
server changes.
"""

from __future__ import annotations

import io
import os
import subprocess
from typing import Any

import numpy as np
import soundfile as sf
from fastapi import FastAPI, Header, HTTPException
from fastapi.responses import Response, StreamingResponse
from pydantic import BaseModel, Field

APP_API_KEY = os.environ.get("KOKORO_API_KEY", "").strip()
DEFAULT_VOICE = os.environ.get("KOKORO_DEFAULT_VOICE", "af_heart")
LANG_CODE = os.environ.get("KOKORO_LANG", "a")
MAX_INPUT_CHARS = int(os.environ.get("KOKORO_MAX_INPUT_CHARS", "10000"))

SAMPLE_RATE = 24000

# request response_format -> (ffmpeg muxer args, media type)
FORMATS: dict[str, tuple[list[str], str]] = {
    "wav": (["-f", "wav"], "audio/wav"),
    "mp3": (["-f", "mp3", "-b:a", "96k"], "audio/mpeg"),
    "opus": (["-f", "opus", "-b:a", "64k"], "audio/ogg"),
    "flac": (["-f", "flac"], "audio/flac"),
    "aac": (["-f", "adts", "-b:a", "96k"], "audio/aac"),
    "m4a": (
        ["-f", "ipod", "-b:a", "96k", "-movflags", "frag_keyframe+empty_moov+default_base_moof"],
        "audio/mp4",
    ),
    "pcm": (["-f", "s16le"], "audio/pcm"),
}

# Formats whose per-segment output can be concatenated into a streamable body.
STREAMABLE = {"mp3", "opus", "aac", "flac", "pcm"}

app = FastAPI(title="Athenaeum Kokoro Sidecar", version="2.0.0")

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
    """OpenAI-compatible speech request. `text` is accepted as a legacy
    alias for `input`. `model` is accepted and ignored: this sidecar only
    serves Kokoro."""

    model: str | None = None
    input: str | None = Field(default=None, max_length=MAX_INPUT_CHARS)
    text: str | None = Field(default=None, max_length=MAX_INPUT_CHARS)
    voice: str | None = None
    response_format: str = "wav"
    speed: float = Field(default=1.0, ge=0.25, le=4.0)
    stream: bool = False


def check_auth(authorization: str | None) -> None:
    if not APP_API_KEY:
        return
    if not authorization or not authorization.startswith("Bearer "):
        raise HTTPException(status_code=401, detail="missing bearer token")
    token = authorization.removeprefix("Bearer ").strip()
    if token != APP_API_KEY:
        raise HTTPException(status_code=401, detail="invalid api key")


def to_wav_bytes(wav: np.ndarray) -> bytes:
    buf = io.BytesIO()
    sf.write(buf, wav, SAMPLE_RATE, format="WAV")
    return buf.getvalue()


def encode(wav: np.ndarray, fmt: str) -> bytes:
    """Encode a float32 mono waveform through ffmpeg."""
    if fmt == "wav":
        return to_wav_bytes(wav)
    args, _ = FORMATS[fmt]
    pcm = np.asarray(wav, dtype=np.float32).tobytes()
    proc = subprocess.run(
        [
            "ffmpeg", "-v", "error",
            "-f", "f32le", "-ar", str(SAMPLE_RATE), "-ac", "1", "-i", "pipe:0",
            *args, "pipe:1",
        ],
        input=pcm,
        capture_output=True,
        check=False,
    )
    if proc.returncode != 0:
        raise RuntimeError(f"ffmpeg failed: {proc.stderr.decode(errors='replace')[:300]}")
    return proc.stdout


@app.get("/health")
def health() -> dict[str, Any]:
    return {"ok": True, "engine": "kokoro"}


@app.get("/voices")
def voices(authorization: str | None = Header(default=None)) -> dict[str, Any]:
    check_auth(authorization)
    return {"voices": KNOWN_VOICES}


@app.get("/v1/audio/voices")
def voices_v1(authorization: str | None = Header(default=None)) -> dict[str, Any]:
    return voices(authorization)


@app.post("/v1/audio/speech")
def speech(body: SpeechRequest, authorization: str | None = Header(default=None)) -> Response:
    check_auth(authorization)
    text = (body.input or body.text or "").strip()
    if not text:
        raise HTTPException(status_code=400, detail="input is required")
    fmt = (body.response_format or "wav").lower()
    if fmt not in FORMATS:
        raise HTTPException(
            status_code=400,
            detail=f"unsupported response_format {fmt!r}; expected one of {sorted(FORMATS)}",
        )
    voice = (body.voice or DEFAULT_VOICE).strip() or DEFAULT_VOICE
    _, media_type = FORMATS[fmt]

    try:
        pipeline = get_pipeline()

        if body.stream and fmt in STREAMABLE:
            def gen():
                for _gs, _ps, audio in pipeline(text, voice=voice, speed=body.speed):
                    yield encode(np.asarray(audio, dtype=np.float32), fmt)

            return StreamingResponse(gen(), media_type=media_type)

        chunks: list[np.ndarray] = []
        for _gs, _ps, audio in pipeline(text, voice=voice, speed=body.speed):
            chunks.append(np.asarray(audio, dtype=np.float32))
        if not chunks:
            raise HTTPException(status_code=502, detail="no audio generated")
        return Response(content=encode(np.concatenate(chunks), fmt), media_type=media_type)
    except HTTPException:
        raise
    except Exception as exc:  # noqa: BLE001
        raise HTTPException(status_code=502, detail=str(exc)) from exc
