.PHONY: build clean test vet

# FTS5 is required for BM25 keyword search (chunks_fts virtual table).
# - CGO_CFLAGS injects SQLITE_ENABLE_FTS5 into go-sqlite3's embedded SQLite C source.
# - -tags fts5 is the go-sqlite3 recommended build tag to enable FTS5 at the driver level.
#
# WARNING: Do NOT run plain `go build .` — it will produce a binary without FTS5 support
# and the server will crash at startup with "no such module: fts5".
CGO_FLAGS := TMPDIR=/tmp GOCACHE=/tmp/go-cache CGO_CFLAGS="-DSQLITE_ENABLE_FTS5"
BUILD_TAGS := -tags "fts5"
BINARY     := mcp-codebase-search-go

build:
	$(CGO_FLAGS) go build $(BUILD_TAGS) -o $(BINARY) .

clean:
	rm -f $(BINARY)

test:
	$(CGO_FLAGS) go test $(BUILD_TAGS) ./... -v

vet:
	$(CGO_FLAGS) go vet $(BUILD_TAGS) ./...
