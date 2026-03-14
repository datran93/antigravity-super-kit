package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	openai "github.com/sashabaranov/go-openai"
)

// ensureIndex loads the skill cache, detects new/modified skills via section-level
// Merkle hashes, re-embeds only changed sections, and persists the updated cache.
//
// Section-level Merkle diff: each section has a sha256(content) hash stored in
// SkillSection.Hash. If the hash matches cache, the embedding is reused. Only
// changed sections (and by extension their parent SkillDoc.Embedding) are re-embedded.
func ensureIndex() (map[string]SkillDoc, error) {
	mu.Lock()
	defer mu.Unlock()

	client, err := getOpenAIClient()
	if err != nil {
		return nil, err
	}

	cache, err := loadCache()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading cache: %v\n", err)
		cache = make(map[string]SkillDoc)
	}

	skillDirs, err := os.ReadDir(skillsDir)
	if err != nil {
		return nil, fmt.Errorf("error reading skills dir: %v", err)
	}

	currentIDs := make(map[string]bool)
	// Texts to embed: map from (skillID, sectionIdx) → text
	type embedJob struct {
		skillID    string
		sectionIdx int
		isFullDoc  bool // true → embed the whole doc text (for backward compat),
		//                  false → embed individual section
		text string
	}
	var jobs []embedJob

	// Per-skill parsed docs (fresh from disk)
	freshDocs := make(map[string]*SkillDoc)

	for _, dir := range skillDirs {
		if !dir.IsDir() {
			continue
		}
		skillPath := filepath.Join(skillsDir, dir.Name(), "SKILL.md")
		if _, err := os.Stat(skillPath); os.IsNotExist(err) {
			continue
		}
		doc, err := parseSkillFile(skillPath)
		if err != nil {
			continue
		}
		currentIDs[doc.ID] = true
		freshDocs[doc.ID] = doc

		cached, exists := cache[doc.ID]

		// ── Merkle section-level diff ──────────────────────────────────────
		// Merge cached section embeddings into fresh doc sections
		cachedSectionEmbs := make(map[string][]float32) // key: section hash
		if exists {
			for _, cs := range cached.Sections {
				if len(cs.Embedding) > 0 {
					cachedSectionEmbs[cs.Hash] = cs.Embedding
				}
			}
		}

		needsDocEmbed := false
		for i := range doc.Sections {
			sec := &doc.Sections[i]
			if emb, ok := cachedSectionEmbs[sec.Hash]; ok {
				sec.Embedding = emb // reuse cached embedding for unchanged section
			} else {
				// Section is new or changed → schedule re-embed
				jobs = append(jobs, embedJob{doc.ID, i, false, sec.Content})
				needsDocEmbed = true
			}
		}

		// Full-doc embedding: reuse if file hash unchanged, else re-embed
		if exists && cached.Metadata.Hash == doc.Metadata.Hash && len(cached.Embedding) > 0 && !needsDocEmbed {
			doc.Embedding = cached.Embedding
		} else {
			jobs = append(jobs, embedJob{doc.ID, -1, true, doc.Text})
		}
	}

	// Remove stale entries
	changed := false
	for id := range cache {
		if !currentIDs[id] {
			delete(cache, id)
			changed = true
		}
	}

	if len(jobs) > 0 {
		fmt.Fprintf(os.Stderr, "Indexing %d embedding jobs (section-level Merkle diff)...\n", len(jobs))

		texts := make([]string, len(jobs))
		for i, j := range jobs {
			texts[i] = j.text
		}

		batchSize := 20
		allEmbs := make([][]float32, len(jobs))

		for i := 0; i < len(texts); i += batchSize {
			end := i + batchSize
			if end > len(texts) {
				end = len(texts)
			}
			resp, err := client.CreateEmbeddings(context.Background(), openai.EmbeddingRequest{
				Input: texts[i:end],
				Model: openai.SmallEmbedding3,
			})
			if err != nil {
				// Log and continue rather than aborting the entire index — partial re-embed
				// is preferred over losing all cached data on a single rate-limit or timeout.
				fmt.Fprintf(os.Stderr, "Warning: embedding batch %d/%d failed: %v — skipping this batch\n",
					i/batchSize+1, (len(texts)+batchSize-1)/batchSize, err)
				// Leave allEmbs[i:end] as nil; those embeddings will be retried next run
				time.Sleep(500 * time.Millisecond) // back-off before continuing
				continue
			}
			for j, emb := range resp.Data {
				allEmbs[i+j] = emb.Embedding
			}
			time.Sleep(100 * time.Millisecond)
		}

		// Apply results back
		for i, job := range jobs {
			doc := freshDocs[job.skillID]
			if job.isFullDoc {
				doc.Embedding = allEmbs[i]
			} else {
				doc.Sections[job.sectionIdx].Embedding = allEmbs[i]
			}
		}
		changed = true
	}

	// Commit fresh docs to cache
	for id, doc := range freshDocs {
		cache[id] = *doc
	}

	if changed {
		saveCache(cache)
	}

	return cache, nil
}
