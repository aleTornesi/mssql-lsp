package handler

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/atornesi/tsql-ls/pkg/lsp"
)

func TestDocumentHighlight(t *testing.T) {
	tests := []struct {
		name  string
		input string
		pos   lsp.Position
		want  []lsp.DocumentHighlight
	}{
		{
			name:  "variable highlights",
			input: "DECLARE @x INT\nSELECT @x",
			pos:   lsp.Position{Line: 1, Character: 7},
			want: []lsp.DocumentHighlight{
				{Range: lsp.Range{Start: lsp.Position{Line: 0, Character: 8}, End: lsp.Position{Line: 0, Character: 10}}, Kind: lsp.DocumentHighlightKindWrite},
				{Range: lsp.Range{Start: lsp.Position{Line: 1, Character: 7}, End: lsp.Position{Line: 1, Character: 9}}, Kind: lsp.DocumentHighlightKindRead},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := lsp.DocumentHighlightParams{
				TextDocumentPositionParams: lsp.TextDocumentPositionParams{
					TextDocument: lsp.TextDocumentIdentifier{URI: testFileURI},
					Position:     tt.pos,
				},
			}
			got, err := documentHighlight(tt.input, params)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("highlight mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
