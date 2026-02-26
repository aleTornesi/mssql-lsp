package handler

import (
	"testing"

	"github.com/aleTornesi/mssql-lsp/pkg/lsp"
	"github.com/aleTornesi/mssql-lsp/parser"
	"github.com/aleTornesi/mssql-lsp/parser/parseutil"
)

func TestCollectBlockSymbols(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantCount int
	}{
		{
			name:      "BEGIN END",
			input:     "BEGIN\n  SELECT 1\nEND",
			wantCount: 1,
		},
		{
			name:      "nested blocks",
			input:     "BEGIN\n  BEGIN\n    SELECT 1\n  END\nEND",
			wantCount: 1, // parser flattens nested BEGIN/END
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
			var symbols []lsp.DocumentSymbol
			collectBlockSymbols(parsed, 0, &symbols)
			if len(symbols) != tt.wantCount {
				t.Errorf("got %d block symbols, want %d", len(symbols), tt.wantCount)
			}
		})
	}
}

func TestDocumentSymbols_Variables(t *testing.T) {
	input := "DECLARE @x INT, @y VARCHAR(50)"
	parsed, err := parser.Parse(input)
	if err != nil {
		t.Fatal(err)
	}

	st := parseutil.ExtractSymbols(parsed)
	varCount := 0
	for _, sym := range st.Symbols {
		if sym.Kind == parseutil.SymbolVariable {
			varCount++
		}
	}
	if varCount != 2 {
		t.Errorf("expected 2 variables, got %d", varCount)
	}
}

func TestDocumentSymbols_CTE(t *testing.T) {
	input := "WITH cte1 AS (SELECT 1), cte2 AS (SELECT 2) SELECT * FROM cte1"
	parsed, err := parser.Parse(input)
	if err != nil {
		t.Fatal(err)
	}

	st := parseutil.ExtractSymbols(parsed)
	cteCount := 0
	for _, sym := range st.Symbols {
		if sym.Kind == parseutil.SymbolCTE {
			cteCount++
		}
	}
	if cteCount != 2 {
		t.Errorf("expected 2 CTEs, got %d", cteCount)
	}
}
