package handler

import (
	"testing"

	"github.com/atornesi/tsql-ls/ast"
	"github.com/atornesi/tsql-ls/internal/lsp"
	"github.com/atornesi/tsql-ls/parser"
)

func TestCollectFoldingRanges(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantCount int
	}{
		{
			name:      "BEGIN END block",
			input:     "BEGIN\n  SELECT 1\nEND",
			wantCount: 1,
		},
		{
			name:      "nested BEGIN END",
			input:     "BEGIN\n  BEGIN\n    SELECT 1\n  END\nEND",
			wantCount: 1, // parser flattens nested BEGIN/END
		},
		{
			name:      "IF statement",
			input:     "IF 1=1\n  SELECT 1",
			wantCount: 1,
		},
		{
			name:      "no blocks",
			input:     "SELECT 1",
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, err := parser.Parse(tt.input)
			if err != nil {
				t.Fatal(err)
			}
			var ranges []lsp.FoldingRange
			collectFoldingRanges(parsed, 0, &ranges)
			if len(ranges) != tt.wantCount {
				t.Errorf("got %d folding ranges, want %d", len(ranges), tt.wantCount)
			}
		})
	}
}

func TestCollectCommentFoldingRanges(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantCount int
	}{
		{
			name:      "multiline comment",
			input:     "/* line 1\n   line 2 */\nSELECT 1",
			wantCount: 1,
		},
		{
			name:      "single line comment",
			input:     "-- comment\nSELECT 1",
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ranges []lsp.FoldingRange
			collectCommentFoldingRanges(tt.input, &ranges)
			if len(ranges) != tt.wantCount {
				t.Errorf("got %d folding ranges, want %d", len(ranges), tt.wantCount)
			}
		})
	}
}

func TestAddFoldingRange_singleLine(t *testing.T) {
	// Single line node should not produce a folding range
	parsed, _ := parser.Parse("SELECT 1")
	var ranges []lsp.FoldingRange
	for _, node := range parsed.GetTokens() {
		if _, ok := node.(ast.TokenList); ok {
			addFoldingRange(node, 0, lsp.FoldingRangeKindRegion, &ranges)
		}
	}
	if len(ranges) != 0 {
		t.Errorf("expected 0 folding ranges for single line, got %d", len(ranges))
	}
}
