FROM python:3.12-slim-bookworm

RUN apt-get update \
  && apt-get install -y --no-install-recommends espeak-ng libsndfile1 \
  && rm -rf /var/lib/apt/lists/*

WORKDIR /app

RUN pip install --no-cache-dir \
  "kokoro>=0.9.2" \
  "soundfile>=0.12" \
  "fastapi>=0.115" \
  "uvicorn[standard]>=0.32" \
  "numpy>=1.26"

COPY app.py /app/app.py

ENV KOKORO_LANG=a
ENV KOKORO_DEFAULT_VOICE=af_heart
EXPOSE 8880

CMD ["uvicorn", "app:app", "--host", "0.0.0.0", "--port", "8880"]
