package parseutil

import (
	"testing"

	"github.com/aleTornesi/mssql-lsp/parser"
)

func TestExtractSymbols_Variables(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantLen int
		wantSym []struct {
			name     string
			dataType string
			kind     SymbolKind
		}
	}{
		{
			name:    "single variable",
			input:   "DECLARE @x INT",
			wantLen: 1,
			wantSym: []struct {
				name     string
				dataType string
				kind     SymbolKind
			}{
				{name: "@x", dataType: "INT", kind: SymbolVariable},
			},
		},
		{
			name:    "multiple variables",
			input:   "DECLARE @x INT, @y VARCHAR(50)",
			wantLen: 2,
			wantSym: []struct {
				name     string
				dataType string
				kind     SymbolKind
			}{
				{name: "@x", dataType: "INT", kind: SymbolVariable},
				{name: "@y", dataType: "VARCHAR(50)", kind: SymbolVariable},
			},
		},
		{
			name:    "variable with assignment",
			input:   "DECLARE @x INT = 5",
			wantLen: 1,
			wantSym: []struct {
				name     string
				dataType string
				kind     SymbolKind
			}{
				{name: "@x", dataType: "INT", kind: SymbolVariable},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, err := parser.Parse(tt.input)
			if err != nil {
				t.Fatal(err)
			}
			st := ExtractSymbols(parsed)
			if len(st.Symbols) != tt.wantLen {
				t.Fatalf("got %d symbols, want %d", len(st.Symbols), tt.wantLen)
			}
			for i, want := range tt.wantSym {
				got := st.Symbols[i]
				if got.Name != want.name {
					t.Errorf("symbol[%d].Name = %q, want %q", i, got.Name, want.name)
				}
				if got.DataType != want.dataType {
					t.Errorf("symbol[%d].DataType = %q, want %q", i, got.DataType, want.dataType)
				}
				if got.Kind != want.kind {
					t.Errorf("symbol[%d].Kind = %d, want %d", i, got.Kind, want.kind)
				}
			}
		})
	}
}

func TestExtractSymbols_CTEs(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantLen int
		wantNames []string
	}{
		{
			name:      "single CTE",
			input:     "WITH cte AS (SELECT 1) SELECT * FROM cte",
			wantLen:   1,
			wantNames: []string{"cte"},
		},
		{
			name:      "multiple CTEs",
			input:     "WITH a AS (SELECT 1), b AS (SELECT 2) SELECT * FROM a JOIN b ON 1=1",
			wantLen:   2,
			wantNames: []string{"a", "b"},
		},
		{
			name:      "WITH NOLOCK is not a CTE",
			input:     "SELECT * FROM t WITH (NOLOCK)",
			wantLen:   0,
			wantNames: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, err := parser.Parse(tt.input)
			if err != nil {
				t.Fatal(err)
			}
			st := ExtractSymbols(parsed)
			cteSyms := []*Symbol{}
			for _, s := range st.Symbols {
				if s.Kind == SymbolCTE {
					cteSyms = append(cteSyms, s)
				}
			}
			if len(cteSyms) != tt.wantLen {
				t.Fatalf("got %d CTE symbols, want %d", len(cteSyms), tt.wantLen)
			}
			for i, wantName := range tt.wantNames {
				if cteSyms[i].Name != wantName {
					t.Errorf("CTE[%d].Name = %q, want %q", i, cteSyms[i].Name, wantName)
				}
			}
		})
	}
}

func TestExtractSymbols_TempTables(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantLen   int
		wantNames []string
	}{
		{
			name:      "CREATE TABLE #tmp",
			input:     "CREATE TABLE #tmp (id INT)",
			wantLen:   1,
			wantNames: []string{"#tmp"},
		},
		{
			name:      "SELECT INTO #tmp",
			input:     "SELECT * INTO #tmp FROM t",
			wantLen:   1,
			wantNames: []string{"#tmp"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, err := parser.Parse(tt.input)
			if err != nil {
				t.Fatal(err)
			}
			st := ExtractSymbols(parsed)
			tempSyms := []*Symbol{}
			for _, s := range st.Symbols {
				if s.Kind == SymbolTempTable {
					tempSyms = append(tempSyms, s)
				}
			}
			if len(tempSyms) != tt.wantLen {
				t.Fatalf("got %d temp table symbols, want %d", len(tempSyms), tt.wantLen)
			}
			for i, wantName := range tt.wantNames {
				if tempSyms[i].Name != wantName {
					t.Errorf("TempTable[%d].Name = %q, want %q", i, tempSyms[i].Name, wantName)
				}
			}
		})
	}
}

func TestSymbolTable_Lookup(t *testing.T) {
	input := "DECLARE @x INT; SELECT @x"
	parsed, err := parser.Parse(input)
	if err != nil {
		t.Fatal(err)
	}
	st := ExtractSymbols(parsed)
	sym := st.Lookup("@x")
	if sym == nil {
		t.Fatal("expected to find @x")
	}
	if sym.DataType != "INT" {
		t.Errorf("got DataType=%q, want INT", sym.DataType)
	}

	// Case insensitive
	sym2 := st.Lookup("@X")
	if sym2 == nil {
		t.Fatal("expected case-insensitive lookup to find @X")
	}

	// Not found
	sym3 := st.Lookup("@z")
	if sym3 != nil {
		t.Error("expected nil for undeclared @z")
	}
}

func TestSymbolTable_FindReferences(t *testing.T) {
	input := "DECLARE @x INT; SELECT @x; SET @x = 1"
	parsed, err := parser.Parse(input)
	if err != nil {
		t.Fatal(err)
	}
	st := ExtractSymbols(parsed)
	refs := st.FindReferences(parsed, "@x")
	if len(refs) < 3 {
		t.Errorf("expected at least 3 references to @x, got %d", len(refs))
	}
}
