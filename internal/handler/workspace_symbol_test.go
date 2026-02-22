package handler

import (
	"strings"
	"testing"

	"github.com/atornesi/tsql-ls/internal/lsp"
	"github.com/atornesi/tsql-ls/parser"
	"github.com/atornesi/tsql-ls/parser/parseutil"
)

func TestWorkspaceSymbol(t *testing.T) {
	s := &Server{files: make(map[string]*File)}
	s.files[testFileURI] = &File{
		LanguageID: "sql",
		Text:       "DECLARE @x INT, @name VARCHAR(50)",
	}

	tests := []struct {
		name      string
		query     string
		wantCount int
	}{
		{name: "empty query returns all", query: "", wantCount: 2},
		{name: "filter by name", query: "name", wantCount: 1},
		{name: "no match", query: "zzz", wantCount: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := lsp.WorkspaceSymbolParams{Query: tt.query}
			// Call the logic directly by simulating
			result := workspaceSymbolDirect(s, params)
			if len(result) != tt.wantCount {
				t.Errorf("got %d symbols, want %d", len(result), tt.wantCount)
			}
		})
	}
}

// workspaceSymbolDirect is a test helper that extracts workspace symbol logic.
func workspaceSymbolDirect(s *Server, params lsp.WorkspaceSymbolParams) []lsp.SymbolInformation {
	var symbols []lsp.SymbolInformation
	query := params.Query

	for uri, f := range s.files {
		batches := parser.SplitBatches(f.Text)
		for _, batch := range batches {
			parsed, err := parser.Parse(batch.Text)
			if err != nil {
				continue
			}
			st := parseutil.ExtractSymbols(parsed)
			for _, sym := range st.Symbols {
				if query != "" && !strings.Contains(strings.ToUpper(sym.Name), strings.ToUpper(query)) {
					continue
				}
				var kind lsp.SymbolKind
				switch sym.Kind {
				case parseutil.SymbolVariable:
					kind = lsp.SymbolKindVariable
				case parseutil.SymbolCTE:
					kind = lsp.SymbolKindClass
				case parseutil.SymbolTempTable:
					kind = lsp.SymbolKindClass
				}
				r := lsp.Range{
					Start: lsp.Position{Line: sym.Pos.Line + batch.StartLine, Character: sym.Pos.Col},
					End:   lsp.Position{Line: sym.EndPos.Line + batch.StartLine, Character: sym.EndPos.Col},
				}
				symbols = append(symbols, lsp.SymbolInformation{
					Name: sym.Name,
					Kind: kind,
					Location: lsp.Location{
						URI:   uri,
						Range: r,
					},
				})
			}
		}
	}
	return symbols
}
