// Package indexer provides embedding generation for code chunks.
package indexer

import (
	"context"
	"fmt"
	"time"

	"mcp-codebase-explorer-go/search"
)

// Embedder generates dense vector embeddings for code chunks.
// Uses ONNX-first → OpenAI-fallback chain.
type Embedder struct {
	onnx *search.OnnxEmbedder // nil if ONNX is unavailable
}

// singleton embedder — created once per process lifetime.
var (
	globalEmbedder *Embedder
	embedOnce      func()
	embedDone      bool
)

func init() {
	embedOnce = func() {
		if embedDone {
			return
		}
		embedDone = true

		onnx, _ := search.NewOnnxEmbedder()
		globalEmbedder = &Embedder{onnx: onnx}
	}
}

// NewEmbedder returns the package-level singleton Embedder.
// Never returns nil — always returns a valid Embedder (may have onnx=nil for OpenAI-only mode).
// Thread-safe: the first call initialises; subsequent calls return the cached instance.
func NewEmbedder() (*Embedder, error) {
	embedOnce()
	return globalEmbedder, nil
}

// EmbedBatch embeds a slice of texts and returns [][]float32 in the same order.
// Uses ONNX locally if available, otherwise falls back to OpenAI API.
func (e *Embedder) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, nil
	}

	// Try ONNX first
	if e.onnx != nil {
		results, err := e.onnx.EmbedBatch(ctx, texts)
		if err == nil {
			return results, nil
		}
		// ONNX failed — fall through to OpenAI
	}

	// Fallback to OpenAI
	results, err := search.OpenAIFallbackEmbed(ctx, texts)
	if err != nil {
		return nil, err
	}
	if results == nil {
		// No OpenAI key either — return nil (BM25-only mode)
		return nil, nil
	}

	return results, nil
}

// EmbedSingle embeds a single text. Convenience wrapper over EmbedBatch.
func (e *Embedder) EmbedSingle(ctx context.Context, text string) ([]float32, error) {
	results, err := e.EmbedBatch(ctx, []string{text})
	if err != nil || results == nil || len(results) == 0 {
		return nil, err
	}
	return results[0], nil
}

// BuildChunkText creates the text that will be embedded for a CodeChunk.
// Combines symbol context + content for richer semantic signal.
func BuildChunkText(chunk CodeChunk) string {
	header := fmt.Sprintf("File: %s | Symbol: %s (%s) | Lines %d-%d\n\n",
		chunk.RelPath, chunk.SymbolName, chunk.SymbolKind, chunk.LineStart, chunk.LineEnd)
	return header + chunk.Content
}

// SleepBetweenBatches is kept for backward compatibility but is a no-op for ONNX.
func SleepBetweenBatches() {
	time.Sleep(100 * time.Millisecond)
}
