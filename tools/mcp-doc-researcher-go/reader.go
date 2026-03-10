package main

import (
	"fmt"
	"os/exec"
	"strings"

	"code.sajari.com/docconv"
)

// ReadLocalDoc Extracts text from a local .doc, .docx, or .pdf file.
// For PDFs, uses pdftotext (poppler) for accurate Unicode extraction.
// For .doc/.docx, uses docconv (supports OLE2 and OOXML formats).
func ReadLocalDoc(filePath string) (string, error) {
	lower := strings.ToLower(filePath)
	if strings.HasSuffix(lower, ".pdf") {
		return readPdfWithPoppler(filePath)
	}

	res, err := docconv.ConvertPath(filePath)
	if err != nil {
		return "", fmt.Errorf("error reading %s: %v", filePath, err)
	}
	return strings.TrimSpace(res.Body), nil
}

// readPdfWithPoppler uses pdftotext (poppler) to extract text from a PDF.
// This correctly handles custom-embedded fonts, CIDFonts, and Unicode PDFs
// that the pure-Go ledongthuc/pdf library cannot decode.
func readPdfWithPoppler(path string) (string, error) {
	pdfToText, err := exec.LookPath("pdftotext")
	if err != nil {
		return "", fmt.Errorf("pdftotext (poppler) is not installed. Install it with: brew install poppler (macOS) or apt install poppler-utils (Linux)")
	}

	// pdftotext -layout path - outputs text to stdout
	cmd := exec.Command(pdfToText, "-layout", path, "-")
	out, err := cmd.Output()
	if err != nil {
		// Try without -layout flag as a fallback
		cmd2 := exec.Command(pdfToText, path, "-")
		out2, err2 := cmd2.Output()
		if err2 != nil {
			return "", fmt.Errorf("pdftotext failed on %s: %v", path, err2)
		}
		out = out2
	}

	text := strings.TrimSpace(string(out))
	if text == "" {
		return "", fmt.Errorf("pdftotext extracted no text from %s (possibly a scanned/image-only PDF)", path)
	}
	return text, nil
}
