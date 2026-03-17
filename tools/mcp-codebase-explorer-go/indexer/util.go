package indexer

import "os"

// readFileSafe reads a file and returns its content, or error.
func readFileSafe(absPath string) ([]byte, error) {
	return os.ReadFile(absPath)
}
