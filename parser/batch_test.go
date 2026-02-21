package parser

import (
	"testing"
)

func TestSplitBatches(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantCount  int
		wantStarts []int
		wantTexts  []string
	}{
		{
			name:       "no GO separator",
			input:      "SELECT 1\nSELECT 2",
			wantCount:  1,
			wantStarts: []int{0},
			wantTexts:  []string{"SELECT 1\nSELECT 2"},
		},
		{
			name:       "single GO",
			input:      "SELECT 1\nGO\nSELECT 2",
			wantCount:  2,
			wantStarts: []int{0, 2},
			wantTexts:  []string{"SELECT 1", "SELECT 2"},
		},
		{
			name:       "GO case insensitive",
			input:      "SELECT 1\ngo\nSELECT 2",
			wantCount:  2,
			wantStarts: []int{0, 2},
			wantTexts:  []string{"SELECT 1", "SELECT 2"},
		},
		{
			name:       "GO with whitespace",
			input:      "SELECT 1\n  GO  \nSELECT 2",
			wantCount:  2,
			wantStarts: []int{0, 2},
			wantTexts:  []string{"SELECT 1", "SELECT 2"},
		},
		{
			name:       "multiple GO",
			input:      "SELECT 1\nGO\nSELECT 2\nGO\nSELECT 3",
			wantCount:  3,
			wantStarts: []int{0, 2, 4},
			wantTexts:  []string{"SELECT 1", "SELECT 2", "SELECT 3"},
		},
		{
			name:       "GO at start",
			input:      "GO\nSELECT 1",
			wantCount:  2,
			wantStarts: []int{0, 1},
			wantTexts:  []string{"", "SELECT 1"},
		},
		{
			name:       "GO at end",
			input:      "SELECT 1\nGO",
			wantCount:  2,
			wantStarts: []int{0, 2},
			wantTexts:  []string{"SELECT 1", ""},
		},
		{
			name:       "GO inside word not matched",
			input:      "SELECT GOPHER FROM animals",
			wantCount:  1,
			wantStarts: []int{0},
			wantTexts:  []string{"SELECT GOPHER FROM animals"},
		},
		{
			name:       "multiline batch",
			input:      "CREATE TABLE t1 (\n  id INT\n)\nGO\nINSERT INTO t1 VALUES (1)",
			wantCount:  2,
			wantStarts: []int{0, 4},
			wantTexts:  []string{"CREATE TABLE t1 (\n  id INT\n)", "INSERT INTO t1 VALUES (1)"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			batches := SplitBatches(tt.input)
			if len(batches) != tt.wantCount {
				t.Fatalf("got %d batches, want %d", len(batches), tt.wantCount)
			}
			for i, b := range batches {
				if b.StartLine != tt.wantStarts[i] {
					t.Errorf("batch %d: StartLine = %d, want %d", i, b.StartLine, tt.wantStarts[i])
				}
				if b.Text != tt.wantTexts[i] {
					t.Errorf("batch %d: Text = %q, want %q", i, b.Text, tt.wantTexts[i])
				}
			}
		})
	}
}

func TestBatchAtLine(t *testing.T) {
	input := "SELECT 1\nGO\nSELECT 2\nSELECT 3\nGO\nSELECT 4"
	// Lines: 0=SELECT 1, 1=GO, 2=SELECT 2, 3=SELECT 3, 4=GO, 5=SELECT 4

	tests := []struct {
		line         int
		wantText     string
		wantAdjusted int
	}{
		{0, "SELECT 1", 0},
		{2, "SELECT 2\nSELECT 3", 0},
		{3, "SELECT 2\nSELECT 3", 1},
		{5, "SELECT 4", 0},
	}

	for _, tt := range tests {
		text, adj := BatchAtLine(input, tt.line)
		if text != tt.wantText {
			t.Errorf("line %d: text = %q, want %q", tt.line, text, tt.wantText)
		}
		if adj != tt.wantAdjusted {
			t.Errorf("line %d: adjusted = %d, want %d", tt.line, adj, tt.wantAdjusted)
		}
	}
}
