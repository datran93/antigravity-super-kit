// Package search implements Reciprocal Rank Fusion (RRF) hybrid search
// over BM25 (FTS5) and dense vector (cosine similarity) results.
package search

import (
	"math"
	"sort"
)

const rrfK = 60.0 // standard RRF constant

// ChunkResult is one ranked search result.
type ChunkResult struct {
	ID         string
	RelPath    string
	SymbolName string
	SymbolKind string
	Content    string
	LineStart  int
	LineEnd    int
	RRFScore   float64
	BM25Rank   int // -1 if not in BM25 results
	VecRank    int // -1 if not in vector results
}

// BM25Input is one FTS result with its rank position.
type BM25Input struct {
	ID         string
	RelPath    string
	SymbolName string
	BM25Rank   int // 0 = best
}

// VecInput is one cosine-similarity result with score.
type VecInput struct {
	ID         string
	RelPath    string
	SymbolName string
	Score      float32 // higher = more similar
}

// RRFFuse merges BM25 and vector results using Reciprocal Rank Fusion.
// bm25Results and vecResults are already sorted (best first).
// topK: how many final results to return.
func RRFFuse(bm25Results []BM25Input, vecResults []VecInput, topK int) []ChunkResult {
	type entry struct {
		id         string
		relPath    string
		symbolName string
		bm25Rank   int
		vecRank    int
		score      float64
	}

	scores := make(map[string]*entry)

	// Accumulate BM25 scores
	for i, r := range bm25Results {
		e, ok := scores[r.ID]
		if !ok {
			e = &entry{id: r.ID, relPath: r.RelPath, symbolName: r.SymbolName, bm25Rank: -1, vecRank: -1}
			scores[r.ID] = e
		}
		e.bm25Rank = i
		e.score += 1.0 / (rrfK + float64(i+1))
	}

	// Accumulate vector scores
	for i, r := range vecResults {
		e, ok := scores[r.ID]
		if !ok {
			e = &entry{id: r.ID, relPath: r.RelPath, symbolName: r.SymbolName, bm25Rank: -1, vecRank: -1}
			scores[r.ID] = e
		}
		e.vecRank = i
		e.score += 1.0 / (rrfK + float64(i+1))
	}

	// Sort by RRF score descending
	ranked := make([]*entry, 0, len(scores))
	for _, e := range scores {
		ranked = append(ranked, e)
	}
	sort.Slice(ranked, func(i, j int) bool {
		return ranked[i].score > ranked[j].score
	})

	if topK > 0 && len(ranked) > topK {
		ranked = ranked[:topK]
	}

	results := make([]ChunkResult, len(ranked))
	for i, e := range ranked {
		results[i] = ChunkResult{
			ID:         e.id,
			RelPath:    e.relPath,
			SymbolName: e.symbolName,
			RRFScore:   e.score,
			BM25Rank:   e.bm25Rank,
			VecRank:    e.vecRank,
		}
	}
	return results
}

// CosineSimilarity computes cosine similarity between two float32 vectors.
func CosineSimilarity(a, b []float32) float32 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	var dot, normA, normB float32
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	for i := 0; i < n; i++ {
		dot += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}
	denom := float32(math.Sqrt(float64(normA))) * float32(math.Sqrt(float64(normB)))
	if denom == 0 {
		return 0
	}
	return dot / denom
}
