# syntax=docker/dockerfile:1

FROM node:26-alpine@sha256:2d984a15c9b54fd0aeb608b8e0d0d83529eb34d2966db27a1fb4f1edc3d298a3 AS web
WORKDIR /src/web
RUN npm install -g pnpm@11.17.0
COPY web/package.json web/pnpm-lock.yaml web/pnpm-workspace.yaml ./
RUN pnpm install --frozen-lockfile --config.dangerouslyAllowAllBuilds=true
COPY web/ ./
RUN pnpm build:fast

FROM golang:1.26-alpine@sha256:28d89ee9cc0ff9fec75c82ca201e6bf7fdf9a679d4b7b24dfa04f2bb766bb468 AS go
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

FROM alpine:3.24.1@sha256:28bd5fe8b56d1bd048e5babf5b10710ebe0bae67db86916198a6eec434943f8b
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
