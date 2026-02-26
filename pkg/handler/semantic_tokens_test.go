package handler

import (
	"testing"

	"github.com/aleTornesi/mssql-lsp/pkg/lsp"
)

func TestEncodeSemanticTokens(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int // expected number of tokens (data length / 5)
	}{
		{
			name:  "simple SELECT",
			input: "SELECT 1",
			want:  2, // SELECT (keyword) + 1 (number)
		},
		{
			name:  "variable",
			input: "DECLARE @x INT",
			want:  3, // DECLARE + @x + INT
		},
		{
			name:  "string literal",
			input: "SELECT 'hello'",
			want:  2, // SELECT + 'hello'
		},
		{
			name:  "comment",
			input: "-- comment\nSELECT 1",
			want:  3, // comment + SELECT + 1
		},
		{
			name:  "operators",
			input: "SELECT 1 + 2",
			want:  4, // SELECT + 1 + + + 2
		},
		{
			name:  "data type",
			input: "CAST(1 AS VARCHAR)",
			want:  4, // CAST(function) + 1(number) + AS(keyword) + VARCHAR(type)
		},
		{
			name:  "function",
			input: "SELECT GETDATE()",
			want:  2, // SELECT + GETDATE
		},
		{
			name:  "multiline",
			input: "SELECT\n  1",
			want:  2,
		},
		{
			name:  "batches",
			input: "SELECT 1\nGO\nSELECT 2",
			want:  4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := encodeSemanticTokens(tt.input, nil)
			got := len(data) / 5
			if got != tt.want {
				t.Errorf("got %d tokens, want %d (data: %v)", got, tt.want, data)
			}
		})
	}
}

func TestEncodeSemanticTokens_deltaEncoding(t *testing.T) {
	// "SELECT 1" -> SELECT at (0,0) len 6, 1 at (0,7) len 1
	data := encodeSemanticTokens("SELECT 1", nil)
	if len(data) != 10 {
		t.Fatalf("expected 10 data items, got %d: %v", len(data), data)
	}

	// First token: deltaLine=0, deltaStart=0, length=6, type=keyword(0)
	if data[0] != 0 || data[1] != 0 || data[2] != 6 || data[3] != uint32(semKeyword) {
		t.Errorf("first token wrong: %v", data[0:5])
	}

	// Second token: deltaLine=0, deltaStart=7, length=1, type=number(5)
	if data[5] != 0 || data[6] != 7 || data[7] != 1 || data[8] != uint32(semNumber) {
		t.Errorf("second token wrong: %v", data[5:10])
	}
}

func TestEncodeSemanticTokens_range(t *testing.T) {
	input := "SELECT 1\nFROM dbo.t"
	// Only request line 1
	rng := testRange(1, 0, 1, 20)
	data := encodeSemanticTokens(input, &rng)
	// Should only have FROM token from line 1 (dbo and t are unmatched identifiers)
	got := len(data) / 5
	if got != 1 {
		t.Errorf("got %d tokens, want 1 (data: %v)", got, data)
	}
}

func testRange(startLine, startChar, endLine, endChar int) lsp.Range {
	return lsp.Range{
		Start: lsp.Position{Line: startLine, Character: startChar},
		End:   lsp.Position{Line: endLine, Character: endChar},
	}
}

func TestComputeSemanticTokenEdits(t *testing.T) {
	tests := []struct {
		name    string
		oldData []uint32
		newData []uint32
		want    int // number of edits
	}{
		{
			name:    "identical data produces no edits",
			oldData: []uint32{0, 0, 6, 0, 0, 0, 7, 1, 5, 0},
			newData: []uint32{0, 0, 6, 0, 0, 0, 7, 1, 5, 0},
			want:    0,
		},
		{
			name:    "changed data produces one edit",
			oldData: []uint32{0, 0, 6, 0, 0, 0, 7, 1, 5, 0},
			newData: []uint32{0, 0, 6, 0, 0, 0, 7, 2, 5, 0},
			want:    1,
		},
		{
			name:    "new data longer than old",
			oldData: []uint32{0, 0, 6, 0, 0},
			newData: []uint32{0, 0, 6, 0, 0, 0, 7, 1, 5, 0},
			want:    1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			edits := computeSemanticTokenEdits(tt.oldData, tt.newData)
			if len(edits) != tt.want {
				t.Errorf("got %d edits, want %d: %+v", len(edits), tt.want, edits)
			}
		})
	}
}

func TestClassifyToken_systemVariable(t *testing.T) {
	// @@IDENTITY should be classified as variable
	input := "SELECT @@IDENTITY"
	data := encodeSemanticTokens(input, nil)
	if len(data) < 10 {
		t.Fatalf("expected at least 2 tokens, got %d", len(data)/5)
	}
	// Second token should be variable type
	if data[8] != uint32(semVariable) {
		t.Errorf("@@IDENTITY should be variable type, got %d", data[8])
	}
}
