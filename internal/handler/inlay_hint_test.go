package handler

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/atornesi/tsql-ls/internal/lsp"
)

func TestInlayHints(t *testing.T) {
	tests := []struct {
		name  string
		input string
		rng   lsp.Range
		want  []lsp.InlayHint
	}{
		{
			name:  "variable type hint",
			input: "DECLARE @x INT\nSELECT @x",
			rng:   lsp.Range{Start: lsp.Position{Line: 0, Character: 0}, End: lsp.Position{Line: 1, Character: 20}},
			want: []lsp.InlayHint{
				{
					Position:    lsp.Position{Line: 1, Character: 9},
					Label:       ": INT",
					Kind:        lsp.InlayHintKindType,
					PaddingLeft: true,
				},
			},
		},
		{
			name:  "no hints for declaration itself",
			input: "DECLARE @x INT",
			rng:   lsp.Range{Start: lsp.Position{Line: 0, Character: 0}, End: lsp.Position{Line: 0, Character: 20}},
			want:  []lsp.InlayHint{},
		},
		{
			name:  "no hints for variables without type",
			input: "SELECT @x",
			rng:   lsp.Range{Start: lsp.Position{Line: 0, Character: 0}, End: lsp.Position{Line: 0, Character: 20}},
			want:  []lsp.InlayHint{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := lsp.InlayHintParams{
				TextDocument: lsp.TextDocumentIdentifier{URI: testFileURI},
				Range:        tt.rng,
			}
			got, err := inlayHints(tt.input, params)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
