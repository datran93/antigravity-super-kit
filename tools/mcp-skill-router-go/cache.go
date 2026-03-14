package main

import (
	"encoding/json"
	"fmt"
	"os"
)

// loadCache reads the JSON skill cache from disk.
// Returns an empty map (not error) if the file doesn't exist yet.
func loadCache() (map[string]SkillDoc, error) {
	cache := make(map[string]SkillDoc)
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return cache, nil
	}
	data, err := os.ReadFile(dbFile)
	if err != nil {
		return nil, err
	}
	var docs []SkillDoc
	if err := json.Unmarshal(data, &docs); err != nil {
		return nil, err
	}
	for _, d := range docs {
		cache[d.ID] = d
	}
	return cache, nil
}

// saveCache persists the in-memory skill index to the JSON file.
func saveCache(cache map[string]SkillDoc) error {
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return fmt.Errorf("failed to create db dir: %w", err)
	}
	var docs []SkillDoc
	for _, d := range cache {
		docs = append(docs, d)
	}
	data, err := json.MarshalIndent(docs, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(dbFile, data, 0644)
}
