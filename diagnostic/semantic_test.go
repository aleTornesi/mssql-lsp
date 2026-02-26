package diagnostic

import (
	"testing"

	"github.com/aleTornesi/mssql-lsp/parser"
)

func TestCheckSemantics_DuplicateVariable(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantCode  string
		wantCount int
	}{
		{
			name:      "duplicate variable",
			input:     "DECLARE @x INT; DECLARE @x VARCHAR(10)",
			wantCode:  CodeDuplicateVariable,
			wantCount: 1,
		},
		{
			name:      "no duplicate",
			input:     "DECLARE @x INT, @y VARCHAR(10)",
			wantCode:  CodeDuplicateVariable,
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, err := parser.Parse(tt.input)
			if err != nil {
				t.Fatal(err)
			}
			c := &Collector{}
			CheckSemantics(parsed, c)
			count := 0
			for _, d := range c.Diagnostics {
				if d.Code == tt.wantCode {
					count++
				}
			}
			if count != tt.wantCount {
				t.Errorf("got %d diagnostics with code %s, want %d", count, tt.wantCode, tt.wantCount)
			}
		})
	}
}

func TestCheckSemantics_UndefinedVariable(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantCode  string
		wantCount int
	}{
		{
			name:      "undefined variable",
			input:     "SELECT @x",
			wantCode:  CodeUndefinedVariable,
			wantCount: 1,
		},
		{
			name:      "declared variable",
			input:     "DECLARE @x INT; SELECT @x",
			wantCode:  CodeUndefinedVariable,
			wantCount: 0,
		},
		{
			name:      "builtin @@ROWCOUNT",
			input:     "SELECT @@ROWCOUNT",
			wantCode:  CodeUndefinedVariable,
			wantCount: 0,
		},
		{
			name:      "builtin @@ERROR",
			input:     "IF @@ERROR <> 0 SELECT 1",
			wantCode:  CodeUndefinedVariable,
			wantCount: 0,
		},
		{
			name:      "multiple undefined",
			input:     "SELECT @a, @b",
			wantCode:  CodeUndefinedVariable,
			wantCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, err := parser.Parse(tt.input)
			if err != nil {
				t.Fatal(err)
			}
			c := &Collector{}
			CheckSemantics(parsed, c)
			count := 0
			for _, d := range c.Diagnostics {
				if d.Code == tt.wantCode {
					count++
				}
			}
			if count != tt.wantCount {
				t.Errorf("got %d diagnostics with code %s, want %d", count, tt.wantCode, tt.wantCount)
			}
		})
	}
}

func TestCheckSemantics_SetUndeclaredVariable(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantCount int
	}{
		{
			name:      "SET on undeclared",
			input:     "SET @x = 1",
			wantCount: 1,
		},
		{
			name:      "SET on declared",
			input:     "DECLARE @x INT\nSET @x = 1",
			wantCount: 0,
		},
		{
			name:      "SET on builtin",
			input:     "SET @x = @@ROWCOUNT",
			wantCount: 1, // @x is undeclared, @@ROWCOUNT is builtin
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, err := parser.Parse(tt.input)
			if err != nil {
				t.Fatal(err)
			}
			c := &Collector{}
			CheckSemantics(parsed, c)
			count := 0
			for _, d := range c.Diagnostics {
				if d.Code == CodeSetUndeclaredVariable {
					count++
				}
			}
			if count != tt.wantCount {
				t.Errorf("got %d diagnostics with code %s, want %d", count, CodeSetUndeclaredVariable, tt.wantCount)
			}
		})
	}
}

func TestCheckSemantics_TransactionMismatch(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantCount int
	}{
		{
			name:      "unmatched BEGIN TRAN",
			input:     "BEGIN TRANSACTION",
			wantCount: 1,
		},
		{
			name:      "matched BEGIN TRAN COMMIT",
			input:     "BEGIN TRANSACTION\nCOMMIT",
			wantCount: 0,
		},
		{
			name:      "matched BEGIN TRAN ROLLBACK",
			input:     "BEGIN TRAN\nROLLBACK",
			wantCount: 0,
		},
		{
			name:      "no transaction",
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
			c := &Collector{}
			CheckSemantics(parsed, c)
			count := 0
			for _, d := range c.Diagnostics {
				if d.Code == CodeTransactionMismatch {
					count++
				}
			}
			if count != tt.wantCount {
				t.Errorf("got %d diagnostics with code %s, want %d", count, CodeTransactionMismatch, tt.wantCount)
			}
		})
	}
}

func TestCheckSemantics_DuplicateCTE(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantCount int
	}{
		{
			name:      "duplicate CTE",
			input:     "WITH cte AS (SELECT 1), cte AS (SELECT 2) SELECT 1",
			wantCount: 1,
		},
		{
			name:      "unique CTEs",
			input:     "WITH a AS (SELECT 1), b AS (SELECT 2) SELECT 1",
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, err := parser.Parse(tt.input)
			if err != nil {
				t.Fatal(err)
			}
			c := &Collector{}
			CheckSemantics(parsed, c)
			count := 0
			for _, d := range c.Diagnostics {
				if d.Code == CodeDuplicateCTE {
					count++
				}
			}
			if count != tt.wantCount {
				t.Errorf("got %d diagnostics with code %s, want %d", count, CodeDuplicateCTE, tt.wantCount)
			}
		})
	}
}

func TestCheckSemantics_UnreferencedCTE(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantCode  string
		wantCount int
	}{
		{
			name:      "unreferenced CTE",
			input:     "WITH cte AS (SELECT 1) SELECT 1",
			wantCode:  CodeUnreferencedCTE,
			wantCount: 1,
		},
		{
			name:      "referenced CTE",
			input:     "WITH cte AS (SELECT 1) SELECT * FROM cte",
			wantCode:  CodeUnreferencedCTE,
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, err := parser.Parse(tt.input)
			if err != nil {
				t.Fatal(err)
			}
			c := &Collector{}
			CheckSemantics(parsed, c)
			count := 0
			for _, d := range c.Diagnostics {
				if d.Code == tt.wantCode {
					count++
				}
			}
			if count != tt.wantCount {
				t.Errorf("got %d diagnostics with code %s, want %d", count, tt.wantCode, tt.wantCount)
			}
		})
	}
}
