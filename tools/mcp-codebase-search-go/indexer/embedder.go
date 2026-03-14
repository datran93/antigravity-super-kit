// Package indexer provides embedding generation for code chunks.
package indexer

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	openai "github.com/sashabaranov/go-openai"
)

const defaultEmbedModel = openai.SmallEmbedding3

// Embedder generates dense vector embeddings for code chunks.
type Embedder struct {
	client *openai.Client
	model  openai.EmbeddingModel
}

// singleton embedder — created once per process lifetime.
var (
	globalEmbedder *Embedder
	embedOnce      sync.Once
)

// NewEmbedder returns the package-level singleton Embedder.
// Returns nil (with no error) if OPENAI_API_KEY is absent — caller must check.
// Thread-safe: the first call initialises; subsequent calls return the cached instance.
func NewEmbedder() (*Embedder, error) {
	embedOnce.Do(func() {
		apiKey := os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			return // leave globalEmbedder nil
		}
		model := openai.EmbeddingModel(os.Getenv("EMBEDDING_MODEL"))
		if model == "" {
			model = defaultEmbedModel
		}
		globalEmbedder = &Embedder{
			client: openai.NewClient(apiKey),
			model:  model,
		}
	})
	return globalEmbedder, nil
}

// EmbedBatch embeds a slice of texts and returns [][]float32 in the same order.
// Calls are batched (20 per request) with 100ms inter-batch sleep to respect rate limits.
func (e *Embedder) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, nil
	}

	results := make([][]float32, len(texts))
	batchSize := 20

	for i := 0; i < len(texts); i += batchSize {
		end := i + batchSize
		if end > len(texts) {
			end = len(texts)
		}
		batch := texts[i:end]

		resp, err := e.client.CreateEmbeddings(ctx, openai.EmbeddingRequest{
			Input: batch,
			Model: e.model,
		})
		if err != nil {
			return nil, fmt.Errorf("embedding API error at batch %d: %w", i/batchSize, err)
		}
		for j, d := range resp.Data {
			results[i+j] = d.Embedding
		}

		if end < len(texts) {
			time.Sleep(100 * time.Millisecond)
		}
	}

	return results, nil
}

// BuildChunkText creates the text that will be embedded for a CodeChunk.
// Combines symbol context + content for richer semantic signal.
func BuildChunkText(chunk CodeChunk) string {
	header := fmt.Sprintf("File: %s | Symbol: %s (%s) | Lines %d-%d\n\n",
		chunk.RelPath, chunk.SymbolName, chunk.SymbolKind, chunk.LineStart, chunk.LineEnd)
	return header + chunk.Content
}
