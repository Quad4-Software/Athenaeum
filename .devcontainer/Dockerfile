# Devcontainer base: Go + common packages. Node/pnpm come from features.
FROM mcr.microsoft.com/devcontainers/base:ubuntu-24.04@sha256:d94c97dd9cacf183d0a6fd12a8e87b526e9e928307674ae9c94139139c0c6eae

RUN apt-get update \
  && apt-get install -y --no-install-recommends \
    curl \
    ca-certificates \
    git \
    build-essential \
    sqlite3 \
  && rm -rf /var/lib/apt/lists/*

# Task runner (pinned release, checksum-verified)
ARG TARGETARCH
RUN set -eux; \
  TASK_VERSION="3.54.0"; \
  case "${TARGETARCH:-amd64}" in \
    amd64) TASK_SHA256="680859dbb4d881a9c72d4d9a8f510825450849af8567deacd7302c01124416fb" ;; \
    arm64) TASK_SHA256="d7189d439d4a6058e39e1b2acd7c8b27acd4856691b408f23b242e4294e1df5b" ;; \
    *) echo "unsupported arch ${TARGETARCH}"; exit 1 ;; \
  esac; \
  curl -fsSLo /tmp/task.tar.gz \
    "https://github.com/go-task/task/releases/download/v${TASK_VERSION}/task_linux_${TARGETARCH:-amd64}.tar.gz"; \
  echo "${TASK_SHA256}  /tmp/task.tar.gz" | sha256sum -c -; \
  tar -xzf /tmp/task.tar.gz -C /usr/local/bin task; \
  rm /tmp/task.tar.gz
