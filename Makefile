# Athenaeum - single-binary EPUB/PDF/audiobook library server.
#
# Common targets:
#   make build         Build the frontend and compile the binary into bin/athenaeum
#   make install       Install the binary and man pages (PREFIX=/usr/local)
#   make uninstall     Remove the installed binary and man pages
#   make man           (no-op; man pages are pre-written under man/)
#   make clean         Remove build artifacts
#
# Honors the standard PREFIX and DESTDIR variables for packaging.

SHELL := /bin/sh

BIN     := athenaeum
PKG     := ./...
CMD     := ./cmd/athenaeum
BINDIR  := bin
BINPATH := $(BINDIR)/$(BIN)

PREFIX      ?= /usr/local
DESTDIR     ?=
EXEC_PREFIX ?= $(PREFIX)
BINPREFIX   ?= $(EXEC_PREFIX)/bin
DATAROOTDIR ?= $(PREFIX)/share
MANDIR      ?= $(DATAROOTDIR)/man
MAN1DIR     ?= $(MANDIR)/man1

INSTALL         ?= install
INSTALL_PROGRAM ?= $(INSTALL) -m 0755
INSTALL_DATA    ?= $(INSTALL) -m 0644
INSTALL_DIR     ?= $(INSTALL) -d -m 0755

GO ?= go

# Version embedded into the binary, taken from web/package.json so the API and
# UI report the same release. Falls back to "dev" when it cannot be read.
VERSION := $(shell sed -n 's/.*"version"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' web/package.json 2>/dev/null | head -n1)
ifeq ($(strip $(VERSION)),)
VERSION := dev
endif

LDFLAGS := -s -w \
	-X athenaeum/internal/version.Version=$(VERSION) \
	-X athenaeum/internal/version.WebVersion=$(VERSION)

MAN1 := man/athenaeum.1 man/athenaeum-users.1

.DEFAULT_GOAL := build

.PHONY: all build build-slim build-web build-web-slim build-go build-go-slim \
	build-cross build-cross-slim build-cross-all self-check install install-bin \
	install-man uninstall man clean help

all: build

## build: Build the frontend bundle and compile the single binary.
build: build-web build-go

## build-slim: Build without in-browser Kokoro WASM into bin/athenaeum-slim.
build-slim: build-web-slim build-go-slim

## build-web: Build the production frontend embedded into the binary.
build-web:
	cd web && pnpm install && pnpm build

## build-web-slim: Build frontend without Kokoro WASM (smaller embed).
build-web-slim:
	cd web && pnpm install && pnpm build:slim

## build-go: Compile the Go binary into bin/athenaeum (assumes frontend is built).
build-go:
	$(INSTALL_DIR) $(BINDIR)
	$(GO) build -trimpath -ldflags "$(LDFLAGS)" -o $(BINPATH) $(CMD)

## build-go-slim: Compile bin/athenaeum-slim (assumes slim frontend is built).
build-go-slim:
	$(INSTALL_DIR) $(BINDIR)
	$(GO) build -trimpath -ldflags "$(LDFLAGS)" -o $(BINDIR)/$(BIN)-slim $(CMD)

## build-cross: Cross-compile all release targets into bin/.
build-cross:
	$(INSTALL_DIR) $(BINDIR)
	@VERSION="$(VERSION)" OUT_DIR="$(BINDIR)" bash scripts/build-release.sh

## build-cross-slim: Cross-compile slim binaries (no Kokoro WASM) into bin/.
build-cross-slim: build-web-slim
	$(INSTALL_DIR) $(BINDIR)
	@SLIM=1 VERSION="$(VERSION)" OUT_DIR="$(BINDIR)" bash scripts/build-release.sh

## build-cross-all: Cross-compile full and slim binaries into bin/.
build-cross-all: build-cross-slim build-web build-cross

## self-check: Build (if needed) and run athenaeum --self-check.
self-check: build-go
	./$(BINPATH) --self-check

## install: Install the binary and man pages under $(PREFIX).
install: install-bin install-man

## install-bin: Install the compiled binary to $(DESTDIR)$(BINPREFIX).
install-bin: build-go
	$(INSTALL_DIR) "$(DESTDIR)$(BINPREFIX)"
	$(INSTALL_PROGRAM) $(BINPATH) "$(DESTDIR)$(BINPREFIX)/$(BIN)"

## install-man: Install man pages to $(DESTDIR)$(MAN1DIR).
install-man:
	$(INSTALL_DIR) "$(DESTDIR)$(MAN1DIR)"
	$(INSTALL_DATA) $(MAN1) "$(DESTDIR)$(MAN1DIR)/"

## uninstall: Remove the installed binary and man pages.
uninstall:
	rm -f "$(DESTDIR)$(BINPREFIX)/$(BIN)"
	rm -f "$(DESTDIR)$(MAN1DIR)/athenaeum.1" "$(DESTDIR)$(MAN1DIR)/athenaeum-users.1"

## man: Man pages are maintained by hand under man/; this target is a no-op.
man:
	@echo "man pages live in man/ (athenaeum.1, athenaeum-users.1)"

## clean: Remove build artifacts.
clean:
	rm -rf $(BINDIR) web/dist

## help: List available targets.
help:
	@grep -E '^## ' $(MAKEFILE_LIST) | sed -e 's/^## //' | awk -F': ' '{printf "  %-14s %s\n", $$1, $$2}'
