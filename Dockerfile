# syntax=docker/dockerfile:1

FROM node:26-alpine@sha256:c6ad9e9c1d80f6eb607c26b4b0a132ff7a2344fad1c78ea3baa2a7d4c8de6d94 AS web
WORKDIR /src/web
RUN apk add --no-cache bash curl \
  && npm install -g pnpm@12.16.1
COPY web/package.json web/pnpm-lock.yaml web/pnpm-workspace.yaml ./
RUN pnpm install --frozen-lockfile
COPY web/ ./
COPY scripts/fetch-kokoro-models.sh /usr/local/bin/fetch-kokoro-models.sh
ENV KOKORO_MODEL_DIR=/src/web/vendor/kokoro-model/Kokoro-82M-v1.0-ONNX
RUN bash /usr/local/bin/fetch-kokoro-models.sh \
  && pnpm build:fast

FROM golang:1.27-alpine@sha256:d200d2dd76b03a99b52651fd1325aff7dcd35422c35c01bac56571b0507f3deb AS go
WORKDIR /src
RUN apk add --no-cache git
COPY go.mod go.sum ./
COPY vendor/ vendor/
COPY cmd/ cmd/
COPY internal/ internal/
COPY --from=web /src/internal/assets/dist ./internal/assets/dist
ARG VERSION=dev
ARG TARGETOS=linux
ARG TARGETARCH=amd64
ARG TARGETVARIANT=
RUN set -eux; \
    export CGO_ENABLED=0 GOOS="${TARGETOS}" GOARCH="${TARGETARCH}"; \
    case "${TARGETARCH}" in \
      arm) \
        case "${TARGETVARIANT}" in \
          v6) export GOARM=6 ;; \
          *) export GOARM=7 ;; \
        esac ;; \
    esac; \
    go build -mod=vendor -trimpath \
      -ldflags "-s -w -X athenaeum/internal/version.Version=${VERSION} -X athenaeum/internal/version.WebVersion=${VERSION}" \
      -o /out/athenaeum ./cmd/athenaeum

FROM alpine:3.24.3@sha256:be2656061b536c1883c4034790e21d0a09c5aaaf51968b89368a3c0371ea2c9a
ARG VERSION=dev
ARG REVISION=
ARG CREATED=
LABEL org.opencontainers.image.title="Athenaeum" \
      org.opencontainers.image.description="Self-hosted library for EPUB, PDF, comics, and audiobooks" \
      org.opencontainers.image.url="https://github.com/Quad4-Software/Athenaeum" \
      org.opencontainers.image.source="https://github.com/Quad4-Software/Athenaeum" \
      org.opencontainers.image.documentation="https://github.com/Quad4-Software/Athenaeum" \
      org.opencontainers.image.licenses="MIT" \
      org.opencontainers.image.vendor="Quad4 Software" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.revision="${REVISION}" \
      org.opencontainers.image.created="${CREATED}"
RUN apk add --no-cache ca-certificates tzdata wget \
    && addgroup -S athenaeum \
    && adduser -S -G athenaeum -h /app -s /sbin/nologin athenaeum \
    && mkdir -p /data /library \
    && chown -R athenaeum:athenaeum /data /library /app
COPY --from=go /out/athenaeum /usr/local/bin/athenaeum
USER athenaeum:athenaeum
WORKDIR /app
ENV ATHENAEUM_ADDR=:8080 \
    ATHENAEUM_DATA=/data \
    ATHENAEUM_LIBRARY=/library
VOLUME ["/data", "/library"]
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=5s --start-period=15s --retries=3 \
    CMD wget -qO- http://127.0.0.1:8080/api/health >/dev/null || exit 1
ENTRYPOINT ["/usr/local/bin/athenaeum"]
