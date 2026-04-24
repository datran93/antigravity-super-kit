# agk-cli Makefile
# Usage: make build | make install | make clean

BINARY    := agk
BUILD_DIR := .
INSTALL   := $(HOME)/.local/bin/$(BINARY)
GOFLAGS   := -trimpath -ldflags="-s -w"

.PHONY: build install clean vet test

## build: Compile the agk binary
build:
	go build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY) .

## install: Build and symlink to ~/.local/bin/agk
install: build
	@mkdir -p $(HOME)/.local/bin
	@ln -sf $(CURDIR)/$(BINARY) $(INSTALL)
	@echo "✅ Installed: $(INSTALL) → $(CURDIR)/$(BINARY)"

## clean: Remove built binary
clean:
	rm -f $(BUILD_DIR)/$(BINARY)

## vet: Run go vet
vet:
	go vet ./...

## test: Run tests
test:
	go test ./...

## help: Show this help
help:
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## //' | column -t -s ':'
