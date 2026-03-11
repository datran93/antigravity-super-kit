package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// ── YAML / Skill parsing ───────────────────────────────────────────────────────

func TestParseSkillFile_Basic(t *testing.T) {
	tmpDir := t.TempDir()
	skillDir := filepath.Join(tmpDir, "my-test-skill")
	_ = os.MkdirAll(skillDir, 0755)

	skillContent := `---
name: my-test-skill
description: A test skill for unit testing purposes.
tags:
  - testing
  - go
---

# My Test Skill

This skill is for testing.
`
	skillPath := filepath.Join(skillDir, "SKILL.md")
	if err := os.WriteFile(skillPath, []byte(skillContent), 0644); err != nil {
		t.Fatalf("failed to write skill: %v", err)
	}

	doc, err := parseSkillFile(skillPath)
	if err != nil {
		t.Fatalf("parseSkillFile failed: %v", err)
	}

	if doc.ID != "my-test-skill" {
		t.Errorf("Expected ID 'my-test-skill', got '%s'", doc.ID)
	}
	if doc.Metadata.Name != "my-test-skill" {
		t.Errorf("Expected Name 'my-test-skill', got '%s'", doc.Metadata.Name)
	}
	if doc.Metadata.Description == "" {
		t.Errorf("Expected non-empty description")
	}
	if doc.Metadata.Tags == "" {
		t.Errorf("Expected non-empty tags")
	}
	if doc.Metadata.Hash == "" {
		t.Errorf("Expected non-empty hash")
	}
	if doc.Metadata.Preview == "" {
		t.Errorf("Expected non-empty preview")
	}
}

func TestParseSkillFile_NoFrontmatter(t *testing.T) {
	tmpDir := t.TempDir()
	skillDir := filepath.Join(tmpDir, "no-frontmatter-skill")
	_ = os.MkdirAll(skillDir, 0755)

	skillContent := `# No Frontmatter Skill

This skill has no YAML frontmatter at all.
It should still parse without errors.
`
	skillPath := filepath.Join(skillDir, "SKILL.md")
	_ = os.WriteFile(skillPath, []byte(skillContent), 0644)

	doc, err := parseSkillFile(skillPath)
	if err != nil {
		t.Fatalf("parseSkillFile with no frontmatter failed: %v", err)
	}
	if doc.ID != "no-frontmatter-skill" {
		t.Errorf("Expected ID 'no-frontmatter-skill', got '%s'", doc.ID)
	}
}

func TestParseSkillFile_StringTags(t *testing.T) {
	tmpDir := t.TempDir()
	skillDir := filepath.Join(tmpDir, "string-tags-skill")
	_ = os.MkdirAll(skillDir, 0755)

	skillContent := `---
description: Skill with string tags
tags: frontend, react, typescript
---

# String Tags Skill
Content here.
`
	skillPath := filepath.Join(skillDir, "SKILL.md")
	_ = os.WriteFile(skillPath, []byte(skillContent), 0644)

	doc, err := parseSkillFile(skillPath)
	if err != nil {
		t.Fatalf("parseSkillFile failed: %v", err)
	}
	if doc.Metadata.Tags == "" {
		t.Errorf("Expected non-empty tags for string-formatted tags")
	}
}

func TestParseSkillFile_LongDescription(t *testing.T) {
	tmpDir := t.TempDir()
	skillDir := filepath.Join(tmpDir, "long-desc-skill")
	_ = os.MkdirAll(skillDir, 0755)

	// Description longer than 150 chars
	longDesc := "This is a very long description that exceeds the 150-character limit set in the parseSkillFile function, it should be truncated at 150 characters."
	skillContent := "---\ndescription: " + longDesc + "\n---\n\nContent\n"
	skillPath := filepath.Join(skillDir, "SKILL.md")
	_ = os.WriteFile(skillPath, []byte(skillContent), 0644)

	doc, err := parseSkillFile(skillPath)
	if err != nil {
		t.Fatalf("parseSkillFile failed: %v", err)
	}
	if len(doc.Metadata.Description) > 150 {
		t.Errorf("Expected description to be truncated to 150 chars, got %d", len(doc.Metadata.Description))
	}
}

func TestParseSkillFile_NonExistent(t *testing.T) {
	_, err := parseSkillFile("/nonexistent/SKILL.md")
	if err == nil {
		t.Errorf("Expected error for non-existent skill file, got nil")
	}
}

// ── Cosine Similarity ──────────────────────────────────────────────────────────

func TestCosineSimilarity_Identical(t *testing.T) {
	vec := []float32{1.0, 0.0, 0.0}
	score := cosineSimilarity(vec, vec)
	if score < 0.9999 || score > 1.0001 {
		t.Errorf("Expected cosine similarity of identical vectors to be ~1.0, got %f", score)
	}
}

func TestCosineSimilarity_Orthogonal(t *testing.T) {
	a := []float32{1.0, 0.0}
	b := []float32{0.0, 1.0}
	score := cosineSimilarity(a, b)
	if score > 0.0001 {
		t.Errorf("Expected cosine similarity of orthogonal vectors to be ~0.0, got %f", score)
	}
}

func TestCosineSimilarity_ZeroVector(t *testing.T) {
	a := []float32{0.0, 0.0}
	b := []float32{1.0, 1.0}
	score := cosineSimilarity(a, b)
	if score != 0.0 {
		t.Errorf("Expected 0.0 for zero vector, got %f", score)
	}
}

func TestCosineSimilarity_DifferentLengths(t *testing.T) {
	a := []float32{1.0, 0.0, 0.0}
	b := []float32{1.0, 0.0}
	// Should compute similarity only up to the shorter length
	score := cosineSimilarity(a, b)
	if score < 0.9999 || score > 1.0001 {
		t.Errorf("Expected similarity ~1.0 for truncated identical vectors, got %f", score)
	}
}

// ── Cache ──────────────────────────────────────────────────────────────────────

func TestCacheSaveLoad(t *testing.T) {
	tmpDir := t.TempDir()
	origDbDir, origDbFile := dbDir, dbFile
	dbDir = tmpDir
	dbFile = filepath.Join(tmpDir, "test_cache.json")
	defer func() {
		dbDir = origDbDir
		dbFile = origDbFile
	}()

	testDoc := SkillDoc{
		ID:   "test-skill",
		Text: "Test skill content",
		Metadata: SkillMetadata{
			Name:        "test-skill",
			Description: "A test skill",
			Tags:        "testing",
			Path:        "/tmp/test-skill/SKILL.md",
			Preview:     "Test skill content",
			Hash:        "abc123",
		},
		Embedding: []float32{0.1, 0.2, 0.3},
	}

	cache := map[string]SkillDoc{"test-skill": testDoc}
	if err := saveCache(cache); err != nil {
		t.Fatalf("saveCache failed: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		t.Fatalf("cache file was not created")
	}

	// Load it back
	loaded, err := loadCache()
	if err != nil {
		t.Fatalf("loadCache failed: %v", err)
	}

	if len(loaded) != 1 {
		t.Fatalf("Expected 1 item in cache, got %d", len(loaded))
	}
	doc, ok := loaded["test-skill"]
	if !ok {
		t.Fatalf("Expected 'test-skill' in cache")
	}
	if doc.Metadata.Description != "A test skill" {
		t.Errorf("Expected description 'A test skill', got '%s'", doc.Metadata.Description)
	}
}

func TestLoadCache_NonExistent(t *testing.T) {
	origDbFile := dbFile
	dbFile = "/nonexistent/path/cache.json"
	defer func() { dbFile = origDbFile }()

	cache, err := loadCache()
	if err != nil {
		t.Fatalf("Expected no error for non-existent cache file, got: %v", err)
	}
	if len(cache) != 0 {
		t.Errorf("Expected empty cache for non-existent file, got %d items", len(cache))
	}
}

func TestLoadCache_CorruptedFile(t *testing.T) {
	tmpDir := t.TempDir()
	origDbFile := dbFile
	dbFile = filepath.Join(tmpDir, "corrupt.json")
	defer func() { dbFile = origDbFile }()

	_ = os.WriteFile(dbFile, []byte("this is not valid json!!!"), 0644)

	_, err := loadCache()
	if err == nil {
		t.Error("Expected error for corrupted cache file")
	}
}

// ── SkillDoc JSON roundtrip ────────────────────────────────────────────────────

func TestSkillDocJSON_Roundtrip(t *testing.T) {
	doc := SkillDoc{
		ID:   "roundtrip-skill",
		Text: "Some content",
		Metadata: SkillMetadata{
			Name:        "roundtrip-skill",
			Description: "A roundtrip test",
			Tags:        "go, testing",
			Path:        "/tmp/roundtrip/SKILL.md",
			Hash:        "def456",
		},
		Embedding: []float32{0.5, 0.5},
	}

	data, err := json.Marshal(doc)
	if err != nil {
		t.Fatalf("Failed to marshal SkillDoc: %v", err)
	}

	var out SkillDoc
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("Failed to unmarshal SkillDoc: %v", err)
	}

	if out.ID != doc.ID {
		t.Errorf("Expected ID %q, got %q", doc.ID, out.ID)
	}
	if out.Metadata.Tags != doc.Metadata.Tags {
		t.Errorf("Expected Tags %q, got %q", doc.Metadata.Tags, out.Metadata.Tags)
	}
}
