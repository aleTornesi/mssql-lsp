package formatter

import (
	"testing"

	"github.com/atornesi/tsql-ls/internal/config"
	"github.com/atornesi/tsql-ls/internal/lsp"
)

func TestFormatRange(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		r        lsp.Range
		wantNil  bool
		wantLine int // expected start line of returned edit
	}{
		{
			name:  "format single line range",
			input: "SELECT 1\nselect    *   from   users\nSELECT 2",
			r: lsp.Range{
				Start: lsp.Position{Line: 1, Character: 0},
				End:   lsp.Position{Line: 1, Character: 25},
			},
			wantNil:  false,
			wantLine: 1,
		},
		{
			name:  "empty range",
			input: "SELECT 1\n\nSELECT 2",
			r: lsp.Range{
				Start: lsp.Position{Line: 1, Character: 0},
				End:   lsp.Position{Line: 1, Character: 0},
			},
			wantNil: true,
		},
	}

	cfg := &config.Config{LowercaseKeywords: false}
	opts := lsp.FormattingOptions{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			edits, err := FormatRange(tt.input, tt.r, opts, cfg)
			if err != nil {
				t.Fatal(err)
			}
			if tt.wantNil {
				if edits != nil {
					t.Errorf("expected nil edits, got %v", edits)
				}
				return
			}
			if len(edits) != 1 {
				t.Fatalf("expected 1 edit, got %d", len(edits))
			}
			if edits[0].Range.Start.Line != tt.wantLine {
				t.Errorf("expected start line %d, got %d", tt.wantLine, edits[0].Range.Start.Line)
			}
		})
	}
}
