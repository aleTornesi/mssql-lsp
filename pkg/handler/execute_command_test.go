package handler

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/aleTornesi/mssql-lsp/pkg/lsp"
)

func Test_extractRangeText(t *testing.T) {
	type args struct {
		text      string
		startLine int
		startChar int
		endLine   int
		endChar   int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "extract single line",
			args: args{
				text:      "select * from city",
				startLine: 0,
				startChar: 0,
				endLine:   0,
				endChar:   8,
			},
			want: "select *",
		},
		{
			name: "extract multi line with not equal start end",
			args: args{
				text:      "select 1;\nselect 2;\nselect 3;",
				startLine: 0,
				startChar: 7,
				endLine:   2,
				endChar:   8,
			},
			want: "1;\nselect 2;\nselect 3",
		},
		{
			name: "extract multi line with equal start end",
			args: args{
				text:      "select 1;\nselect 2;\nselect 3;",
				startLine: 1,
				startChar: 2,
				endLine:   1,
				endChar:   6,
			},
			want: "lect",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := extractRangeText(tt.args.text, tt.args.startLine, tt.args.startChar, tt.args.endLine, tt.args.endChar); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func Test_codeActionRemoveCTE(t *testing.T) {
	uri := "file:///test.sql"
	tests := []struct {
		name     string
		input    string
		diagName string
		diagRange lsp.Range
		wantEdit *lsp.Range // nil means no action returned
	}{
		{
			name:     "single CTE removes WITH clause",
			input:    "WITH cte AS (SELECT 1)\nSELECT 1",
			diagName: "cte",
			diagRange: lsp.Range{
				Start: lsp.Position{Line: 0, Character: 5},
				End:   lsp.Position{Line: 0, Character: 8},
			},
			wantEdit: &lsp.Range{
				Start: lsp.Position{Line: 0, Character: 0},
				End:   lsp.Position{Line: 0, Character: 22},
			},
		},
		{
			name:     "first of two CTEs",
			input:    "WITH a AS (SELECT 1), b AS (SELECT 2)\nSELECT * FROM b",
			diagName: "a",
			diagRange: lsp.Range{
				Start: lsp.Position{Line: 0, Character: 5},
				End:   lsp.Position{Line: 0, Character: 6},
			},
			wantEdit: &lsp.Range{
				Start: lsp.Position{Line: 0, Character: 5},
				End:   lsp.Position{Line: 0, Character: 21},
			},
		},
		{
			name:     "last of two CTEs",
			input:    "WITH a AS (SELECT 1), b AS (SELECT 2)\nSELECT * FROM a",
			diagName: "b",
			diagRange: lsp.Range{
				Start: lsp.Position{Line: 0, Character: 22},
				End:   lsp.Position{Line: 0, Character: 23},
			},
			wantEdit: &lsp.Range{
				Start: lsp.Position{Line: 0, Character: 20},
				End:   lsp.Position{Line: 0, Character: 37},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := "TSQL011"
			diag := lsp.Diagnostic{
				Range: tt.diagRange,
				Code:  &code,
			}
			action := codeActionRemoveCTE(uri, tt.input, diag)
			if tt.wantEdit == nil {
				if action != nil {
					t.Errorf("expected nil action")
				}
				return
			}
			if action == nil {
				t.Fatal("expected non-nil action")
			}
			if action.Edit == nil {
				t.Fatal("expected edit in action")
			}
			edits := action.Edit.Changes[uri]
			if len(edits) != 1 {
				t.Fatalf("expected 1 edit, got %d", len(edits))
			}
			if diff := cmp.Diff(*tt.wantEdit, edits[0].Range); diff != "" {
				t.Errorf("edit range mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
