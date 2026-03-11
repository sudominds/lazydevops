package application

import (
	"strings"
	"testing"
)

func TestHighlightLogSeverityKeywords(t *testing.T) {
	line := "WARNING: one failed step with an ERROR"

	highlightedLine := highlightLogSeverityKeywords(line)

	if !strings.Contains(highlightedLine, "\x1b[33mWARNING\x1b[0m") {
		t.Fatalf("expected warning to be highlighted yellow, got %q", highlightedLine)
	}
	if !strings.Contains(highlightedLine, "\x1b[31mfailed\x1b[0m") {
		t.Fatalf("expected failed to be highlighted red, got %q", highlightedLine)
	}
	if !strings.Contains(highlightedLine, "\x1b[31mERROR\x1b[0m") {
		t.Fatalf("expected error to be highlighted red, got %q", highlightedLine)
	}
}
