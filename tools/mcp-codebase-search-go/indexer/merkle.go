// Package indexer provides Merkle-tree-based incremental file diff detection.
package indexer

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strings"
)

// MerkleTree tracks sha256 hashes of each file for incremental diff.
// The root hash represents the entire indexed state of the codebase.
type MerkleTree struct {
	// FileHashes maps relative file path → sha256 of file content
	FileHashes map[string]string
}

// NewMerkleTree initialises an empty Merkle tree.
func NewMerkleTree() *MerkleTree {
	return &MerkleTree{FileHashes: make(map[string]string)}
}

// Set records (or updates) the hash for a file.
func (m *MerkleTree) Set(relPath, hash string) {
	m.FileHashes[relPath] = hash
}

// Root computes the deterministic root hash of all tracked files.
// Sorted by path for determinism.
func (m *MerkleTree) Root() string {
	keys := make([]string, 0, len(m.FileHashes))
	for k := range m.FileHashes {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var sb strings.Builder
	for _, k := range keys {
		sb.WriteString(k)
		sb.WriteString(":")
		sb.WriteString(m.FileHashes[k])
		sb.WriteString("\n")
	}
	h := sha256.Sum256([]byte(sb.String()))
	return hex.EncodeToString(h[:])
}

// Diff compares fresh file entries against the current tree.
// Returns slices of added, changed, and removed relative paths.
func (m *MerkleTree) Diff(fresh []FileEntry) (added, changed, removed []string) {
	freshMap := make(map[string]string, len(fresh))
	for _, f := range fresh {
		freshMap[f.RelPath] = hashFile(f.AbsPath)
	}

	// Detect added and changed
	for rel, hash := range freshMap {
		existing, exists := m.FileHashes[rel]
		if !exists {
			added = append(added, rel)
		} else if existing != hash {
			changed = append(changed, rel)
		}
	}

	// Detect removed
	for rel := range m.FileHashes {
		if _, ok := freshMap[rel]; !ok {
			removed = append(removed, rel)
		}
	}

	return added, changed, removed
}

// Apply updates the tree with the result of a Diff.
func (m *MerkleTree) Apply(fresh []FileEntry) {
	for _, f := range fresh {
		m.FileHashes[f.RelPath] = hashFile(f.AbsPath)
	}
}

// Remove deletes a file entry from the tree.
func (m *MerkleTree) Remove(relPath string) {
	delete(m.FileHashes, relPath)
}

// hashFile returns the sha256 of a file's content.
// Returns empty string on read error (file will be treated as "new" next time).
func hashFile(absPath string) string {
	data, err := readFileSafe(absPath)
	if err != nil {
		return ""
	}
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}
