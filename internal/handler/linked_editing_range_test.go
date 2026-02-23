package handler

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/atornesi/tsql-ls/internal/lsp"
)

func TestLinkedEditingRange(t *testing.T) {
	tests := []struct {
		name  string
		input string
		pos   lsp.Position
		want  *lsp.LinkedEditingRanges
	}{
		{
			name:  "BEGIN/END - cursor on BEGIN",
			input: "BEGIN\n  SELECT 1\nEND",
			pos:   lsp.Position{Line: 0, Character: 2},
			want: &lsp.LinkedEditingRanges{
				Ranges: []lsp.Range{
					{Start: lsp.Position{Line: 0, Character: 0}, End: lsp.Position{Line: 0, Character: 5}},
					{Start: lsp.Position{Line: 2, Character: 0}, End: lsp.Position{Line: 2, Character: 3}},
				},
			},
		},
		{
			name:  "BEGIN/END - cursor on END",
			input: "BEGIN\n  SELECT 1\nEND",
			pos:   lsp.Position{Line: 2, Character: 1},
			want: &lsp.LinkedEditingRanges{
				Ranges: []lsp.Range{
					{Start: lsp.Position{Line: 0, Character: 0}, End: lsp.Position{Line: 0, Character: 5}},
					{Start: lsp.Position{Line: 2, Character: 0}, End: lsp.Position{Line: 2, Character: 3}},
				},
			},
		},
		{
			name:  "cursor not on keyword",
			input: "BEGIN\n  SELECT 1\nEND",
			pos:   lsp.Position{Line: 1, Character: 5},
			want:  nil,
		},
		{
			name:  "CASE/END",
			input: "SELECT CASE WHEN 1=1 THEN 'a' END",
			pos:   lsp.Position{Line: 0, Character: 8},
			want: &lsp.LinkedEditingRanges{
				Ranges: []lsp.Range{
					{Start: lsp.Position{Line: 0, Character: 7}, End: lsp.Position{Line: 0, Character: 11}},
					{Start: lsp.Position{Line: 0, Character: 30}, End: lsp.Position{Line: 0, Character: 33}},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := lsp.LinkedEditingRangeParams{
				TextDocumentPositionParams: lsp.TextDocumentPositionParams{
					TextDocument: lsp.TextDocumentIdentifier{URI: testFileURI},
					Position:     tt.pos,
				},
			}
			got, err := linkedEditingRange(tt.input, params)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
