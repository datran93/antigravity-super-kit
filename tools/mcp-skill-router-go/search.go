package main

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

// SearchResult holds a ranked skill match.
type SearchResult struct {
	Doc   SkillDoc
	Score float32
}

// cosineSimilarity computes cosine similarity between two float32 vectors.
func cosineSimilarity(a, b []float32) float32 {
	var dotProduct, normA, normB float32
	for i := 0; i < len(a) && i < len(b); i++ {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dotProduct / (float32(math.Sqrt(float64(normA))) * float32(math.Sqrt(float64(normB))))
}

// searchSkills performs semantic search over indexed skills.
// Scores each skill using the full-doc embedding (cosine similarity).
// Section embeddings are stored for future fine-grained retrieval.
func searchSkills(ctx context.Context, query, tagsFilter string, topK int) ([]SearchResult, error) {
	client, err := getOpenAIClient()
	if err != nil {
		return nil, err
	}

	cache, err := ensureIndex()
	if err != nil {
		return nil, fmt.Errorf("error initializing index: %v", err)
	}
	if len(cache) == 0 {
		return nil, nil
	}

	resp, err := client.CreateEmbeddings(ctx, openai.EmbeddingRequest{
		Input: []string{query},
		Model: openai.SmallEmbedding3,
	})
	if err != nil {
		return nil, fmt.Errorf("error embedding query: %v", err)
	}
	queryEmb := resp.Data[0].Embedding

	// Parse tag filters
	var filterTags []string
	if tagsFilter != "" {
		for _, p := range strings.Split(tagsFilter, ",") {
			f := strings.TrimSpace(strings.ToLower(p))
			if f != "" {
				filterTags = append(filterTags, f)
			}
		}
	}

	var results []SearchResult
	for _, doc := range cache {
		if len(filterTags) > 0 {
			skillTags := strings.ToLower(doc.Metadata.Tags)
			valid := true
			for _, ft := range filterTags {
				if !strings.Contains(skillTags, ft) {
					valid = false
					break
				}
			}
			if !valid {
				continue
			}
		}
		if len(doc.Embedding) == 0 {
			continue
		}
		score := cosineSimilarity(queryEmb, doc.Embedding)
		results = append(results, SearchResult{Doc: doc, Score: score})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})
	if len(results) > topK {
		results = results[:topK]
	}
	return results, nil
}

// formatSearchResults formats search results for the MCP tool response.
func formatSearchResults(query string, results []SearchResult) string {
	if len(results) == 0 {
		return "❌ No relevant skills found."
	}
	out := []string{fmt.Sprintf("🎯 SEMANTIC SEARCH RESULTS FOR QUERY: '%s'", query)}
	for _, r := range results {
		res := fmt.Sprintf("\n🔹 **%s**", r.Doc.Metadata.Name)
		res += fmt.Sprintf("\n   - Path: `%s`", r.Doc.Metadata.Path)
		res += fmt.Sprintf("\n   - Tags: %s", r.Doc.Metadata.Tags)
		res += fmt.Sprintf("\n   - Description: %s", r.Doc.Metadata.Description)
		if r.Doc.Metadata.Preview != "" {
			res += fmt.Sprintf("\n   - Preview: %s", r.Doc.Metadata.Preview)
		}
		out = append(out, res)
	}
	out = append(out, "\n💡 ADVICE FOR AGENT: Use the `view_file` tool on the Path provided above to read the skill details before starting your task.")
	return strings.Join(out, "\n")
}
