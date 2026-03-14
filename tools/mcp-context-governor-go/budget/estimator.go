// Package budget provides token cost estimation for files and text.
package budget

import (
	"os"
)

// EstimateTokens returns a heuristic token count for a text string.
// Rule of thumb: 1 token ≈ 4 bytes for English/code content.
func EstimateTokens(text string) int {
	return max1(len(text)/4, 1)
}

// EstimateFileTokens reads a file and estimates its token count.
// Returns 0 and no error if the file cannot be read.
func EstimateFileTokens(absPath string) int {
	data, err := os.ReadFile(absPath)
	if err != nil {
		return 0
	}
	return max1(len(data)/4, 1)
}

// EstimateContextLoad estimates the token cost of loading a list of files.
func EstimateContextLoad(files []string) int {
	total := 0
	for _, f := range files {
		total += EstimateFileTokens(f)
	}
	return total
}

func max1(a, b int) int {
	if a > b {
		return a
	}
	return b
}
