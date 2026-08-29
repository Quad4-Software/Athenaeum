# Devcontainer base: Go + common packages. Node/pnpm come from features.
FROM mcr.microsoft.com/devcontainers/base:ubuntu-24.04

RUN apt-get update \
  && apt-get install -y --no-install-recommends \
    curl \
    ca-certificates \
    git \
    build-essential \
    sqlite3 \
  && rm -rf /var/lib/apt/lists/*

# Task runner
RUN curl -fsSL https://taskfile.dev/install.sh | sh -s -- -d -b /usr/local/bin
