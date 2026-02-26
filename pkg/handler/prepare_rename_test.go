package handler

import (
	"testing"

	"github.com/atornesi/tsql-ls/pkg/lsp"
)

func TestPrepareRename(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		pos         lsp.Position
		wantName    string
		wantErr     bool
		wantNilResult bool
	}{
		{
			name:     "variable",
			input:    "DECLARE @x INT\nSELECT @x",
			pos:      lsp.Position{Line: 1, Character: 8},
			wantName: "@x",
		},
		{
			name:     "table name",
			input:    "SELECT id FROM users",
			pos:      lsp.Position{Line: 0, Character: 17},
			wantName: "users",
		},
		{
			name:        "keyword - not renameable",
			input:       "SELECT 1",
			pos:         lsp.Position{Line: 0, Character: 0},
			wantNilResult: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := lsp.PrepareRenameParams{
				TextDocumentPositionParams: lsp.TextDocumentPositionParams{
					TextDocument: lsp.TextDocumentIdentifier{URI: "file:///test.sql"},
					Position:     tt.pos,
				},
			}
			res, err := prepareRename(tt.input, params)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if tt.wantNilResult {
				// prepareRename returns error for non-renameable, not nil
				// The error is a jsonrpc2 error
				if err == nil && res != nil {
					t.Error("expected error or nil result for non-renameable element")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if res == nil {
				t.Fatal("expected result, got nil")
			}
			if res.Placeholder != tt.wantName {
				t.Errorf("placeholder = %q, want %q", res.Placeholder, tt.wantName)
			}
		})
	}
}
