package handler

import (
	"testing"

	"github.com/atornesi/tsql-ls/internal/lsp"
)

func TestOnTypeFormatting(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		pos      lsp.Position
		ch       string
		wantEdit bool
	}{
		{
			name:     "after BEGIN adds indent",
			input:    "BEGIN\n",
			pos:      lsp.Position{Line: 1, Character: 0},
			ch:       "\n",
			wantEdit: true,
		},
		{
			name:     "after THEN adds indent",
			input:    "CASE WHEN 1=1 THEN\n",
			pos:      lsp.Position{Line: 1, Character: 0},
			ch:       "\n",
			wantEdit: true,
		},
		{
			name:     "after ELSE adds indent",
			input:    "ELSE\n",
			pos:      lsp.Position{Line: 1, Character: 0},
			ch:       "\n",
			wantEdit: true,
		},
		{
			name:     "after BEGIN TRY adds indent",
			input:    "BEGIN TRY\n",
			pos:      lsp.Position{Line: 1, Character: 0},
			ch:       "\n",
			wantEdit: true,
		},
		{
			name:     "normal line no indent",
			input:    "SELECT 1\n",
			pos:      lsp.Position{Line: 1, Character: 0},
			ch:       "\n",
			wantEdit: false,
		},
		{
			name:     "non-newline char ignored",
			input:    "SELECT 1;",
			pos:      lsp.Position{Line: 0, Character: 9},
			ch:       ";",
			wantEdit: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := lsp.DocumentOnTypeFormattingParams{
				TextDocument: lsp.TextDocumentIdentifier{URI: "file:///test.sql"},
				Position:     tt.pos,
				Ch:           tt.ch,
				Options:      lsp.FormattingOptions{TabSize: 4, InsertSpaces: true},
			}
			edits := onTypeFormatting(tt.input, params)
			hasEdits := len(edits) > 0
			if hasEdits != tt.wantEdit {
				t.Errorf("got edits=%v, want edits=%v", hasEdits, tt.wantEdit)
			}
		})
	}
}

func TestOnTypeFormatting_indentContent(t *testing.T) {
	input := "BEGIN\n"
	params := lsp.DocumentOnTypeFormattingParams{
		TextDocument: lsp.TextDocumentIdentifier{URI: "file:///test.sql"},
		Position:     lsp.Position{Line: 1, Character: 0},
		Ch:           "\n",
		Options:      lsp.FormattingOptions{TabSize: 4, InsertSpaces: true},
	}
	edits := onTypeFormatting(input, params)
	if len(edits) != 1 {
		t.Fatalf("expected 1 edit, got %d", len(edits))
	}
	if edits[0].NewText != "    " {
		t.Errorf("expected 4-space indent, got %q", edits[0].NewText)
	}
}

func TestOnTypeFormatting_tabs(t *testing.T) {
	input := "BEGIN\n"
	params := lsp.DocumentOnTypeFormattingParams{
		TextDocument: lsp.TextDocumentIdentifier{URI: "file:///test.sql"},
		Position:     lsp.Position{Line: 1, Character: 0},
		Ch:           "\n",
		Options:      lsp.FormattingOptions{TabSize: 4, InsertSpaces: false},
	}
	edits := onTypeFormatting(input, params)
	if len(edits) != 1 {
		t.Fatalf("expected 1 edit, got %d", len(edits))
	}
	if edits[0].NewText != "\t" {
		t.Errorf("expected tab indent, got %q", edits[0].NewText)
	}
}
