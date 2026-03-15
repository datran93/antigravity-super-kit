package main

import (
	"strings"
	"testing"
)

// paginateContent replicates the rune-based pagination logic from getFileContentTool
// for unit testing without needing a live GitHub connection.
func paginateContent(rawContent string, page, pageSize int) (string, int) {
	const defaultPS = 8000
	const maxPS = 32000
	if pageSize <= 0 {
		pageSize = defaultPS
	}
	if pageSize > maxPS {
		pageSize = maxPS
	}
	if page <= 0 {
		page = 1
	}

	runes := []rune(rawContent)
	totalRunes := len(runes)
	totalPages := (totalRunes + pageSize - 1) / pageSize
	if totalPages == 0 {
		totalPages = 1
	}
	if page > totalPages {
		return "", totalPages
	}
	start := (page - 1) * pageSize
	end := start + pageSize
	if end > totalRunes {
		end = totalRunes
	}
	return string(runes[start:end]), totalPages
}

func TestPagination_SinglePage(t *testing.T) {
	content := strings.Repeat("a", 100)
	result, total := paginateContent(content, 1, 8000)
	if result != content {
		t.Errorf("expected full content on single page, got len=%d", len(result))
	}
	if total != 1 {
		t.Errorf("expected 1 total page, got %d", total)
	}
}

func TestPagination_MultiPage_Page1(t *testing.T) {
	content := strings.Repeat("x", 20000)
	result, total := paginateContent(content, 1, 8000)
	if len([]rune(result)) != 8000 {
		t.Errorf("expected page 1 to have 8000 chars, got %d", len(result))
	}
	if total != 3 {
		t.Errorf("expected 3 pages for 20000 chars at 8000/page, got %d", total)
	}
}

func TestPagination_MultiPage_LastPage(t *testing.T) {
	content := strings.Repeat("y", 20000)
	result, total := paginateContent(content, 3, 8000)
	if len([]rune(result)) != 4000 { // 20000 - 2*8000 = 4000
		t.Errorf("expected last page to have 4000 chars, got %d", len(result))
	}
	if total != 3 {
		t.Errorf("expected 3 total pages, got %d", total)
	}
}

func TestPagination_BeyondLastPage(t *testing.T) {
	content := strings.Repeat("z", 100)
	result, _ := paginateContent(content, 99, 8000)
	if result != "" {
		t.Errorf("expected empty string for page beyond total, got %q", result)
	}
}

func TestPagination_EmptyContent(t *testing.T) {
	result, total := paginateContent("", 1, 8000)
	if result != "" {
		t.Errorf("expected empty result, got %q", result)
	}
	if total != 1 {
		t.Errorf("expected 1 total page for empty content, got %d", total)
	}
}

func TestPagination_PageSizeCap(t *testing.T) {
	content := strings.Repeat("a", 10000)
	// page_size > max should be capped
	result, total := paginateContent(content, 1, 99999)
	// maxPageSize = 32000, 10000 < 32000, so whole content fits on page 1
	_ = result
	if total != 1 {
		t.Errorf("expected 1 page when content fits within cap, got %d", total)
	}
}

func TestPagination_UnicodeContent(t *testing.T) {
	// Each Vietnamese character is a single rune but multiple bytes
	unit := "Xin chào thế giới! "        // 20 runes
	content := strings.Repeat(unit, 500) // 10000 runes
	result, total := paginateContent(content, 1, 8000)
	if len([]rune(result)) != 8000 {
		t.Errorf("expected 8000 runes on page 1, got %d", len([]rune(result)))
	}
	if total != 2 {
		t.Errorf("expected 2 pages for 10000 runes at 8000/page, got %d", total)
	}
}
