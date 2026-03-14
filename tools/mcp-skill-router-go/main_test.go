package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ── parseSkillFile tests ───────────────────────────────────────────────────────

func TestParseSkillFile_Basic(t *testing.T) {
	dir := t.TempDir()
	skillDir := filepath.Join(dir, "my-skill")
	os.MkdirAll(skillDir, 0755)

	content := `---
name: My Skill
description: Does something useful.
tags: [go, backend]
---
# My Skill

## Usage
Use this when you need something.

## Steps
1. Do step one
2. Do step two
`
	path := filepath.Join(skillDir, "SKILL.md")
	os.WriteFile(path, []byte(content), 0644)

	doc, err := parseSkillFile(path)
	if err != nil {
		t.Fatalf("parseSkillFile failed: %v", err)
	}
	if doc.ID != "my-skill" {
		t.Errorf("expected ID 'my-skill', got '%s'", doc.ID)
	}
	if !strings.Contains(doc.Metadata.Description, "useful") {
		t.Errorf("description not parsed: %s", doc.Metadata.Description)
	}
	if doc.Metadata.Tags == "" {
		t.Error("tags should be parsed")
	}
	if doc.Metadata.Hash == "" {
		t.Error("hash should be set")
	}
}

func TestParseSkillFile_NoFrontmatter(t *testing.T) {
	dir := t.TempDir()
	skillDir := filepath.Join(dir, "bare-skill")
	os.MkdirAll(skillDir, 0755)

	path := filepath.Join(skillDir, "SKILL.md")
	os.WriteFile(path, []byte("# Bare skill\nJust some content."), 0644)

	doc, err := parseSkillFile(path)
	if err != nil {
		t.Fatalf("parseSkillFile failed: %v", err)
	}
	if doc.ID != "bare-skill" {
		t.Errorf("expected 'bare-skill', got '%s'", doc.ID)
	}
}

func TestParseSkillFile_NonExistent(t *testing.T) {
	_, err := parseSkillFile("/tmp/nonexistent/SKILL.md")
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

// ── extractSections tests ──────────────────────────────────────────────────────

func TestExtractSections_WithHeaders(t *testing.T) {
	body := `## Usage
Use this when you need it.

## Steps
1. Step one
2. Step two
`
	sections := extractSections("test-skill", "A description", "go", body)

	// Should have: description + usage + steps
	if len(sections) < 3 {
		t.Errorf("expected at least 3 sections, got %d", len(sections))
	}

	// Verify description section exists
	var found bool
	for _, s := range sections {
		if s.Section == "description" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected a 'description' section")
	}
}

func TestExtractSections_NoHeaders(t *testing.T) {
	body := "Just plain text without any headers."
	sections := extractSections("plain-skill", "desc", "", body)
	// Should have: description + body
	if len(sections) != 2 {
		t.Errorf("expected 2 sections (description + body), got %d: %+v", len(sections), sections)
	}
}

func TestExtractSections_HashDeterminism(t *testing.T) {
	body := "## Usage\nSame content."
	s1 := extractSections("skill", "desc", "", body)
	s2 := extractSections("skill", "desc", "", body)
	if len(s1) != len(s2) {
		t.Fatal("different number of sections for same input")
	}
	for i := range s1 {
		if s1[i].Hash != s2[i].Hash {
			t.Errorf("section %d hash not deterministic: %s vs %s", i, s1[i].Hash, s2[i].Hash)
		}
	}
}

func TestExtractSections_SectionHashChangesOnEdit(t *testing.T) {
	body1 := "## Usage\nOriginal content."
	body2 := "## Usage\nModified content."
	s1 := extractSections("skill", "desc", "", body1)
	s2 := extractSections("skill", "desc", "", body2)

	if len(s1) == 0 || len(s2) == 0 {
		t.Fatal("empty sections")
	}
	// Find the usage sections
	var h1, h2 string
	for _, s := range s1 {
		if s.Section == "usage" {
			h1 = s.Hash
		}
	}
	for _, s := range s2 {
		if s.Section == "usage" {
			h2 = s.Hash
		}
	}
	if h1 == "" || h2 == "" {
		t.Fatal("usage section not found")
	}
	if h1 == h2 {
		t.Error("hash should differ when content changes")
	}
}

// ── normalizeSectionName tests ─────────────────────────────────────────────────

func TestNormalizeSectionName(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"## Usage", "usage"},
		{"### Use When", "usage"},
		{"## Examples", "examples"},
		{"## Step-by-Step Guide", "steps"},
		{"## Prerequisites", "prerequisites"},
		{"## My Custom Section", "my_custom_section"},
	}
	for _, tc := range cases {
		got := normalizeSectionName(tc.input)
		if got != tc.expected {
			t.Errorf("normalizeSectionName(%q) = %q, want %q", tc.input, got, tc.expected)
		}
	}
}

// ── cosineSimilarity tests ─────────────────────────────────────────────────────

func TestCosineSimilarity_Identical(t *testing.T) {
	v := []float32{1, 0, 0, 0}
	if got := cosineSimilarity(v, v); got < 0.999 {
		t.Errorf("identical: expected ~1.0, got %f", got)
	}
}

func TestCosineSimilarity_Orthogonal(t *testing.T) {
	a := []float32{1, 0}
	b := []float32{0, 1}
	if got := cosineSimilarity(a, b); got != 0 {
		t.Errorf("orthogonal: expected 0, got %f", got)
	}
}

func TestCosineSimilarity_ZeroVector(t *testing.T) {
	a := []float32{0, 0}
	b := []float32{1, 1}
	if got := cosineSimilarity(a, b); got != 0 {
		t.Errorf("zero vector: expected 0, got %f", got)
	}
}

func TestCosineSimilarity_DifferentLengths(t *testing.T) {
	a := []float32{1, 2, 3}
	b := []float32{1, 2}
	// Should not panic; compute with shorter length
	_ = cosineSimilarity(a, b)
}

// ── cache tests ────────────────────────────────────────────────────────────────

func TestCacheSaveLoad(t *testing.T) {
	// Override dbDir/dbFile for test isolation
	tmpDir := t.TempDir()
	origDbDir := dbDir
	origDbFile := dbFile
	dbDir = tmpDir
	dbFile = filepath.Join(tmpDir, "skills_cache.json")
	defer func() {
		dbDir = origDbDir
		dbFile = origDbFile
	}()

	doc := SkillDoc{
		ID:   "test-skill",
		Text: "Test skill text",
		Metadata: SkillMetadata{
			Name:        "test-skill",
			Description: "A test skill",
			Hash:        "abc123",
		},
		Embedding: []float32{0.1, 0.2, 0.3},
	}

	cache := map[string]SkillDoc{"test-skill": doc}
	if err := saveCache(cache); err != nil {
		t.Fatalf("saveCache failed: %v", err)
	}

	loaded, err := loadCache()
	if err != nil {
		t.Fatalf("loadCache failed: %v", err)
	}
	if _, ok := loaded["test-skill"]; !ok {
		t.Error("test-skill not found in loaded cache")
	}
}

func TestLoadCache_NonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	origDbFile := dbFile
	dbFile = filepath.Join(tmpDir, "nonexistent.json")
	defer func() { dbFile = origDbFile }()

	cache, err := loadCache()
	if err != nil {
		t.Fatalf("expected no error for non-existent cache, got: %v", err)
	}
	if len(cache) != 0 {
		t.Error("expected empty cache")
	}
}
