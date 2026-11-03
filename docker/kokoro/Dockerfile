FROM python:3.12-slim-trixie@sha256:b18b261e1b4974b4f321451ad5251ded85abb22c2b8b262da527dda1d031ab30

LABEL org.opencontainers.image.title="Athenaeum Kokoro TTS sidecar" \
      org.opencontainers.image.description="Optional Kokoro TTS sidecar for Athenaeum narration" \
      org.opencontainers.image.url="https://github.com/Quad4-Software/Athenaeum" \
      org.opencontainers.image.source="https://github.com/Quad4-Software/Athenaeum" \
      org.opencontainers.image.licenses="MIT" \
      org.opencontainers.image.vendor="Quad4 Software" \
      org.opencontainers.image.base.name="docker.io/library/python:3.12-slim-trixie"

RUN apt-get update \
  && apt-get install -y --no-install-recommends espeak-ng libsndfile1 ffmpeg \
  && rm -rf /var/lib/apt/lists/*

WORKDIR /app

RUN pip install --no-cache-dir \
  "kokoro>=0.9.2" \
  "soundfile>=0.12" \
  "fastapi>=0.115" \
  "uvicorn[standard]>=0.32" \
  "numpy>=1.26"

COPY app.py /app/app.py

RUN groupadd -g 10001 kokoro \
  && useradd -u 10001 -g kokoro -d /app -s /usr/sbin/nologin -M kokoro \
  && chown -R 10001:10001 /app

ENV HOME=/app
ENV HF_HOME=/app/.cache/huggingface
ENV KOKORO_LANG=a
ENV KOKORO_DEFAULT_VOICE=af_heart
EXPOSE 8880

USER 10001:10001

HEALTHCHECK --interval=30s --timeout=10s --start-period=120s --retries=5 \
  CMD ["python", "-c", "import urllib.request; urllib.request.urlopen('http://127.0.0.1:8880/health')"]

CMD ["uvicorn", "app:app", "--host", "0.0.0.0", "--port", "8880"]
