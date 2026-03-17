.PHONY: build clean test vet

# FTS5 + Tree-sitter build.
# CGO is required for both go-sqlite3 (FTS5) and go-tree-sitter.
CGO_FLAGS := TMPDIR=/tmp GOCACHE=/tmp/go-cache CGO_CFLAGS="-DSQLITE_ENABLE_FTS5"
BUILD_TAGS := -tags "fts5"
BINARY     := mcp-codebase-explorer-go

build:
	$(CGO_FLAGS) go build $(BUILD_TAGS) -o $(BINARY) .

clean:
	rm -f $(BINARY)

test:
	$(CGO_FLAGS) go test $(BUILD_TAGS) ./... -v

vet:
	$(CGO_FLAGS) go vet $(BUILD_TAGS) ./...
