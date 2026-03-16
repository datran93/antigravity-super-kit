.PHONY: build clean test

# FTS5 is required for knowledge BM25 search.
# - CGO_CFLAGS injects SQLITE_ENABLE_FTS5 into go-sqlite3's embedded SQLite.
# - -tags fts5 is the go-sqlite3 recommended build tag for FTS5 support.
CGO_FLAGS := CGO_CFLAGS="-DSQLITE_ENABLE_FTS5"
BUILD_TAGS := -tags "fts5"
BINARY     := mcp-context-manager-go

build:
	$(CGO_FLAGS) go build $(BUILD_TAGS) -o $(BINARY) .

clean:
	rm -f $(BINARY)

test:
	$(CGO_FLAGS) go test $(BUILD_TAGS) ./... -v

vet:
	$(CGO_FLAGS) go vet $(BUILD_TAGS) ./...
