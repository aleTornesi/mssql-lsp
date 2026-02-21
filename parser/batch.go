package parser

import (
	"regexp"
	"strings"
)

// goSeparator matches a GO batch separator line (case-insensitive, standalone).
var goSeparator = regexp.MustCompile(`(?im)^\s*GO\s*$`)

// Batch represents a single T-SQL batch separated by GO.
type Batch struct {
	Text      string // batch text content
	StartLine int    // 0-based line offset in original document
}

// SplitBatches splits a T-SQL document into batches separated by GO lines.
// Each batch knows its starting line offset in the original document.
func SplitBatches(text string) []Batch {
	lines := strings.Split(text, "\n")
	var batches []Batch
	batchStart := 0
	var batchLines []string

	for i, line := range lines {
		if goSeparator.MatchString(line) {
			batches = append(batches, Batch{
				Text:      strings.Join(batchLines, "\n"),
				StartLine: batchStart,
			})
			batchLines = nil
			batchStart = i + 1
		} else {
			batchLines = append(batchLines, line)
		}
	}

	// Last batch (after final GO or entire file if no GO)
	batches = append(batches, Batch{
		Text:      strings.Join(batchLines, "\n"),
		StartLine: batchStart,
	})

	return batches
}

// BatchAtLine returns the batch containing the given 0-based line number,
// along with the line number adjusted to be relative to the batch start.
func BatchAtLine(text string, line int) (batchText string, adjustedLine int) {
	batches := SplitBatches(text)
	for i := len(batches) - 1; i >= 0; i-- {
		if line >= batches[i].StartLine {
			return batches[i].Text, line - batches[i].StartLine
		}
	}
	// Fallback: return full text
	return text, line
}
