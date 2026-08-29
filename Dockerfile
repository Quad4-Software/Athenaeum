# syntax=docker/dockerfile:1

FROM node:alpine AS web
WORKDIR /src/web
RUN npm install -g pnpm@11.17.0
COPY web/package.json web/pnpm-lock.yaml web/pnpm-workspace.yaml ./
RUN pnpm install --frozen-lockfile --config.dangerouslyAllowAllBuilds=true
COPY web/ ./
RUN pnpm build:fast

FROM golang:1.26-alpine AS go
WORKDIR /src
RUN apk add --no-cache git
COPY go.mod go.sum ./
COPY vendor/ vendor/
COPY cmd/ cmd/
COPY internal/ internal/
COPY --from=web /src/internal/assets/dist ./internal/assets/dist
ARG VERSION=dev
RUN CGO_ENABLED=0 GOOS=linux go build -mod=vendor -trimpath \
    -ldflags "-s -w -X athenaeum/internal/version.Version=${VERSION} -X athenaeum/internal/version.WebVersion=${VERSION}" \
    -o /out/athenaeum ./cmd/athenaeum

FROM alpine:latest
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
