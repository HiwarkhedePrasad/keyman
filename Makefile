BINARY     := keyman
MODULE     := github.com/keyman/keyman
VERSION    ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT     ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS := -s -w \
	-X main.version=$(VERSION) \
	-X main.commit=$(COMMIT) \
	-X main.buildTime=$(BUILD_TIME)

GO      := go
GOFLAGS := -trimpath
OUT_DIR := dist

.PHONY: all build run install uninstall test lint clean release

all: build

## Build for the current platform
build:
	@echo "→ Building $(BINARY) $(VERSION)"
	$(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BINARY) .

## Run without installing
run:
	$(GO) run . 

## Install to /usr/local/bin
install: build
	@echo "→ Installing to /usr/local/bin/$(BINARY)"
	install -m 0755 $(BINARY) /usr/local/bin/$(BINARY)

## Remove the installed binary
uninstall:
	rm -f /usr/local/bin/$(BINARY)

## Run all tests
test:
	$(GO) test ./... -v -race -timeout 30s

## Lint (requires golangci-lint)
lint:
	golangci-lint run ./...

## Cross-compile release binaries
release: clean
	@mkdir -p $(OUT_DIR)
	@echo "→ Building release binaries"
	GOOS=linux   GOARCH=amd64  $(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(OUT_DIR)/$(BINARY)-linux-amd64        .
	GOOS=linux   GOARCH=arm64  $(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(OUT_DIR)/$(BINARY)-linux-arm64        .
	GOOS=darwin  GOARCH=amd64  $(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(OUT_DIR)/$(BINARY)-darwin-amd64       .
	GOOS=darwin  GOARCH=arm64  $(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(OUT_DIR)/$(BINARY)-darwin-arm64       .
	GOOS=windows GOARCH=amd64  $(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(OUT_DIR)/$(BINARY)-windows-amd64.exe  .
	@echo "→ Checksums"
	cd $(OUT_DIR) && sha256sum * > checksums.txt
	@echo "→ Done. Binaries in $(OUT_DIR)/"

## Remove build artifacts
clean:
	rm -rf $(BINARY) $(OUT_DIR)

## Show help
help:
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## /  /'
