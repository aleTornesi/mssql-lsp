package handler

import (
	"testing"

	"github.com/aleTornesi/mssql-lsp/pkg/lsp"
)

func TestSelectionRange(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		pos       lsp.Position
		wantDepth int // how many parent links
	}{
		{
			name:      "simple SELECT",
			input:     "SELECT 1",
			pos:       lsp.Position{Line: 0, Character: 0},
			wantDepth: 1, // at least token → statement
		},
		{
			name:      "nested in BEGIN END",
			input:     "BEGIN\n  SELECT 1\nEND",
			pos:       lsp.Position{Line: 1, Character: 2},
			wantDepth: 2, // token → statement → begin/end
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sr := selectionRange(tt.input, tt.pos)
			depth := 0
			p := sr.Parent
			for p != nil {
				depth++
				p = p.Parent
			}
			if depth < tt.wantDepth {
				t.Errorf("got depth %d, want at least %d", depth, tt.wantDepth)
			}
		})
	}
}
