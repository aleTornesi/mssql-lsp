package handler

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/atornesi/tsql-ls/internal/lsp"
)

func TestReferences(t *testing.T) {
	tests := []struct {
		name               string
		input              string
		pos                lsp.Position
		includeDeclaration bool
		want               []lsp.Location
	}{
		{
			name:               "variable refs include declaration",
			input:              "DECLARE @x INT\nSELECT @x\nSET @x = 1",
			pos:                lsp.Position{Line: 1, Character: 7},
			includeDeclaration: true,
			want: []lsp.Location{
				{URI: testFileURI, Range: lsp.Range{Start: lsp.Position{Line: 0, Character: 8}, End: lsp.Position{Line: 0, Character: 10}}},
				{URI: testFileURI, Range: lsp.Range{Start: lsp.Position{Line: 1, Character: 7}, End: lsp.Position{Line: 1, Character: 9}}},
				{URI: testFileURI, Range: lsp.Range{Start: lsp.Position{Line: 2, Character: 4}, End: lsp.Position{Line: 2, Character: 6}}},
			},
		},
		{
			name:               "variable refs exclude declaration",
			input:              "DECLARE @x INT\nSELECT @x\nSET @x = 1",
			pos:                lsp.Position{Line: 1, Character: 7},
			includeDeclaration: false,
			want: []lsp.Location{
				{URI: testFileURI, Range: lsp.Range{Start: lsp.Position{Line: 1, Character: 7}, End: lsp.Position{Line: 1, Character: 9}}},
				{URI: testFileURI, Range: lsp.Range{Start: lsp.Position{Line: 2, Character: 4}, End: lsp.Position{Line: 2, Character: 6}}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := lsp.ReferenceParams{
				TextDocumentPositionParams: lsp.TextDocumentPositionParams{
					TextDocument: lsp.TextDocumentIdentifier{URI: testFileURI},
					Position:     tt.pos,
				},
				Context: lsp.ReferenceContext{IncludeDeclaration: tt.includeDeclaration},
			}
			got, err := references(testFileURI, tt.input, params)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("references mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
