package main

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// SkillMetadata holds parsed YAML frontmatter and file info.
type SkillMetadata struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Tags        string `json:"tags"`
	Path        string `json:"path"`
	Preview     string `json:"preview"`
	Hash        string `json:"hash"` // md5 of entire SKILL.md (for backward compat)
}

// SkillDoc is a full skill with embedding — one per SKILL.md file.
type SkillDoc struct {
	ID        string        `json:"id"`
	Text      string        `json:"text"` // concatenated sections for embedding
	Metadata  SkillMetadata `json:"metadata"`
	Embedding []float32     `json:"embedding"`
	// Section-level data (populated by section chunker)
	Sections []SkillSection `json:"sections,omitempty"`
}

// SkillSection represents one logical section of a SKILL.md file.
type SkillSection struct {
	SkillName string    `json:"skill_name"`
	Section   string    `json:"section"` // "description" | "usage" | "steps" | "examples" | "body"
	Content   string    `json:"content"`
	Hash      string    `json:"hash"` // sha256 of content (for section-level Merkle diff)
	Embedding []float32 `json:"embedding,omitempty"`
}

var frontmatterRE = regexp.MustCompile(`(?s)^---\n(.*?)\n---`)

// parseSkillFile parses a SKILL.md into a SkillDoc with section chunks.
func parseSkillFile(path string) (*SkillDoc, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	contentStr := string(content)

	// File-level md5 hash (kept for backward compat with existing cache)
	hashBytes := md5.Sum(content)
	fileHash := hex.EncodeToString(hashBytes[:])

	// Parse YAML frontmatter
	var meta map[string]any
	match := frontmatterRE.FindStringSubmatch(contentStr)
	if len(match) > 1 {
		_ = yaml.Unmarshal([]byte(match[1]), &meta)
	}

	desc := ""
	if d, ok := meta["description"].(string); ok {
		desc = d
	}

	tagsStr := ""
	if t, ok := meta["tags"]; ok {
		switch v := t.(type) {
		case []interface{}:
			var ts []string
			for _, item := range v {
				ts = append(ts, fmt.Sprintf("%v", item))
			}
			tagsStr = strings.Join(ts, ", ")
		case string:
			tagsStr = v
		}
	}

	// Strip frontmatter for body
	body := strings.TrimSpace(frontmatterRE.ReplaceAllString(contentStr, ""))

	// Build preview (first 250 chars of body)
	words := strings.Fields(body)
	preview := strings.Join(words, " ")
	if len(preview) > 250 {
		preview = preview[:250] + "..."
	}

	skillName := filepath.Base(filepath.Dir(path))

	// Full search text for embedding
	descClean := strings.ReplaceAll(desc, "\n", " ")
	if len(descClean) > 150 {
		descClean = descClean[:150]
	}
	searchText := fmt.Sprintf("Skill: %s\nTags: %s\nDescription: %s\n\nPreview: %s",
		skillName, tagsStr, desc, preview)

	// Section chunking
	sections := extractSections(skillName, desc, tagsStr, body)

	return &SkillDoc{
		ID:   skillName,
		Text: searchText,
		Metadata: SkillMetadata{
			Name:        skillName,
			Description: descClean,
			Tags:        tagsStr,
			Path:        path,
			Preview:     preview,
			Hash:        fileHash,
		},
		Sections: sections,
	}, nil
}

// extractSections splits a SKILL.md body into logical sections.
// Sections: "description" (from frontmatter), then H2/H3 headers as named sections,
// with a catch-all "body" section for any remaining content.
func extractSections(skillName, desc, tags, body string) []SkillSection {
	var sections []SkillSection

	// Always include description as its own section
	if desc != "" {
		sections = append(sections, makeSection(skillName, "description", desc))
	}

	// Split body by markdown H2/H3 headers
	headerRE := regexp.MustCompile(`(?m)^#{2,3} .+`)
	locs := headerRE.FindAllStringIndex(body, -1)

	if len(locs) == 0 {
		// No headers — treat whole body as "body" section
		if body != "" {
			sections = append(sections, makeSection(skillName, "body", body))
		}
		return sections
	}

	// Content before first header
	if locs[0][0] > 0 {
		preamble := strings.TrimSpace(body[:locs[0][0]])
		if preamble != "" {
			sections = append(sections, makeSection(skillName, "body", preamble))
		}
	}

	for i, loc := range locs {
		headerLine := body[loc[0]:loc[1]]
		sectionName := normalizeSectionName(headerLine)

		var sectionBody string
		if i+1 < len(locs) {
			sectionBody = strings.TrimSpace(body[loc[1]:locs[i+1][0]])
		} else {
			sectionBody = strings.TrimSpace(body[loc[1]:])
		}

		combined := headerLine + "\n" + sectionBody
		if combined != "" {
			sections = append(sections, makeSection(skillName, sectionName, combined))
		}
	}

	return sections
}

// normalizeSectionName converts a markdown header into a lowercase slug.
func normalizeSectionName(header string) string {
	// Strip leading #s and spaces
	text := strings.TrimLeft(header, "# ")
	text = strings.ToLower(text)
	text = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(text, "_")
	text = strings.Trim(text, "_")
	if text == "" {
		return "section"
	}
	// Normalise common section names to canonical forms
	switch {
	case strings.Contains(text, "usage") || strings.Contains(text, "use_when"):
		return "usage"
	case strings.Contains(text, "example"):
		return "examples"
	case strings.Contains(text, "step"):
		return "steps"
	case strings.Contains(text, "prereq"):
		return "prerequisites"
	default:
		if len(text) > 40 {
			text = text[:40]
		}
		return text
	}
}

// makeSection creates a SkillSection with sha256 content hash.
func makeSection(skillName, sectionName, content string) SkillSection {
	h := sha256.Sum256([]byte(content))
	return SkillSection{
		SkillName: skillName,
		Section:   sectionName,
		Content:   content,
		Hash:      hex.EncodeToString(h[:]),
	}
}
